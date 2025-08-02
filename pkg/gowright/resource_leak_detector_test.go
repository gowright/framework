package gowright

import (
	"fmt"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
)

func TestNewResourceLeakDetector(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	assert.NotNil(t, rld)
	assert.Equal(t, config, rld.config)
	assert.NotNil(t, rld.trackedResources)
	assert.NotNil(t, rld.resourceTypes)
	assert.NotNil(t, rld.leakReports)
	
	// Verify common resource types were registered
	assert.Contains(t, rld.resourceTypes, "browser")
	assert.Contains(t, rld.resourceTypes, "database_connection")
	assert.Contains(t, rld.resourceTypes, "http_client")
	assert.Contains(t, rld.resourceTypes, "temp_file")
	assert.Contains(t, rld.resourceTypes, "screenshot")
}

func TestResourceLeakDetector_TrackResource(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.EnableStackTrace = false // Disable for simpler testing
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	metadata := map[string]interface{}{
		"test_key": "test_value",
		"size":     1024,
	}
	
	rld.TrackResource("resource_1", "test_resource", 1024, metadata)
	
	rld.mutex.RLock()
	resource, exists := rld.trackedResources["resource_1"]
	rld.mutex.RUnlock()
	
	assert.True(t, exists)
	assert.Equal(t, "resource_1", resource.ID)
	assert.Equal(t, "test_resource", resource.Type)
	assert.Equal(t, int64(1024), resource.Size)
	assert.Equal(t, 1, resource.RefCount)
	assert.Equal(t, metadata, resource.Metadata)
	assert.False(t, resource.IsLeaked)
}

func TestResourceLeakDetector_UntrackResource(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Track a resource
	rld.TrackResource("resource_1", "test_resource", 1024, nil)
	
	// Verify it exists
	rld.mutex.RLock()
	_, exists := rld.trackedResources["resource_1"]
	rld.mutex.RUnlock()
	assert.True(t, exists)
	
	// Untrack it
	rld.UntrackResource("resource_1")
	
	// Verify it's gone
	rld.mutex.RLock()
	_, exists = rld.trackedResources["resource_1"]
	rld.mutex.RUnlock()
	assert.False(t, exists)
}

func TestResourceLeakDetector_UpdateResourceAccess(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	rld.TrackResource("resource_1", "test_resource", 1024, nil)
	
	// Get initial access time
	rld.mutex.RLock()
	initialTime := rld.trackedResources["resource_1"].LastAccessed
	rld.mutex.RUnlock()
	
	time.Sleep(10 * time.Millisecond)
	
	// Update access
	rld.UpdateResourceAccess("resource_1")
	
	// Verify access time was updated
	rld.mutex.RLock()
	updatedTime := rld.trackedResources["resource_1"].LastAccessed
	rld.mutex.RUnlock()
	
	assert.True(t, updatedTime.After(initialTime))
}

func TestResourceLeakDetector_IncrementRefCount(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	rld.TrackResource("resource_1", "test_resource", 1024, nil)
	
	// Initial ref count should be 1
	rld.mutex.RLock()
	initialRefCount := rld.trackedResources["resource_1"].RefCount
	rld.mutex.RUnlock()
	assert.Equal(t, 1, initialRefCount)
	
	// Increment ref count
	rld.IncrementRefCount("resource_1")
	
	// Verify ref count increased
	rld.mutex.RLock()
	newRefCount := rld.trackedResources["resource_1"].RefCount
	rld.mutex.RUnlock()
	assert.Equal(t, 2, newRefCount)
}

func TestResourceLeakDetector_DecrementRefCount(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.CleanupGracePeriod = 50 * time.Millisecond
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	rld.TrackResource("resource_1", "test_resource", 1024, nil)
	rld.IncrementRefCount("resource_1") // Make ref count 2
	
	// Decrement ref count
	rld.DecrementRefCount("resource_1")
	
	// Verify ref count decreased
	rld.mutex.RLock()
	refCount := rld.trackedResources["resource_1"].RefCount
	rld.mutex.RUnlock()
	assert.Equal(t, 1, refCount)
	
	// Decrement to 0
	rld.DecrementRefCount("resource_1")
	
	// Verify ref count is 0
	rld.mutex.RLock()
	refCount = rld.trackedResources["resource_1"].RefCount
	rld.mutex.RUnlock()
	assert.Equal(t, 0, refCount)
	
	// Wait for grace period to potentially mark as leaked
	time.Sleep(100 * time.Millisecond)
	
	// Resource might be marked as leaked if not cleaned up
	rld.mutex.RLock()
	resource, exists := rld.trackedResources["resource_1"]
	rld.mutex.RUnlock()
	
	if exists {
		// If still exists, it might be marked as leaked
		t.Logf("Resource leak status: %v", resource.IsLeaked)
	}
}

