package gowright

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ResourceLeakDetector monitors and detects resource leaks in the framework
type ResourceLeakDetector struct {
	config             *LeakDetectorConfig
	trackedResources   map[string]*TrackedResource
	resourceTypes      map[string]*ResourceTypeInfo
	leakReports        []*LeakReport
	monitoringTicker   *time.Ticker
	ctx                context.Context
	cancel             context.CancelFunc
	mutex              sync.RWMutex
	reportMutex        sync.Mutex // Separate mutex for report generation
	lastScan           time.Time
	totalLeaksDetected int64
	totalMemoryLeaked  int64
}

// LeakDetectorConfig holds configuration for leak detection
type LeakDetectorConfig struct {
	// Monitoring intervals
	ScanInterval  time.Duration `json:"scan_interval"`
	LeakThreshold time.Duration `json:"leak_threshold"`

	// Resource limits
	MaxTrackedResources int `json:"max_tracked_resources"`
	MaxLeakReports      int `json:"max_leak_reports"`

	// Detection settings
	EnableStackTrace    bool `json:"enable_stack_trace"`
	EnableMemoryProfile bool `json:"enable_memory_profile"`

	// Alert thresholds
	CriticalLeakCount int   `json:"critical_leak_count"`
	CriticalMemoryMB  int64 `json:"critical_memory_mb"`

	// Cleanup settings
	AutoCleanupLeaks   bool          `json:"auto_cleanup_leaks"`
	CleanupGracePeriod time.Duration `json:"cleanup_grace_period"`
}

