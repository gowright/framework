package gowright

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkResourceManagerAcquisition benchmarks resource acquisition performance
func BenchmarkResourceManagerAcquisition(b *testing.B) {
	config := DefaultParallelRunnerConfig()
	config.BrowserPoolSize = 10
	config.HTTPClientPoolSize = 20
	config.DatabasePoolSize = 15

	rm := NewResourceManager(config)
	ctx := context.Background()

	err := rm.Initialize(ctx)
	if err != nil {
		b.Fatalf("Failed to initialize resource manager: %v", err)
	}
	defer func() { _ = rm.Cleanup() }()

	mockTest := &MockTestPerf{name: "benchmark_test"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			handle, err := rm.AcquireResources(ctx, mockTest)
			if err != nil {
				b.Errorf("Failed to acquire resources: %v", err)
				continue
			}

			// Simulate some work
			time.Sleep(1 * time.Millisecond)

			err = rm.ReleaseResources(handle)
			if err != nil {
				b.Errorf("Failed to release resources: %v", err)
			}
		}
	})
}

// BenchmarkResourceCleanupManager benchmarks cleanup operations
func BenchmarkResourceCleanupManager(b *testing.B) {
	config := DefaultResourceCleanupConfig()
	config.AutoCleanupInterval = 1 * time.Hour // Disable auto cleanup for benchmark

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Register temp files
			for i := 0; i < 10; i++ {
				tempFile := fmt.Sprintf("/tmp/benchmark_file_%d_%d", b.N, i)
				rcm.RegisterTempFile(tempFile)
			}

			// Track resources
			for i := 0; i < 5; i++ {
				resourceID := fmt.Sprintf("resource_%d_%d", b.N, i)
				rcm.TrackResource("test_resource", resourceID, 1024, nil)
			}
		}
	})
}

// BenchmarkMemoryEfficientCapture benchmarks memory-efficient capture operations
func BenchmarkMemoryEfficientCapture(b *testing.B) {
	config := DefaultMemoryEfficientConfig()
	config.AutoCleanupEnabled = false // Disable auto cleanup for benchmark

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Create test data
	testData := make([]byte, 1024*1024) // 1MB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		testNum := 0
		for pb.Next() {
			testName := fmt.Sprintf("benchmark_test_%d", testNum)
			testNum++

			result, err := mecm.CaptureDataStreamOptimized(testData, testName, "benchmark_data")
			if err != nil {
				b.Errorf("Failed to capture data: %v", err)
				continue
			}

			// Clean up immediately to avoid memory buildup
			_ = mecm.ReleaseCaptureMemory(result.ID)
		}
	})
}

// BenchmarkResourceLeakDetector benchmarks leak detection performance
func BenchmarkResourceLeakDetector(b *testing.B) {
	config := DefaultLeakDetectorConfig()
	config.ScanInterval = 1 * time.Hour // Disable auto scanning for benchmark

	rld := NewResourceLeakDetector(config)
	defer func() { _ = rld.Shutdown() }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		resourceNum := 0
		for pb.Next() {
			resourceID := fmt.Sprintf("benchmark_resource_%d", resourceNum)
			resourceNum++

			// Track resource
			rld.TrackResource(resourceID, "test_resource", 1024, map[string]interface{}{
				"test": "benchmark",
			})

			// Simulate some access
			rld.UpdateResourceAccess(resourceID)
			rld.IncrementRefCount(resourceID)
			rld.DecrementRefCount(resourceID)

			// Untrack resource
			rld.UntrackResource(resourceID)
		}
	})
}

