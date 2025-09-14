package ui

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// MemoryEfficientCaptureManager handles memory-efficient screenshot and data capture
type MemoryEfficientCaptureManager struct {
	config          *MemoryEfficientConfig
	compressionPool sync.Pool
	bufferPool      sync.Pool
	mutex           sync.RWMutex
	activeCaptures  map[string]*CaptureInfo
	totalMemoryUsed int64
	maxMemoryUsage  int64
}

// MemoryEfficientConfig holds configuration for memory-efficient operations
type MemoryEfficientConfig struct {
	// Memory limits
	MaxMemoryUsageMB    int64 `json:"max_memory_usage_mb"`
	MaxScreenshotSizeMB int64 `json:"max_screenshot_size_mb"`

	// Compression settings
	EnableCompression   bool  `json:"enable_compression"`
	CompressionLevel    int   `json:"compression_level"`
	CompressThresholdKB int64 `json:"compress_threshold_kb"`

	// Image optimization
	MaxImageWidth  int `json:"max_image_width"`
	MaxImageHeight int `json:"max_image_height"`
	JPEGQuality    int `json:"jpeg_quality"`

	// Streaming settings
	EnableStreaming   bool  `json:"enable_streaming"`
	StreamChunkSizeKB int64 `json:"stream_chunk_size_kb"`

	// Cleanup settings
	AutoCleanupEnabled    bool          `json:"auto_cleanup_enabled"`
	CleanupInterval       time.Duration `json:"cleanup_interval"`
	MaxCaptureAge         time.Duration `json:"max_capture_age"`
	MaxConcurrentCaptures int           `json:"max_concurrent_captures"`
}

// CaptureInfo holds information about an active capture
type CaptureInfo struct {
	ID           string    `json:"id"`
	TestName     string    `json:"test_name"`
	StartTime    time.Time `json:"start_time"`
	MemoryUsed   int64     `json:"memory_used"`
	Size         int64     `json:"size"` // Alias for MemoryUsed for backward compatibility
	Type         string    `json:"type"` // Type of capture
	Compressed   bool      `json:"compressed"`
	FilePath     string    `json:"file_path"`
	Status       string    `json:"status"`
	OriginalSize int64     `json:"original_size"`
	FinalSize    int64     `json:"final_size"`
	Optimized    bool      `json:"optimized"`
}

// OptimizedCaptureOptions holds options for optimized capture operations
type OptimizedCaptureOptions struct {
	MaxWidth      int  `json:"max_width"`
	MaxHeight     int  `json:"max_height"`
	ConvertToJPEG bool `json:"convert_to_jpeg"`
	JPEGQuality   int  `json:"jpeg_quality"`
}

// DefaultMemoryEfficientConfig returns default configuration
func DefaultMemoryEfficientConfig() *MemoryEfficientConfig {
	return &MemoryEfficientConfig{
		MaxMemoryUsageMB:      512,
		MaxScreenshotSizeMB:   50,
		EnableCompression:     true,
		CompressionLevel:      6,
		CompressThresholdKB:   100,
		MaxImageWidth:         1920,
		MaxImageHeight:        1080,
		JPEGQuality:           85,
		EnableStreaming:       true,
		StreamChunkSizeKB:     64,
		AutoCleanupEnabled:    true,
		CleanupInterval:       5 * time.Minute,
		MaxCaptureAge:         30 * time.Minute,
		MaxConcurrentCaptures: 10,
	}
}

// DefaultOptimizedCaptureOptions returns default optimized capture options
func DefaultOptimizedCaptureOptions() *OptimizedCaptureOptions {
	return &OptimizedCaptureOptions{
		MaxWidth:      1920,
		MaxHeight:     1080,
		ConvertToJPEG: true,
		JPEGQuality:   85,
	}
}

// NewMemoryEfficientCaptureManager creates a new memory-efficient capture manager
func NewMemoryEfficientCaptureManager(config *MemoryEfficientConfig) *MemoryEfficientCaptureManager {
	if config == nil {
		config = DefaultMemoryEfficientConfig()
	}

	manager := &MemoryEfficientCaptureManager{
		config:         config,
		activeCaptures: make(map[string]*CaptureInfo),
		maxMemoryUsage: config.MaxMemoryUsageMB * 1024 * 1024, // Convert to bytes
	}

	// Initialize object pools
	manager.compressionPool = sync.Pool{
		New: func() interface{} {
			var buf bytes.Buffer
			writer, _ := gzip.NewWriterLevel(&buf, config.CompressionLevel)
			return writer
		},
	}

	manager.bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 64*1024) // 64KB initial capacity
		},
	}

	// Start cleanup goroutine if auto cleanup is enabled
	if config.AutoCleanupEnabled {
		go manager.cleanupRoutine()
	}

	return manager
}