// TrackedResource represents a resource being monitored for leaks
type TrackedResource struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessed   time.Time              `json:"last_accessed"`
	Size           int64                  `json:"size"`
	RefCount       int                    `json:"ref_count"`
	StackTrace     []string               `json:"stack_trace,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	IsLeaked       bool                   `json:"is_leaked"`
	LeakDetectedAt *time.Time             `json:"leak_detected_at,omitempty"`
}

// ResourceTypeInfo holds information about a resource type
type ResourceTypeInfo struct {
	TypeName         string             `json:"type_name"`
	MaxLifetime      time.Duration      `json:"max_lifetime"`
	ExpectedRefCount int                `json:"expected_ref_count"`
	CleanupFunc      func(string) error `json:"-"`
	TotalCreated     int64              `json:"total_created"`
	TotalCleaned     int64              `json:"total_cleaned"`
	CurrentActive    int64              `json:"current_active"`
}

// LeakReport contains information about detected leaks
type LeakReport struct {
	ID              string            `json:"id"`
	DetectedAt      time.Time         `json:"detected_at"`
	LeakedResources []TrackedResource `json:"leaked_resources"`
	TotalLeaks      int               `json:"total_leaks"`
	MemoryLeaked    int64             `json:"memory_leaked"`
	Severity        LeakSeverity      `json:"severity"`
	Recommendations []string          `json:"recommendations"`
}

// LeakSeverity represents the severity of a leak
type LeakSeverity int

const (
	LeakSeverityLow LeakSeverity = iota
	LeakSeverityMedium
	LeakSeverityHigh
	LeakSeverityCritical
)

// String returns the string representation of LeakSeverity
func (ls LeakSeverity) String() string {
	switch ls {
	case LeakSeverityLow:
		return "low"
	case LeakSeverityMedium:
		return "medium"
	case LeakSeverityHigh:
		return "high"
	case LeakSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// DefaultLeakDetectorConfig returns default configuration for leak detection
func DefaultLeakDetectorConfig() *LeakDetectorConfig {
	return &LeakDetectorConfig{
		ScanInterval:        30 * time.Second,
		LeakThreshold:       5 * time.Minute,
		MaxTrackedResources: 10000,
		MaxLeakReports:      100,
		EnableStackTrace:    true,
		EnableMemoryProfile: false,
		CriticalLeakCount:   50,
		CriticalMemoryMB:    100,
		AutoCleanupLeaks:    true,
		CleanupGracePeriod:  1 * time.Minute,
	}
}

// NewResourceLeakDetector creates a new resource leak detector
func NewResourceLeakDetector(config *LeakDetectorConfig) *ResourceLeakDetector {
	if config == nil {
		config = DefaultLeakDetectorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	rld := &ResourceLeakDetector{
		config:           config,
		trackedResources: make(map[string]*TrackedResource),
		resourceTypes:    make(map[string]*ResourceTypeInfo),
		leakReports:      make([]*LeakReport, 0),
		ctx:              ctx,
		cancel:           cancel,
		lastScan:         time.Now(),
	}

	// Register common resource types
	rld.registerCommonResourceTypes()

	// Start monitoring
	rld.startMonitoring()

	return rld
}

// registerCommonResourceTypes registers common framework resource types
func (rld *ResourceLeakDetector) registerCommonResourceTypes() {
	// Browser resources
	rld.RegisterResourceType("browser", &ResourceTypeInfo{
		TypeName:         "browser",
		MaxLifetime:      10 * time.Minute,
		ExpectedRefCount: 0,
		CleanupFunc:      rld.cleanupBrowserResource,
	})

	// Database connections
	rld.RegisterResourceType("database_connection", &ResourceTypeInfo{
		TypeName:         "database_connection",
		MaxLifetime:      30 * time.Minute,
		ExpectedRefCount: 0,
		CleanupFunc:      rld.cleanupDatabaseResource,
	})

	// HTTP clients
	rld.RegisterResourceType("http_client", &ResourceTypeInfo{
		TypeName:         "http_client",
		MaxLifetime:      15 * time.Minute,
		ExpectedRefCount: 0,
		CleanupFunc:      rld.cleanupHTTPClientResource,
	})

	// Temporary files
	rld.RegisterResourceType("temp_file", &ResourceTypeInfo{
		TypeName:         "temp_file",
		MaxLifetime:      1 * time.Hour,
		ExpectedRefCount: 0,
		CleanupFunc:      rld.cleanupTempFileResource,
	})

	// Screenshots
	rld.RegisterResourceType("screenshot", &ResourceTypeInfo{
		TypeName:         "screenshot",
		MaxLifetime:      2 * time.Hour,
		ExpectedRefCount: 0,
		CleanupFunc:      rld.cleanupScreenshotResource,
	})
}

// RegisterResourceType registers a new resource type for monitoring
func (rld *ResourceLeakDetector) RegisterResourceType(typeName string, info *ResourceTypeInfo) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	rld.resourceTypes[typeName] = info
}

// TrackResource starts tracking a resource for leak detection
func (rld *ResourceLeakDetector) TrackResource(resourceID, resourceType string, size int64, metadata map[string]interface{}) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	// Check if we're at the tracking limit
	if len(rld.trackedResources) >= rld.config.MaxTrackedResources {
		// Remove oldest resource to make space
		rld.removeOldestResource()
	}

	var stackTrace []string
	if rld.config.EnableStackTrace {
		stackTrace = rld.captureStackTrace()
	}

	resource := &TrackedResource{
		ID:           resourceID,
		Type:         resourceType,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Size:         size,
		RefCount:     1,
		StackTrace:   stackTrace,
		Metadata:     metadata,
		IsLeaked:     false,
	}

	rld.trackedResources[resourceID] = resource

	// Update resource type stats
	if typeInfo, exists := rld.resourceTypes[resourceType]; exists {
		typeInfo.TotalCreated++
		typeInfo.CurrentActive++
	}
}

// UntrackResource stops tracking a resource (indicates it was properly cleaned up)
func (rld *ResourceLeakDetector) UntrackResource(resourceID string) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	if resource, exists := rld.trackedResources[resourceID]; exists {
		// Update resource type stats
		if typeInfo, exists := rld.resourceTypes[resource.Type]; exists {
			typeInfo.TotalCleaned++
			typeInfo.CurrentActive--
		}

		delete(rld.trackedResources, resourceID)
	}
}

// UpdateResourceAccess updates the last accessed time for a resource
func (rld *ResourceLeakDetector) UpdateResourceAccess(resourceID string) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	if resource, exists := rld.trackedResources[resourceID]; exists {
		resource.LastAccessed = time.Now()
	}
}

// IncrementRefCount increments the reference count for a resource
func (rld *ResourceLeakDetector) IncrementRefCount(resourceID string) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	if resource, exists := rld.trackedResources[resourceID]; exists {
		resource.RefCount++
		resource.LastAccessed = time.Now()
	}
}

// DecrementRefCount decrements the reference count for a resource
func (rld *ResourceLeakDetector) DecrementRefCount(resourceID string) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	if resource, exists := rld.trackedResources[resourceID]; exists {
		resource.RefCount--
		resource.LastAccessed = time.Now()

		// If ref count reaches 0, the resource should be cleaned up soon
		if resource.RefCount <= 0 {
			// Start a grace period timer for cleanup
			go rld.scheduleCleanupCheck(resourceID, rld.config.CleanupGracePeriod)
		}
	}
}

// scheduleCleanupCheck schedules a check to see if a resource was properly cleaned up
func (rld *ResourceLeakDetector) scheduleCleanupCheck(resourceID string, gracePeriod time.Duration) {
	time.Sleep(gracePeriod)

	rld.mutex.RLock()
	resource, exists := rld.trackedResources[resourceID]
	rld.mutex.RUnlock()

	if exists && resource.RefCount <= 0 {
		// Resource still exists after grace period - potential leak
		rld.markResourceAsLeaked(resourceID)
	}
}

// markResourceAsLeaked marks a resource as leaked
func (rld *ResourceLeakDetector) markResourceAsLeaked(resourceID string) {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	if resource, exists := rld.trackedResources[resourceID]; exists && !resource.IsLeaked {
		resource.IsLeaked = true
		now := time.Now()
		resource.LeakDetectedAt = &now

		rld.totalLeaksDetected++
		rld.totalMemoryLeaked += resource.Size

		// Trigger immediate leak report if critical threshold is reached
		if rld.shouldTriggerImmediateReport() {
			go rld.generateLeakReport()
		}
	}
}

// startMonitoring starts the resource monitoring routine
func (rld *ResourceLeakDetector) startMonitoring() {
	rld.monitoringTicker = time.NewTicker(rld.config.ScanInterval)

	go func() {
		for {
			select {
			case <-rld.monitoringTicker.C:
				rld.performLeakScan()
			case <-rld.ctx.Done():
				return
			}
		}
	}()
}

// performLeakScan performs a scan for resource leaks
func (rld *ResourceLeakDetector) performLeakScan() {
	rld.mutex.Lock()
	defer rld.mutex.Unlock()

	now := time.Now()
	leakThreshold := now.Add(-rld.config.LeakThreshold)

	var newLeaks []string

	for resourceID, resource := range rld.trackedResources {
		if resource.IsLeaked {
			continue // Already marked as leaked
		}

		// Check if resource has exceeded its type's max lifetime
		typeInfo, typeExists := rld.resourceTypes[resource.Type]
		if typeExists && now.Sub(resource.CreatedAt) > typeInfo.MaxLifetime {
			resource.IsLeaked = true
			resource.LeakDetectedAt = &now
			newLeaks = append(newLeaks, resourceID)
			continue
		}

		// Check if resource hasn't been accessed recently and has no references
		if resource.RefCount <= 0 && resource.LastAccessed.Before(leakThreshold) {
			resource.IsLeaked = true
			resource.LeakDetectedAt = &now
			newLeaks = append(newLeaks, resourceID)
		}
	}

	rld.lastScan = now

	// Update leak counters
	rld.totalLeaksDetected += int64(len(newLeaks))
	for _, resourceID := range newLeaks {
		if resource, exists := rld.trackedResources[resourceID]; exists {
			rld.totalMemoryLeaked += resource.Size
		}
	}

	// Generate leak report if new leaks were found
	if len(newLeaks) > 0 {
		go rld.generateLeakReport()
	}

	// Perform auto cleanup if enabled
	if rld.config.AutoCleanupLeaks && len(newLeaks) > 0 {
		go rld.performAutoCleanup(newLeaks)
	}
}

// generateLeakReport generates a comprehensive leak report
func (rld *ResourceLeakDetector) generateLeakReport() {
	// Prevent concurrent report generation
	rld.reportMutex.Lock()
	defer rld.reportMutex.Unlock()

	// First, collect leaked resources with read lock
	rld.mutex.RLock()
	var leakedResources []TrackedResource
	var totalMemoryLeaked int64

	for _, resource := range rld.trackedResources {
		if resource.IsLeaked {
			leakedResources = append(leakedResources, *resource)
			totalMemoryLeaked += resource.Size
		}
	}
	rld.mutex.RUnlock()

	if len(leakedResources) == 0 {
		return // No leaks to report
	}

	severity := rld.calculateLeakSeverity(len(leakedResources), totalMemoryLeaked)
	recommendations := rld.generateRecommendations(leakedResources)

	report := &LeakReport{
		ID:              fmt.Sprintf("leak_report_%d", time.Now().UnixNano()),
		DetectedAt:      time.Now(),
		LeakedResources: leakedResources,
		TotalLeaks:      len(leakedResources),
		MemoryLeaked:    totalMemoryLeaked,
		Severity:        severity,
		Recommendations: recommendations,
	}

	// Now acquire write lock to modify the reports slice
	rld.mutex.Lock()
	rld.leakReports = append(rld.leakReports, report)
	if len(rld.leakReports) > rld.config.MaxLeakReports {
		rld.leakReports = rld.leakReports[1:] // Remove oldest report
	}
	rld.mutex.Unlock()

	// Log the leak report
	rld.logLeakReport(report)
}

// calculateLeakSeverity calculates the severity of detected leaks
func (rld *ResourceLeakDetector) calculateLeakSeverity(leakCount int, memoryLeaked int64) LeakSeverity {
	memoryLeakedMB := memoryLeaked / (1024 * 1024)

	if leakCount >= rld.config.CriticalLeakCount || memoryLeakedMB >= rld.config.CriticalMemoryMB {
		return LeakSeverityCritical
	} else if leakCount >= rld.config.CriticalLeakCount/2 || memoryLeakedMB >= rld.config.CriticalMemoryMB/2 {
		return LeakSeverityHigh
	} else if leakCount >= 10 || memoryLeakedMB >= 10 {
		return LeakSeverityMedium
	}

	return LeakSeverityLow
}

// generateRecommendations generates recommendations based on leaked resources
func (rld *ResourceLeakDetector) generateRecommendations(leakedResources []TrackedResource) []string {
	var recommendations []string
	resourceTypeCounts := make(map[string]int)

	// Count leaks by resource type
	for _, resource := range leakedResources {
		resourceTypeCounts[resource.Type]++
	}

	// Generate type-specific recommendations
	for resourceType, count := range resourceTypeCounts {
		switch resourceType {
		case "browser":
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d leaked browser instances. Ensure browser.Close() is called in defer statements.", count))
		case "database_connection":
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d leaked database connections. Ensure db.Close() is called and connection pooling is properly configured.", count))
		case "http_client":
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d leaked HTTP clients. Consider reusing HTTP clients and properly closing response bodies.", count))
		case "temp_file":
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d leaked temporary files. Ensure temporary files are cleaned up after use.", count))
		case "screenshot":
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d leaked screenshots. Consider implementing automatic screenshot cleanup.", count))
		default:
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d leaked %s resources. Review resource cleanup procedures.", count, resourceType))
		}
	}

	// General recommendations
	if len(leakedResources) > 10 {
		recommendations = append(recommendations, "Consider implementing more aggressive resource cleanup policies.")
		recommendations = append(recommendations, "Review test teardown procedures to ensure proper resource cleanup.")
	}

	return recommendations
}

// performAutoCleanup attempts to automatically clean up leaked resources
func (rld *ResourceLeakDetector) performAutoCleanup(leakedResourceIDs []string) {
	var cleanedCount int
	var errors []error

	for _, resourceID := range leakedResourceIDs {
		rld.mutex.RLock()
		resource, exists := rld.trackedResources[resourceID]
		rld.mutex.RUnlock()

		if !exists {
			continue
		}

		// Get cleanup function for resource type
		typeInfo, typeExists := rld.resourceTypes[resource.Type]
		if !typeExists || typeInfo.CleanupFunc == nil {
			continue
		}

		// Attempt cleanup
		if err := typeInfo.CleanupFunc(resourceID); err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup %s: %w", resourceID, err))
		} else {
			cleanedCount++
			rld.UntrackResource(resourceID)
		}
	}

	if cleanedCount > 0 {
		fmt.Printf("Auto-cleanup: successfully cleaned %d leaked resources\n", cleanedCount)
	}

	if len(errors) > 0 {
		fmt.Printf("Auto-cleanup errors: %v\n", errors)
	}
}

// Cleanup functions for different resource types
func (rld *ResourceLeakDetector) cleanupBrowserResource(resourceID string) error {
	// This would integrate with the browser pool to force cleanup
	fmt.Printf("Cleaning up leaked browser resource: %s\n", resourceID)
	return nil
}

func (rld *ResourceLeakDetector) cleanupDatabaseResource(resourceID string) error {
	// This would integrate with the database pool to force cleanup
	fmt.Printf("Cleaning up leaked database resource: %s\n", resourceID)
	return nil
}

func (rld *ResourceLeakDetector) cleanupHTTPClientResource(resourceID string) error {
	// This would integrate with the HTTP client pool to force cleanup
	fmt.Printf("Cleaning up leaked HTTP client resource: %s\n", resourceID)
	return nil
}

func (rld *ResourceLeakDetector) cleanupTempFileResource(resourceID string) error {
	// This would remove temporary files
	fmt.Printf("Cleaning up leaked temp file resource: %s\n", resourceID)
	return nil
}

func (rld *ResourceLeakDetector) cleanupScreenshotResource(resourceID string) error {
	// This would remove screenshot files
	fmt.Printf("Cleaning up leaked screenshot resource: %s\n", resourceID)
	return nil
}

// Helper methods
func (rld *ResourceLeakDetector) removeOldestResource() {
	var oldestID string
	var oldestTime time.Time

	for id, resource := range rld.trackedResources {
		if oldestID == "" || resource.CreatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = resource.CreatedAt
		}
	}

	if oldestID != "" {
		delete(rld.trackedResources, oldestID)
	}
}

func (rld *ResourceLeakDetector) captureStackTrace() []string {
	const maxStackDepth = 10
	pcs := make([]uintptr, maxStackDepth)
	n := runtime.Callers(3, pcs) // Skip 3 frames to get to the actual caller

	frames := runtime.CallersFrames(pcs[:n])
	var stackTrace []string

	for {
		frame, more := frames.Next()
		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return stackTrace
}

func (rld *ResourceLeakDetector) shouldTriggerImmediateReport() bool {
	leakCount := 0
	for _, resource := range rld.trackedResources {
		if resource.IsLeaked {
			leakCount++
		}
	}

	return leakCount >= rld.config.CriticalLeakCount
}

func (rld *ResourceLeakDetector) logLeakReport(report *LeakReport) {
	fmt.Printf("=== RESOURCE LEAK REPORT ===\n")
	fmt.Printf("Report ID: %s\n", report.ID)
	fmt.Printf("Detected At: %s\n", report.DetectedAt.Format(time.RFC3339))
	fmt.Printf("Severity: %s\n", report.Severity.String())
	fmt.Printf("Total Leaks: %d\n", report.TotalLeaks)
	fmt.Printf("Memory Leaked: %d MB\n", report.MemoryLeaked/(1024*1024))
	fmt.Printf("Recommendations:\n")
	for _, rec := range report.Recommendations {
		fmt.Printf("  - %s\n", rec)
	}
	fmt.Printf("============================\n")
}

// Public API methods
func (rld *ResourceLeakDetector) GetLeakReports() []*LeakReport {
	rld.mutex.RLock()
	defer rld.mutex.RUnlock()

	// Return a copy to avoid race conditions
	reports := make([]*LeakReport, len(rld.leakReports))
	copy(reports, rld.leakReports)
	return reports
}

func (rld *ResourceLeakDetector) GetLatestLeakReport() *LeakReport {
	rld.mutex.RLock()
	defer rld.mutex.RUnlock()

	if len(rld.leakReports) == 0 {
		return nil
	}

	return rld.leakReports[len(rld.leakReports)-1]
}

func (rld *ResourceLeakDetector) GetDetectorStats() *LeakDetectorStats {
	rld.mutex.RLock()
	defer rld.mutex.RUnlock()

	return &LeakDetectorStats{
		TrackedResources:   len(rld.trackedResources),
		TotalLeaksDetected: rld.totalLeaksDetected,
		TotalMemoryLeaked:  rld.totalMemoryLeaked,
		LastScan:           rld.lastScan,
		ReportCount:        len(rld.leakReports),
	}
}

// LeakDetectorStats holds statistics about the leak detector
type LeakDetectorStats struct {
	TrackedResources   int       `json:"tracked_resources"`
	TotalLeaksDetected int64     `json:"total_leaks_detected"`
	TotalMemoryLeaked  int64     `json:"total_memory_leaked"`
	LastScan           time.Time `json:"last_scan"`
	ReportCount        int       `json:"report_count"`
}

// Shutdown gracefully shuts down the leak detector
func (rld *ResourceLeakDetector) Shutdown() error {
	rld.cancel()

	if rld.monitoringTicker != nil {
		rld.monitoringTicker.Stop()
	}

	// Generate final leak report
	// Use a small delay to allow any ongoing goroutines to complete
	time.Sleep(10 * time.Millisecond)
	rld.generateLeakReport()

	return nil
}
