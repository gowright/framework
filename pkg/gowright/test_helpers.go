package gowright

import (
	"fmt"
	"time"
)

// TestExecutor provides a convenient way to execute tests with assertion logging
type TestExecutor struct {
	assertion *TestAssertion
	startTime time.Time
}

// NewTestExecutor creates a new test executor with assertion logging
func NewTestExecutor(testName string) *TestExecutor {
	return &TestExecutor{
		assertion: NewTestAssertion(testName),
		startTime: time.Now(),
	}
}

// Assert returns the TestAssertion instance for making assertions
func (te *TestExecutor) Assert() *TestAssertion {
	return te.assertion
}

// Log adds a log entry to the test
func (te *TestExecutor) Log(message string) {
	te.assertion.Log(message)
}

// Logf adds a formatted log entry to the test
func (te *TestExecutor) Logf(format string, args ...interface{}) {
	te.assertion.Logf(format, args...)
}

// Complete finalizes the test execution and returns the result
func (te *TestExecutor) Complete(testName string) *TestCaseResult {
	endTime := time.Now()
	
	// Determine overall test status
	status := TestStatusPassed
	var testError error
	
	if te.assertion.HasFailures() {
		status = TestStatusFailed
		_, failed := te.assertion.GetSummary()
		testError = fmt.Errorf("test failed with %d assertion failures", failed)
	}
	
	return &TestCaseResult{
		Name:      testName,
		Status:    status,
		Duration:  endTime.Sub(te.startTime),
		Error:     testError,
		StartTime: te.startTime,
		EndTime:   endTime,
		Logs:      te.assertion.GetLogs(),
		Steps:     te.assertion.GetSteps(),
	}
}

// ExecuteTestWithAssertions is a helper function to execute a test function with assertion logging
func ExecuteTestWithAssertions(testName string, testFunc func(*TestAssertion)) *TestCaseResult {
	executor := NewTestExecutor(testName)
	
	// Execute the test function
	testFunc(executor.assertion)
	
	return executor.Complete(testName)
}

// TestSuiteExecutor provides utilities for executing test suites with assertion reporting
type TestSuiteExecutor struct {
	suiteName string
	results   []TestCaseResult
	startTime time.Time
}

// NewTestSuiteExecutor creates a new test suite executor
func NewTestSuiteExecutor(suiteName string) *TestSuiteExecutor {
	return &TestSuiteExecutor{
		suiteName: suiteName,
		results:   make([]TestCaseResult, 0),
		startTime: time.Now(),
	}
}

// AddTest adds a test result to the suite
func (tse *TestSuiteExecutor) AddTest(result *TestCaseResult) {
	tse.results = append(tse.results, *result)
}

// ExecuteTest executes a test function and adds the result to the suite
func (tse *TestSuiteExecutor) ExecuteTest(testName string, testFunc func(*TestAssertion)) {
	result := ExecuteTestWithAssertions(testName, testFunc)
	tse.AddTest(result)
}

// GetResults returns the complete test results for the suite
func (tse *TestSuiteExecutor) GetResults() *TestResults {
	endTime := time.Now()
	
	// Calculate statistics
	totalTests := len(tse.results)
	passedTests := 0
	failedTests := 0
	skippedTests := 0
	errorTests := 0
	
	for _, result := range tse.results {
		switch result.Status {
		case TestStatusPassed:
			passedTests++
		case TestStatusFailed:
			failedTests++
		case TestStatusSkipped:
			skippedTests++
		case TestStatusError:
			errorTests++
		}
	}
	
	return &TestResults{
		SuiteName:    tse.suiteName,
		StartTime:    tse.startTime,
		EndTime:      endTime,
		TotalTests:   totalTests,
		PassedTests:  passedTests,
		FailedTests:  failedTests,
		SkippedTests: skippedTests,
		ErrorTests:   errorTests,
		TestCases:    tse.results,
	}
}

// GenerateReports generates reports for the test suite
func (tse *TestSuiteExecutor) GenerateReports(config *ReportConfig) error {
	results := tse.GetResults()
	reportManager := NewReportManager(config)
	summary := reportManager.GenerateReports(results)
	
	// Return error if no reports were successful
	if summary.SuccessfulReports == 0 {
		if summary.FallbackError != nil {
			return summary.FallbackError
		}
		if len(summary.Results) > 0 {
			return summary.Results[0].Error
		}
		return NewGowrightError(ReportingError, "all reporting failed", nil)
	}
	
	return nil
}