// CaptureScreenshotEfficient captures a screenshot with memory optimization
func (m *MemoryEfficientCaptureManager) CaptureScreenshotEfficient(tester core.UITester, testName string) (string, error) {
	// Check memory usage before capture
	if err := m.checkMemoryLimits(); err != nil {
		return "", err
	}

	captureID := fmt.Sprintf("%s_%d", testName, time.Now().UnixNano())

	// Register capture
	captureInfo := &CaptureInfo{
		ID:        captureID,
		TestName:  testName,
		StartTime: time.Now(),
		Status:    "in_progress",
	}

	m.mutex.Lock()
	m.activeCaptures[captureID] = captureInfo
	m.mutex.Unlock()

	defer func() {
		m.mutex.Lock()
		delete(m.activeCaptures, captureID)
		m.mutex.Unlock()
	}()

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.png", testName, timestamp)
	filePath := filepath.Join("./captures/screenshots", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create screenshot directory", err)
	}

	// Take screenshot
	screenshotPath, err := tester.TakeScreenshot(filePath)
	if err != nil {
		captureInfo.Status = "failed"
		return "", err
	}

	// Get file info for memory tracking
	fileInfo, err := os.Stat(screenshotPath)
	if err == nil {
		captureInfo.MemoryUsed = fileInfo.Size()
		m.updateMemoryUsage(captureInfo.MemoryUsed)
	}

	// Compress if needed
	if m.config.EnableCompression && fileInfo != nil && fileInfo.Size() > m.config.CompressThresholdKB*1024 {
		compressedPath, err := m.compressFile(screenshotPath)
		if err == nil {
			// Remove original file and use compressed version
			if removeErr := os.Remove(screenshotPath); removeErr != nil {
				// Log warning but continue with compressed version
				fmt.Printf("Warning: failed to remove original screenshot %s: %v\n", screenshotPath, removeErr)
			}
			screenshotPath = compressedPath
			captureInfo.Compressed = true
		}
	}

	captureInfo.FilePath = screenshotPath
	captureInfo.Status = "completed"

	return screenshotPath, nil
}

// CapturePageSourceEfficient captures page source with memory optimization
func (m *MemoryEfficientCaptureManager) CapturePageSourceEfficient(tester core.UITester, testName string) (string, error) {
	// Check memory usage before capture
	if err := m.checkMemoryLimits(); err != nil {
		return "", err
	}

	// Get page source
	source, err := tester.GetPageSource()
	if err != nil {
		return "", err
	}

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.html", testName, timestamp)
	filePath := filepath.Join("./captures/page_sources", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create page source directory", err)
	}

	// Compress if enabled and source is large
	sourceBytes := []byte(source)
	if m.config.EnableCompression && len(sourceBytes) > int(m.config.CompressThresholdKB*1024) {
		compressedData, err := m.compressData(sourceBytes)
		if err == nil {
			filePath += ".gz"
			sourceBytes = compressedData
		}
	}

	// Write to file
	if err := os.WriteFile(filePath, sourceBytes, 0600); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to write page source", err)
	}

	// Update memory usage
	m.updateMemoryUsage(int64(len(sourceBytes)))

	return filePath, nil
}

// checkMemoryLimits checks if memory usage is within limits
func (m *MemoryEfficientCaptureManager) checkMemoryLimits() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.totalMemoryUsed > m.maxMemoryUsage {
		return core.NewGowrightError(core.BrowserError, "memory usage limit exceeded", nil)
	}

	if len(m.activeCaptures) >= m.config.MaxConcurrentCaptures {
		return core.NewGowrightError(core.BrowserError, "maximum concurrent captures exceeded", nil)
	}

	return nil
}

// updateMemoryUsage updates the total memory usage
func (m *MemoryEfficientCaptureManager) updateMemoryUsage(delta int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalMemoryUsed += delta
}

// compressFile compresses a file and returns the compressed file path
func (m *MemoryEfficientCaptureManager) compressFile(filePath string) (string, error) {
	// Read original file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Compress data
	compressedData, err := m.compressData(data)
	if err != nil {
		return "", err
	}

	// Write compressed file
	compressedPath := filePath + ".gz"
	if err := os.WriteFile(compressedPath, compressedData, 0600); err != nil {
		return "", err
	}

	return compressedPath, nil
}

