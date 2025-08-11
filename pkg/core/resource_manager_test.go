package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceManager(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	assert.NotNil(t, rm)
	assert.Equal(t, config, rm.config)
	assert.NotNil(t, rm.resources)
	assert.NotNil(t, rm.resourceStats)
}

func TestResourceManager_RegisterResource(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	resource := &MockResource{
		id:        "test_resource_1",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	err := rm.RegisterResource(resource)
	assert.NoError(t, err)

	// Verify resource was registered
	retrievedResource, err := rm.GetResource("test_resource_1")
	assert.NoError(t, err)
	assert.Equal(t, resource, retrievedResource)

	// Verify stats were updated
	stats := rm.GetStats()
	browserStats := stats[ResourceTypeBrowser]
	assert.Equal(t, 1, browserStats.Active)
	assert.Equal(t, 1, browserStats.Created)
}

func TestResourceManager_RegisterResource_Duplicate(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	resource1 := &MockResource{
		id:        "duplicate_resource",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	resource2 := &MockResource{
		id:        "duplicate_resource",
		resType:   ResourceTypeDatabase,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	// First registration should succeed
	err := rm.RegisterResource(resource1)
	assert.NoError(t, err)

	// Second registration with same ID should fail
	err = rm.RegisterResource(resource2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestResourceManager_RegisterResource_MaxLimit(t *testing.T) {
	config := DefaultResourceManagerConfig()
	config.MaxResources = 2 // Small limit for testing
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	// Register resources up to the limit
	for i := 0; i < 2; i++ {
		resource := &MockResource{
			id:        fmt.Sprintf("resource_%d", i),
			resType:   ResourceTypeBrowser,
			createdAt: time.Now(),
			lastUsed:  time.Now(),
			active:    true,
		}
		err := rm.RegisterResource(resource)
		assert.NoError(t, err)
	}

	// Try to register one more (should fail)
	extraResource := &MockResource{
		id:        "extra_resource",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}
	err := rm.RegisterResource(extraResource)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum number of resources exceeded")
}

func TestResourceManager_UnregisterResource(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	resource := &MockResource{
		id:        "test_resource_1",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	// Register resource
	err := rm.RegisterResource(resource)
	require.NoError(t, err)

	// Unregister resource
	err = rm.UnregisterResource("test_resource_1")
	assert.NoError(t, err)

	// Verify resource was unregistered
	_, err = rm.GetResource("test_resource_1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Verify stats were updated
	stats := rm.GetStats()
	browserStats := stats[ResourceTypeBrowser]
	assert.Equal(t, 0, browserStats.Active)
	assert.Equal(t, 1, browserStats.Cleaned)
}

func TestResourceManager_UnregisterResource_NotFound(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	err := rm.UnregisterResource("nonexistent_resource")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResourceManager_GetResourcesByType(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	// Register resources of different types
	browserResource := &MockResource{
		id:        "browser_1",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	dbResource := &MockResource{
		id:        "db_1",
		resType:   ResourceTypeDatabase,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	anotherBrowserResource := &MockResource{
		id:        "browser_2",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}

	err := rm.RegisterResource(browserResource)
	require.NoError(t, err)
	err = rm.RegisterResource(dbResource)
	require.NoError(t, err)
	err = rm.RegisterResource(anotherBrowserResource)
	require.NoError(t, err)

	// Get browser resources
	browserResources := rm.GetResourcesByType(ResourceTypeBrowser)
	assert.Len(t, browserResources, 2)

	// Get database resources
	dbResources := rm.GetResourcesByType(ResourceTypeDatabase)
	assert.Len(t, dbResources, 1)

	// Get non-existent type
	mobileResources := rm.GetResourcesByType(ResourceTypeMobile)
	assert.Len(t, mobileResources, 0)
}

func TestResourceManager_CleanupAll(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	// Register multiple resources
	for i := 0; i < 3; i++ {
		resource := &MockResource{
			id:        fmt.Sprintf("resource_%d", i),
			resType:   ResourceTypeBrowser,
			createdAt: time.Now(),
			lastUsed:  time.Now(),
			active:    true,
		}
		err := rm.RegisterResource(resource)
		require.NoError(t, err)
	}

	// Verify resources exist
	stats := rm.GetStats()
	assert.Equal(t, 3, stats[ResourceTypeBrowser].Active)

	// Cleanup all
	err := rm.CleanupAll()
	assert.NoError(t, err)

	// Verify all resources were cleaned up
	stats = rm.GetStats()
	assert.Equal(t, 0, stats[ResourceTypeBrowser].Active)
}

func TestResourceManager_Shutdown(t *testing.T) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)

	// Register a resource
	resource := &MockResource{
		id:        "test_resource",
		resType:   ResourceTypeBrowser,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		active:    true,
	}
	err := rm.RegisterResource(resource)
	require.NoError(t, err)

	// Shutdown should cleanup all resources
	err = rm.Shutdown()
	assert.NoError(t, err)

	// Verify resources were cleaned up
	stats := rm.GetStats()
	assert.Equal(t, 0, stats[ResourceTypeBrowser].Active)
}

func TestDefaultResourceManagerConfig(t *testing.T) {
	config := DefaultResourceManagerConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.MaxResources, 0)
	assert.Greater(t, config.CleanupInterval, time.Duration(0))
	assert.Greater(t, config.ResourceTimeout, time.Duration(0))
	assert.Greater(t, config.MaxIdleTime, time.Duration(0))
	assert.True(t, config.EnableLeakDetection)
	assert.Greater(t, config.LeakDetectionInterval, time.Duration(0))
}

// MockResource implements the Resource interface for testing
type MockResource struct {
	id         string
	resType    ResourceType
	createdAt  time.Time
	lastUsed   time.Time
	active     bool
	cleanupErr error
}

func (m *MockResource) GetID() string {
	return m.id
}

func (m *MockResource) GetType() ResourceType {
	return m.resType
}

func (m *MockResource) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m *MockResource) GetLastUsed() time.Time {
	return m.lastUsed
}

func (m *MockResource) IsActive() bool {
	return m.active
}

func (m *MockResource) Cleanup() error {
	m.active = false
	return m.cleanupErr
}
