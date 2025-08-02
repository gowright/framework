package gowright

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ParallelMockTest implements the Test interface for parallel testing
type ParallelMockTest struct {
	mock.Mock
	name     string
	duration time.Duration
	status   TestStatus
	err      error
}

func (mt *ParallelMockTest) GetName() string {
	return mt.name
}

func (mt *ParallelMockTest) Execute() *TestCaseResult {
	mt.Called()
	
	// Simulate test execution time
	if mt.duration > 0 {
		time.Sleep(mt.duration)
	}
	
	return &TestCaseResult{
		Name:      mt.name,
		Status:    mt.status,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(mt.duration),
		Duration:  mt.duration,
		Error:     mt.err,
	}
}

func TestNewParallelRunner(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := DefaultParallelRunnerConfig()
	
	runner := NewParallelRunner(config, runnerConfig)
	
	assert.NotNil(t, runner)
	assert.Equal(t, runnerConfig.MaxConcurrency, runner.maxConcurrency)
	assert.NotNil(t, runner.resourceManager)
	assert.NotNil(t, runner.semaphore)
	assert.Equal(t, runnerConfig.MaxConcurrency, cap(runner.semaphore))
}

func TestNewParallelRunnerWithNilConfig(t *testing.T) {
	config := DefaultConfig()
	
	runner := NewParallelRunner(config, nil)
	
	assert.NotNil(t, runner)
	assert.NotNil(t, runner.resourceManager)
}

func TestParallelRunner_ExecuteTestsParallel_EmptyTests(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := DefaultParallelRunnerConfig()
	runner := NewParallelRunner(config, runnerConfig)
	
	results, err := runner.ExecuteTestsParallel([]Test{})
	
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, results.TotalTests)
	assert.Equal(t, 0, len(results.TestCases))
}

func TestParallelRunner_ExecuteTestsParallel_SingleTest(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := &ParallelRunnerConfig{
		MaxConcurrency:     1,
		ResourceTimeout:    5 * time.Second,
		BrowserPoolSize:    1,
		DatabasePoolSize:   1,
		HTTPClientPoolSize: 1,
		GracefulShutdown:   5 * time.Second,
	}
	runner := NewParallelRunner(config, runnerConfig)
	
	mockTest := &ParallelMockTest{
		name:     "test1",
		duration: 100 * time.Millisecond,
		status:   TestStatusPassed,
	}
	mockTest.On("Execute").Return()
	
	tests := []Test{mockTest}
	results, err := runner.ExecuteTestsParallel(tests)
	
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 1, results.TotalTests)
	assert.Equal(t, 1, results.PassedTests)
	assert.Equal(t, 0, results.FailedTests)
	assert.Equal(t, 1, len(results.TestCases))
	assert.Equal(t, "test1", results.TestCases[0].Name)
	assert.Equal(t, TestStatusPassed, results.TestCases[0].Status)
	
	mockTest.AssertExpectations(t)
}

func TestParallelRunner_ExecuteTestsParallel_MultipleTests(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := &ParallelRunnerConfig{
		MaxConcurrency:     3,
		ResourceTimeout:    5 * time.Second,
		BrowserPoolSize:    3,
		DatabasePoolSize:   3,
		HTTPClientPoolSize: 3,
		GracefulShutdown:   5 * time.Second,
	}
	runner := NewParallelRunner(config, runnerConfig)
	
	// Create multiple mock tests
	tests := make([]Test, 5)
	for i := 0; i < 5; i++ {
		mockTest := &ParallelMockTest{
			name:     fmt.Sprintf("test%d", i+1),
			duration: 100 * time.Millisecond,
			status:   TestStatusPassed,
		}
		mockTest.On("Execute").Return()
		tests[i] = mockTest
	}
	
	startTime := time.Now()
	results, err := runner.ExecuteTestsParallel(tests)
	executionTime := time.Since(startTime)
	
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 5, results.TotalTests)
	assert.Equal(t, 5, results.PassedTests)
	assert.Equal(t, 0, results.FailedTests)
	assert.Equal(t, 5, len(results.TestCases))
	
	// Verify parallel execution - should be faster than sequential
	// With 3 concurrent workers and 5 tests of 100ms each, it should take around 200ms
	assert.Less(t, executionTime, 400*time.Millisecond, "Parallel execution should be faster than sequential")
	
	// Verify all mock expectations
	for _, test := range tests {
		mockTest := test.(*ParallelMockTest)
		mockTest.AssertExpectations(t)
	}
}

