package core

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCleanupManager(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	assert.NotNil(t, cm)
	assert.Equal(t, config, cm.config)
	assert.NotNil(t, cm.cleanupFuncs)
	assert.True(t, cm.active)
}

func TestCleanupManager_RegisterCleanup(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	executed := false
	cleanupFunc := func() error {
		executed = true
		return nil
	}

	cm.RegisterCleanup("test_cleanup", cleanupFunc, 10)

	// Verify cleanup function was registered
	assert.Len(t, cm.cleanupFuncs, 1)
	assert.Equal(t, "test_cleanup", cm.cleanupFuncs[0].Name)
	assert.Equal(t, 10, cm.cleanupFuncs[0].Priority)
	assert.Equal(t, config.DefaultTimeout, cm.cleanupFuncs[0].Timeout)

	// Execute cleanup to verify function works
	metrics := cm.ExecuteCleanup()
	assert.True(t, executed)
	assert.Equal(t, 1, metrics.TotalCleanups)
	assert.Equal(t, 1, metrics.SuccessfulCleanups)
	assert.Equal(t, 0, metrics.FailedCleanups)
}

func TestCleanupManager_RegisterCleanupWithTimeout(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	customTimeout := 5 * time.Second
	cleanupFunc := func() error { return nil }

	cm.RegisterCleanupWithTimeout("test_cleanup", cleanupFunc, 10, customTimeout)

	assert.Len(t, cm.cleanupFuncs, 1)
	assert.Equal(t, customTimeout, cm.cleanupFuncs[0].Timeout)
}

func TestCleanupManager_RegisterCleanup_PriorityOrder(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	// Register cleanups with different priorities
	cm.RegisterCleanup("low_priority", func() error { return nil }, 1)
	cm.RegisterCleanup("high_priority", func() error { return nil }, 10)
	cm.RegisterCleanup("medium_priority", func() error { return nil }, 5)

	// Verify they are ordered by priority (highest first)
	assert.Equal(t, "high_priority", cm.cleanupFuncs[0].Name)
	assert.Equal(t, "medium_priority", cm.cleanupFuncs[1].Name)
	assert.Equal(t, "low_priority", cm.cleanupFuncs[2].Name)
}

func TestCleanupManager_ExecuteCleanup_Sequential(t *testing.T) {
	config := DefaultCleanupConfig()
	config.MaxConcurrent = 1 // Force sequential execution
	cm := NewCleanupManager(config)

	var executionOrder []string
	var mutex sync.Mutex

	// Register multiple cleanup functions
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("cleanup_%d", i)
		cm.RegisterCleanup(name, func(n string) func() error {
			return func() error {
				mutex.Lock()
				executionOrder = append(executionOrder, n)
				mutex.Unlock()
				return nil
			}
		}(name), 10-i) // Decreasing priority
	}

	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 3, metrics.TotalCleanups)
	assert.Equal(t, 3, metrics.SuccessfulCleanups)
	assert.Equal(t, 0, metrics.FailedCleanups)
	assert.Len(t, executionOrder, 3)
}

func TestCleanupManager_ExecuteCleanup_Concurrent(t *testing.T) {
	config := DefaultCleanupConfig()
	config.MaxConcurrent = 3 // Allow concurrent execution
	cm := NewCleanupManager(config)

	var executedCount int32
	var mutex sync.Mutex

	// Register multiple cleanup functions
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("cleanup_%d", i)
		cm.RegisterCleanup(name, func() error {
			time.Sleep(10 * time.Millisecond) // Simulate work
			mutex.Lock()
			executedCount++
			mutex.Unlock()
			return nil
		}, 10)
	}

	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 5, metrics.TotalCleanups)
	assert.Equal(t, 5, metrics.SuccessfulCleanups)
	assert.Equal(t, 0, metrics.FailedCleanups)
	assert.Equal(t, int32(5), executedCount)
}

func TestCleanupManager_ExecuteCleanup_WithErrors(t *testing.T) {
	config := DefaultCleanupConfig()
	config.ContinueOnError = true
	cm := NewCleanupManager(config)

	// Register successful cleanup
	cm.RegisterCleanup("success", func() error { return nil }, 10)

	// Register failing cleanup
	cm.RegisterCleanup("failure", func() error {
		return errors.New("cleanup failed")
	}, 9)

	// Register another successful cleanup
	cm.RegisterCleanup("success2", func() error { return nil }, 8)

	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 3, metrics.TotalCleanups)
	assert.Equal(t, 2, metrics.SuccessfulCleanups)
	assert.Equal(t, 1, metrics.FailedCleanups)
	assert.Len(t, metrics.Errors, 1)
	assert.Contains(t, metrics.Errors[0], "failure")
}

