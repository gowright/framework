package gowright

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryEfficientCaptureManager(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	assert.NotNil(t, mecm)
	assert.Equal(t, config, mecm.config)
	assert.NotNil(t, mecm.activeCaptures)
	assert.Equal(t, config.MaxMemoryUsageMB*1024*1024, mecm.maxMemoryUsage)
}

func TestMemoryEfficientCaptureManager_CaptureDataStreamOptimized(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.EnableCompression = true
	config.CompressThresholdKB = 1 // Compress everything

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Create test data with repetitive pattern (compressible)
	testData := make([]byte, 10*1024) // 10KB
	pattern := []byte("This is a test pattern that repeats. ")
	for i := 0; i < len(testData); i++ {
		testData[i] = pattern[i%len(pattern)]
	}

	result, err := mecm.CaptureDataStreamOptimized(testData, "test_capture", "test_data")
	require.NoError(t, err)
	defer func() { _ = mecm.ReleaseCaptureMemory(result.ID) }()

	assert.NotEmpty(t, result.ID)
	assert.NotEmpty(t, result.FilePath)
	assert.Equal(t, int64(len(testData)), result.OriginalSize)
	assert.True(t, result.Compressed)
	assert.True(t, result.Optimized)
	assert.Less(t, result.FinalSize, result.OriginalSize)

	// Verify file exists and contains data
	_, err = os.Stat(result.FilePath)
	assert.NoError(t, err)

	// Verify the data can be read back
	fileData, err := os.ReadFile(result.FilePath)
	require.NoError(t, err)

	// Decompress and verify
	reader := bytes.NewReader(fileData)
	gzipReader, err := gzip.NewReader(reader)
	require.NoError(t, err)
	defer func() { _ = gzipReader.Close() }()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(gzipReader)
	require.NoError(t, err)

	assert.Equal(t, testData, decompressed.Bytes())
}

func TestMemoryEfficientCaptureManager_CaptureDataStreamOptimized_NoCompression(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.EnableCompression = false

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	testData := []byte("Small test data that won't be compressed")

	result, err := mecm.CaptureDataStreamOptimized(testData, "test_no_compression", "test_data")
	require.NoError(t, err)
	defer func() { _ = mecm.ReleaseCaptureMemory(result.ID) }()

	assert.False(t, result.Compressed)
	assert.Equal(t, int64(len(testData)), result.OriginalSize)
	assert.Equal(t, result.OriginalSize, result.FinalSize)

	// Verify file contains original data
	fileData, err := os.ReadFile(result.FilePath)
	require.NoError(t, err)
	assert.Equal(t, testData, fileData)
}

func TestMemoryEfficientCaptureManager_writeCompressedStream(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.CompressionLevel = 6
	config.StreamChunkSizeKB = 4 // Small chunks for testing

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Create test data
	testData := make([]byte, 20*1024) // 20KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	var output bytes.Buffer
	written, err := mecm.writeCompressedStream(&output, testData)
	require.NoError(t, err)

	assert.Greater(t, written, int64(0))
	assert.Less(t, written, int64(len(testData))) // Should be compressed

	// Verify we can decompress the data
	reader := bytes.NewReader(output.Bytes())
	gzipReader, err := gzip.NewReader(reader)
	require.NoError(t, err)
	defer func() { _ = gzipReader.Close() }()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(gzipReader)
	require.NoError(t, err)

	assert.Equal(t, testData, decompressed.Bytes())
}

func TestMemoryEfficientCaptureManager_writeDataStream(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.StreamChunkSizeKB = 4 // Small chunks for testing

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	testData := []byte("This is test data for streaming without compression")

	var output bytes.Buffer
	written, err := mecm.writeDataStream(&output, testData)
	require.NoError(t, err)

	assert.Equal(t, int64(len(testData)), written)
	assert.Equal(t, testData, output.Bytes())
}

func TestMemoryEfficientCaptureManager_checkMemoryUsage(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.MaxMemoryUsageMB = 1       // Very low limit for testing
	config.AutoCleanupEnabled = false // Disable auto cleanup for this test

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Initially should pass
	err := mecm.checkMemoryUsage()
	assert.NoError(t, err)

	// Add memory usage beyond threshold
	mecm.mutex.Lock()
	mecm.totalMemoryUsed = 2 * 1024 * 1024 // 2MB
	mecm.mutex.Unlock()

	// Should now fail
	err = mecm.checkMemoryUsage()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory usage exceeded")
}