func TestParallelRunner_ExecuteTestsParallel_ConcurrencyLimit(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := &ParallelRunnerConfig{
		MaxConcurrency:     2, // Limit to 2 concurrent tests
		ResourceTimeout:    5 * time.Second,
		BrowserPoolSize:    2,
		DatabasePoolSize:   2,
		HTTPClientPoolSize: 2,
		GracefulShutdown:   5 * time.Second,
	}
	runner := NewParallelRunner(config, runnerConfig)
	
	// Track concurrent executions
	var concurrentCount int32
	var maxConcurrent int32
	
	// Create tests that track concurrency
	tests := make([]Test, 4)
	for i := 0; i < 4; i++ {
		mockTest := &ParallelMockTest{
			name:     fmt.Sprintf("test%d", i+1),
			duration: 200 * time.Millisecond,
			status:   TestStatusPassed,
		}
		
		mockTest.On("Execute").Return().Run(func(args mock.Arguments) {
			current := atomic.AddInt32(&concurrentCount, 1)
			
			// Update max concurrent safely
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			}
			
			// Simulate work
			time.Sleep(200 * time.Millisecond)
			
			atomic.AddInt32(&concurrentCount, -1)
		})
		
		tests[i] = mockTest
	}
	
	results, err := runner.ExecuteTestsParallel(tests)
	
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 4, results.TotalTests)
	assert.LessOrEqual(t, maxConcurrent, int32(2), "Should not exceed max concurrency limit")
	
	// Verify all mock expectations
	for _, test := range tests {
		mockTest := test.(*ParallelMockTest)
		mockTest.AssertExpectations(t)
	}
}

func TestParallelRunner_ExecuteTestsParallel_MixedResults(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := DefaultParallelRunnerConfig()
	runner := NewParallelRunner(config, runnerConfig)
	
	// Create tests with different outcomes
	tests := []Test{
		&ParallelMockTest{name: "passed_test", status: TestStatusPassed},
		&ParallelMockTest{name: "failed_test", status: TestStatusFailed},
		&ParallelMockTest{name: "error_test", status: TestStatusError},
		&ParallelMockTest{name: "skipped_test", status: TestStatusSkipped},
	}
	
	for _, test := range tests {
		mockTest := test.(*ParallelMockTest)
		mockTest.On("Execute").Return()
	}
	
	results, err := runner.ExecuteTestsParallel(tests)
	
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 4, results.TotalTests)
	assert.Equal(t, 1, results.PassedTests)
	assert.Equal(t, 1, results.FailedTests)
	assert.Equal(t, 1, results.ErrorTests)
	assert.Equal(t, 1, results.SkippedTests)
	
	// Verify all mock expectations
	for _, test := range tests {
		mockTest := test.(*ParallelMockTest)
		mockTest.AssertExpectations(t)
	}
}

func TestParallelRunner_Shutdown(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := &ParallelRunnerConfig{
		MaxConcurrency:   2,
		GracefulShutdown: 1 * time.Second,
	}
	runner := NewParallelRunner(config, runnerConfig)
	
	// Start some long-running tests
	tests := []Test{
		&ParallelMockTest{name: "long_test1", duration: 2 * time.Second, status: TestStatusPassed},
		&ParallelMockTest{name: "long_test2", duration: 2 * time.Second, status: TestStatusPassed},
	}
	
	for _, test := range tests {
		mockTest := test.(*ParallelMockTest)
		mockTest.On("Execute").Return()
	}
	
	// Start execution in background
	go func() {
		runner.ExecuteTestsParallel(tests)
	}()
	
	// Give tests time to start
	time.Sleep(100 * time.Millisecond)
	
	// Shutdown should complete within reasonable time
	startTime := time.Now()
	err := runner.Shutdown()
	shutdownTime := time.Since(startTime)
	
	assert.NoError(t, err)
	assert.Less(t, shutdownTime, 2*time.Second, "Shutdown should complete quickly")
}

func TestParallelRunner_GetStats(t *testing.T) {
	config := DefaultConfig()
	runnerConfig := DefaultParallelRunnerConfig()
	runner := NewParallelRunner(config, runnerConfig)
	
	stats := runner.GetStats()
	
	assert.NotNil(t, stats)
	assert.Equal(t, runnerConfig.MaxConcurrency, stats.MaxConcurrency)
	assert.Equal(t, runnerConfig.MaxConcurrency, stats.ActiveTests) // All slots available initially
	assert.NotNil(t, stats.ResourceStats)
}

func TestDefaultParallelRunnerConfig(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	
	assert.NotNil(t, config)
	assert.Greater(t, config.MaxConcurrency, 0)
	assert.Greater(t, config.ResourceTimeout, time.Duration(0))
	assert.Greater(t, config.BrowserPoolSize, 0)
	assert.Greater(t, config.DatabasePoolSize, 0)
	assert.Greater(t, config.HTTPClientPoolSize, 0)
	assert.Greater(t, config.GracefulShutdown, time.Duration(0))
}

// Benchmark tests
func BenchmarkParallelRunner_ExecuteTestsParallel(b *testing.B) {
	config := DefaultConfig()
	runnerConfig := DefaultParallelRunnerConfig()
	runner := NewParallelRunner(config, runnerConfig)
	
	// Create a set of mock tests
	tests := make([]Test, 10)
	for i := 0; i < 10; i++ {
		mockTest := &ParallelMockTest{
			name:     fmt.Sprintf("benchmark_test_%d", i),
			duration: 10 * time.Millisecond,
			status:   TestStatusPassed,
		}
		mockTest.On("Execute").Return()
		tests[i] = mockTest
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Reset mock expectations for each iteration
		for _, test := range tests {
			mockTest := test.(*ParallelMockTest)
			mockTest.ExpectedCalls = nil
			mockTest.On("Execute").Return()
		}
		
		_, err := runner.ExecuteTestsParallel(tests)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}