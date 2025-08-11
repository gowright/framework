package core

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// CleanupManager handles cleanup operations for test resources
type CleanupManager struct {
	cleanupFuncs []CleanupFunc
	mutex        sync.RWMutex
	config       *CleanupConfig
	active       bool
}

// CleanupFunc represents a cleanup function
type CleanupFunc struct {
	Name     string
	Function func() error
	Priority int // Higher priority runs first
	Timeout  time.Duration
}

// CleanupConfig holds configuration for cleanup operations
type CleanupConfig struct {
	DefaultTimeout  time.Duration `json:"default_timeout"`
	MaxConcurrent   int           `json:"max_concurrent"`
	ContinueOnError bool          `json:"continue_on_error"`
	ForceGC         bool          `json:"force_gc"`
	GCInterval      time.Duration `json:"gc_interval"`
	EnableMetrics   bool          `json:"enable_metrics"`
}

// CleanupMetrics holds metrics about cleanup operations
type CleanupMetrics struct {
	TotalCleanups      int           `json:"total_cleanups"`
	SuccessfulCleanups int           `json:"successful_cleanups"`
	FailedCleanups     int           `json:"failed_cleanups"`
	AverageTime        time.Duration `json:"average_time"`
	LastCleanupTime    time.Time     `json:"last_cleanup_time"`
	Errors             []string      `json:"errors,omitempty"`
}

// DefaultCleanupConfig returns default cleanup configuration
func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		DefaultTimeout:  30 * time.Second,
		MaxConcurrent:   5,
		ContinueOnError: true,
		ForceGC:         true,
		GCInterval:      1 * time.Minute,
		EnableMetrics:   true,
	}
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(config *CleanupConfig) *CleanupManager {
	if config == nil {
		config = DefaultCleanupConfig()
	}

	return &CleanupManager{
		cleanupFuncs: make([]CleanupFunc, 0),
		config:       config,
		active:       true,
	}
}

// RegisterCleanup registers a cleanup function
func (cm *CleanupManager) RegisterCleanup(name string, fn func() error, priority int) {
	cm.RegisterCleanupWithTimeout(name, fn, priority, cm.config.DefaultTimeout)
}

// RegisterCleanupWithTimeout registers a cleanup function with a specific timeout
func (cm *CleanupManager) RegisterCleanupWithTimeout(name string, fn func() error, priority int, timeout time.Duration) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.active {
		return // Don't register new cleanups if manager is shutting down
	}

	cleanupFunc := CleanupFunc{
		Name:     name,
		Function: fn,
		Priority: priority,
		Timeout:  timeout,
	}

	// Insert in priority order (higher priority first)
	inserted := false
	for i, existing := range cm.cleanupFuncs {
		if priority > existing.Priority {
			// Insert at this position
			cm.cleanupFuncs = append(cm.cleanupFuncs[:i], append([]CleanupFunc{cleanupFunc}, cm.cleanupFuncs[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		// Append at the end
		cm.cleanupFuncs = append(cm.cleanupFuncs, cleanupFunc)
	}
}

// ExecuteCleanup executes all registered cleanup functions
func (cm *CleanupManager) ExecuteCleanup() *CleanupMetrics {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	metrics := &CleanupMetrics{
		LastCleanupTime: time.Now(),
		Errors:          make([]string, 0),
	}

	if len(cm.cleanupFuncs) == 0 {
		return metrics
	}

	startTime := time.Now()

	// Execute cleanup functions based on concurrency settings
	if cm.config.MaxConcurrent <= 1 {
		// Sequential execution
		cm.executeSequential(metrics)
	} else {
		// Concurrent execution
		cm.executeConcurrent(metrics)
	}

	// Calculate metrics
	metrics.TotalCleanups = len(cm.cleanupFuncs)
	metrics.AverageTime = time.Since(startTime) / time.Duration(len(cm.cleanupFuncs))

	// Force garbage collection if enabled
	if cm.config.ForceGC {
		runtime.GC()
	}

	return metrics
}

// executeSequential executes cleanup functions sequentially
func (cm *CleanupManager) executeSequential(metrics *CleanupMetrics) {
	for _, cleanupFunc := range cm.cleanupFuncs {
		if err := cm.executeCleanupFunc(cleanupFunc); err != nil {
			metrics.FailedCleanups++
			metrics.Errors = append(metrics.Errors, fmt.Sprintf("%s: %v", cleanupFunc.Name, err))
			if !cm.config.ContinueOnError {
				break
			}
		} else {
			metrics.SuccessfulCleanups++
		}
	}
}

// executeConcurrent executes cleanup functions concurrently
func (cm *CleanupManager) executeConcurrent(metrics *CleanupMetrics) {
	semaphore := make(chan struct{}, cm.config.MaxConcurrent)
	var wg sync.WaitGroup
	var resultMutex sync.Mutex

	for _, cleanupFunc := range cm.cleanupFuncs {
		wg.Add(1)
		go func(cf CleanupFunc) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := cm.executeCleanupFunc(cf); err != nil {
				resultMutex.Lock()
				metrics.FailedCleanups++
				metrics.Errors = append(metrics.Errors, fmt.Sprintf("%s: %v", cf.Name, err))
				resultMutex.Unlock()
			} else {
				resultMutex.Lock()
				metrics.SuccessfulCleanups++
				resultMutex.Unlock()
			}
		}(cleanupFunc)
	}

	wg.Wait()
}

// executeCleanupFunc executes a single cleanup function with timeout
func (cm *CleanupManager) executeCleanupFunc(cleanupFunc CleanupFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupFunc.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- cleanupFunc.Function()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return NewGowrightError(ConfigurationError,
			fmt.Sprintf("cleanup function '%s' timed out after %v", cleanupFunc.Name, cleanupFunc.Timeout),
			ctx.Err())
	}
}

// Shutdown shuts down the cleanup manager and executes all cleanup functions
func (cm *CleanupManager) Shutdown() *CleanupMetrics {
	cm.mutex.Lock()
	cm.active = false
	cm.mutex.Unlock()

	return cm.ExecuteCleanup()
}
