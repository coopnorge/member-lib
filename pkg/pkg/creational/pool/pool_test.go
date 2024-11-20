package pool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test creation and manipulation of resources
func TestCreateAndManipulateResources(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 2, factory)

	// Attempt to acquire a newResource
	unitContext := context.TODO()
	newResource, firstAcqErr := manager.AcquireResource(unitContext, false)
	assert.NoError(t, firstAcqErr)
	assert.Equal(t, "NewOne", newResource.SomeValue)

	// Check pool
	manager.pool.Range(func(k, v any) bool {
		mr, ok := v.(*managedResource[stubResource])
		assert.True(t, ok)
		assert.True(t, mr.isAcquired)
		return true
	})

	// Try to get more
	_, secondAcqErr := manager.AcquireResource(unitContext, false)
	assert.NotNil(t, secondAcqErr, "expected to be error since pool size is 1")

	// Do some work
	newResource.SomeWork = true
	newResource.SomeValue = "Mutated"
	manager.ReleaseResource(newResource)

	// Get it again - reuse
	acqResource, acqErr := manager.AcquireResource(unitContext, false)
	assert.Equal(t, "Mutated", acqResource.SomeValue)
	assert.NoError(t, acqErr)
	acqResource.SomeWork = true
	manager.ReleaseResource(acqResource)

	acqResource, acqErr = manager.AcquireResource(unitContext, false)
	assert.NoError(t, acqErr)
	assert.Equal(t, "NewOne", acqResource.SomeValue)
}

func TestCreateAndManipulateResourcesNoLimitOfUsages(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	all10 := 0
	unitContext := context.TODO()
	initRes, initResErr := manager.AcquireResource(unitContext, false)
	assert.NoError(t, initResErr)
	assert.NotNil(t, initRes)
	initRes.SomeValue = "old"
	manager.ReleaseResource(initRes)

	for i := byte(0); i < 10; i++ {
		res, err := manager.AcquireResource(unitContext, false)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		manager.ReleaseResource(res)
		if res != nil && res.SomeValue == "old" {
			all10++
		}
	}

	assert.True(t, all10 == 10, "expected to be no limitations on AcquireResource")
}

func TestAcquireAndRelease(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	unitContext := context.TODO()
	workErr := manager.AcquireAndReleaseResource(unitContext, func(resource *stubResource) error {
		resource.SomeWork = true
		resource.SomeValue = "unit_test"

		return nil
	})

	assert.NoError(t, workErr)
	updatedResource, ackErr := manager.AcquireResource(unitContext, false)
	assert.NoError(t, ackErr)
	assert.True(t, updatedResource.SomeWork)
	assert.True(t, updatedResource.SomeValue == "unit_test")
}

func TestReleaseNotExistingResource(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	notManagedResource := &stubResource{
		SomeWork:  true,
		SomeValue: "not_exist",
	}
	manager.ReleaseResource(notManagedResource)

	unitContext := context.TODO()
	ackRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.NoError(t, ackErr)
	assert.False(t, ackRes.SomeWork == notManagedResource.SomeWork)
	assert.False(t, ackRes.SomeValue == notManagedResource.SomeValue)
}

func TestEmptyPool(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](0, 0, factory)

	unitContext := context.TODO()
	ackRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.NotNil(t, ackErr)
	assert.Nil(t, ackRes)
}

func TestLimitedResourceAndLimitedUsages(t *testing.T) {
	testCases := []struct {
		name                 string
		resourceCap          uint8
		resourceUsageCap     uint8
		callForResource      uint8
		factory              *stubFactory
		allResourceAreReused bool
		mu                   sync.Mutex
	}{
		{
			"Test case 1: Limited Resource And Limited Usages",
			2,
			10,
			20,
			&stubFactory{},
			true,
			sync.Mutex{},
		},
		{
			"Test case 2: Limited Resource And Unlimited Usages",
			2,
			0,
			20,
			&stubFactory{},
			true,
			sync.Mutex{},
		},
	}

	for i := range testCases {
		tc := &testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			manager := NewResourcePoolManager[stubResource](tc.resourceCap, tc.resourceUsageCap, tc.factory)

			var wg sync.WaitGroup

			for i := uint8(0); i < tc.callForResource; i++ {
				wg.Add(1)
				go func() {
					tc.mu.Lock()
					defer wg.Done()
					defer tc.mu.Unlock()

					ackRes, ackErr := manager.AcquireResource(context.TODO(), false)
					if ackRes == nil || ackErr != nil {
						tc.allResourceAreReused = false
					}

					manager.ReleaseResource(ackRes)
				}()
			}

			wg.Wait()

			tc.mu.Lock()
			defer tc.mu.Unlock()
			if !tc.allResourceAreReused {
				t.Fatal("expected to use all requested resource in test case")
			}
		})
	}
}

