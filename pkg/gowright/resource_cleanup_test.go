package gowright

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceCleanupManager(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	assert.NotNil(t, rcm)
	assert.Equal(t, config, rcm.config)
	assert.NotNil(t, rcm.tempFiles)
	assert.NotNil(t, rcm.tempDirs)
	assert.NotNil(t, rcm.activeResources)
}

func TestResourceCleanupManager_RegisterTempFile(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	config.MaxTempFiles = 2 // Low limit for testing
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create actual temp files
	tempFile1, err := os.CreateTemp("", "test_cleanup_1_")
	require.NoError(t, err)
	_ = tempFile1.Close()
	defer func() { _ = os.Remove(tempFile1.Name()) }()

	tempFile2, err := os.CreateTemp("", "test_cleanup_2_")
	require.NoError(t, err)
	_ = tempFile2.Close()
	defer func() { _ = os.Remove(tempFile2.Name()) }()

	tempFile3, err := os.CreateTemp("", "test_cleanup_3_")
	require.NoError(t, err)
	_ = tempFile3.Close()
	defer func() { _ = os.Remove(tempFile3.Name()) }()

	// Register files
	rcm.RegisterTempFile(tempFile1.Name())
	rcm.RegisterTempFile(tempFile2.Name())

	stats := rcm.GetCleanupStats()
	assert.Equal(t, 2, stats.TempFilesTracked)

	// Adding third file should trigger cleanup of oldest
	rcm.RegisterTempFile(tempFile3.Name())

	// Give some time for cleanup goroutine
	time.Sleep(100 * time.Millisecond)

	stats = rcm.GetCleanupStats()
	assert.LessOrEqual(t, stats.TempFilesTracked, 2)
}

func TestResourceCleanupManager_RegisterTempDir(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	tempDir, err := os.MkdirTemp("", "test_cleanup_dir_")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	rcm.RegisterTempDir(tempDir)

	stats := rcm.GetCleanupStats()
	assert.Equal(t, 1, stats.TempDirsTracked)
}

func TestResourceCleanupManager_TrackResource(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	metadata := map[string]interface{}{
		"type": "test_resource",
		"size": 1024,
	}

	rcm.TrackResource("test_resource", "resource_1", 1024, metadata)

	stats := rcm.GetCleanupStats()
	assert.Equal(t, 1, stats.ActiveResources)
}

func TestResourceCleanupManager_UntrackResource(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	rcm.TrackResource("test_resource", "resource_1", 1024, nil)

	stats := rcm.GetCleanupStats()
	assert.Equal(t, 1, stats.ActiveResources)

	rcm.UntrackResource("resource_1")

	stats = rcm.GetCleanupStats()
	assert.Equal(t, 0, stats.ActiveResources)
}

func TestResourceCleanupManager_UpdateResourceAccess(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	rcm.TrackResource("test_resource", "resource_1", 1024, nil)

	// Get initial access time
	rcm.mutex.RLock()
	initialTime := rcm.activeResources["resource_1"].LastAccessed
	rcm.mutex.RUnlock()

	time.Sleep(10 * time.Millisecond)

	rcm.UpdateResourceAccess("resource_1")

	// Check that access time was updated
	rcm.mutex.RLock()
	updatedTime := rcm.activeResources["resource_1"].LastAccessed
	rcm.mutex.RUnlock()

	assert.True(t, updatedTime.After(initialTime))
}

func TestResourceCleanupManager_DetectResourceLeaks(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create a resource that will appear as leaked
	rcm.TrackResource("test_resource", "leaked_resource", 2048, nil)

	// Manually set the last accessed time to make it appear old
	rcm.mutex.Lock()
	rcm.activeResources["leaked_resource"].LastAccessed = time.Now().Add(-2 * time.Hour)
	rcm.mutex.Unlock()

	report := rcm.DetectResourceLeaks()

	assert.NotNil(t, report)
	assert.Equal(t, 1, report.TotalLeaks)
	assert.Equal(t, int64(2048), report.MemoryLeaked)
	assert.Len(t, report.LeakedResources, 1)
	assert.Equal(t, "leaked_resource", report.LeakedResources[0].ResourceID)
}