func TestMemoryEfficientCaptureManager_performMemoryCleanup(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Create some captures with old access times
	oldTime := time.Now().Add(-2 * time.Hour)

	// Create temp files for testing
	tempFile1, err := os.CreateTemp("", "memory_cleanup_1_")
	require.NoError(t, err)
	_ = tempFile1.Close()

	tempFile2, err := os.CreateTemp("", "memory_cleanup_2_")
	require.NoError(t, err)
	_ = tempFile2.Close()

	// Add captures to manager
	mecm.mutex.Lock()
	mecm.activeCaptures["old_capture_1"] = &CaptureInfo{
		ID:           "old_capture_1",
		Type:         "test",
		Size:         1024,
		LastAccessed: oldTime,
		FilePath:     tempFile1.Name(),
	}
	mecm.activeCaptures["old_capture_2"] = &CaptureInfo{
		ID:           "old_capture_2",
		Type:         "test",
		Size:         2048,
		LastAccessed: oldTime,
		FilePath:     tempFile2.Name(),
	}
	mecm.totalMemoryUsed = 3072
	mecm.mutex.Unlock()

	initialCount := len(mecm.activeCaptures)

	// Perform cleanup
	mecm.performMemoryCleanup()

	// Check that old captures were removed
	mecm.mutex.RLock()
	finalCount := len(mecm.activeCaptures)
	finalMemoryUsed := mecm.totalMemoryUsed
	mecm.mutex.RUnlock()

	assert.Less(t, finalCount, initialCount)
	assert.Less(t, finalMemoryUsed, int64(3072))

	// Verify files were removed
	_, err = os.Stat(tempFile1.Name())
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(tempFile2.Name())
	assert.True(t, os.IsNotExist(err))
}

func TestMemoryEfficientCaptureManager_ReleaseCaptureMemory(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Create a temp file
	tempFile, err := os.CreateTemp("", "release_test_")
	require.NoError(t, err)
	_ = tempFile.Close()

	// Add capture to manager
	captureID := "test_capture"
	mecm.mutex.Lock()
	mecm.activeCaptures[captureID] = &CaptureInfo{
		ID:       captureID,
		Type:     "test",
		Size:     1024,
		FilePath: tempFile.Name(),
	}
	mecm.totalMemoryUsed = 1024
	mecm.mutex.Unlock()

	// Release the capture
	err = mecm.ReleaseCaptureMemory(captureID)
	assert.NoError(t, err)

	// Verify capture was removed
	mecm.mutex.RLock()
	_, exists := mecm.activeCaptures[captureID]
	memoryUsed := mecm.totalMemoryUsed
	mecm.mutex.RUnlock()

	assert.False(t, exists)
	assert.Equal(t, int64(0), memoryUsed)

	// Verify file was removed
	_, err = os.Stat(tempFile.Name())
	assert.True(t, os.IsNotExist(err))
}

func TestMemoryEfficientCaptureManager_ReleaseCaptureMemory_NotFound(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	err := mecm.ReleaseCaptureMemory("nonexistent_capture")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryEfficientCaptureManager_GetMemoryStats(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Add some captures
	mecm.mutex.Lock()
	mecm.activeCaptures["capture_1"] = &CaptureInfo{Size: 1024}
	mecm.activeCaptures["capture_2"] = &CaptureInfo{Size: 2048}
	mecm.totalMemoryUsed = 3072
	mecm.mutex.Unlock()

	stats := mecm.GetMemoryStats()

	assert.Equal(t, 2, stats.ActiveCaptures)
	assert.Equal(t, int64(0), stats.TotalMemoryUsedMB) // 3072 bytes = 0MB (rounded down)
	assert.Equal(t, config.MaxMemoryUsageMB, stats.MaxMemoryUsageMB)
	assert.GreaterOrEqual(t, stats.SystemMemoryMB, int64(0))
	assert.GreaterOrEqual(t, stats.GCCount, int64(0))
}

func TestMemoryEfficientCaptureManager_Cleanup(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)

	// Create temp files
	tempFile1, err := os.CreateTemp("", "cleanup_test_1_")
	require.NoError(t, err)
	_ = tempFile1.Close()

	tempFile2, err := os.CreateTemp("", "cleanup_test_2_")
	require.NoError(t, err)
	_ = tempFile2.Close()

	// Add captures
	mecm.mutex.Lock()
	mecm.activeCaptures["capture_1"] = &CaptureInfo{
		ID:       "capture_1",
		FilePath: tempFile1.Name(),
		Size:     1024,
	}
	mecm.activeCaptures["capture_2"] = &CaptureInfo{
		ID:       "capture_2",
		FilePath: tempFile2.Name(),
		Size:     2048,
	}
	mecm.totalMemoryUsed = 3072
	mecm.mutex.Unlock()

	// Perform cleanup
	err = mecm.Cleanup()
	assert.NoError(t, err)

	// Verify all captures were removed
	mecm.mutex.RLock()
	captureCount := len(mecm.activeCaptures)
	memoryUsed := mecm.totalMemoryUsed
	mecm.mutex.RUnlock()

	assert.Equal(t, 0, captureCount)
	assert.Equal(t, int64(0), memoryUsed)

	// Verify files were removed
	_, err = os.Stat(tempFile1.Name())
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(tempFile2.Name())
	assert.True(t, os.IsNotExist(err))
}

