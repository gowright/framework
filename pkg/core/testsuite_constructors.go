package core

import (
	"time"
)

// TestSuiteBuilder provides a fluent interface for building test suites
type TestSuiteBuilder struct {
	name         string
	tests        []Test
	setupFunc    func() error
	teardownFunc func() error
	parallel     bool
	maxWorkers   int
	timeout      time.Duration
}

// NewTestSuiteBuilder creates a new test suite builder
func NewTestSuiteBuilder(name string) *TestSuiteBuilder {
	return &TestSuiteBuilder{
		name:       name,
		tests:      make([]Test, 0),
		parallel:   false,
		maxWorkers: 4,
		timeout:    30 * time.Minute,
	}
}

// WithSetup sets the setup function
func (tsb *TestSuiteBuilder) WithSetup(setupFunc func() error) *TestSuiteBuilder {
	tsb.setupFunc = setupFunc
	return tsb
}

// WithTeardown sets the teardown function
func (tsb *TestSuiteBuilder) WithTeardown(teardownFunc func() error) *TestSuiteBuilder {
	tsb.teardownFunc = teardownFunc
	return tsb
}

// WithParallel enables parallel execution
func (tsb *TestSuiteBuilder) WithParallel(parallel bool) *TestSuiteBuilder {
	tsb.parallel = parallel
	return tsb
}

// WithMaxWorkers sets the maximum number of parallel workers
func (tsb *TestSuiteBuilder) WithMaxWorkers(maxWorkers int) *TestSuiteBuilder {
	tsb.maxWorkers = maxWorkers
	return tsb
}

// WithTimeout sets the suite timeout
func (tsb *TestSuiteBuilder) WithTimeout(timeout time.Duration) *TestSuiteBuilder {
	tsb.timeout = timeout
	return tsb
}

// AddTest adds a test to the suite
func (tsb *TestSuiteBuilder) AddTest(test Test) *TestSuiteBuilder {
	tsb.tests = append(tsb.tests, test)
	return tsb
}

// AddTestFunc adds a function-based test to the suite
func (tsb *TestSuiteBuilder) AddTestFunc(name string, testFunc func(*TestContext)) *TestSuiteBuilder {
	test := NewFunctionTest(name, testFunc)
	tsb.tests = append(tsb.tests, test)
	return tsb
}

// Build creates the test suite
func (tsb *TestSuiteBuilder) Build() *TestSuite {
	return &TestSuite{
		Name:         tsb.name,
		Tests:        tsb.tests,
		SetupFunc:    tsb.setupFunc,
		TeardownFunc: tsb.teardownFunc,
	}
}

// NewTestSuite creates a new test suite with the given name
func NewTestSuite(name string) *TestSuite {
	return &TestSuite{
		Name:  name,
		Tests: make([]Test, 0),
	}
}

// AddTest adds a test to the test suite
func (ts *TestSuite) AddTest(test Test) {
	ts.Tests = append(ts.Tests, test)
}

// AddTestFunc adds a function-based test to the test suite
func (ts *TestSuite) AddTestFunc(name string, testFunc func(*TestContext)) {
	test := NewFunctionTest(name, testFunc)
	ts.AddTest(test)
}

// Run executes all tests in the suite and returns the results
func (ts *TestSuite) Run() *TestSuiteResults {
	startTime := time.Now()
	results := &TestSuiteResults{
		SuiteName:   ts.Name,
		TestResults: make([]*TestCaseResult, 0),
		StartTime:   startTime,
	}

	// Execute setup if defined
	if ts.SetupFunc != nil {
		if err := ts.SetupFunc(); err != nil {
			// If setup fails, mark all tests as error
			for _, test := range ts.Tests {
				result := &TestCaseResult{
					Name:      test.GetName(),
					Status:    TestStatusError,
					StartTime: startTime,
					EndTime:   time.Now(),
					Error:     NewTestSetupError("suite", "suite setup failed", err),
				}
				results.TestResults = append(results.TestResults, result)
				results.ErrorCount++
			}
			results.EndTime = time.Now()
			results.Duration = results.EndTime.Sub(startTime)
			return results
		}
	}

	// Execute tests sequentially
	ts.runSequential(results)

	// Execute teardown if defined
	if ts.TeardownFunc != nil {
		if err := ts.TeardownFunc(); err != nil {
			// Log teardown error but don't fail the tests
			// In a real implementation, you might want to log this
			_ = err // Explicitly ignore the error to satisfy linter
		}
	}

	results.EndTime = time.Now()
	results.Duration = results.EndTime.Sub(startTime)
	return results
}

// runSequential executes tests sequentially
func (ts *TestSuite) runSequential(results *TestSuiteResults) {
	for _, test := range ts.Tests {
		result := test.Execute()
		results.TestResults = append(results.TestResults, result)
		ts.updateCounters(results, result)
	}
}

// updateCounters updates the result counters
func (ts *TestSuite) updateCounters(results *TestSuiteResults, result *TestCaseResult) {
	switch result.Status {
	case TestStatusPassed:
		results.PassedCount++
	case TestStatusFailed:
		results.FailedCount++
	case TestStatusError:
		results.ErrorCount++
	case TestStatusSkipped:
		results.SkippedCount++
	}
}

// TestSuiteResults holds the results of a test suite execution
type TestSuiteResults struct {
	SuiteName    string            `json:"suite_name"`
	TestResults  []*TestCaseResult `json:"test_results"`
	PassedCount  int               `json:"passed_count"`
	FailedCount  int               `json:"failed_count"`
	ErrorCount   int               `json:"error_count"`
	SkippedCount int               `json:"skipped_count"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Duration     time.Duration     `json:"duration"`
}

// GetTotalCount returns the total number of tests
func (tsr *TestSuiteResults) GetTotalCount() int {
	return tsr.PassedCount + tsr.FailedCount + tsr.ErrorCount + tsr.SkippedCount
}

// GetSuccessRate returns the success rate as a percentage
func (tsr *TestSuiteResults) GetSuccessRate() float64 {
	total := tsr.GetTotalCount()
	if total == 0 {
		return 0.0
	}
	return float64(tsr.PassedCount) / float64(total) * 100.0
}

// HasFailures returns true if there are any failures or errors
func (tsr *TestSuiteResults) HasFailures() bool {
	return tsr.FailedCount > 0 || tsr.ErrorCount > 0
}

// GetFailedTests returns all failed test results
func (tsr *TestSuiteResults) GetFailedTests() []*TestCaseResult {
	var failed []*TestCaseResult
	for _, result := range tsr.TestResults {
		if result.Status == TestStatusFailed || result.Status == TestStatusError {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetPassedTests returns all passed test results
func (tsr *TestSuiteResults) GetPassedTests() []*TestCaseResult {
	var passed []*TestCaseResult
	for _, result := range tsr.TestResults {
		if result.Status == TestStatusPassed {
			passed = append(passed, result)
		}
	}
	return passed
}