func TestResourceCleanupManager_ForceCleanup(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create temp files
	tempFile1, err := os.CreateTemp("", "force_cleanup_1_")
	require.NoError(t, err)
	_ = tempFile1.Close()

	tempFile2, err := os.CreateTemp("", "force_cleanup_2_")
	require.NoError(t, err)
	_ = tempFile2.Close()

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "force_cleanup_dir_")
	require.NoError(t, err)

	// Register with cleanup manager
	rcm.RegisterTempFile(tempFile1.Name())
	rcm.RegisterTempFile(tempFile2.Name())
	rcm.RegisterTempDir(tempDir)

	// Verify files exist
	_, err = os.Stat(tempFile1.Name())
	assert.NoError(t, err)
	_, err = os.Stat(tempFile2.Name())
	assert.NoError(t, err)
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)

	// Force cleanup
	err = rcm.ForceCleanup()
	assert.NoError(t, err)

	// Verify files were removed
	_, err = os.Stat(tempFile1.Name())
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(tempFile2.Name())
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))

	// Verify tracking was cleared
	stats := rcm.GetCleanupStats()
	assert.Equal(t, 0, stats.TempFilesTracked)
	assert.Equal(t, 0, stats.TempDirsTracked)
}

func TestResourceCleanupManager_MemoryThresholdCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory threshold test in short mode")
	}

	config := DefaultResourceCleanupConfig()
	config.MemoryThresholdMB = 1 // Very low threshold
	config.CleanupOnLowMemory = true
	config.AutoCleanupInterval = 100 * time.Millisecond

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create temp files to trigger cleanup
	var tempFiles []string
	for i := 0; i < 10; i++ {
		tempFile, err := os.CreateTemp("", "memory_cleanup_")
		require.NoError(t, err)
		_ = tempFile.Close()

		tempFiles = append(tempFiles, tempFile.Name())
		rcm.RegisterTempFile(tempFile.Name())
	}

	// Wait for potential cleanup
	time.Sleep(300 * time.Millisecond)

	// Check if some files were cleaned up
	cleanedCount := 0
	for _, filePath := range tempFiles {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			cleanedCount++
		} else {
			// Clean up remaining files
			_ = os.Remove(filePath)
		}
	}

	t.Logf("Cleaned up %d out of %d files due to memory threshold", cleanedCount, len(tempFiles))
}

func TestResourceCleanupManager_GetCleanupStats(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Add some tracked items
	tempFile, err := os.CreateTemp("", "stats_test_")
	require.NoError(t, err)
	_ = tempFile.Close()
	defer func() { _ = os.Remove(tempFile.Name()) }()

	tempDir, err := os.MkdirTemp("", "stats_test_dir_")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	rcm.RegisterTempFile(tempFile.Name())
	rcm.RegisterTempDir(tempDir)
	rcm.TrackResource("test_resource", "resource_1", 1024, nil)

	stats := rcm.GetCleanupStats()

	assert.Equal(t, 1, stats.TempFilesTracked)
	assert.Equal(t, 1, stats.TempDirsTracked)
	assert.Equal(t, 1, stats.ActiveResources)
	assert.True(t, stats.CurrentMemoryMB >= 0) // Memory can be 0 in some test environments
	assert.Equal(t, config.MemoryThresholdMB, stats.MemoryThresholdMB)
	assert.False(t, stats.LastCleanup.IsZero())
	assert.True(t, stats.NextCleanup.After(stats.LastCleanup) || stats.NextCleanup.Equal(stats.LastCleanup.Add(config.AutoCleanupInterval)))
}

func TestResourceCleanupManager_AutoCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping auto cleanup test in short mode")
	}

	config := DefaultResourceCleanupConfig()
	config.AutoCleanupInterval = 100 * time.Millisecond
	config.TempFileMaxAge = 200 * time.Millisecond

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create temp files
	var tempFiles []string
	for i := 0; i < 5; i++ {
		tempFile, err := os.CreateTemp("", "auto_cleanup_")
		require.NoError(t, err)
		_ = tempFile.Close()

		tempFiles = append(tempFiles, tempFile.Name())
		rcm.RegisterTempFile(tempFile.Name())
	}

	// Wait for files to age and cleanup to occur
	time.Sleep(400 * time.Millisecond)

	// Check that files were cleaned up
	cleanedCount := 0
	for _, filePath := range tempFiles {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			cleanedCount++
		} else {
			// Clean up remaining files
			_ = os.Remove(filePath)
		}
	}

	t.Logf("Auto cleanup removed %d out of %d files", cleanedCount, len(tempFiles))

	// At least some files should have been cleaned up
	assert.Greater(t, cleanedCount, 0, "Auto cleanup should have removed some files")
}

