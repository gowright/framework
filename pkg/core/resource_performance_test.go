package core

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// BenchmarkResourceManagerOperations benchmarks resource manager operations
func BenchmarkResourceManagerOperations(b *testing.B) {
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		resourceNum := 0
		for pb.Next() {
			resourceID := fmt.Sprintf("benchmark_resource_%d", resourceNum)
			resourceNum++

			resource := &MockResource{
				id:        resourceID,
				resType:   ResourceTypeBrowser,
				createdAt: time.Now(),
				lastUsed:  time.Now(),
				active:    true,
			}

			// Register resource
			err := rm.RegisterResource(resource)
			if err != nil {
				b.Errorf("Failed to register resource: %v", err)
				continue
			}

			// Simulate some work
			time.Sleep(1 * time.Microsecond)

			// Unregister resource
			err = rm.UnregisterResource(resourceID)
			if err != nil {
				b.Errorf("Failed to unregister resource: %v", err)
			}
		}
	})
}

// BenchmarkCleanupManagerOperations benchmarks cleanup manager operations
func BenchmarkCleanupManagerOperations(b *testing.B) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		cleanupNum := 0
		for pb.Next() {
			cleanupName := fmt.Sprintf("benchmark_cleanup_%d", cleanupNum)
			cleanupNum++

			// Register cleanup function
			cm.RegisterCleanup(cleanupName, func() error {
				// Simulate cleanup work
				time.Sleep(1 * time.Microsecond)
				return nil
			}, 10)
		}
	})

	// Execute all cleanups at the end
	metrics := cm.ExecuteCleanup()
	b.Logf("Executed %d cleanup functions", metrics.TotalCleanups)
}

// BenchmarkResourceLeakDetectorOperations benchmarks leak detector operations
func BenchmarkResourceLeakDetectorOperations(b *testing.B) {
	b.Skip("ResourceLeakDetector not implemented yet")
	// config := DefaultLeakDetectorConfig()
	// config.Enabled = true
	// rld := NewResourceLeakDetector(config)
	// defer rld.Stop()

	// b.ResetTimer()
	// b.RunParallel(func(pb *testing.PB) {
	// 	resourceNum := 0
	// 	for pb.Next() {
	// 		resourceID := fmt.Sprintf("benchmark_resource_%d", resourceNum)
	// 		resourceNum++

	// 		// Track resource
	// 		rld.TrackResource(resourceID, ResourceTypeBrowser, map[string]interface{}{
	// 			"test": "benchmark",
	// 		})

	// 		// Simulate some access
	// 		rld.UpdateResourceAccess(resourceID)

	// 		// Untrack resource
	// 		rld.UntrackResource(resourceID)
	// 	}
	// })
}

// TestResourceManagerConcurrentStress tests resource manager under concurrent stress
func TestResourceManagerConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := DefaultResourceManagerConfig()
	config.MaxResources = 100 // Reasonable limit for testing
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

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
				resourceID := fmt.Sprintf("stress_resource_%d_%d", goroutineID, j)
				resource := &MockResource{
					id:        resourceID,
					resType:   ResourceTypeBrowser,
					createdAt: time.Now(),
					lastUsed:  time.Now(),
					active:    true,
				}

				err := rm.RegisterResource(resource)
				if err != nil {
					mutex.Lock()
					errorCount++
					mutex.Unlock()
					continue
				}

				// Simulate variable work duration
				workDuration := time.Duration(j%10) * time.Microsecond
				time.Sleep(workDuration)

				err = rm.UnregisterResource(resourceID)
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

// TestCleanupManagerConcurrentStress tests cleanup manager under concurrent stress
func TestCleanupManagerConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				cleanupName := fmt.Sprintf("stress_cleanup_%d_%d", goroutineID, j)

				cm.RegisterCleanup(cleanupName, func() error {
					// Simulate cleanup work
					time.Sleep(1 * time.Microsecond)
					return nil
				}, j%10) // Variable priority
			}
		}(i)
	}

	wg.Wait()

	// Execute all cleanups
	metrics := cm.ExecuteCleanup()

	expectedCleanups := numGoroutines * operationsPerGoroutine
	t.Logf("Expected cleanups: %d", expectedCleanups)
	t.Logf("Total cleanups: %d", metrics.TotalCleanups)
	t.Logf("Successful cleanups: %d", metrics.SuccessfulCleanups)
	t.Logf("Failed cleanups: %d", metrics.FailedCleanups)

	if metrics.TotalCleanups != expectedCleanups {
		t.Errorf("Expected %d cleanups, got %d", expectedCleanups, metrics.TotalCleanups)
	}

	if metrics.SuccessfulCleanups != expectedCleanups {
		t.Errorf("Expected %d successful cleanups, got %d", expectedCleanups, metrics.SuccessfulCleanups)
	}
}