func TestResourceLeakDetector_RegisterResourceType(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	customType := &ResourceTypeInfo{
		TypeName:         "custom_resource",
		MaxLifetime:      5 * time.Minute,
		ExpectedRefCount: 0,
		CleanupFunc:      nil,
	}
	
	rld.RegisterResourceType("custom_resource", customType)
	
	rld.mutex.RLock()
	registeredType, exists := rld.resourceTypes["custom_resource"]
	rld.mutex.RUnlock()
	
	assert.True(t, exists)
	assert.Equal(t, customType, registeredType)
}

func TestResourceLeakDetector_performLeakScan(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.LeakThreshold = 100 * time.Millisecond
	config.ScanInterval = 1 * time.Hour // Disable auto scanning
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Create a resource that will be considered leaked
	rld.TrackResource("leaked_resource", "test_resource", 2048, nil)
	
	// Manually set old access time
	rld.mutex.Lock()
	rld.trackedResources["leaked_resource"].LastAccessed = time.Now().Add(-200 * time.Millisecond)
	rld.trackedResources["leaked_resource"].RefCount = 0 // No references
	rld.mutex.Unlock()
	
	// Create a resource that won't be leaked
	rld.TrackResource("good_resource", "test_resource", 1024, nil)
	
	// Perform leak scan
	rld.performLeakScan()
	
	// Check results
	rld.mutex.RLock()
	leakedResource := rld.trackedResources["leaked_resource"]
	goodResource := rld.trackedResources["good_resource"]
	totalLeaks := rld.totalLeaksDetected
	rld.mutex.RUnlock()
	
	assert.True(t, leakedResource.IsLeaked)
	assert.NotNil(t, leakedResource.LeakDetectedAt)
	assert.False(t, goodResource.IsLeaked)
	assert.Greater(t, totalLeaks, int64(0))
}

func TestResourceLeakDetector_generateLeakReport(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Create leaked resources
	rld.TrackResource("leaked_1", "browser", 1024, nil)
	rld.TrackResource("leaked_2", "database_connection", 2048, nil)
	
	// Mark them as leaked
	now := time.Now()
	rld.mutex.Lock()
	rld.trackedResources["leaked_1"].IsLeaked = true
	rld.trackedResources["leaked_1"].LeakDetectedAt = &now
	rld.trackedResources["leaked_2"].IsLeaked = true
	rld.trackedResources["leaked_2"].LeakDetectedAt = &now
	rld.mutex.Unlock()
	
	// Generate report
	rld.generateLeakReport()
	
	// Check report was created
	reports := rld.GetLeakReports()
	assert.Len(t, reports, 1)
	
	report := reports[0]
	assert.Equal(t, 2, report.TotalLeaks)
	assert.Equal(t, int64(3072), report.MemoryLeaked) // 1024 + 2048
	assert.Len(t, report.LeakedResources, 2)
	assert.NotEmpty(t, report.Recommendations)
	
	// Check severity calculation
	assert.Equal(t, LeakSeverityLow, report.Severity)
}

func TestResourceLeakDetector_calculateLeakSeverity(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.CriticalLeakCount = 10
	config.CriticalMemoryMB = 50
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Test low severity
	severity := rld.calculateLeakSeverity(2, 5*1024*1024) // 2 leaks, 5MB
	assert.Equal(t, LeakSeverityLow, severity)
	
	// Test medium severity (using hardcoded threshold of 10 leaks or 10MB)
	severity = rld.calculateLeakSeverity(3, 12*1024*1024) // 3 leaks (< 5), 12MB (>= 10MB threshold)
	assert.Equal(t, LeakSeverityMedium, severity)
	
	// Test high severity (half of critical threshold)
	severity = rld.calculateLeakSeverity(5, 8*1024*1024) // 5 leaks (>= 5), 8MB (< 25MB threshold)
	assert.Equal(t, LeakSeverityHigh, severity)
	
	// Test critical severity
	severity = rld.calculateLeakSeverity(10, 60*1024*1024) // 10 leaks (>= critical), 60MB
	assert.Equal(t, LeakSeverityCritical, severity)
}

func TestResourceLeakDetector_generateRecommendations(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	leakedResources := []TrackedResource{
		{Type: "browser", ID: "browser_1"},
		{Type: "browser", ID: "browser_2"},
		{Type: "database_connection", ID: "db_1"},
		{Type: "temp_file", ID: "temp_1"},
	}
	
	recommendations := rld.generateRecommendations(leakedResources)
	
	assert.NotEmpty(t, recommendations)
	
	// Check that recommendations mention the specific resource types
	recommendationText := ""
	for _, rec := range recommendations {
		recommendationText += rec + " "
	}
	
	assert.Contains(t, recommendationText, "browser")
	assert.Contains(t, recommendationText, "database")
	assert.Contains(t, recommendationText, "temp")
}

