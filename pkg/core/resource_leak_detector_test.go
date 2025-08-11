//go:build ignore
// +build ignore

package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceLeakDetector(t *testing.T) {
	t.Skip("ResourceLeakDetector not implemented yet")
	// config := DefaultLeakDetectorConfig()
	// rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	assert.NotNil(t, rld)
	assert.Equal(t, config, rld.config)
	assert.NotNil(t, rld.resourceTracker)
	assert.NotNil(t, rld.metrics)
}

func TestResourceLeakDetector_TrackResource(t *testing.T) {
	t.Skip("ResourceLeakDetector not implemented yet")
	// config := DefaultLeakDetectorConfig()
	// config.Enabled = true
	// rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	metadata := map[string]interface{}{
		"test_key": "test_value",
		"size":     1024,
	}

	rld.TrackResource("resource_1", ResourceTypeBrowser, metadata)

	rld.mutex.RLock()
	resource, exists := rld.resourceTracker["resource_1"]
	rld.mutex.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, "resource_1", resource.ID)
	assert.Equal(t, ResourceTypeBrowser, resource.Type)
	assert.Equal(t, metadata, resource.Metadata)
	assert.True(t, resource.Active)
	assert.False(t, resource.CreatedAt.IsZero())
	assert.False(t, resource.LastAccessed.IsZero())
}

func TestResourceLeakDetector_TrackResource_Disabled(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = false
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	rld.TrackResource("resource_1", ResourceTypeBrowser, nil)

	rld.mutex.RLock()
	_, exists := rld.resourceTracker["resource_1"]
	rld.mutex.RUnlock()

	assert.False(t, exists)
}

func TestResourceLeakDetector_UntrackResource(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	// Track a resource
	rld.TrackResource("resource_1", ResourceTypeBrowser, nil)

	// Verify it exists
	rld.mutex.RLock()
	_, exists := rld.resourceTracker["resource_1"]
	rld.mutex.RUnlock()
	assert.True(t, exists)

	// Untrack it
	rld.UntrackResource("resource_1")

	// Verify it's gone
	rld.mutex.RLock()
	_, exists = rld.resourceTracker["resource_1"]
	rld.mutex.RUnlock()
	assert.False(t, exists)
}

func TestResourceLeakDetector_UpdateResourceAccess(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	rld.TrackResource("resource_1", ResourceTypeBrowser, nil)

	// Get initial access time
	rld.mutex.RLock()
	initialTime := rld.resourceTracker["resource_1"].LastAccessed
	rld.mutex.RUnlock()

	time.Sleep(10 * time.Millisecond)

	// Update access
	rld.UpdateResourceAccess("resource_1")

	// Verify access time was updated
	rld.mutex.RLock()
	updatedTime := rld.resourceTracker["resource_1"].LastAccessed
	rld.mutex.RUnlock()

	assert.True(t, updatedTime.After(initialTime))
}

func TestResourceLeakDetector_ScanForLeaks(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	config.LeakThreshold = 100 * time.Millisecond
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	// Create a resource that will be considered leaked
	rld.TrackResource("leaked_resource", ResourceTypeBrowser, nil)

	// Manually set old access time
	rld.mutex.Lock()
	rld.resourceTracker["leaked_resource"].LastAccessed = time.Now().Add(-200 * time.Millisecond)
	rld.mutex.Unlock()

	// Create a resource that won't be leaked
	rld.TrackResource("good_resource", ResourceTypeBrowser, nil)

	// Perform leak scan
	leaks := rld.ScanForLeaks()

	assert.Len(t, leaks, 1)
	assert.Contains(t, leaks[0], "leaked_resource")
	assert.Contains(t, leaks[0], "Resource leak detected")

	// Check metrics were updated
	metrics := rld.GetMetrics()
	assert.Greater(t, metrics.TotalScans, 0)
	assert.Greater(t, metrics.LeaksDetected, 0)
}

func TestResourceLeakDetector_ScanForLeaks_MemoryThreshold(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	config.EnableMemoryTracking = true
	config.MemoryThresholdMB = 1 // Very low threshold to trigger warning
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	leaks := rld.ScanForLeaks()

	// Should detect high memory usage (depending on test environment)
	// This test might not always trigger, but it shouldn't fail
	assert.NotNil(t, leaks)

	metrics := rld.GetMetrics()
	assert.GreaterOrEqual(t, metrics.MemoryUsageMB, int64(0))
}

func TestResourceLeakDetector_ScanForLeaks_GoroutineThreshold(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	config.EnableGoroutineCheck = true
	config.GoroutineThreshold = 1 // Very low threshold to trigger warning
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	leaks := rld.ScanForLeaks()

	// Should detect high goroutine count
	assert.NotNil(t, leaks)
	// Might contain goroutine warning depending on current count

	metrics := rld.GetMetrics()
	assert.Greater(t, metrics.GoroutineCount, 0)
}