// TestResourceLeakDetectorConcurrentStress tests leak detector under concurrent stress
func TestResourceLeakDetectorConcurrentStress(t *testing.T) {
	t.Skip("ResourceLeakDetector not implemented yet")
	/*
		if testing.Short() {
			t.Skip("Skipping stress test in short mode")
		}

		// config := DefaultLeakDetectorConfig()
		// config.Enabled = true
		// config.ScanInterval = 1 * time.Hour // Disable auto scanning for this test
		// rld := NewResourceLeakDetector(config)
		// defer rld.Stop()

		const numGoroutines = 15
		const operationsPerGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					resourceID := fmt.Sprintf("stress_resource_%d_%d", goroutineID, j)

					// Track resource
					rld.TrackResource(resourceID, ResourceTypeBrowser, map[string]interface{}{
						"goroutine": goroutineID,
						"operation": j,
					})

					// Simulate some access
					rld.UpdateResourceAccess(resourceID)

					// Untrack resource
					rld.UntrackResource(resourceID)
				}
			}(i)
		}

		wg.Wait()

		// Verify final state
		metrics := rld.GetMetrics()
		t.Logf("Final tracked resources: %d", metrics.ResourcesTracked)
		t.Logf("Total scans: %d", metrics.TotalScans)

		// All resources should have been untracked
		if metrics.ResourcesTracked != 0 {
			t.Errorf("Expected 0 tracked resources, got %d", metrics.ResourcesTracked)
		}
	*/
}

// BenchmarkParallelResourceOperations benchmarks concurrent resource operations across all managers
func BenchmarkParallelResourceOperations(b *testing.B) {
	// Resource manager
	rmConfig := DefaultResourceManagerConfig()
	rm := NewResourceManager(rmConfig)
	defer func() { _ = rm.Shutdown() }()

	// Cleanup manager
	cleanupConfig := DefaultCleanupConfig()
	cm := NewCleanupManager(cleanupConfig)

	// Leak detector
	// leakConfig := DefaultLeakDetectorConfig()
	// leakConfig.Enabled = true
	// leakConfig.ScanInterval = 1 * time.Hour // Disable for benchmark
	// rld := NewResourceLeakDetector(leakConfig)
	// defer rld.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		operationNum := 0
		for pb.Next() {
			opID := fmt.Sprintf("parallel_op_%d", operationNum)
			operationNum++

			// Resource manager operations
			resource := &MockResource{
				id:        opID,
				resType:   ResourceTypeBrowser,
				createdAt: time.Now(),
				lastUsed:  time.Now(),
				active:    true,
			}

			err := rm.RegisterResource(resource)
			if err == nil {
				_ = rm.UnregisterResource(opID)
			}

			// Cleanup manager operations
			cm.RegisterCleanup(opID, func() error {
				return nil
			}, 10)

			// Leak detector operations (commented out)
			// rld.TrackResource(opID, ResourceTypeBrowser, nil)
			// rld.UpdateResourceAccess(opID)
			// rld.UntrackResource(opID)
		}
	})

	// Execute cleanups at the end
	metrics := cm.ExecuteCleanup()
	b.Logf("Executed %d cleanup functions", metrics.TotalCleanups)
}

// TestResourceManagerMemoryEfficiency tests memory usage patterns
func TestResourceManagerMemoryEfficiency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory efficiency test in short mode")
	}

	config := DefaultResourceManagerConfig()
	config.MaxResources = 1000
	rm := NewResourceManager(config)
	defer func() { _ = rm.Shutdown() }()

	// Create and register many resources
	const numResources = 500
	resourceIDs := make([]string, numResources)

	for i := 0; i < numResources; i++ {
		resourceID := fmt.Sprintf("memory_test_resource_%d", i)
		resourceIDs[i] = resourceID

		resource := &MockResource{
			id:        resourceID,
			resType:   ResourceTypeBrowser,
			createdAt: time.Now(),
			lastUsed:  time.Now(),
			active:    true,
		}

		err := rm.RegisterResource(resource)
		if err != nil {
			t.Fatalf("Failed to register resource %d: %v", i, err)
		}
	}

	// Verify all resources are registered
	stats := rm.GetStats()
	if stats[ResourceTypeBrowser].Active != numResources {
		t.Errorf("Expected %d active resources, got %d", numResources, stats[ResourceTypeBrowser].Active)
	}

	// Clean up all resources
	for _, resourceID := range resourceIDs {
		err := rm.UnregisterResource(resourceID)
		if err != nil {
			t.Errorf("Failed to unregister resource %s: %v", resourceID, err)
		}
	}

	// Verify all resources are cleaned up
	finalStats := rm.GetStats()
	if finalStats[ResourceTypeBrowser].Active != 0 {
		t.Errorf("Expected 0 active resources after cleanup, got %d", finalStats[ResourceTypeBrowser].Active)
	}

	if finalStats[ResourceTypeBrowser].Cleaned != numResources {
		t.Errorf("Expected %d cleaned resources, got %d", numResources, finalStats[ResourceTypeBrowser].Cleaned)
	}
}