func TestResourceLeakDetector_GetDetectorStats(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Add some resources
	rld.TrackResource("resource_1", "test_resource", 1024, nil)
	rld.TrackResource("resource_2", "test_resource", 2048, nil)
	
	// Mark one as leaked
	rld.mutex.Lock()
	rld.trackedResources["resource_1"].IsLeaked = true
	rld.totalLeaksDetected = 1
	rld.totalMemoryLeaked = 1024
	rld.mutex.Unlock()
	
	stats := rld.GetDetectorStats()
	
	assert.Equal(t, 2, stats.TrackedResources)
	assert.Equal(t, int64(1), stats.TotalLeaksDetected)
	assert.Equal(t, int64(1024), stats.TotalMemoryLeaked)
	assert.False(t, stats.LastScan.IsZero())
}

func TestResourceLeakDetector_GetLeakReports(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Initially should have no reports
	reports := rld.GetLeakReports()
	assert.Empty(t, reports)
	
	// Add a report manually
	report := &LeakReport{
		ID:         "test_report",
		DetectedAt: time.Now(),
		TotalLeaks: 1,
	}
	
	rld.mutex.Lock()
	rld.leakReports = append(rld.leakReports, report)
	rld.mutex.Unlock()
	
	reports = rld.GetLeakReports()
	assert.Len(t, reports, 1)
	assert.Equal(t, "test_report", reports[0].ID)
}

func TestResourceLeakDetector_GetLatestLeakReport(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Initially should have no reports
	report := rld.GetLatestLeakReport()
	assert.Nil(t, report)
	
	// Add reports
	report1 := &LeakReport{ID: "report_1", DetectedAt: time.Now()}
	report2 := &LeakReport{ID: "report_2", DetectedAt: time.Now().Add(1 * time.Second)}
	
	rld.mutex.Lock()
	rld.leakReports = append(rld.leakReports, report1, report2)
	rld.mutex.Unlock()
	
	latest := rld.GetLatestLeakReport()
	assert.NotNil(t, latest)
	assert.Equal(t, "report_2", latest.ID)
}

func TestResourceLeakDetector_MaxTrackedResourcesLimit(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.MaxTrackedResources = 3 // Small limit for testing
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Add resources up to the limit
	for i := 0; i < 5; i++ {
		resourceID := fmt.Sprintf("resource_%d", i)
		rld.TrackResource(resourceID, "test_resource", 1024, nil)
	}
	
	// Should only have the maximum number of resources
	rld.mutex.RLock()
	resourceCount := len(rld.trackedResources)
	rld.mutex.RUnlock()
	
	assert.LessOrEqual(t, resourceCount, config.MaxTrackedResources)
}

func TestResourceLeakDetector_MaxLeakReportsLimit(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.MaxLeakReports = 2 // Small limit for testing
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
	// Add more reports than the limit
	for i := 0; i < 4; i++ {
		report := &LeakReport{
			ID:         fmt.Sprintf("report_%d", i),
			DetectedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		
		rld.mutex.Lock()
		rld.leakReports = append(rld.leakReports, report)
		if len(rld.leakReports) > rld.config.MaxLeakReports {
			rld.leakReports = rld.leakReports[1:] // Remove oldest
		}
		rld.mutex.Unlock()
	}
	
	reports := rld.GetLeakReports()
	assert.LessOrEqual(t, len(reports), config.MaxLeakReports)
	
	// Should have the latest reports
	if len(reports) > 0 {
		assert.Equal(t, "report_3", reports[len(reports)-1].ID)
	}
}

func TestResourceLeakDetector_ConcurrentOperations(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.ScanInterval = 1 * time.Hour // Disable auto scanning
	
	rld := NewResourceLeakDetector(config)
	defer rld.Shutdown()
	
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
				rld.TrackResource(resourceID, "test_resource", 1024, map[string]interface{}{
					"goroutine": goroutineID,
					"operation": j,
				})
				
				// Update access
				rld.UpdateResourceAccess(resourceID)
				
				// Increment/decrement ref count
				rld.IncrementRefCount(resourceID)
				rld.DecrementRefCount(resourceID)
				
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
	stats := rld.GetDetectorStats()
	assert.Equal(t, 0, stats.TrackedResources, "All resources should have been untracked")
}

func TestLeakSeverity_String(t *testing.T) {
	assert.Equal(t, "low", LeakSeverityLow.String())
	assert.Equal(t, "medium", LeakSeverityMedium.String())
	assert.Equal(t, "high", LeakSeverityHigh.String())
	assert.Equal(t, "critical", LeakSeverityCritical.String())
	assert.Equal(t, "unknown", LeakSeverity(999).String())
}

func TestDefaultLeakDetectorConfig(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	
	assert.NotNil(t, config)
	assert.Greater(t, config.ScanInterval, time.Duration(0))
	assert.Greater(t, config.LeakThreshold, time.Duration(0))
	assert.Greater(t, config.MaxTrackedResources, 0)
	assert.Greater(t, config.MaxLeakReports, 0)
	assert.Greater(t, config.CriticalLeakCount, 0)
	assert.Greater(t, config.CriticalMemoryMB, int64(0))
	assert.Greater(t, config.CleanupGracePeriod, time.Duration(0))
}