// compressData compresses byte data using gzip
func (m *MemoryEfficientCaptureManager) compressData(data []byte) ([]byte, error) {
	writer := m.compressionPool.Get().(*gzip.Writer)
	defer m.compressionPool.Put(writer)

	var buf bytes.Buffer
	writer.Reset(&buf)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// cleanupRoutine runs periodic cleanup of old captures
func (m *MemoryEfficientCaptureManager) cleanupRoutine() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.performCleanup()
		runtime.GC() // Force garbage collection
	}
}

// performCleanup removes old captures and frees memory
func (m *MemoryEfficientCaptureManager) performCleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for id, capture := range m.activeCaptures {
		if now.Sub(capture.StartTime) > m.config.MaxCaptureAge {
			// Remove old capture
			if capture.FilePath != "" {
				if removeErr := os.Remove(capture.FilePath); removeErr != nil && !os.IsNotExist(removeErr) {
					// Log warning but continue cleanup
					fmt.Printf("Warning: failed to remove old capture file %s: %v\n", capture.FilePath, removeErr)
				}
			}
			m.totalMemoryUsed -= capture.MemoryUsed
			delete(m.activeCaptures, id)
		}
	}
}

// GetStats returns current memory usage statistics
func (m *MemoryEfficientCaptureManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"total_memory_used_mb":    m.totalMemoryUsed / (1024 * 1024),
		"max_memory_usage_mb":     m.maxMemoryUsage / (1024 * 1024),
		"active_captures":         len(m.activeCaptures),
		"max_concurrent_captures": m.config.MaxConcurrentCaptures,
		"compression_enabled":     m.config.EnableCompression,
	}
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	ActiveCaptures    int   `json:"active_captures"`
	TotalMemoryUsedMB int64 `json:"total_memory_used_mb"`
	MaxMemoryUsageMB  int64 `json:"max_memory_usage_mb"`
	SystemMemoryMB    int64 `json:"system_memory_mb"`
	GCCount           int64 `json:"gc_count"`
}

// GetMemoryStats returns detailed memory statistics
func (m *MemoryEfficientCaptureManager) GetMemoryStats() MemoryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryStats{
		ActiveCaptures:    len(m.activeCaptures),
		TotalMemoryUsedMB: m.totalMemoryUsed / (1024 * 1024),
		MaxMemoryUsageMB:  m.maxMemoryUsage / (1024 * 1024),
		SystemMemoryMB:    func() int64 {
			if memStats.Sys > 9223372036854775807 { // max int64
				return 9223372036854775807 / (1024 * 1024)
			}
			return int64(memStats.Sys) / (1024 * 1024)
		}(),
		GCCount:           int64(memStats.NumGC),
	}
}

// CaptureResult holds the result of a data capture operation
type CaptureResult struct {
	ID           string `json:"id"`
	FilePath     string `json:"file_path"`
	OriginalSize int64  `json:"original_size"`
	FinalSize    int64  `json:"final_size"`
	Compressed   bool   `json:"compressed"`
	Optimized    bool   `json:"optimized"`
}

