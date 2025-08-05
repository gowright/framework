package gowright

import (
	"fmt"
	"sync"
	"time"
)

// TestSuiteManager manages test suite execution and orchestration
type TestSuiteManager struct {
	suite   *TestSuite
	config  *Config
	results *TestResults
	mutex   sync.RWMutex
}

// NewTestSuiteManager creates a new test suite manager
func NewTestSuiteManager(suite *TestSuite, config *Config) *TestSuiteManager {
	return &TestSuiteManager{
		suite:  suite,
		config: config,
		results: &TestResults{
			SuiteName: suite.Name,
			TestCases: make([]TestCaseResult, 0),
		},
	}
}

// RegisterTest registers a test with the test suite
func (tsm *TestSuiteManager) RegisterTest(test Test) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	if tsm.suite.Tests == nil {
		tsm.suite.Tests = make([]Test, 0)
	}

	tsm.suite.Tests = append(tsm.suite.Tests, test)
}

// RegisterTests registers multiple tests with the test suite
func (tsm *TestSuiteManager) RegisterTests(tests []Test) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	if tsm.suite.Tests == nil {
		tsm.suite.Tests = make([]Test, 0)
	}

	tsm.suite.Tests = append(tsm.suite.Tests, tests...)
}

// ExecuteTestSuite executes the complete test suite with setup and teardown
func (tsm *TestSuiteManager) ExecuteTestSuite() (*TestResults, error) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	tsm.results.StartTime = time.Now()

	// Execute setup
	if tsm.suite.SetupFunc != nil {
		if err := tsm.suite.SetupFunc(); err != nil {
			return nil, NewGowrightError(ConfigurationError, "test suite setup failed", err)
		}
	}

	// Ensure teardown is called even if tests fail
	defer func() {
		if tsm.suite.TeardownFunc != nil {
			if err := tsm.suite.TeardownFunc(); err != nil {
				// Log teardown error but don't fail the suite
				fmt.Printf("Warning: test suite teardown failed: %v\n", err)
			}
		}
	}()

	// Execute tests
	if tsm.config.Parallel {
		tsm.executeTestsParallel()
	} else {
		tsm.executeTestsSequential()
	}

	tsm.results.EndTime = time.Now()
	tsm.calculateSummary()

	return tsm.results, nil
}

// executeTestsSequential executes tests one by one
func (tsm *TestSuiteManager) executeTestsSequential() {
	for _, test := range tsm.suite.Tests {
		result := tsm.executeTest(test)
		tsm.results.TestCases = append(tsm.results.TestCases, *result)
	}
}

// executeTestsParallel executes tests in parallel using the enhanced parallel runner
func (tsm *TestSuiteManager) executeTestsParallel() {
	// Create parallel runner configuration
	runnerConfig := DefaultParallelRunnerConfig()

	// Create parallel runner
	parallelRunner := NewParallelRunner(tsm.config, runnerConfig)
	defer func() {
		if err := parallelRunner.Shutdown(); err != nil {
			fmt.Printf("Warning: parallel runner shutdown failed: %v\n", err)
		}
	}()

	// Execute tests using the parallel runner
	results, err := parallelRunner.ExecuteTestsParallel(tsm.suite.Tests)
	if err != nil {
		fmt.Printf("Warning: parallel execution encountered errors: %v\n", err)
	}

	// Merge results
	if results != nil {
		tsm.results.TestCases = append(tsm.results.TestCases, results.TestCases...)
	}
}

// executeTest executes a single test with error handling and retry logic
func (tsm *TestSuiteManager) executeTest(test Test) *TestCaseResult {
	result := &TestCaseResult{
		Name:      test.GetName(),
		StartTime: time.Now(),
	}

	var lastError error
	var lastStatus TestStatus

	// Retry logic - only retry on errors, not on failures
	for attempt := 0; attempt <= tsm.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Add delay between retries
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		testResult := test.Execute()
		lastStatus = testResult.Status
		lastError = testResult.Error

		// If test passed or failed (but not error), don't retry
		if testResult.Status == TestStatusPassed || testResult.Status == TestStatusFailed || testResult.Status == TestStatusSkipped {
			result.Status = testResult.Status
			result.Duration = testResult.Duration
			result.Screenshots = testResult.Screenshots
			result.Logs = testResult.Logs
			result.Error = testResult.Error
			result.EndTime = time.Now()
			return result
		}

		// Only retry on errors
		if testResult.Status == TestStatusError && testResult.Error != nil {
			continue
		}

		// If no error but status is error, treat as final result
		result.Status = testResult.Status
		result.Duration = testResult.Duration
		result.Screenshots = testResult.Screenshots
		result.Logs = testResult.Logs
		result.Error = testResult.Error
		result.EndTime = time.Now()
		return result
	}

	// All retries failed
	result.Status = lastStatus
	result.Error = lastError
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// calculateSummary calculates test execution summary
func (tsm *TestSuiteManager) calculateSummary() {
	tsm.results.TotalTests = len(tsm.results.TestCases)

	for _, testCase := range tsm.results.TestCases {
		switch testCase.Status {
		case TestStatusPassed:
			tsm.results.PassedTests++
		case TestStatusFailed:
			tsm.results.FailedTests++
		case TestStatusSkipped:
			tsm.results.SkippedTests++
		case TestStatusError:
			tsm.results.ErrorTests++
		}
	}
}

// GetTestCount returns the number of registered tests
func (tsm *TestSuiteManager) GetTestCount() int {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()

	if tsm.suite.Tests == nil {
		return 0
	}

	return len(tsm.suite.Tests)
}

// GetTestSuite returns the test suite
func (tsm *TestSuiteManager) GetTestSuite() *TestSuite {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()
	return tsm.suite
}

// GetResults returns the current test results
func (tsm *TestSuiteManager) GetResults() *TestResults {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()
	return tsm.results
}

// ClearTests removes all registered tests
func (tsm *TestSuiteManager) ClearTests() {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	tsm.suite.Tests = make([]Test, 0)
}

// SetSetupFunc sets the setup function for the test suite
func (tsm *TestSuiteManager) SetSetupFunc(setupFunc func() error) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	tsm.suite.SetupFunc = setupFunc
}

// SetTeardownFunc sets the teardown function for the test suite
func (tsm *TestSuiteManager) SetTeardownFunc(teardownFunc func() error) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	tsm.suite.TeardownFunc = teardownFunc
}

// ExecuteTestByName executes a specific test by name
func (tsm *TestSuiteManager) ExecuteTestByName(testName string) (*TestCaseResult, error) {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()

	for _, test := range tsm.suite.Tests {
		if test.GetName() == testName {
			return tsm.executeTest(test), nil
		}
	}

	return nil, fmt.Errorf("test '%s' not found in suite", testName)
}

// GetTestNames returns the names of all registered tests
func (tsm *TestSuiteManager) GetTestNames() []string {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()

	names := make([]string, 0, len(tsm.suite.Tests))
	for _, test := range tsm.suite.Tests {
		names = append(names, test.GetName())
	}

	return names
}
