package pool

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrorPoolLimitReached thrown when ResourcePoolManager current pool size is reached to Maximum allowed size.
	ErrorPoolLimitReached = errors.New("resource - pool limit reached")
	// ErrorContextCanceled thrown when client (passed) context is canceled and operation must be canceled.
	ErrorContextCanceled = context.DeadlineExceeded
)

const defaultRetryOnResourceDelay = time.Second

type (
	// Resource which will be managed by the ResourcePoolManager.
	Resource interface {
		any
	}

	// ResourceFactory used to construct new Resource's in ResourcePoolManager and monitor their usage.
	ResourceFactory[T Resource] interface {
		Construct() *T
	}

	// managedResource is a struct that represents a resource managed within the ResourcePoolManager.
	// It contains a flags for resource being acquired and usage count along with the resource itself.
	managedResource[T Resource] struct {
		isAcquired bool
		usageCount uint8
		resource   *T
		mu         sync.Mutex
	}

	// ResourcePoolManager represents a pool of resources with encapsulated logic.
	// It controls pool behavior and attributes such as maximum pool size and resource usage limit.
	ResourcePoolManager[T Resource] struct {
		factory            ResourceFactory[T]
		pool               sync.Map
		maxPoolSize        uint8
		resourceUsageLimit uint8
		// retryOnResourceDelay used on resource manipulation (AcquireResource).
		retryOnResourceDelay time.Duration
		mu                   sync.RWMutex
	}
)

// NewResourcePoolManager is a constructor function for creating a new ResourcePoolManager.
// It accepts three parameters:
//   - poolSize: a uint8 representing the maximum size of the pool. This limits the number of resources that can be managed by the pool.
//   - resourceUsageLimit: a uint8 representing the usage limit for each resource in the pool. Setting this to '0' indicates there is no usage limit.
//   - resourceFactory: a function of type ResourceFactory[T], which the ResourcePoolManager uses to create new resources if there are none available in the pool.
//
// The purpose of the ResourcePoolManager is to manage a pool of resources, ensuring there are always resources available up to the maximum pool size.
// Each resource can be used multiple times, controlled by the resourceUsageLimit, before being discarded or renewed.
func NewResourcePoolManager[T Resource](poolSize, resourceUsageLimit uint8, resourceFactory ResourceFactory[T]) *ResourcePoolManager[T] {
	return &ResourcePoolManager[T]{
		factory:              resourceFactory,
		resourceUsageLimit:   resourceUsageLimit,
		maxPoolSize:          poolSize,
		retryOnResourceDelay: defaultRetryOnResourceDelay,
	}
}

// AcquireResource retrieves an available resource from the pool.
// ctx context.Context - controlling code flow, if `isNeedToRetryOnTaken` will be true
// ResourcePoolManager will try to obtain Resource when it will be available recursively until context.Context will be canceled.
// If there is no need to re-try, pass `isNeedToRetryOnTaken` as false.
func (rpm *ResourcePoolManager[T]) AcquireResource(ctx context.Context, isNeedToRetryOnTaken bool) (*T, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	resource, getResourceErr := rpm.getResource()
	if getResourceErr == nil {
		return resource, nil
	} else if !isNeedToRetryOnTaken && getResourceErr != nil {
		return nil, getResourceErr
	}

	// NOTE: Ignore pool limit since retry will handle resource acquiring in recursion.
	if !errors.Is(getResourceErr, ErrorPoolLimitReached) {
		return nil, getResourceErr
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	select {
	case <-ctx.Done():
		return nil, ErrorContextCanceled
	case <-time.After(rpm.retryOnResourceDelay):
		return rpm.AcquireResource(ctx, isNeedToRetryOnTaken)
	}
}

// getResource if no resource is available, a new one is created using the provided factory method.
// Acquisition of resources is thread-safe.
func (rpm *ResourcePoolManager[T]) getResource() (*T, error) {
	var acqManagedResource *managedResource[T]

	rpm.pool.Range(func(_, value any) bool {
		mr, ok := value.(*managedResource[T])
		if !ok {
			return false
		}

		mr.mu.Lock()
		defer mr.mu.Unlock()
		if !mr.isAcquired {
			// NOTE: If usageCount not 0 then it's set to be unlimited amount of usages.
			if rpm.resourceUsageLimit == 0 || mr.usageCount < rpm.resourceUsageLimit {
				acqManagedResource = mr
			}
		}

		return true
	})

	if acqManagedResource == nil {
		if err := rpm.verifyCurrentPoolSize(); err != nil {
			return nil, err
		}
		acqManagedResource = rpm.createManagedResource()
	}

	acqManagedResource.mu.Lock()
	acqManagedResource.usageCount++
	acqManagedResource.isAcquired = true
	acqManagedResource.mu.Unlock()
	rpm.pool.Store(acqManagedResource.resource, acqManagedResource)

	return acqManagedResource.resource, nil
}

// ReleaseResource releases a given resource back to the pool.
// If a resource exceeds the usage limit it gets removed from the pool.
// Releasing of resources is thread-safe.
func (rpm *ResourcePoolManager[T]) ReleaseResource(releasedResource *T) {
	value, ok := rpm.pool.Load(releasedResource)
	if !ok {
		return
	}

	managedResource, _ := value.(*managedResource[T]) //nolint:errcheck // value is already found by key, and it's strictly controlled how it's stored.
	managedResource.mu.Lock()
	if rpm.resourceUsageLimit == 0 || managedResource.usageCount < rpm.resourceUsageLimit {
		managedResource.isAcquired = false
		rpm.pool.Store(releasedResource, managedResource)
	} else {
		rpm.pool.Delete(releasedResource)
	}

	managedResource.mu.Unlock()
}

// AcquireAndReleaseResource allows to execute action with needed Resource -> T.
func (rpm *ResourcePoolManager[T]) AcquireAndReleaseResource(ctx context.Context, action func(resource *T) error) error {
	r, rErr := rpm.AcquireResource(ctx, true)
	if rErr != nil {
		return rErr
	}
	defer rpm.ReleaseResource(r)

	return action(r)
}

func (rpm *ResourcePoolManager[T]) createManagedResource() *managedResource[T] {
	return &managedResource[T]{
		resource: rpm.factory.Construct(),
	}
}

func (rpm *ResourcePoolManager[T]) verifyCurrentPoolSize() error {
	var currentPoolSize uint8

	// NOTE: Allow max size of type
	if rpm.maxPoolSize == ^uint8(0) {
		return nil
	}

	rpm.pool.Range(func(k, v any) bool {
		currentPoolSize++
		return true
	})

	if currentPoolSize >= rpm.maxPoolSize {
		return ErrorPoolLimitReached
	}

	return nil
}

// GetRetryOnResourceDelay returns the delay duration before the next retry attempt.
func (rpm *ResourcePoolManager[T]) GetRetryOnResourceDelay() time.Duration {
	rpm.mu.RLock()
	defer rpm.mu.RUnlock()

	return rpm.retryOnResourceDelay
}

// SetRetryOnResourceDelay configures the delay duration for subsequent retry attempts.
func (rpm *ResourcePoolManager[T]) SetRetryOnResourceDelay(retryOnResourceDelay time.Duration) {
	rpm.mu.Lock()
	defer rpm.mu.Unlock()

	rpm.retryOnResourceDelay = retryOnResourceDelay
}
