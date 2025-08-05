package gowright

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ParallelRunner manages concurrent test execution with resource management
type ParallelRunner struct {
	config          *Config
	maxConcurrency  int
	resourceManager *ResourceManager
	semaphore       chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
	mutex           sync.RWMutex
}

// ParallelRunnerConfig holds configuration for parallel test execution
type ParallelRunnerConfig struct {
	MaxConcurrency     int           `json:"max_concurrency"`
	ResourceTimeout    time.Duration `json:"resource_timeout"`
	BrowserPoolSize    int           `json:"browser_pool_size"`
	DatabasePoolSize   int           `json:"database_pool_size"`
	HTTPClientPoolSize int           `json:"http_client_pool_size"`
	GracefulShutdown   time.Duration `json:"graceful_shutdown"`
}

// DefaultParallelRunnerConfig returns default configuration for parallel runner
func DefaultParallelRunnerConfig() *ParallelRunnerConfig {
	return &ParallelRunnerConfig{
		MaxConcurrency:     runtime.NumCPU(),
		ResourceTimeout:    30 * time.Second,
		BrowserPoolSize:    5,
		DatabasePoolSize:   10,
		HTTPClientPoolSize: 20,
		GracefulShutdown:   10 * time.Second,
	}
}

// NewParallelRunner creates a new parallel test runner
func NewParallelRunner(config *Config, runnerConfig *ParallelRunnerConfig) *ParallelRunner {
	if runnerConfig == nil {
		runnerConfig = DefaultParallelRunnerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ParallelRunner{
		config:          config,
		maxConcurrency:  runnerConfig.MaxConcurrency,
		resourceManager: NewResourceManager(runnerConfig),
		semaphore:       make(chan struct{}, runnerConfig.MaxConcurrency),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// ExecuteTestsParallel executes tests in parallel with resource management
func (pr *ParallelRunner) ExecuteTestsParallel(tests []Test) (*TestResults, error) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if len(tests) == 0 {
		return &TestResults{
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}, nil
	}

	// Initialize resource manager
	if err := pr.resourceManager.Initialize(pr.ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize resource manager: %w", err)
	}

	defer func() {
		if err := pr.resourceManager.Cleanup(); err != nil {
			fmt.Printf("Warning: resource manager cleanup failed: %v\n", err)
		}
	}()

	results := &TestResults{
		StartTime: time.Now(),
		TestCases: make([]TestCaseResult, 0, len(tests)),
	}

	// Create channels for communication
	resultsChan := make(chan *TestCaseResult, len(tests))
	errorsChan := make(chan error, len(tests))

	var wg sync.WaitGroup

	// Execute tests concurrently
	for _, test := range tests {
		wg.Add(1)
		go func(t Test) {
			defer wg.Done()
			pr.executeTestWithResourceManagement(t, resultsChan, errorsChan)
		}(test)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// Collect results
	var executionErrors []error
	for {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				results.TestCases = append(results.TestCases, *result)
			}
		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
			} else if err != nil {
				executionErrors = append(executionErrors, err)
			}
		}

		if resultsChan == nil && errorsChan == nil {
			break
		}
	}

	results.EndTime = time.Now()
	pr.calculateSummary(results)

	if len(executionErrors) > 0 {
		return results, fmt.Errorf("parallel execution errors: %v", executionErrors)
	}

	return results, nil
}

// executeTestWithResourceManagement executes a single test with proper resource management
func (pr *ParallelRunner) executeTestWithResourceManagement(test Test, resultsChan chan<- *TestCaseResult, errorsChan chan<- error) {
	// Acquire semaphore to limit concurrency
	select {
	case pr.semaphore <- struct{}{}:
		defer func() { <-pr.semaphore }()
	case <-pr.ctx.Done():
		errorsChan <- fmt.Errorf("test execution cancelled: %w", pr.ctx.Err())
		return
	}

	// Get resources for the test
	resources, err := pr.resourceManager.AcquireResources(pr.ctx, test)
	if err != nil {
		result := &TestCaseResult{
			Name:      test.GetName(),
			Status:    TestStatusError,
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Error:     fmt.Errorf("failed to acquire resources: %w", err),
		}
		resultsChan <- result
		return
	}

	defer func() {
		if err := pr.resourceManager.ReleaseResources(resources); err != nil {
			errorsChan <- fmt.Errorf("failed to release resources for test %s: %w", test.GetName(), err)
		}
	}()

	// Execute the test with timeout
	ctx, cancel := context.WithTimeout(pr.ctx, pr.config.BrowserConfig.Timeout)
	defer cancel()

	resultChan := make(chan *TestCaseResult, 1)

	go func() {
		result := test.Execute()
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		resultsChan <- result
	case <-ctx.Done():
		result := &TestCaseResult{
			Name:      test.GetName(),
			Status:    TestStatusError,
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Error:     fmt.Errorf("test execution timeout: %w", ctx.Err()),
		}
		resultsChan <- result
	}
}

// calculateSummary calculates test execution summary
func (pr *ParallelRunner) calculateSummary(results *TestResults) {
	results.TotalTests = len(results.TestCases)

	for _, testCase := range results.TestCases {
		switch testCase.Status {
		case TestStatusPassed:
			results.PassedTests++
		case TestStatusFailed:
			results.FailedTests++
		case TestStatusSkipped:
			results.SkippedTests++
		case TestStatusError:
			results.ErrorTests++
		}
	}
}

// Shutdown gracefully shuts down the parallel runner
func (pr *ParallelRunner) Shutdown() error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	// Cancel context to stop new executions
	pr.cancel()

	// Wait for ongoing executions to complete with timeout
	done := make(chan struct{})
	go func() {
		// Wait for all semaphore slots to be available (all tests completed)
		for i := 0; i < pr.maxConcurrency; i++ {
			pr.semaphore <- struct{}{}
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(pr.resourceManager.config.GracefulShutdown):
		return fmt.Errorf("graceful shutdown timeout exceeded")
	}
}

// GetStats returns current execution statistics
func (pr *ParallelRunner) GetStats() *ParallelRunnerStats {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	return &ParallelRunnerStats{
		MaxConcurrency: pr.maxConcurrency,
		ActiveTests:    pr.maxConcurrency - len(pr.semaphore),
		ResourceStats:  pr.resourceManager.GetStats(),
	}
}

// ParallelRunnerStats holds statistics about parallel execution
type ParallelRunnerStats struct {
	MaxConcurrency int            `json:"max_concurrency"`
	ActiveTests    int            `json:"active_tests"`
	ResourceStats  *ResourceStats `json:"resource_stats"`
}
