package gowright

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ResourceCleanupManager handles automatic cleanup of resources and temporary files
type ResourceCleanupManager struct {
	config           *ResourceCleanupConfig
	tempDirs         map[string]time.Time
	tempFiles        map[string]time.Time
	activeResources  map[string]*ResourceTracker
	cleanupTicker    *time.Ticker
	ctx              context.Context
	cancel           context.CancelFunc
	mutex            sync.RWMutex
	memoryThreshold  int64
	diskThreshold    int64
	lastCleanup      time.Time
}

// ResourceCleanupConfig holds configuration for resource cleanup
type ResourceCleanupConfig struct {
	// Cleanup intervals
	AutoCleanupInterval    time.Duration `json:"auto_cleanup_interval"`
	TempFileMaxAge         time.Duration `json:"temp_file_max_age"`
	ScreenshotMaxAge       time.Duration `json:"screenshot_max_age"`
	
	// Memory management
	MemoryThresholdMB      int64         `json:"memory_threshold_mb"`
	DiskThresholdMB        int64         `json:"disk_threshold_mb"`
	MaxScreenshotSizeMB    int64         `json:"max_screenshot_size_mb"`
	
	// Resource limits
	MaxTempFiles           int           `json:"max_temp_files"`
	MaxScreenshots         int           `json:"max_screenshots"`
	MaxBrowserInstances    int           `json:"max_browser_instances"`
	
	// Cleanup behavior
	AggressiveCleanup      bool          `json:"aggressive_cleanup"`
	CleanupOnLowMemory     bool          `json:"cleanup_on_low_memory"`
	CompressOldScreenshots bool          `json:"compress_old_screenshots"`
}

// ResourceTracker tracks resource usage for leak detection
type ResourceTracker struct {
	ResourceType  string                 `json:"resource_type"`
	ResourceID    string                 `json:"resource_id"`
	CreatedAt     time.Time              `json:"created_at"`
	LastAccessed  time.Time              `json:"last_accessed"`
	Size          int64                  `json:"size"`
	Metadata      map[string]interface{} `json:"metadata"`
	RefCount      int                    `json:"ref_count"`
}

