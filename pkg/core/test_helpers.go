package core

import (
	"fmt"
	"time"

	"github.com/gowright/framework/pkg/assertions"
)

// TestExecutor provides a convenient way to execute tests with assertion logging
type TestExecutor struct {
	asserter  *assertions.Asserter
	testName  string
	startTime time.Time
	logs      []string
}

// NewTestExecutor creates a new test executor with assertion logging
func NewTestExecutor(testName string) *TestExecutor {
	return &TestExecutor{
		asserter:  assertions.NewAsserter(),
		testName:  testName,
		startTime: time.Now(),
		logs:      make([]string, 0),
	}
}

// Assert returns the Asserter instance for making assertions
func (te *TestExecutor) Assert() *assertions.Asserter {
	return te.asserter
}

// Log adds a log entry to the test
func (te *TestExecutor) Log(message string) {
	te.logs = append(te.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), message))
}

// Logf adds a formatted log entry to the test
func (te *TestExecutor) Logf(format string, args ...interface{}) {
	te.Log(fmt.Sprintf(format, args...))
}

// Complete finalizes the test execution and returns the result
func (te *TestExecutor) Complete() *TestCaseResult {
	endTime := time.Now()

	// Determine overall test status
	status := TestStatusPassed
	var testError error

	if te.asserter.HasFailures() {
		status = TestStatusFailed
		testError = fmt.Errorf("test failed with assertion failures")
	}

	return &TestCaseResult{
		Name:      te.testName,
		Status:    status,
		Duration:  endTime.Sub(te.startTime),
		Error:     testError,
		StartTime: te.startTime,
		EndTime:   endTime,
		Steps:     te.asserter.GetSteps(),
		Logs:      te.logs,
	}
}

// GetTestName returns the test name
func (te *TestExecutor) GetTestName() string {
	return te.testName
}

// GetDuration returns the current test duration
func (te *TestExecutor) GetDuration() time.Duration {
	return time.Since(te.startTime)
}

// GetLogs returns all log entries
func (te *TestExecutor) GetLogs() []string {
	return te.logs
}

// Reset resets the test executor for reuse
func (te *TestExecutor) Reset(testName string) {
	te.testName = testName
	te.startTime = time.Now()
	te.logs = make([]string, 0)
	te.asserter.Reset()
}

// SimpleTest is a basic implementation of the Test interface for testing
type SimpleTest struct {
	name     string
	testFunc func() *TestCaseResult
}

// NewSimpleTest creates a new simple test
func NewSimpleTest(name string, testFunc func() *TestCaseResult) *SimpleTest {
	return &SimpleTest{
		name:     name,
		testFunc: testFunc,
	}
}

// GetName returns the test name
func (st *SimpleTest) GetName() string {
	return st.name
}

// Execute runs the test function
func (st *SimpleTest) Execute() *TestCaseResult {
	if st.testFunc != nil {
		return st.testFunc()
	}

	return &TestCaseResult{
		Name:      st.name,
		Status:    TestStatusPassed,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  0,
	}
}

// TestBuilder provides a fluent interface for building tests
type TestBuilder struct {
	name        string
	setupFunc   func() error
	testFunc    func(*TestExecutor) error
	cleanupFunc func() error
}

// NewTestBuilder creates a new test builder
func NewTestBuilder(name string) *TestBuilder {
	return &TestBuilder{
		name: name,
	}
}

// WithSetup sets the setup function
func (tb *TestBuilder) WithSetup(setupFunc func() error) *TestBuilder {
	tb.setupFunc = setupFunc
	return tb
}

// WithTest sets the test function
func (tb *TestBuilder) WithTest(testFunc func(*TestExecutor) error) *TestBuilder {
	tb.testFunc = testFunc
	return tb
}

// WithCleanup sets the cleanup function
func (tb *TestBuilder) WithCleanup(cleanupFunc func() error) *TestBuilder {
	tb.cleanupFunc = cleanupFunc
	return tb
}

// Build creates the test
func (tb *TestBuilder) Build() Test {
	return NewSimpleTest(tb.name, func() *TestCaseResult {
		executor := NewTestExecutor(tb.name)

		// Execute setup
		if tb.setupFunc != nil {
			if err := tb.setupFunc(); err != nil {
				return &TestCaseResult{
					Name:      tb.name,
					Status:    TestStatusError,
					Error:     NewGowrightError(ConfigurationError, "setup failed", err),
					StartTime: executor.startTime,
					EndTime:   time.Now(),
					Duration:  time.Since(executor.startTime),
				}
			}
		}

		// Ensure cleanup runs
		defer func() {
			if tb.cleanupFunc != nil {
				if err := tb.cleanupFunc(); err != nil {
					executor.Logf("Cleanup failed: %v", err)
				}
			}
		}()

		// Execute test
		if tb.testFunc != nil {
			if err := tb.testFunc(executor); err != nil {
				return &TestCaseResult{
					Name:      tb.name,
					Status:    TestStatusError,
					Error:     err,
					StartTime: executor.startTime,
					EndTime:   time.Now(),
					Duration:  time.Since(executor.startTime),
					Logs:      executor.logs,
				}
			}
		}

		return executor.Complete()
	})
}