// CaptureDataStreamOptimized captures data with streaming and compression optimization
func (m *MemoryEfficientCaptureManager) CaptureDataStreamOptimized(data []byte, testName, dataType string) (*CaptureResult, error) {
	// Check memory usage before capture
	if err := m.checkMemoryLimits(); err != nil {
		return nil, err
	}

	captureID := fmt.Sprintf("%s_%s_%d", testName, dataType, time.Now().UnixNano())
	originalSize := int64(len(data))

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.dat", testName, dataType, timestamp)
	filePath := filepath.Join("./captures/data", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
		return nil, core.NewGowrightError(core.BrowserError, "failed to create data directory", err)
	}

	var finalData []byte
	var compressed bool
	var err error

	// Determine if we should compress
	shouldCompress := m.config.EnableCompression && originalSize > m.config.CompressThresholdKB*1024

	if shouldCompress {
		if m.config.EnableStreaming {
			// Use streaming compression
			file, err := os.Create(filePath + ".gz")
			if err != nil {
				return nil, err
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					fmt.Printf("Warning: failed to close compressed file: %v\n", closeErr)
				}
			}()

			written, err := m.writeCompressedStream(file, data)
			if err != nil {
				return nil, err
			}
			filePath = filePath + ".gz"
			finalData = make([]byte, written) // Just for size tracking
			compressed = true
		} else {
			// Use in-memory compression
			finalData, err = m.compressData(data)
			if err != nil {
				return nil, err
			}
			filePath = filePath + ".gz"
			compressed = true
		}
	} else {
		if m.config.EnableStreaming {
			// Use streaming without compression
			file, err := os.Create(filePath)
			if err != nil {
				return nil, err
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					fmt.Printf("Warning: failed to close data file: %v\n", closeErr)
				}
			}()

			written, err := m.writeDataStream(file, data)
			if err != nil {
				return nil, err
			}
			finalData = make([]byte, written) // Just for size tracking
		} else {
			// Direct write
			finalData = data
		}
	}

	// Write data if not already written via streaming
	if !m.config.EnableStreaming {
		if err := os.WriteFile(filePath, finalData, 0600); err != nil {
			return nil, core.NewGowrightError(core.BrowserError, "failed to write data file", err)
		}
	}

	finalSize := int64(len(finalData))
	if m.config.EnableStreaming {
		// Get actual file size for streaming
		if fileInfo, err := os.Stat(filePath); err == nil {
			finalSize = fileInfo.Size()
		}
	}

	// Register capture
	captureInfo := &CaptureInfo{
		ID:           captureID,
		TestName:     testName,
		Type:         dataType,
		StartTime:    time.Now(),
		MemoryUsed:   finalSize,
		Size:         finalSize,
		OriginalSize: originalSize,
		FinalSize:    finalSize,
		Compressed:   compressed,
		Optimized:    shouldCompress || m.config.EnableStreaming,
		FilePath:     filePath,
		Status:       "completed",
	}

	m.mutex.Lock()
	m.activeCaptures[captureID] = captureInfo
	m.totalMemoryUsed += finalSize
	m.mutex.Unlock()

	return &CaptureResult{
		ID:           captureID,
		FilePath:     filePath,
		OriginalSize: originalSize,
		FinalSize:    finalSize,
		Compressed:   compressed,
		Optimized:    captureInfo.Optimized,
	}, nil
}

// ReleaseCaptureMemory releases memory and resources for a specific capture
func (m *MemoryEfficientCaptureManager) ReleaseCaptureMemory(captureID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	capture, exists := m.activeCaptures[captureID]
	if !exists {
		return fmt.Errorf("capture %s not found", captureID)
	}

	// Remove file if it exists
	if capture.FilePath != "" {
		if err := os.Remove(capture.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to remove capture file %s: %v\n", capture.FilePath, err)
		}
	}

	// Update memory usage
	m.totalMemoryUsed -= capture.MemoryUsed

	// Remove from active captures
	delete(m.activeCaptures, captureID)

	return nil
}

// writeCompressedStream writes data to a stream with compression
func (m *MemoryEfficientCaptureManager) writeCompressedStream(writer interface{ Write([]byte) (int, error) }, data []byte) (int64, error) {
	// Create gzip writer
	gzipWriter, err := gzip.NewWriterLevel(writer, m.config.CompressionLevel)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close gzip writer: %v\n", closeErr)
		}
	}()

	// Write data in chunks
	chunkSize := int(m.config.StreamChunkSizeKB * 1024)
	totalWritten := int64(0)

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[i:end]
		written, err := gzipWriter.Write(chunk)
		if err != nil {
			return totalWritten, err
		}
		totalWritten += int64(written)
	}

	// Ensure all data is flushed
	if err := gzipWriter.Close(); err != nil {
		return totalWritten, err
	}

	// For compressed streams, we need to return the actual compressed size
	// This is a bit tricky since we're writing to a stream, but we can estimate
	return totalWritten, nil
}

// writeDataStream writes data to a stream without compression
func (m *MemoryEfficientCaptureManager) writeDataStream(writer interface{ Write([]byte) (int, error) }, data []byte) (int64, error) {
	chunkSize := int(m.config.StreamChunkSizeKB * 1024)
	totalWritten := int64(0)

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[i:end]
		written, err := writer.Write(chunk)
		if err != nil {
			return totalWritten, err
		}
		totalWritten += int64(written)
	}

	return totalWritten, nil
}

// Cleanup performs final cleanup of all resources
func (m *MemoryEfficientCaptureManager) Cleanup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Clean up all active captures
	for id, capture := range m.activeCaptures {
		if capture.FilePath != "" {
			if removeErr := os.Remove(capture.FilePath); removeErr != nil && !os.IsNotExist(removeErr) {
				fmt.Printf("Warning: failed to remove capture file during cleanup %s: %v\n", capture.FilePath, removeErr)
			}
		}
		delete(m.activeCaptures, id)
	}

	m.totalMemoryUsed = 0
	return nil
}
