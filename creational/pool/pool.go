// Package pool or Object Pool design pattern is a way to manage the allocation and reuse of
// objects, especially when creating new instances of these objects is costly in terms of resources or performance.
//
// More about Object pool can be found https://en.wikipedia.org/wiki/Object_pool_pattern in wiki.
//
// Possible use cases:
//   - You might need to have `Connection Management`.
//     Then you could use this component to balanced workload, reduces the initialisation overhead for
//     each connection, and improves connection management efficiency.
//   - `Messages Tracking` - Manager will subsequently handle resource creation and
//     destruction processes. This use-case improves tracking, prevents message
//     loss, and ensures effective resource allocation and deallocation.
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

	// ResourceBuilder used to construct new Resource's in ResourcePoolManager or destroy them.
	ResourceBuilder[T Resource] interface {
		// Construct will be called to create new instance of Resource.
		Construct() *T
		// Deconstruction will ber called when ResourcePool will need gracefully destroy/clean object.
		Deconstruction(*T)
	}

	// ResourcePool functionality.
	ResourcePool[T Resource] interface {
		// AcquireResource retrieves an available resource from the pool.
		AcquireResource(ctx context.Context, isNeedToRetryOnTaken bool) (*T, error)
		// ReleaseResource releases a given resource back to the pool.
		ReleaseResource(releasedResource *T)
		// DetachResource will move out current resources from the management of ResourcePool.
		DetachResource(resource *T)
		// CleanUpManagedResources all created resource in ResourcePool.
		CleanUpManagedResources(ctx context.Context) error
		// AcquireAndReleaseResource allows to execute action with needed Resource -> T.
		AcquireAndReleaseResource(ctx context.Context, action func(resource *T) error) error
		// GetRetryOnResourceDelay returns the delay duration before the next retry attempt.
		GetRetryOnResourceDelay() time.Duration
		// SetRetryOnResourceDelay configures the delay duration for subsequent retry attempts.
		SetRetryOnResourceDelay(retryOnResourceDelay time.Duration)
	}

	// managedResource is a struct that represents a resource managed within the ResourcePoolManager.
	// It contains a flags for resource being acquired and usage count along with the resource itself.
	managedResource[T Resource] struct {
		isAcquired bool
		usageCount uint8
		resource   *T
		mu         sync.Mutex
	}

	// resourceObtainer used to thread safely get/create resources.
	resourceObtainer[T Resource] struct {
		resource *T
		error    error
	}

	// ResourcePoolManager represents a pool of resources with encapsulated logic.
	// It controls pool behavior and attributes such as maximum pool size and resource usage limit.
	ResourcePoolManager[T Resource] struct {
		factory            ResourceBuilder[T]
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
//   - resourceFactory: a function of type ResourceBuilder[T], which the ResourcePoolManager uses to create new resources if there are none available in the pool.
//
// The purpose of the ResourcePoolManager is to manage a pool of resources, ensuring there are always resources available up to the maximum pool size.
// Each resource can be used multiple times, controlled by the resourceUsageLimit, before being discarded or renewed.
func NewResourcePoolManager[T Resource](poolSize, resourceUsageLimit uint8, resourceFactory ResourceBuilder[T]) *ResourcePoolManager[T] {
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

	obtainedResource := make(chan resourceObtainer[T], 1)
	go rpm.getResource(obtainedResource)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case managedResource := <-obtainedResource:
		if managedResource.error == nil {
			return managedResource.resource, nil
		} else if !isNeedToRetryOnTaken && managedResource.error != nil {
			return nil, managedResource.error
		}

		// NOTE: Ignore pool limit since retry will handle resource acquiring in recursion.
		if !errors.Is(managedResource.error, ErrorPoolLimitReached) {
			return nil, managedResource.error
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(rpm.retryOnResourceDelay):
		return rpm.AcquireResource(ctx, isNeedToRetryOnTaken)
	}
}

// getResource if no resource is available, a new one is created using the provided factory method.
// Acquisition of resources is thread-safe.
func (rpm *ResourcePoolManager[T]) getResource(resourceObtained chan<- resourceObtainer[T]) {
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
			resourceObtained <- resourceObtainer[T]{error: err}
			return
		}
		acqManagedResource = rpm.createManagedResource()
	}

	acqManagedResource.mu.Lock()
	defer acqManagedResource.mu.Unlock()

	acqManagedResource.usageCount++
	acqManagedResource.isAcquired = true
	rpm.pool.Store(acqManagedResource.resource, acqManagedResource)

	resourceObtained <- resourceObtainer[T]{resource: acqManagedResource.resource}
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
		rpm.destroyManagedResource(releasedResource)
	}

	managedResource.mu.Unlock()
}

// DetachResource will move out current resources from the management of ResourcePool.
func (rpm *ResourcePoolManager[T]) DetachResource(resource *T) {
	r, ok := rpm.pool.Load(resource)
	if !ok { // Not found, already not managed / deleted from pool
		return
	}

	managedResource, _ := r.(*managedResource[T]) //nolint:errcheck // value is already found by key, and it's strictly controlled how it's stored.
	managedResource.mu.Lock()
	defer managedResource.mu.Unlock()

	rpm.pool.Delete(resource)
}

// CleanUpManagedResources all created resource in ResourcePool.
func (rpm *ResourcePoolManager[T]) CleanUpManagedResources(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	for i := uint8(0); i < rpm.maxPoolSize; i++ {
		var acqManagedResource *managedResource[T]
		rpm.pool.Range(func(_, value any) bool {
			mr, ok := value.(*managedResource[T])
			if !ok {
				return false
			}

			acqManagedResource = mr

			return true
		})
		if acqManagedResource != nil {
			rpm.destroyManagedResource(acqManagedResource.resource)
		}
	}

	return nil
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

func (rpm *ResourcePoolManager[T]) destroyManagedResource(releasedResource *T) {
	rpm.pool.Delete(releasedResource)
	rpm.factory.Deconstruction(releasedResource)
}

func (rpm *ResourcePoolManager[T]) verifyCurrentPoolSize() error {
	var currentPoolSize uint8

	// NOTE: Allow max size of type
	if rpm.maxPoolSize == ^uint8(0) {
		return nil
	}

	rpm.pool.Range(func(_, _ any) bool {
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