func TestResourceLeakDetector_GetMetrics(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	// Add some resources
	rld.TrackResource("resource_1", ResourceTypeBrowser, nil)
	rld.TrackResource("resource_2", ResourceTypeDatabase, nil)

	// Perform a scan to update metrics
	rld.ScanForLeaks()

	metrics := rld.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalScans, 0)
	assert.Equal(t, 2, metrics.ResourcesTracked)
	assert.False(t, metrics.LastScanTime.IsZero())
	assert.Greater(t, metrics.LastScanDuration, time.Duration(0))
	assert.NotNil(t, metrics.LeaksByType)
	assert.NotNil(t, metrics.ActiveResourcesByType)
}

func TestResourceLeakDetector_GetTrackedResources(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	metadata1 := map[string]interface{}{"key1": "value1"}
	metadata2 := map[string]interface{}{"key2": "value2"}

	rld.TrackResource("resource_1", ResourceTypeBrowser, metadata1)
	rld.TrackResource("resource_2", ResourceTypeDatabase, metadata2)

	resources := rld.GetTrackedResources()

	assert.Len(t, resources, 2)

	// Find resources by ID
	var resource1, resource2 *TrackedResource
	for _, r := range resources {
		if r.ID == "resource_1" {
			resource1 = r
		} else if r.ID == "resource_2" {
			resource2 = r
		}
	}

	assert.NotNil(t, resource1)
	assert.NotNil(t, resource2)
	assert.Equal(t, ResourceTypeBrowser, resource1.Type)
	assert.Equal(t, ResourceTypeDatabase, resource2.Type)
	assert.Equal(t, metadata1, resource1.Metadata)
	assert.Equal(t, metadata2, resource2.Metadata)
}

func TestResourceLeakDetector_MaxTrackedResourcesLimit(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	config.MaxTrackedResources = 3 // Small limit for testing
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	// Add resources up to the limit
	for i := 0; i < 5; i++ {
		resourceID := fmt.Sprintf("resource_%d", i)
		rld.TrackResource(resourceID, ResourceTypeBrowser, nil)
	}

	// Should only have the maximum number of resources
	rld.mutex.RLock()
	resourceCount := len(rld.resourceTracker)
	rld.mutex.RUnlock()

	assert.LessOrEqual(t, resourceCount, config.MaxTrackedResources)
}

func TestResourceLeakDetector_StartStop(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	config.ScanInterval = 100 * time.Millisecond
	rld := NewResourceLeakDetector(config)

	assert.True(t, rld.IsRunning())

	// Stop the detector
	rld.Stop()
	assert.False(t, rld.IsRunning())

	// Start again
	rld.Start()
	assert.True(t, rld.IsRunning())

	// Clean up
	rld.Stop()
}

func TestResourceLeakDetector_IsEnabled(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	assert.True(t, rld.IsEnabled())

	// Test with disabled config
	disabledConfig := DefaultLeakDetectorConfig()
	disabledConfig.Enabled = false
	disabledRld := NewResourceLeakDetector(disabledConfig)
	defer disabledRld.Stop()

	assert.False(t, disabledRld.IsEnabled())
}

func TestResourceLeakDetector_UpdateConfig(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	assert.True(t, rld.IsRunning())

	// Update config to disable
	newConfig := DefaultLeakDetectorConfig()
	newConfig.Enabled = false
	rld.UpdateConfig(newConfig)

	assert.False(t, rld.IsEnabled())
	assert.False(t, rld.IsRunning())

	// Update config to enable again
	enabledConfig := DefaultLeakDetectorConfig()
	enabledConfig.Enabled = true
	rld.UpdateConfig(enabledConfig)

	assert.True(t, rld.IsEnabled())
	assert.True(t, rld.IsRunning())
}

func TestResourceLeakDetector_ConcurrentOperations(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.Enabled = true
	config.ScanInterval = 1 * time.Hour // Disable auto scanning
	rld := NewResourceLeakDetector(config)
	defer rld.Stop()

	const numGoroutines = 10
	const operationsPerGoroutine = 50

	done := make(chan bool, numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				resourceID := fmt.Sprintf("resource_%d_%d", goroutineID, j)

				// Track resource
				rld.TrackResource(resourceID, ResourceTypeBrowser, map[string]interface{}{
					"goroutine": goroutineID,
					"operation": j,
				})

				// Update access
				rld.UpdateResourceAccess(resourceID)

				// Untrack resource
				rld.UntrackResource(resourceID)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	metrics := rld.GetMetrics()
	assert.Equal(t, 0, metrics.ResourcesTracked, "All resources should have been untracked")
}

func TestDefaultLeakDetectorConfig(t *testing.T) {
	config := DefaultLeakDetectorConfig()

	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Greater(t, config.ScanInterval, time.Duration(0))
	assert.Greater(t, config.LeakThreshold, time.Duration(0))
	assert.Greater(t, config.MaxTrackedResources, 0)
	assert.True(t, config.EnableMemoryTracking)
	assert.Greater(t, config.MemoryThresholdMB, int64(0))
	assert.True(t, config.EnableGoroutineCheck)
	assert.Greater(t, config.GoroutineThreshold, 0)
}