// TestResourceManagerMemoryUsage tests memory usage under load
func TestResourceManagerMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	config := DefaultParallelRunnerConfig()
	config.BrowserPoolSize = 5
	config.HTTPClientPoolSize = 10
	config.DatabasePoolSize = 8

	rm := NewResourceManager(config)
	ctx := context.Background()

	err := rm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize resource manager: %v", err)
	}
	defer func() { _ = rm.Cleanup() }()

	// Measure initial memory
	runtime.GC()
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// Simulate high load
	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			mockTest := &MockTestPerf{name: fmt.Sprintf("load_test_%d", goroutineID)}

			for j := 0; j < operationsPerGoroutine; j++ {
				handle, err := rm.AcquireResources(ctx, mockTest)
				if err != nil {
					t.Errorf("Failed to acquire resources: %v", err)
					continue
				}

				// Simulate work
				time.Sleep(1 * time.Millisecond)

				err = rm.ReleaseResources(handle)
				if err != nil {
					t.Errorf("Failed to release resources: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Measure final memory
	runtime.GC()
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)

	memoryIncrease := finalMem.Alloc - initialMem.Alloc
	memoryIncreaseMB := memoryIncrease / (1024 * 1024)

	t.Logf("Memory usage increase: %d MB", memoryIncreaseMB)

	// Memory increase should be reasonable (less than 100MB for this test)
	if memoryIncreaseMB > 100 {
		t.Errorf("Memory usage increased by %d MB, which is too high", memoryIncreaseMB)
	}
}

// TestResourceCleanupEffectiveness tests the effectiveness of resource cleanup
func TestResourceCleanupEffectiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup effectiveness test in short mode")
	}

	config := DefaultResourceCleanupConfig()
	config.AutoCleanupInterval = 100 * time.Millisecond
	config.TempFileMaxAge = 200 * time.Millisecond
	config.MemoryThresholdMB = 1 // Very low threshold to trigger cleanup

	rcm := NewResourceCleanupManager(config)
	defer func() { _ = rcm.Shutdown() }()

	// Create temporary files
	tempFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		tempFile, err := os.CreateTemp("", "cleanup_test_")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		_ = tempFile.Close()

		tempFiles[i] = tempFile.Name()
		rcm.RegisterTempFile(tempFile.Name())
	}

	// Wait for cleanup to occur
	time.Sleep(500 * time.Millisecond)

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

	t.Logf("Cleaned up %d out of %d temporary files", cleanedCount, len(tempFiles))

	// At least some files should have been cleaned up
	if cleanedCount == 0 {
		t.Error("No temporary files were cleaned up")
	}
}

// TestMemoryEfficientCaptureCompression tests compression effectiveness
func TestMemoryEfficientCaptureCompression(t *testing.T) {
	config := DefaultMemoryEfficientConfig()
	config.EnableCompression = true
	config.CompressThresholdKB = 1 // Compress everything

	mecm := NewMemoryEfficientCaptureManager(config)
	defer func() { _ = mecm.Cleanup() }()

	// Create compressible test data (repeated pattern)
	testData := make([]byte, 100*1024) // 100KB
	pattern := []byte("This is a test pattern that should compress well. ")
	for i := 0; i < len(testData); i++ {
		testData[i] = pattern[i%len(pattern)]
	}

	result, err := mecm.CaptureDataStreamOptimized(testData, "compression_test", "test_data")
	if err != nil {
		t.Fatalf("Failed to capture data: %v", err)
	}
	defer func() { _ = mecm.ReleaseCaptureMemory(result.ID) }()

	compressionRatio := float64(result.OriginalSize) / float64(result.FinalSize)

	t.Logf("Original size: %d bytes", result.OriginalSize)
	t.Logf("Compressed size: %d bytes", result.FinalSize)
	t.Logf("Compression ratio: %.2f:1", compressionRatio)

	// Should achieve at least 2:1 compression on this repetitive data
	if compressionRatio < 2.0 {
		t.Errorf("Compression ratio %.2f is lower than expected", compressionRatio)
	}

	if !result.Compressed {
		t.Error("Data should have been compressed but wasn't")
	}
}

// TestResourceLeakDetection tests leak detection accuracy
func TestResourceLeakDetection(t *testing.T) {
	config := DefaultLeakDetectorConfig()
	config.LeakThreshold = 100 * time.Millisecond
	config.ScanInterval = 50 * time.Millisecond

	rld := NewResourceLeakDetector(config)
	defer func() { _ = rld.Shutdown() }()

	// Create some resources that will be properly cleaned up
	for i := 0; i < 5; i++ {
		resourceID := fmt.Sprintf("good_resource_%d", i)
		rld.TrackResource(resourceID, "test_resource", 1024, nil)

		// Simulate proper cleanup
		go func(id string) {
			time.Sleep(30 * time.Millisecond)
			rld.UntrackResource(id)
		}(resourceID)
	}

	// Create some resources that will leak
	leakedResources := make([]string, 3)
	for i := 0; i < 3; i++ {
		resourceID := fmt.Sprintf("leaked_resource_%d", i)
		leakedResources[i] = resourceID
		rld.TrackResource(resourceID, "test_resource", 2048, nil)
		// Set ref count to 0 to simulate abandoned resources
		rld.DecrementRefCount(resourceID)
		// Don't clean these up - they should be detected as leaks
	}

	// Wait for leak detection - need to wait longer than LeakThreshold + some scan intervals
	// LeakThreshold is 100ms, so wait at least 200ms to ensure resources are considered old enough
	// Plus additional time for multiple scan intervals to occur
	time.Sleep(400 * time.Millisecond)

	// Check leak detection results
	stats := rld.GetDetectorStats()

	t.Logf("Tracked resources: %d", stats.TrackedResources)
	t.Logf("Total leaks detected: %d", stats.TotalLeaksDetected)
	t.Logf("Total memory leaked: %d bytes", stats.TotalMemoryLeaked)

	// Should detect the leaked resources
	if stats.TotalLeaksDetected == 0 {
		t.Error("No leaks were detected, but some resources were intentionally leaked")
	}

	// Get the latest leak report
	report := rld.GetLatestLeakReport()
	if report == nil {
		t.Error("No leak report was generated")
	} else {
		t.Logf("Leak report severity: %s", report.Severity.String())
		t.Logf("Leak report recommendations: %v", report.Recommendations)

		if report.TotalLeaks == 0 {
			t.Error("Leak report shows no leaks, but leaks were expected")
		}
	}
}