func TestAcquireResourceWithCanceledContext(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	unitContext, unitContextCancel := context.WithCancel(context.TODO())
	unitContextCancel()

	ackRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.NotNil(t, ackErr)
	assert.Nil(t, ackRes)

	ackRes, ackErr = manager.AcquireResource(context.TODO(), false)
	assert.NoError(t, ackErr)
	assert.NotNil(t, ackRes)
}

func TestReleaseResourceWithCanceledContext(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	unitContext, unitContextCancel := context.WithCancel(context.TODO())

	firstRes, firstResErr := manager.AcquireResource(unitContext, false)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	manager.ReleaseResource(firstRes)

	unitContextCancel()

	canceledRes, canceledResErr := manager.AcquireResource(unitContext, false)
	assert.NotNil(t, canceledResErr)
	assert.Nil(t, canceledRes)

	newAckRes, newAckResErr := manager.AcquireResource(context.TODO(), false)
	assert.NoError(t, newAckResErr)
	assert.NotNil(t, newAckRes)
	assert.Same(t, firstRes, newAckRes, "Expected to have same address of resource since it's was returned before and reused now")
}

func TestAcquireResourceThatIsTakenButWithRetry(t *testing.T) {
	manager := NewResourcePoolManager[stubResource](1, 0, new(stubFactory))
	manager.retryOnResourceDelay = time.Nanosecond

	unitContext := context.TODO()

	// All good new resource
	firstRes, firstResErr := manager.AcquireResource(unitContext, true)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	// Try to get resource with retry
	waitTimePassed := time.Now()

	time.AfterFunc(time.Nanosecond, func() { manager.ReleaseResource(firstRes) })
	time.Sleep(time.Millisecond)

	secondRes, secondResErr := manager.AcquireResource(unitContext, true)
	assert.Nil(t, secondResErr)
	assert.NotNil(t, secondRes)

	isOverMillisecond := time.Since(waitTimePassed) > time.Millisecond
	assert.True(t, isOverMillisecond, "expected to have more that 1 millisecond to get resource with attempt")
}

func TestAcquireResourceThatIsTakenButContextCanceledOnRetry(t *testing.T) {
	manager := NewResourcePoolManager[stubResource](1, 0, new(stubFactory))
	manager.retryOnResourceDelay = time.Nanosecond

	unitContext, unitContextCancel := context.WithCancel(context.TODO())

	// All good new resource
	firstRes, firstResErr := manager.AcquireResource(unitContext, true)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	time.AfterFunc(time.Nanosecond, func() { unitContextCancel() })
	time.Sleep(time.Millisecond)

	secondRes, secondResErr := manager.AcquireResource(unitContext, true)
	assert.NotNil(t, secondResErr)
	assert.Nil(t, secondRes)
	assert.NotNil(t, unitContext.Err())
}

func TestModifyRetryOnResourceDelay(t *testing.T) {
	manager := NewResourcePoolManager[stubResource](1, 0, new(stubFactory))

	assert.True(t, manager.retryOnResourceDelay == defaultRetryOnResourceDelay)

	manager.SetRetryOnResourceDelay(time.Nanosecond)
	assert.True(t, manager.GetRetryOnResourceDelay() == time.Nanosecond)
}

