package gowright

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// MemoryEfficientCaptureManager handles memory-efficient screenshot and data capture
type MemoryEfficientCaptureManager struct {
	config          *MemoryEfficientConfig
	compressionPool sync.Pool
	imagePool       sync.Pool
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
	AutoCleanupEnabled bool  `json:"auto_cleanup_enabled"`
	CleanupThresholdMB int64 `json:"cleanup_threshold_mb"`
}

// CaptureInfo holds information about active captures
type CaptureInfo struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	Compressed   bool      `json:"compressed"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	FilePath     string    `json:"file_path"`
}

// StreamingCapture represents a streaming capture operation
type StreamingCapture struct {
	// Fields removed as they were unused
}

// DefaultMemoryEfficientConfig returns default configuration
func DefaultMemoryEfficientConfig() *MemoryEfficientConfig {
	return &MemoryEfficientConfig{
		MaxMemoryUsageMB:    512, // 512MB
		MaxScreenshotSizeMB: 10,  // 10MB per screenshot
		EnableCompression:   true,
		CompressionLevel:    6,   // Balanced compression
		CompressThresholdKB: 100, // Compress files > 100KB
		MaxImageWidth:       1920,
		MaxImageHeight:      1080,
		JPEGQuality:         85,
		EnableStreaming:     true,
		StreamChunkSizeKB:   64, // 64KB chunks
		AutoCleanupEnabled:  true,
		CleanupThresholdMB:  256, // Cleanup when 256MB used
	}
}

// NewMemoryEfficientCaptureManager creates a new memory-efficient capture manager
func NewMemoryEfficientCaptureManager(config *MemoryEfficientConfig) *MemoryEfficientCaptureManager {
	if config == nil {
		config = DefaultMemoryEfficientConfig()
	}

	mecm := &MemoryEfficientCaptureManager{
		config:         config,
		activeCaptures: make(map[string]*CaptureInfo),
		maxMemoryUsage: config.MaxMemoryUsageMB * 1024 * 1024,
	}

	// Initialize object pools for memory reuse
	mecm.compressionPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	mecm.imagePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024*1024) // 1MB initial capacity
		},
	}

	mecm.bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, config.StreamChunkSizeKB*1024)
		},
	}

	return mecm
}

// CaptureScreenshotOptimized captures a screenshot with memory optimization
func (mecm *MemoryEfficientCaptureManager) CaptureScreenshotOptimized(tester UITester, testName string, options *OptimizedCaptureOptions) (*CaptureResult, error) {
	if options == nil {
		options = DefaultOptimizedCaptureOptions()
	}

	// Check memory usage before capture
	if err := mecm.checkMemoryUsage(); err != nil {
		return nil, fmt.Errorf("memory check failed: %w", err)
	}

	captureID := fmt.Sprintf("%s_%d", testName, time.Now().UnixNano())

	// Create temporary file for screenshot
	tempFile, err := mecm.createTempFile(captureID, "screenshot", ".png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = tempFile.Close() }()

	// Capture screenshot to temporary file
	screenshotPath, err := tester.TakeScreenshot(tempFile.Name())
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(screenshotPath)
	if err != nil {
		_ = os.Remove(screenshotPath)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	originalSize := fileInfo.Size()

	// Optimize the image if needed
	optimizedPath, finalSize, err := mecm.optimizeImage(screenshotPath, options)
	if err != nil {
		_ = os.Remove(screenshotPath)
		return nil, fmt.Errorf("failed to optimize image: %w", err)
	}

	// Track the capture
	captureInfo := &CaptureInfo{
		ID:           captureID,
		Type:         "screenshot",
		Size:         finalSize,
		Compressed:   finalSize < originalSize,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		FilePath:     optimizedPath,
	}

	mecm.mutex.Lock()
	mecm.activeCaptures[captureID] = captureInfo
	mecm.totalMemoryUsed += finalSize
	mecm.mutex.Unlock()

	return &CaptureResult{
		ID:           captureID,
		FilePath:     optimizedPath,
		OriginalSize: originalSize,
		FinalSize:    finalSize,
		Compressed:   captureInfo.Compressed,
		Optimized:    finalSize < originalSize,
	}, nil
}

// optimizeImage optimizes an image file for memory efficiency
func (mecm *MemoryEfficientCaptureManager) optimizeImage(imagePath string, options *OptimizedCaptureOptions) (string, int64, error) {
	// Open the original image
	file, err := os.Open(imagePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open image: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return "", 0, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if needed
	if options.MaxWidth > 0 || options.MaxHeight > 0 {
		img = mecm.resizeImage(img, options.MaxWidth, options.MaxHeight)
	}

	// Create optimized file
	optimizedPath := imagePath
	if options.ConvertToJPEG && format != "jpeg" {
		optimizedPath = mecm.changeExtension(imagePath, ".jpg")
	}

	// Save optimized image
	optimizedFile, err := os.Create(optimizedPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create optimized file: %w", err)
	}
	defer func() { _ = optimizedFile.Close() }()

	// Encode with optimization
	if options.ConvertToJPEG || format == "jpeg" {
		err = jpeg.Encode(optimizedFile, img, &jpeg.Options{Quality: options.JPEGQuality})
	} else {
		err = png.Encode(optimizedFile, img)
	}

	if err != nil {
		return "", 0, fmt.Errorf("failed to encode optimized image: %w", err)
	}

	// Get final size
	fileInfo, err := optimizedFile.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get optimized file info: %w", err)
	}

	// Remove original if different from optimized
	if optimizedPath != imagePath {
		_ = os.Remove(imagePath)
	}

	return optimizedPath, fileInfo.Size(), nil
}

// resizeImage resizes an image to fit within the specified dimensions
func (mecm *MemoryEfficientCaptureManager) resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling factor
	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Only resize if image is larger than max dimensions
	if scale >= 1.0 {
		return img
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Create new image with scaled dimensions
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest neighbor scaling (for better performance)
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	return newImg
}

// CaptureDataStreamOptimized captures large data with streaming and compression
func (mecm *MemoryEfficientCaptureManager) CaptureDataStreamOptimized(data []byte, testName, dataType string) (*CaptureResult, error) {
	captureID := fmt.Sprintf("%s_%s_%d", testName, dataType, time.Now().UnixNano())

	// Check memory usage
	if err := mecm.checkMemoryUsage(); err != nil {
		return nil, fmt.Errorf("memory check failed: %w", err)
	}

	// Create temporary file
	tempFile, err := mecm.createTempFile(captureID, dataType, ".dat")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = tempFile.Close() }()

	originalSize := int64(len(data))
	var finalSize int64
	var compressed bool

	// Determine if compression should be used
	shouldCompress := mecm.config.EnableCompression &&
		originalSize > mecm.config.CompressThresholdKB*1024

	if shouldCompress {
		// Use streaming compression
		finalSize, err = mecm.writeCompressedStream(tempFile, data)
		if err != nil {
			_ = os.Remove(tempFile.Name())
			return nil, fmt.Errorf("failed to write compressed data: %w", err)
		}
		compressed = true
	} else {
		// Write data directly in chunks to avoid memory spikes
		finalSize, err = mecm.writeDataStream(tempFile, data)
		if err != nil {
			_ = os.Remove(tempFile.Name())
			return nil, fmt.Errorf("failed to write data: %w", err)
		}
	}

	// Track the capture
	captureInfo := &CaptureInfo{
		ID:           captureID,
		Type:         dataType,
		Size:         finalSize,
		Compressed:   compressed,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		FilePath:     tempFile.Name(),
	}

	mecm.mutex.Lock()
	mecm.activeCaptures[captureID] = captureInfo
	mecm.totalMemoryUsed += finalSize
	mecm.mutex.Unlock()

	return &CaptureResult{
		ID:           captureID,
		FilePath:     tempFile.Name(),
		OriginalSize: originalSize,
		FinalSize:    finalSize,
		Compressed:   compressed,
		Optimized:    finalSize < originalSize,
	}, nil
}

// writeCompressedStream writes data using streaming compression
func (mecm *MemoryEfficientCaptureManager) writeCompressedStream(writer io.Writer, data []byte) (int64, error) {
	// Get buffer from pool
	buffer := mecm.compressionPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer mecm.compressionPool.Put(buffer)

	// Create gzip writer
	gzipWriter, err := gzip.NewWriterLevel(buffer, mecm.config.CompressionLevel)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	// Write data in chunks to avoid memory spikes
	chunkSize := mecm.config.StreamChunkSizeKB * 1024
	var totalWritten int64

	for i := int64(0); i < int64(len(data)); i += chunkSize {
		end := i + chunkSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}

		chunk := data[i:end]
		if _, err := gzipWriter.Write(chunk); err != nil {
			_ = gzipWriter.Close()
			return 0, fmt.Errorf("failed to write chunk: %w", err)
		}

		// Flush buffer to writer periodically to avoid memory buildup
		if buffer.Len() > int(chunkSize) {
			if err := gzipWriter.Flush(); err != nil {
				_ = gzipWriter.Close()
				return 0, fmt.Errorf("failed to flush gzip writer: %w", err)
			}

			written, err := writer.Write(buffer.Bytes())
			if err != nil {
				_ = gzipWriter.Close()
				return 0, fmt.Errorf("failed to write to output: %w", err)
			}

			totalWritten += int64(written)
			buffer.Reset()
		}
	}

	// Close gzip writer and write remaining data
	if err := gzipWriter.Close(); err != nil {
		return 0, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if buffer.Len() > 0 {
		written, err := writer.Write(buffer.Bytes())
		if err != nil {
			return 0, fmt.Errorf("failed to write final data: %w", err)
		}
		totalWritten += int64(written)
	}

	return totalWritten, nil
}

// writeDataStream writes data in chunks without compression
func (mecm *MemoryEfficientCaptureManager) writeDataStream(writer io.Writer, data []byte) (int64, error) {
	chunkSize := mecm.config.StreamChunkSizeKB * 1024
	var totalWritten int64

	for i := int64(0); i < int64(len(data)); i += chunkSize {
		end := i + chunkSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}

		chunk := data[i:end]
		written, err := writer.Write(chunk)
		if err != nil {
			return totalWritten, fmt.Errorf("failed to write chunk: %w", err)
		}

		totalWritten += int64(written)
	}

	return totalWritten, nil
}

// checkMemoryUsage checks current memory usage and triggers cleanup if needed
func (mecm *MemoryEfficientCaptureManager) checkMemoryUsage() error {
	mecm.mutex.RLock()
	currentUsage := mecm.totalMemoryUsed
	mecm.mutex.RUnlock()

	if currentUsage > mecm.maxMemoryUsage {
		return fmt.Errorf("memory usage exceeded: %d MB > %d MB",
			currentUsage/(1024*1024), mecm.maxMemoryUsage/(1024*1024))
	}

	// Trigger cleanup if threshold is reached
	if mecm.config.AutoCleanupEnabled &&
		currentUsage > mecm.config.CleanupThresholdMB*1024*1024 {
		go mecm.performMemoryCleanup()
	}

	return nil
}

// performMemoryCleanup performs cleanup to free memory
func (mecm *MemoryEfficientCaptureManager) performMemoryCleanup() {
	mecm.mutex.Lock()
	defer mecm.mutex.Unlock()

	// Find oldest captures to remove
	cutoffTime := time.Now().Add(-1 * time.Hour)
	var toRemove []string

	for id, capture := range mecm.activeCaptures {
		if capture.LastAccessed.Before(cutoffTime) {
			toRemove = append(toRemove, id)
		}
	}

	// Remove old captures
	for _, id := range toRemove {
		if capture, exists := mecm.activeCaptures[id]; exists {
			_ = os.Remove(capture.FilePath)
			mecm.totalMemoryUsed -= capture.Size
			delete(mecm.activeCaptures, id)
		}
	}

	// Force garbage collection
	runtime.GC()
}

// createTempFile creates a temporary file for captures
func (mecm *MemoryEfficientCaptureManager) createTempFile(id, dataType, extension string) (*os.File, error) {
	tempDir := filepath.Join(os.TempDir(), "gowright_captures")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	filename := fmt.Sprintf("%s_%s%s", id, dataType, extension)
	filepath := filepath.Join(tempDir, filename)

	return os.Create(filepath)
}

// changeExtension changes the file extension
func (mecm *MemoryEfficientCaptureManager) changeExtension(filePath, newExt string) string {
	dir := filepath.Dir(filePath)
	name := filepath.Base(filePath)
	nameWithoutExt := name[:len(name)-len(filepath.Ext(name))]
	return filepath.Join(dir, nameWithoutExt+newExt)
}

// OptimizedCaptureOptions holds options for optimized capture
type OptimizedCaptureOptions struct {
	MaxWidth      int  `json:"max_width"`
	MaxHeight     int  `json:"max_height"`
	ConvertToJPEG bool `json:"convert_to_jpeg"`
	JPEGQuality   int  `json:"jpeg_quality"`
}

// DefaultOptimizedCaptureOptions returns default capture options
func DefaultOptimizedCaptureOptions() *OptimizedCaptureOptions {
	return &OptimizedCaptureOptions{
		MaxWidth:      1920,
		MaxHeight:     1080,
		ConvertToJPEG: true,
		JPEGQuality:   85,
	}
}

// CaptureResult holds the result of a capture operation
type CaptureResult struct {
	ID           string `json:"id"`
	FilePath     string `json:"file_path"`
	OriginalSize int64  `json:"original_size"`
	FinalSize    int64  `json:"final_size"`
	Compressed   bool   `json:"compressed"`
	Optimized    bool   `json:"optimized"`
}

// GetMemoryStats returns current memory usage statistics
func (mecm *MemoryEfficientCaptureManager) GetMemoryStats() *MemoryStats {
	mecm.mutex.RLock()
	defer mecm.mutex.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		ActiveCaptures:    len(mecm.activeCaptures),
		TotalMemoryUsedMB: mecm.totalMemoryUsed / (1024 * 1024),
		MaxMemoryUsageMB:  mecm.maxMemoryUsage / (1024 * 1024),
		SystemMemoryMB:    int64(m.Alloc) / (1024 * 1024),
		GCCount:           int64(m.NumGC),
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

// ReleaseCaptureMemory releases memory associated with a capture
func (mecm *MemoryEfficientCaptureManager) ReleaseCaptureMemory(captureID string) error {
	mecm.mutex.Lock()
	defer mecm.mutex.Unlock()

	capture, exists := mecm.activeCaptures[captureID]
	if !exists {
		return fmt.Errorf("capture %s not found", captureID)
	}

	// Remove file
	if err := os.Remove(capture.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove capture file: %w", err)
	}

	// Update memory tracking
	mecm.totalMemoryUsed -= capture.Size
	delete(mecm.activeCaptures, captureID)

	return nil
}

// Cleanup performs cleanup of all managed resources
func (mecm *MemoryEfficientCaptureManager) Cleanup() error {
	mecm.mutex.Lock()
	defer mecm.mutex.Unlock()

	var errors []error

	// Remove all capture files
	for id, capture := range mecm.activeCaptures {
		if err := os.Remove(capture.FilePath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to remove capture %s: %w", id, err))
		}
	}

	// Reset tracking
	mecm.activeCaptures = make(map[string]*CaptureInfo)
	mecm.totalMemoryUsed = 0

	// Force garbage collection
	runtime.GC()

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}