// TestCleanupManagerExecutionOrder tests that cleanup functions execute in priority order
func TestCleanupManagerExecutionOrder(t *testing.T) {
	config := DefaultCleanupConfig()
	config.MaxConcurrent = 1 // Force sequential execution
	cm := NewCleanupManager(config)

	var executionOrder []string
	var mutex sync.Mutex

	// Register cleanup functions with different priorities
	priorities := []int{1, 10, 5, 8, 3}
	expectedOrder := []string{"priority_10", "priority_8", "priority_5", "priority_3", "priority_1"}

	for _, priority := range priorities {
		name := fmt.Sprintf("priority_%d", priority)
		cm.RegisterCleanup(name, func(n string) func() error {
			return func() error {
				mutex.Lock()
				executionOrder = append(executionOrder, n)
				mutex.Unlock()
				return nil
			}
		}(name), priority)
	}

	// Execute cleanups
	metrics := cm.ExecuteCleanup()

	// Verify execution order
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d executions, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected execution order[%d] = %s, got %s", i, expected, executionOrder[i])
		}
	}

	// Verify metrics
	if metrics.TotalCleanups != len(priorities) {
		t.Errorf("Expected %d total cleanups, got %d", len(priorities), metrics.TotalCleanups)
	}

	if metrics.SuccessfulCleanups != len(priorities) {
		t.Errorf("Expected %d successful cleanups, got %d", len(priorities), metrics.SuccessfulCleanups)
	}
}

// TestResourceLeakDetectorAccuracy tests the accuracy of leak detection
func TestResourceLeakDetectorAccuracy(t *testing.T) {
	t.Skip("ResourceLeakDetector not implemented yet")
	/*
		// config := DefaultLeakDetectorConfig()
		// config.Enabled = true
		config.LeakThreshold = 100 * time.Millisecond
		config.ScanInterval = 50 * time.Millisecond
		rld := NewResourceLeakDetector(config)
		defer rld.Stop()

		// Create resources that will be properly cleaned up
		const numGoodResources = 5
		for i := 0; i < numGoodResources; i++ {
			resourceID := fmt.Sprintf("good_resource_%d", i)
			rld.TrackResource(resourceID, ResourceTypeBrowser, nil)

			// Clean up after a short delay
			go func(id string) {
				time.Sleep(30 * time.Millisecond)
				rld.UntrackResource(id)
			}(resourceID)
		}

		// Create resources that will leak
		const numLeakedResources = 3
		leakedResourceIDs := make([]string, numLeakedResources)
		for i := 0; i < numLeakedResources; i++ {
			resourceID := fmt.Sprintf("leaked_resource_%d", i)
			leakedResourceIDs[i] = resourceID
			rld.TrackResource(resourceID, ResourceTypeBrowser, nil)
			// Don't clean these up - they should be detected as leaks
		}

		// Wait for leak detection to occur
		time.Sleep(300 * time.Millisecond)

		// Check leak detection results
		leaks := rld.ScanForLeaks()
		metrics := rld.GetMetrics()

		t.Logf("Detected leaks: %v", leaks)
		t.Logf("Tracked resources: %d", metrics.ResourcesTracked)
		t.Logf("Total leaks detected: %d", metrics.LeaksDetected)

		// Should detect the leaked resources
		if metrics.LeaksDetected == 0 {
			t.Error("No leaks were detected, but some resources were intentionally leaked")
		}

		// Should have detected at least some of the leaked resources
		if len(leaks) == 0 {
			t.Error("ScanForLeaks returned no leaks, but leaks were expected")
		}

		// Clean up leaked resources
		for _, resourceID := range leakedResourceIDs {
			rld.UntrackResource(resourceID)
		}
	*/
}