func TestAcquireResourceToContextExpire(t *testing.T) {
	manager := NewResourcePoolManager[stubResource](1, 1, new(stubFactory))
	manager.retryOnResourceDelay = time.Nanosecond

	unitContext, unitContextCancel := context.WithTimeout(context.TODO(), time.Millisecond)
	defer unitContextCancel()

	// All good new resource
	firstRes, firstResErr := manager.AcquireResource(unitContext, true)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	time.Sleep(time.Microsecond)

	timer := time.NewTimer(time.Millisecond)
	var finalError error

	for {
		select {
		case <-timer.C:
			assert.True(
				t,
				errors.Is(finalError, ErrorContextCanceled),
				fmt.Sprintf("expected that error will be related to canceled context but given: %s", finalError.Error()),
			)
			return
		default:
			secondRes, secondResErr := manager.AcquireResource(unitContext, true)
			assert.NotNil(t, secondResErr)
			assert.Nil(t, secondRes)

			time.Sleep(time.Microsecond)

			finalError = secondResErr
		}
	}
}

func TestAcquireResourceButLimitReachedError(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	unitContext := context.TODO()

	ackRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.Nil(t, ackErr)
	assert.NotNil(t, ackRes)

	ackRes, ackErr = manager.AcquireResource(unitContext, false)
	assert.Nil(t, ackRes)
	assert.ErrorIs(t, ackErr, ErrorPoolLimitReached)
}

func TestDetachResource(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 0, factory)

	unitContext := context.TODO()

	initRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.Nil(t, ackErr)
	assert.NotNil(t, initRes)
	initRes.SomeValue = "Initial resource"

	manager.DetachResource(initRes)
	_, ok := manager.pool.Load(initRes)
	assert.False(t, ok, "Expected that resource is already deleted from pool")
	manager.DetachResource(initRes) // No panic on already deleted

	newRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.Nil(t, ackErr)
	assert.NotNil(t, newRes)
	newRes.SomeValue = "Will be new resource"

	assert.NotSame(t, initRes, newRes, "Expected that resource will be not same")
	assert.NotEqual(t, initRes.SomeValue, newRes.SomeValue)
}

func TestResourceDeconstruction(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 1, factory)
	unitContext := context.TODO()

	initRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.Nil(t, ackErr)
	assert.NotNil(t, initRes)
	initRes.SomeValue = "Initial resource"

	manager.ReleaseResource(initRes)

	newRes, ackErr := manager.AcquireResource(unitContext, false)
	assert.Nil(t, ackErr)
	assert.NotNil(t, newRes)

	assert.False(t, initRes.SomeWork, "expected to be changed after deconstruction")
	assert.True(t, initRes.SomeValue == "", "expected to be changed after deconstruction")

	var storedManagedResource *managedResource[Resource]
	manager.pool.Range(func(_, value any) bool {
		mr, ok := value.(*managedResource[Resource])
		if !ok {
			return false
		}

		storedManagedResource = mr

		return true
	})

	assert.Nil(t, storedManagedResource, "expected to be nil, resource must be deleted from pool after deconstruction when limit of usages was used")
}

func TestFactoryToConstructAndDeconstructResources(t *testing.T) {
	holdResources := make([]*stubResource, 0, 6)

	factory := &stubFactory{}
	manager := NewResourcePoolManager[stubResource](1, 1, factory)
	unitContext := context.TODO()

	for i := 0; i < 6; i++ {
		initRes, ackErr := manager.AcquireResource(unitContext, false)
		assert.Nil(t, ackErr)
		assert.NotNil(t, initRes)

		holdResources = append(holdResources, initRes)
		manager.ReleaseResource(initRes)
	}

	for _, resource := range holdResources {
		assert.Nil(t, resource.someExternalObject, "expected to be destroyed all external resources")
	}
}

func TestTryCatchGetResourceWhenContextWasCanceledAfterMutex(t *testing.T) {
	factory := &stubRemoteConnectionFactory{}

	manager := NewResourcePoolManager[stubRemoteConnectionFactory](1, 5, factory)
	unitContext, unitContextCancel := context.WithTimeout(context.TODO(), time.Millisecond)
	defer unitContextCancel()

	res, resErr := manager.AcquireResource(unitContext, false)
	assert.Nil(t, res)
	assert.NotNil(t, resErr)
	assert.ErrorIs(t, resErr, context.DeadlineExceeded)
}

