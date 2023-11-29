package pool

import (
	"sync"
	"testing"

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
	newResource, firstAcqErr := manager.AcquireResource()
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
	_, secondAcqErr := manager.AcquireResource()
	assert.NotNil(t, secondAcqErr, "expected to be error since pool size is 1")

	// Do some work
	newResource.SomeWork = true
	newResource.SomeValue = "Mutated"
	manager.ReleaseResource(newResource)

	// Get it again - reuse
	acqResource, acqErr := manager.AcquireResource()
	assert.Equal(t, "Mutated", acqResource.SomeValue)
	assert.NoError(t, acqErr)
	acqResource.SomeWork = true
	manager.ReleaseResource(acqResource)

	acqResource, acqErr = manager.AcquireResource()
	assert.NoError(t, acqErr)
	assert.Equal(t, "NewOne", acqResource.SomeValue)
}

func TestCreateAndManipulateResourcesNoLimitOfUsages(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManger[stubResource](1, 0, factory)

	all10 := 0
	initRes, initResErr := manager.AcquireResource()
	assert.NoError(t, initResErr)
	assert.NotNil(t, initRes)
	initRes.SomeValue = "old"
	manager.ReleaseResource(initRes)

	for i := byte(0); i < 10; i++ {
		res, err := manager.AcquireResource()
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

	workErr := manager.AcquireAndReleaseResource(func(resource *stubResource) error {
		resource.SomeWork = true
		resource.SomeValue = "unit_test"

		return nil
	})

	assert.NoError(t, workErr)
	updatedResource, ackErr := manager.AcquireResource()
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
	ackRes, ackErr := manager.AcquireResource()
	assert.NoError(t, ackErr)
	assert.False(t, ackRes.SomeWork == notManagedResource.SomeWork)
	assert.False(t, ackRes.SomeValue == notManagedResource.SomeValue)
}

func TestEmptyPool(t *testing.T) {
	factory := &stubFactory{}
	manager := NewResourcePoolManger[stubResource](0, 0, factory)

	ackRes, ackErr := manager.AcquireResource()
	assert.NotNil(t, ackErr)
	assert.Nil(t, ackRes)
}

func TestLimitedResourceAndLimitedUsages(t *testing.T) {
	testCases := []struct {
		name             string
		resourceCap      uint8
		resourceUsageCap uint8
		callForResource  uint8
		factory          *stubFactory
		assertFn         func(t *testing.T, ackErr error)
	}{
		{
			"Test case 1: Limited Resource And Limited Usages",
			2,
			10,
			20,
			&stubFactory{},
			func(t *testing.T, ackErr error) {
				assert.NoError(t, ackErr)
			},
		},
		{
			"Test case 2: Limited Resource And Unlimited Usages",
			2,
			0,
			20,
			&stubFactory{},
			func(t *testing.T, ackErr error) {
				assert.NoError(t, ackErr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewResourcePoolManger[stubResource](tc.resourceCap, tc.resourceUsageCap, tc.factory)

			var wg sync.WaitGroup

			allResourceAreReused := true
			for i := uint8(0); i < tc.resourceUsageCap; i++ {
				wg.Add(1)
				go func(t *testing.T) {
					defer wg.Done()
					ackRes, ackErr := manager.AcquireResource()

					assert.NotNil(t, ackRes)
					assert.NoError(t, ackErr)
					if ackErr != nil {
						allResourceAreReused = false
					}

					manager.ReleaseResource(ackRes)
				}(t)
			}

			wg.Wait()
			assert.True(t, allResourceAreReused, "expected to use all requested resource in test case")
		})
	}
}