func TestCleanupManager_ExecuteCleanup_StopOnError(t *testing.T) {
	config := DefaultCleanupConfig()
	config.ContinueOnError = false
	config.MaxConcurrent = 1 // Force sequential execution for stop-on-error behavior
	cm := NewCleanupManager(config)

	var executed []string
	var mutex sync.Mutex

	// Register cleanups in priority order
	cm.RegisterCleanup("first", func() error {
		mutex.Lock()
		executed = append(executed, "first")
		mutex.Unlock()
		return nil
	}, 10)

	cm.RegisterCleanup("failing", func() error {
		mutex.Lock()
		executed = append(executed, "failing")
		mutex.Unlock()
		return errors.New("cleanup failed")
	}, 9)

	cm.RegisterCleanup("should_not_execute", func() error {
		mutex.Lock()
		executed = append(executed, "should_not_execute")
		mutex.Unlock()
		return nil
	}, 8)

	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 3, metrics.TotalCleanups)
	assert.Equal(t, 1, metrics.SuccessfulCleanups)
	assert.Equal(t, 1, metrics.FailedCleanups)
	assert.Len(t, executed, 2) // Only first two should execute
	assert.Contains(t, executed, "first")
	assert.Contains(t, executed, "failing")
	assert.NotContains(t, executed, "should_not_execute")
}

func TestCleanupManager_ExecuteCleanup_Timeout(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	// Register cleanup that takes too long
	cm.RegisterCleanupWithTimeout("slow_cleanup", func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, 10, 50*time.Millisecond)

	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 1, metrics.TotalCleanups)
	assert.Equal(t, 0, metrics.SuccessfulCleanups)
	assert.Equal(t, 1, metrics.FailedCleanups)
	assert.Len(t, metrics.Errors, 1)
	assert.Contains(t, metrics.Errors[0], "timed out")
}

func TestCleanupManager_ExecuteCleanup_ForceGC(t *testing.T) {
	config := DefaultCleanupConfig()
	config.ForceGC = true
	cm := NewCleanupManager(config)

	cm.RegisterCleanup("test", func() error { return nil }, 10)

	// This test mainly verifies that ForceGC doesn't cause issues
	// We can't easily test that GC actually ran
	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 1, metrics.TotalCleanups)
	assert.Equal(t, 1, metrics.SuccessfulCleanups)
}

func TestCleanupManager_Shutdown(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	executed := false
	cm.RegisterCleanup("test", func() error {
		executed = true
		return nil
	}, 10)

	assert.True(t, cm.active)

	metrics := cm.Shutdown()

	assert.False(t, cm.active)
	assert.True(t, executed)
	assert.Equal(t, 1, metrics.TotalCleanups)
}

func TestCleanupManager_RegisterCleanup_AfterShutdown(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	// Shutdown first
	cm.Shutdown()

	// Try to register cleanup after shutdown
	cm.RegisterCleanup("after_shutdown", func() error { return nil }, 10)

	// Should not be registered
	assert.Len(t, cm.cleanupFuncs, 0)
}

func TestCleanupManager_SpecializedCleanups(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	var executed []string
	var mutex sync.Mutex

	// Test browser cleanup
	cm.RegisterCleanup("browser_1", func() error {
		mutex.Lock()
		executed = append(executed, "browser")
		mutex.Unlock()
		return nil
	}, 1)

	// Test database cleanup
	cm.RegisterCleanup("db_1", func() error {
		mutex.Lock()
		executed = append(executed, "database")
		mutex.Unlock()
		return nil
	}, 2)

	// Test HTTP client cleanup
	cm.RegisterCleanup("client_1", func() error {
		mutex.Lock()
		executed = append(executed, "http_client")
		mutex.Unlock()
		return nil
	}, 3)

	// Test file cleanup
	cm.RegisterCleanup("/tmp/test", func() error {
		mutex.Lock()
		executed = append(executed, "file")
		mutex.Unlock()
		return nil
	}, 4)

	// Test memory cleanup
	cm.RegisterCleanup("cache", func() error {
		mutex.Lock()
		executed = append(executed, "memory")
		mutex.Unlock()
		return nil
	}, 5)

	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 5, metrics.TotalCleanups)
	assert.Equal(t, 5, metrics.SuccessfulCleanups)
	assert.Len(t, executed, 5)

	// Verify all types were executed
	assert.Contains(t, executed, "browser")
	assert.Contains(t, executed, "database")
	assert.Contains(t, executed, "http_client")
	assert.Contains(t, executed, "file")
	assert.Contains(t, executed, "memory")
}

func TestCleanupManager_EmptyCleanup(t *testing.T) {
	config := DefaultCleanupConfig()
	cm := NewCleanupManager(config)

	// Execute cleanup with no registered functions
	metrics := cm.ExecuteCleanup()

	assert.Equal(t, 0, metrics.TotalCleanups)
	assert.Equal(t, 0, metrics.SuccessfulCleanups)
	assert.Equal(t, 0, metrics.FailedCleanups)
	assert.Empty(t, metrics.Errors)
	assert.False(t, metrics.LastCleanupTime.IsZero())
}

func TestDefaultCleanupConfig(t *testing.T) {
	config := DefaultCleanupConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.DefaultTimeout, time.Duration(0))
	assert.Greater(t, config.MaxConcurrent, 0)
	assert.True(t, config.ContinueOnError)
	assert.True(t, config.ForceGC)
	assert.Greater(t, config.GCInterval, time.Duration(0))
	assert.True(t, config.EnableMetrics)
}
