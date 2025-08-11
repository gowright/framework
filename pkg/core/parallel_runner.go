package core

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/gowright/framework/pkg/config"
)

// ParallelRunner manages concurrent test execution with resource management
type ParallelRunner struct {
	config         *config.Config
	maxConcurrency int
	semaphore      chan struct{}
	ctx            context.Context
	cancel         context.CancelFunc
	mutex          sync.RWMutex
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
func NewParallelRunner(cfg *config.Config, runnerConfig *ParallelRunnerConfig) *ParallelRunner {
	if runnerConfig == nil {
		runnerConfig = DefaultParallelRunnerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ParallelRunner{
		config:         cfg,
		maxConcurrency: runnerConfig.MaxConcurrency,
		semaphore:      make(chan struct{}, runnerConfig.MaxConcurrency),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// ExecuteTestsParallel executes tests in parallel with concurrency control
func (pr *ParallelRunner) ExecuteTestsParallel(tests []Test) (*TestResults, error) {
	if len(tests) == 0 {
		return &TestResults{
			SuiteName: "Parallel Test Suite",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			TestCases: make([]TestCaseResult, 0),
		}, nil
	}

	results := &TestResults{
		SuiteName: "Parallel Test Suite",
		StartTime: time.Now(),
		TestCases: make([]TestCaseResult, 0, len(tests)),
	}

	// Channel to collect results
	resultsChan := make(chan TestCaseResult, len(tests))
	var wg sync.WaitGroup

	// Execute tests in parallel
	for _, test := range tests {
		wg.Add(1)
		go func(t Test) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case pr.semaphore <- struct{}{}:
				defer func() { <-pr.semaphore }()
			case <-pr.ctx.Done():
				resultsChan <- TestCaseResult{
					Name:      t.GetName(),
					Status:    TestStatusError,
					Error:     NewGowrightError(ConfigurationError, "test cancelled", pr.ctx.Err()),
					StartTime: time.Now(),
					EndTime:   time.Now(),
				}
				return
			}

			// Execute test
			result := pr.executeTestWithTimeout(t)
			resultsChan <- *result
		}(test)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results.TestCases = append(results.TestCases, result)
	}

	results.EndTime = time.Now()
	pr.calculateSummary(results)

	return results, nil
}

// executeTestWithTimeout executes a test with timeout
func (pr *ParallelRunner) executeTestWithTimeout(test Test) *TestCaseResult {
	// Create a timeout context
	timeout := pr.config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute // Default timeout
	}

	ctx, cancel := context.WithTimeout(pr.ctx, timeout)
	defer cancel()

	// Channel to receive test result
	resultChan := make(chan *TestCaseResult, 1)

	// Execute test in goroutine
	go func() {
		result := test.Execute()
		select {
		case resultChan <- result:
		case <-ctx.Done():
			// Test execution was cancelled
		}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return &TestCaseResult{
			Name:      test.GetName(),
			Status:    TestStatusError,
			Error:     NewGowrightError(ConfigurationError, "test timed out", ctx.Err()),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
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

	if pr.cancel != nil {
		pr.cancel()
	}

	return nil
}

// GetMaxConcurrency returns the maximum concurrency level
func (pr *ParallelRunner) GetMaxConcurrency() int {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()
	return pr.maxConcurrency
}

// SetMaxConcurrency sets the maximum concurrency level
func (pr *ParallelRunner) SetMaxConcurrency(maxConcurrency int) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if maxConcurrency > 0 {
		pr.maxConcurrency = maxConcurrency
		// Recreate semaphore with new capacity
		pr.semaphore = make(chan struct{}, maxConcurrency)
	}
}

// GetActiveTests returns the number of currently running tests
func (pr *ParallelRunner) GetActiveTests() int {
	return len(pr.semaphore)
}

// IsShutdown returns whether the runner has been shut down
func (pr *ParallelRunner) IsShutdown() bool {
	select {
	case <-pr.ctx.Done():
		return true
	default:
		return false
	}
}