// BenchmarkParallelResourceOperations benchmarks concurrent resource operations
func BenchmarkParallelResourceOperations(b *testing.B) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)
	ctx := context.Background()

	err := rm.Initialize(ctx)
	if err != nil {
		b.Fatalf("Failed to initialize resource manager: %v", err)
	}
	defer func() { _ = rm.Cleanup() }()

	// Also test cleanup manager
	cleanupConfig := DefaultResourceCleanupConfig()
	cleanupConfig.AutoCleanupInterval = 1 * time.Hour // Disable for benchmark
	rcm := NewResourceCleanupManager(cleanupConfig)
	defer func() { _ = rcm.Shutdown() }()

	// And leak detector
	leakConfig := DefaultLeakDetectorConfig()
	leakConfig.ScanInterval = 1 * time.Hour // Disable for benchmark
	rld := NewResourceLeakDetector(leakConfig)
	defer func() { _ = rld.Shutdown() }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		testNum := 0
		for pb.Next() {
			testName := fmt.Sprintf("parallel_test_%d", testNum)
			testNum++

			// Resource manager operations
			mockTest := &MockTestPerf{name: testName}
			handle, err := rm.AcquireResources(ctx, mockTest)
			if err == nil {
				_ = rm.ReleaseResources(handle)
			}

			// Cleanup manager operations
			resourceID := fmt.Sprintf("resource_%s", testName)
			rcm.TrackResource("test_resource", resourceID, 1024, nil)
			rcm.UpdateResourceAccess(resourceID)
			rcm.UntrackResource(resourceID)

			// Leak detector operations
			leakResourceID := fmt.Sprintf("leak_resource_%s", testName)
			rld.TrackResource(leakResourceID, "test_resource", 512, nil)
			rld.UpdateResourceAccess(leakResourceID)
			rld.UntrackResource(leakResourceID)
		}
	})
}

// TestResourceManagerConcurrentStress tests resource manager under concurrent stress
func TestResourceManagerConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := DefaultParallelRunnerConfig()
	config.BrowserPoolSize = 3
	config.HTTPClientPoolSize = 5
	config.DatabasePoolSize = 4

	rm := NewResourceManager(config)
	ctx := context.Background()

	err := rm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize resource manager: %v", err)
	}
	defer func() { _ = rm.Cleanup() }()

	const numGoroutines = 20
	const operationsPerGoroutine = 50

	var wg sync.WaitGroup
	var errorCount int64
	var successCount int64
	var mutex sync.Mutex

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				mockTest := &MockTestPerf{name: fmt.Sprintf("stress_test_%d_%d", goroutineID, j)}

				handle, err := rm.AcquireResources(ctx, mockTest)
				if err != nil {
					mutex.Lock()
					errorCount++
					mutex.Unlock()
					continue
				}

				// Simulate variable work duration
				workDuration := time.Duration(j%10) * time.Millisecond
				time.Sleep(workDuration)

				err = rm.ReleaseResources(handle)
				if err != nil {
					mutex.Lock()
					errorCount++
					mutex.Unlock()
				} else {
					mutex.Lock()
					successCount++
					mutex.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	totalOperations := int64(numGoroutines * operationsPerGoroutine)
	successRate := float64(successCount) / float64(totalOperations) * 100

	t.Logf("Total operations: %d", totalOperations)
	t.Logf("Successful operations: %d", successCount)
	t.Logf("Failed operations: %d", errorCount)
	t.Logf("Success rate: %.2f%%", successRate)

	// Should have a high success rate under stress
	if successRate < 95.0 {
		t.Errorf("Success rate %.2f%% is too low", successRate)
	}

	// Get final stats
	stats := rm.GetStats()
	t.Logf("Final resource stats: %+v", stats)
}

// MockTestPerf implements the Test interface for benchmarking
type MockTestPerf struct {
	name string
}

func (mt *MockTestPerf) GetName() string {
	return mt.name
}

func (mt *MockTestPerf) Execute() *TestCaseResult {
	return &TestCaseResult{
		Name:      mt.name,
		Status:    TestStatusPassed,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  1 * time.Millisecond,
	}
}
