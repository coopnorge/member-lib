package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stubResource struct {
	SomeWork  bool
	SomeValue string
}

type stubFactory struct{}

func (m *stubFactory) Construct() *stubResource {
	return &stubResource{SomeValue: "NewOne"}
}

// Test creation and manipulation of resources
func TestCreateAndManipulateResources(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManger[stubResource](1, 2, factory)

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
	manager := NewResourcePoolManger[stubResource](1, 0, factory)

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
	manager := NewResourcePoolManger[stubResource](1, 0, factory)

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
	manager := NewResourcePoolManger[stubResource](1, 0, factory)

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
	manager := NewResourcePoolManger[stubResource](0, 0, factory)

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewResourcePoolManger[stubResource](tc.resourceCap, tc.resourceUsageCap, tc.factory)

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
	manager := NewResourcePoolManger[stubResource](1, 0, factory)

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
	manager := NewResourcePoolManger[stubResource](1, 0, factory)

	unitContext, unitContextCancel := context.WithCancel(context.TODO())

	firstRes, firstResErr := manager.AcquireResource(unitContext, false)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	unitContextCancel()

	manager.ReleaseResource(firstRes)

	canceledRes, canceledResErr := manager.AcquireResource(unitContext, false)
	assert.NotNil(t, canceledResErr)
	assert.Nil(t, canceledRes)

	newAckRes, newAckResErr := manager.AcquireResource(context.TODO(), false)
	assert.NoError(t, newAckResErr)
	assert.NotNil(t, newAckRes)
	assert.Same(t, firstRes, newAckRes, "Expected to have same address of resource since it's was returned before and reused now")
}

func TestAcquireResourceThatIsTakenButWithRetry(t *testing.T) {
	manager := NewResourcePoolManger[stubResource](1, 0, new(stubFactory))
	manager.retryOnResourceDelay = time.Nanosecond

	unitContext := context.TODO()

	// All good new resource
	firstRes, firstResErr := manager.AcquireResource(unitContext, true)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	// Try to get resource with retry
	waitTimePassed := time.Now()

	time.AfterFunc(time.Millisecond, func() { manager.ReleaseResource(firstRes) })
	secondRes, secondResErr := manager.AcquireResource(unitContext, true)
	assert.Nil(t, secondResErr)
	assert.NotNil(t, secondRes)

	isOverMillisecond := time.Since(waitTimePassed) > time.Millisecond
	assert.True(t, isOverMillisecond, "expected to have more that 1 millisecond to get resource with attempt")
}

func TestAcquireResourceThatIsTakenButContextCanceledOnRetry(t *testing.T) {
	manager := NewResourcePoolManger[stubResource](1, 0, new(stubFactory))
	manager.retryOnResourceDelay = time.Nanosecond

	unitContext, unitContextCancel := context.WithCancel(context.TODO())

	// All good new resource
	firstRes, firstResErr := manager.AcquireResource(unitContext, true)
	assert.NoError(t, firstResErr)
	assert.NotNil(t, firstRes)

	time.AfterFunc(time.Millisecond, func() { unitContextCancel() })
	secondRes, secondResErr := manager.AcquireResource(unitContext, true)
	assert.NotNil(t, secondResErr)
	assert.Nil(t, secondRes)
	assert.NotNil(t, unitContext.Err())
}