func TestMemoryEfficientCaptureManager_createTempFile(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	tempFile, err := mecm.createTempFile("test_id", "test_type", ".dat")
	require.NoError(t, err)
	defer func() { _ = tempFile.Close() }()
	defer func() { _ = os.Remove(tempFile.Name()) }()

	assert.NotEmpty(t, tempFile.Name())
	assert.Contains(t, tempFile.Name(), "test_id")
	assert.Contains(t, tempFile.Name(), "test_type")
	assert.Contains(t, tempFile.Name(), ".dat")

	// Verify file exists
	_, err = os.Stat(tempFile.Name())
	assert.NoError(t, err)
}

func TestMemoryEfficientCaptureManager_changeExtension(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	originalPath := "/path/to/file.png"
	newPath := mecm.changeExtension(originalPath, ".jpg")

	assert.Equal(t, "/path/to/file.jpg", newPath)
}

func TestMemoryEfficientCaptureManager_ConcurrentOperations(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.AutoCleanupEnabled = false // Disable auto cleanup for this test

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	const numGoroutines = 10
	const operationsPerGoroutine = 20

	// Create test data
	testData := make([]byte, 1024) // 1KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	done := make(chan bool, numGoroutines)
	var captureIDs []string
	var captureIDsMutex sync.Mutex

	// Run concurrent capture operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				testName := fmt.Sprintf("concurrent_test_%d_%d", goroutineID, j)

				result, err := mecm.CaptureDataStreamOptimized(testData, testName, "concurrent_data")
				if err != nil {
					t.Errorf("Failed to capture data: %v", err)
					continue
				}

				captureIDsMutex.Lock()
				captureIDs = append(captureIDs, result.ID)
				captureIDsMutex.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all captures were created
	expectedCount := numGoroutines * operationsPerGoroutine
	assert.Equal(t, expectedCount, len(captureIDs))

	// Clean up all captures
	for _, captureID := range captureIDs {
		_ = mecm.ReleaseCaptureMemory(captureID)
	}

	// Verify cleanup
	stats := mecm.GetMemoryStats()
	assert.Equal(t, 0, stats.ActiveCaptures)
	assert.Equal(t, int64(0), stats.TotalMemoryUsedMB)
}

func TestDefaultMemoryEfficientConfig(t *testing.T) {
	config := DefaultMemoryEfficientConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.MaxMemoryUsageMB, int64(0))
	assert.Greater(t, config.MaxScreenshotSizeMB, int64(0))
	assert.True(t, config.EnableCompression)
	assert.Greater(t, config.CompressionLevel, 0)
	assert.Greater(t, config.CompressThresholdKB, int64(0))
	assert.Greater(t, config.MaxImageWidth, 0)
	assert.Greater(t, config.MaxImageHeight, 0)
	assert.Greater(t, config.JPEGQuality, 0)
	assert.LessOrEqual(t, config.JPEGQuality, 100)
	assert.True(t, config.EnableStreaming)
	assert.Greater(t, config.StreamChunkSizeKB, int64(0))
}

func TestDefaultOptimizedCaptureOptions(t *testing.T) {
	options := DefaultOptimizedCaptureOptions()

	assert.NotNil(t, options)
	assert.Greater(t, options.MaxWidth, 0)
	assert.Greater(t, options.MaxHeight, 0)
	assert.True(t, options.ConvertToJPEG)
	assert.Greater(t, options.JPEGQuality, 0)
	assert.LessOrEqual(t, options.JPEGQuality, 100)
}