func TestResourceCleanupManager_ScreenshotCleanup(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	config.ScreenshotMaxAge = 100 * time.Millisecond

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create screenshot directory
	screenshotDir := "./captures/screenshots"
	err := os.MkdirAll(screenshotDir, 0755)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll("./captures") }()

	// Create old screenshot files
	oldScreenshot := filepath.Join(screenshotDir, "old_screenshot.png")
	err = os.WriteFile(oldScreenshot, []byte("fake screenshot data"), 0644)
	require.NoError(t, err)

	// Make the file appear old
	oldTime := time.Now().Add(-200 * time.Millisecond)
	err = os.Chtimes(oldScreenshot, oldTime, oldTime)
	require.NoError(t, err)

	// Create new screenshot file
	newScreenshot := filepath.Join(screenshotDir, "new_screenshot.png")
	err = os.WriteFile(newScreenshot, []byte("fake screenshot data"), 0644)
	require.NoError(t, err)

	// Perform cleanup
	err = rcm.performAutoCleanup()
	assert.NoError(t, err)

	// Check results
	_, err = os.Stat(oldScreenshot)
	assert.True(t, os.IsNotExist(err), "Old screenshot should have been cleaned up")

	_, err = os.Stat(newScreenshot)
	assert.NoError(t, err, "New screenshot should still exist")

	// Clean up remaining file
	_ = os.Remove(newScreenshot)
}

func TestResourceCleanupManager_ConcurrentOperations(t *testing.T) {
	config := DefaultResourceCleanupConfig()
	config.AutoCleanupInterval = 1 * time.Hour // Disable auto cleanup for this test

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	const numGoroutines = 10
	const operationsPerGoroutine = 50

	// Run concurrent operations
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				resourceID := fmt.Sprintf("resource_%d_%d", goroutineID, j)

				// Track resource
				rcm.TrackResource("test_resource", resourceID, 1024, map[string]interface{}{
					"goroutine": goroutineID,
					"operation": j,
				})

				// Update access
				rcm.UpdateResourceAccess(resourceID)

				// Untrack resource
				rcm.UntrackResource(resourceID)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	stats := rcm.GetCleanupStats()
	assert.Equal(t, 0, stats.ActiveResources, "All resources should have been untracked")
}

func TestResourceCleanupManager_MemoryPressureHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory pressure test in short mode")
	}

	config := DefaultResourceCleanupConfig()
	config.MemoryThresholdMB = 50 // Reasonable threshold
	config.CleanupOnLowMemory = true

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Get initial memory stats
	var initialMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMem)

	// Create many temp files to simulate memory pressure
	var tempFiles []string
	for i := 0; i < 100; i++ {
		tempFile, err := os.CreateTemp("", "memory_pressure_")
		require.NoError(t, err)
		_ = tempFile.Close()

		tempFiles = append(tempFiles, tempFile.Name())
		rcm.RegisterTempFile(tempFile.Name())

		// Track resources too
		resourceID := fmt.Sprintf("memory_resource_%d", i)
		rcm.TrackResource("memory_test", resourceID, 1024*1024, nil) // 1MB each
	}

	// Force memory check
	err := rcm.checkMemoryAndCleanup()
	if err != nil {
		t.Logf("Memory cleanup triggered: %v", err)
	}

	// Clean up remaining files
	for _, filePath := range tempFiles {
		_ = os.Remove(filePath)
	}

	// Get final memory stats
	var finalMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMem)

	t.Logf("Initial memory: %d MB", initialMem.Alloc/(1024*1024))
	t.Logf("Final memory: %d MB", finalMem.Alloc/(1024*1024))
}

func TestDefaultResourceCleanupConfig(t *testing.T) {
	config := DefaultResourceCleanupConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.AutoCleanupInterval, time.Duration(0))
	assert.Greater(t, config.TempFileMaxAge, time.Duration(0))
	assert.Greater(t, config.ScreenshotMaxAge, time.Duration(0))
	assert.Greater(t, config.MemoryThresholdMB, int64(0))
	assert.Greater(t, config.DiskThresholdMB, int64(0))
	assert.Greater(t, config.MaxTempFiles, 0)
	assert.Greater(t, config.MaxScreenshots, 0)
}