// ResourceLeakReport contains information about detected resource leaks
type ResourceLeakReport struct {
	LeakedResources []ResourceTracker `json:"leaked_resources"`
	TotalLeaks      int               `json:"total_leaks"`
	MemoryLeaked    int64             `json:"memory_leaked"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

// DefaultResourceCleanupConfig returns default cleanup configuration
func DefaultResourceCleanupConfig() *ResourceCleanupConfig {
	return &ResourceCleanupConfig{
		AutoCleanupInterval:    5 * time.Minute,
		TempFileMaxAge:         1 * time.Hour,
		ScreenshotMaxAge:       24 * time.Hour,
		MemoryThresholdMB:      1024, // 1GB
		DiskThresholdMB:        5120, // 5GB
		MaxScreenshotSizeMB:    10,   // 10MB per screenshot
		MaxTempFiles:           1000,
		MaxScreenshots:         500,
		MaxBrowserInstances:    10,
		AggressiveCleanup:      false,
		CleanupOnLowMemory:     true,
		CompressOldScreenshots: true,
	}
}

// NewResourceCleanupManager creates a new resource cleanup manager
func NewResourceCleanupManager(config *ResourceCleanupConfig) *ResourceCleanupManager {
	if config == nil {
		config = DefaultResourceCleanupConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	rcm := &ResourceCleanupManager{
		config:          config,
		tempDirs:        make(map[string]time.Time),
		tempFiles:       make(map[string]time.Time),
		activeResources: make(map[string]*ResourceTracker),
		ctx:             ctx,
		cancel:          cancel,
		memoryThreshold: config.MemoryThresholdMB * 1024 * 1024,
		diskThreshold:   config.DiskThresholdMB * 1024 * 1024,
		lastCleanup:     time.Now(),
	}
	
	// Start automatic cleanup routine
	rcm.startAutoCleanup()
	
	return rcm
}

// startAutoCleanup starts the automatic cleanup routine
func (rcm *ResourceCleanupManager) startAutoCleanup() {
	rcm.cleanupTicker = time.NewTicker(rcm.config.AutoCleanupInterval)
	
	go func() {
		for {
			select {
			case <-rcm.cleanupTicker.C:
				if err := rcm.performAutoCleanup(); err != nil {
					fmt.Printf("Warning: auto cleanup failed: %v\n", err)
				}
			case <-rcm.ctx.Done():
				return
			}
		}
	}()
}

// RegisterTempFile registers a temporary file for cleanup tracking
func (rcm *ResourceCleanupManager) RegisterTempFile(filePath string) {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	rcm.tempFiles[filePath] = time.Now()
	
	// Check if we need immediate cleanup due to limits
	if len(rcm.tempFiles) > rcm.config.MaxTempFiles {
		go rcm.cleanupOldestTempFiles(len(rcm.tempFiles) - rcm.config.MaxTempFiles)
	}
}

// RegisterTempDir registers a temporary directory for cleanup tracking
func (rcm *ResourceCleanupManager) RegisterTempDir(dirPath string) {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	rcm.tempDirs[dirPath] = time.Now()
}

// TrackResource tracks a resource for leak detection
func (rcm *ResourceCleanupManager) TrackResource(resourceType, resourceID string, size int64, metadata map[string]interface{}) {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	tracker := &ResourceTracker{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Size:         size,
		Metadata:     metadata,
		RefCount:     1,
	}
	
	rcm.activeResources[resourceID] = tracker
}

// UntrackResource removes a resource from tracking
func (rcm *ResourceCleanupManager) UntrackResource(resourceID string) {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	delete(rcm.activeResources, resourceID)
}

// UpdateResourceAccess updates the last accessed time for a resource
func (rcm *ResourceCleanupManager) UpdateResourceAccess(resourceID string) {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	if tracker, exists := rcm.activeResources[resourceID]; exists {
		tracker.LastAccessed = time.Now()
	}
}

// performAutoCleanup performs automatic cleanup of resources
func (rcm *ResourceCleanupManager) performAutoCleanup() error {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	var errors []error
	
	// Check memory usage and trigger cleanup if needed
	if rcm.config.CleanupOnLowMemory {
		if err := rcm.checkMemoryAndCleanup(); err != nil {
			errors = append(errors, fmt.Errorf("memory cleanup failed: %w", err))
		}
	}
	
	// Clean up old temporary files
	if err := rcm.cleanupOldTempFiles(); err != nil {
		errors = append(errors, fmt.Errorf("temp file cleanup failed: %w", err))
	}
	
	// Clean up old temporary directories
	if err := rcm.cleanupOldTempDirs(); err != nil {
		errors = append(errors, fmt.Errorf("temp dir cleanup failed: %w", err))
	}
	
	// Clean up old screenshots
	if err := rcm.cleanupOldScreenshots(); err != nil {
		errors = append(errors, fmt.Errorf("screenshot cleanup failed: %w", err))
	}
	
	rcm.lastCleanup = time.Now()
	
	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}
	
	return nil
}

// checkMemoryAndCleanup checks memory usage and performs cleanup if threshold is exceeded
func (rcm *ResourceCleanupManager) checkMemoryAndCleanup() error {
	var m runtime.MemStats
	runtime.GC() // Force garbage collection to get accurate memory stats
	runtime.ReadMemStats(&m)
	
	currentMemory := int64(m.Alloc)
	
	if currentMemory > rcm.memoryThreshold {
		fmt.Printf("Memory threshold exceeded: %d MB > %d MB, performing cleanup\n", 
			currentMemory/(1024*1024), rcm.memoryThreshold/(1024*1024))
		
		// Perform aggressive cleanup
		if err := rcm.performAggressiveCleanup(); err != nil {
			return fmt.Errorf("aggressive cleanup failed: %w", err)
		}
		
		// Force another GC after cleanup
		runtime.GC()
	}
	
	return nil
}

// performAggressiveCleanup performs aggressive cleanup to free memory
func (rcm *ResourceCleanupManager) performAggressiveCleanup() error {
	var errors []error
	
	// Clean all temporary files older than half the max age
	cutoffTime := time.Now().Add(-rcm.config.TempFileMaxAge / 2)
	for filePath, createdAt := range rcm.tempFiles {
		if createdAt.Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				errors = append(errors, fmt.Errorf("failed to remove temp file %s: %w", filePath, err))
			} else {
				delete(rcm.tempFiles, filePath)
			}
		}
	}
	
	// Clean screenshots older than half the max age
	screenshotCutoff := time.Now().Add(-rcm.config.ScreenshotMaxAge / 2)
	if err := rcm.cleanupScreenshotsOlderThan(screenshotCutoff); err != nil {
		errors = append(errors, fmt.Errorf("screenshot cleanup failed: %w", err))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("aggressive cleanup errors: %v", errors)
	}
	
	return nil
}

// cleanupOldTempFiles removes temporary files older than the configured max age
func (rcm *ResourceCleanupManager) cleanupOldTempFiles() error {
	cutoffTime := time.Now().Add(-rcm.config.TempFileMaxAge)
	var errors []error
	
	for filePath, createdAt := range rcm.tempFiles {
		if createdAt.Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				errors = append(errors, fmt.Errorf("failed to remove temp file %s: %w", filePath, err))
			} else {
				delete(rcm.tempFiles, filePath)
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("temp file cleanup errors: %v", errors)
	}
	
	return nil
}

// cleanupOldTempDirs removes temporary directories older than the configured max age
func (rcm *ResourceCleanupManager) cleanupOldTempDirs() error {
	cutoffTime := time.Now().Add(-rcm.config.TempFileMaxAge)
	var errors []error
	
	for dirPath, createdAt := range rcm.tempDirs {
		if createdAt.Before(cutoffTime) {
			if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
				errors = append(errors, fmt.Errorf("failed to remove temp dir %s: %w", dirPath, err))
			} else {
				delete(rcm.tempDirs, dirPath)
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("temp dir cleanup errors: %v", errors)
	}
	
	return nil
}

// cleanupOldScreenshots removes screenshots older than the configured max age
func (rcm *ResourceCleanupManager) cleanupOldScreenshots() error {
	cutoffTime := time.Now().Add(-rcm.config.ScreenshotMaxAge)
	return rcm.cleanupScreenshotsOlderThan(cutoffTime)
}

// cleanupScreenshotsOlderThan removes screenshots older than the specified time
func (rcm *ResourceCleanupManager) cleanupScreenshotsOlderThan(cutoffTime time.Time) error {
	screenshotDirs := []string{"./captures/screenshots", "./reports/screenshots", "./temp/screenshots"}
	
	var errors []error
	
	for _, dir := range screenshotDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if !info.IsDir() && info.ModTime().Before(cutoffTime) {
				// Check if it's an image file
				ext := filepath.Ext(path)
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					if err := os.Remove(path); err != nil {
						return fmt.Errorf("failed to remove screenshot %s: %w", path, err)
					}
				}
			}
			
			return nil
		})
		
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup screenshots in %s: %w", dir, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("screenshot cleanup errors: %v", errors)
	}
	
	return nil
}

// cleanupOldestTempFiles removes the oldest temporary files
func (rcm *ResourceCleanupManager) cleanupOldestTempFiles(count int) {
	// Create a slice of file paths sorted by creation time
	type fileInfo struct {
		path      string
		createdAt time.Time
	}
	
	var files []fileInfo
	for path, createdAt := range rcm.tempFiles {
		files = append(files, fileInfo{path: path, createdAt: createdAt})
	}
	
	// Sort by creation time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].createdAt.After(files[j].createdAt) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
	
	// Remove the oldest files
	for i := 0; i < count && i < len(files); i++ {
		if err := os.Remove(files[i].path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove old temp file %s: %v\n", files[i].path, err)
		} else {
			delete(rcm.tempFiles, files[i].path)
		}
	}
}

// DetectResourceLeaks detects potential resource leaks
func (rcm *ResourceCleanupManager) DetectResourceLeaks() *ResourceLeakReport {
	rcm.mutex.RLock()
	defer rcm.mutex.RUnlock()
	
	report := &ResourceLeakReport{
		GeneratedAt: time.Now(),
	}
	
	leakThreshold := time.Now().Add(-1 * time.Hour) // Resources older than 1 hour
	
	for _, tracker := range rcm.activeResources {
		if tracker.LastAccessed.Before(leakThreshold) {
			report.LeakedResources = append(report.LeakedResources, *tracker)
			report.MemoryLeaked += tracker.Size
		}
	}
	
	report.TotalLeaks = len(report.LeakedResources)
	
	return report
}

// ForceCleanup performs immediate cleanup of all tracked resources
func (rcm *ResourceCleanupManager) ForceCleanup() error {
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	
	var errors []error
	
	// Clean all temporary files
	for filePath := range rcm.tempFiles {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to remove temp file %s: %w", filePath, err))
		}
	}
	rcm.tempFiles = make(map[string]time.Time)
	
	// Clean all temporary directories
	for dirPath := range rcm.tempDirs {
		if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to remove temp dir %s: %w", dirPath, err))
		}
	}
	rcm.tempDirs = make(map[string]time.Time)
	
	// Force garbage collection
	runtime.GC()
	
	if len(errors) > 0 {
		return fmt.Errorf("force cleanup errors: %v", errors)
	}
	
	return nil
}

// GetCleanupStats returns statistics about cleanup operations
func (rcm *ResourceCleanupManager) GetCleanupStats() *ResourceCleanupStats {
	rcm.mutex.RLock()
	defer rcm.mutex.RUnlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return &ResourceCleanupStats{
		TempFilesTracked:    len(rcm.tempFiles),
		TempDirsTracked:     len(rcm.tempDirs),
		ActiveResources:     len(rcm.activeResources),
		LastCleanup:         rcm.lastCleanup,
		CurrentMemoryMB:     int64(m.Alloc) / (1024 * 1024),
		MemoryThresholdMB:   rcm.memoryThreshold / (1024 * 1024),
		NextCleanup:         rcm.lastCleanup.Add(rcm.config.AutoCleanupInterval),
	}
}

// ResourceCleanupStats holds statistics about cleanup operations
type ResourceCleanupStats struct {
	TempFilesTracked    int       `json:"temp_files_tracked"`
	TempDirsTracked     int       `json:"temp_dirs_tracked"`
	ActiveResources     int       `json:"active_resources"`
	LastCleanup         time.Time `json:"last_cleanup"`
	CurrentMemoryMB     int64     `json:"current_memory_mb"`
	MemoryThresholdMB   int64     `json:"memory_threshold_mb"`
	NextCleanup         time.Time `json:"next_cleanup"`
}

// Shutdown gracefully shuts down the cleanup manager
func (rcm *ResourceCleanupManager) Shutdown() error {
	rcm.cancel()
	
	if rcm.cleanupTicker != nil {
		rcm.cleanupTicker.Stop()
	}
	
	// Perform final cleanup
	return rcm.ForceCleanup()
}