func TestCleanUpManagedResources(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name        string
		ctx         context.Context
		maxPoolSize uint8
		expectErr   error
	}{
		{
			name:        "context canceled before operation",
			ctx:         canceledCtx,
			maxPoolSize: 5,
			expectErr:   context.Canceled,
		},
		{
			name:        "normal operation",
			ctx:         context.Background(),
			maxPoolSize: 5,
			expectErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := &stubFactory{}
			rpm := NewResourcePoolManager[stubResource](tc.maxPoolSize, 1, factory)

			for i := uint8(0); i < tc.maxPoolSize; i++ {
				_, _ = rpm.AcquireResource(tc.ctx, false)
			}

			err := rpm.CleanUpManagedResources(tc.ctx)
			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			}
		})
	}
}

type stubResource struct {
	SomeWork           bool
	SomeValue          string
	someExternalObject context.Context
}

type stubFactory struct{}

func (m *stubFactory) Construct() *stubResource {
	return &stubResource{SomeValue: "NewOne", someExternalObject: context.TODO()}
}

func (m *stubFactory) Deconstruction(r *stubResource) {
	r.SomeWork = false
	r.SomeValue = ""
	r.someExternalObject = nil
}

type stubRemoteConnectionFactory struct{}

func (m *stubRemoteConnectionFactory) Construct() *stubRemoteConnectionFactory {
	time.Sleep(time.Hour)

	return new(stubRemoteConnectionFactory)
}

func (m *stubRemoteConnectionFactory) Deconstruction(_ *stubRemoteConnectionFactory) {}

func Example_schematicallyResourcePoolMangerWorks() {
	/*
	   ┌─────────────────────────────────────────────┐
	   │ Your Project Code                           │     ┌──────────────────────────┐
	   │                                             │     │ Configure                │
	   │                                             │     │                          │
	   │   ┌──────────────────┐                      │     │ Pool Size and Usage Count│
	   │   │DatabasedConnector│                      │     └─────────────┬────────────┘
	   │   └┬─────────────────┘                      │                   │
	   │    │                                        │                   │
	   │   ┌▼──────────────────────────┬┐            │     ┌─────────────▼───────────┐
	   │   │Database Connection Factory│┼────────────┼─────┤► Resource Pool Manager  │
	   │   └───────────────────────────┴┘            │     └──────┬┬─────────────────┤
	   │                                             │            ││                 │
	   │                                             │            ││                 │
	   │                                             │ ┌──────────┤►Acquire Resource │
	   │   ┌──────────────────────────┐              │ │          ││                 │
	   │   │Repository                │              │ │          ││                 │
	   │   ├──────────┬┬──────────────┘              │ │       ┌──┤►Release Resource │
	   │   │ Find User││                             │ │       │  └──────────────────┘
	   │   └──────┬───┼┘                             │ │       │
	   │          │   │  1.Ask to get free Connector │ │       │
	   │          │   └──────────────────────────────┼─┘       │
	   │          │                                  │         │
	   │          │      2.Return back Connector     │         │
	   │          └──────────────────────────────────┼─────────┘
	   │                                             │
	   └─────────────────────────────────────────────┘
	*/
}

func Example_newResourcePoolManger() {
	const maxConn uint8 = 5
	const maxMsgPerConn uint8 = 100
	isNeededToWaitResource := false

	// NewResourcePoolManager returns a new pool manager that will create and manage resources from stubRemoteConnectionFactory
	// Internally pool manager will check if resource used 100 times (maxMsgPerConn) and new resources can be produced by maxConn capacity (5).
	connectionPool := NewResourcePoolManager[stubRemoteConnectionFactory](maxConn, maxMsgPerConn, new(stubRemoteConnectionFactory))

	// Now you can get new or already created resource, if it's needed to wait when it will be available pass true as second arg (isNeededToWaitResource).
	resource, acquireErr := connectionPool.AcquireResource(context.TODO(), isNeededToWaitResource)
	if acquireErr != nil {
		log.Println("Reason...")
		return
	}

	// Now we can return resource back to pool so that another part of code could use it.
	connectionPool.ReleaseResource(resource)
}
