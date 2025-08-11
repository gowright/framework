package core

import (
	"time"

	"github.com/gowright/framework/pkg/assertions"
)

// TestContext provides a context for function-based tests with assertion capabilities
type TestContext struct {
	testName  string
	asserter  *assertions.Asserter
	startTime time.Time
	timeout   time.Duration
}

// NewTestContext creates a new test context
func NewTestContext(testName string) *TestContext {
	return &TestContext{
		testName:  testName,
		asserter:  assertions.NewAsserter(),
		startTime: time.Now(),
		timeout:   30 * time.Second, // default timeout
	}
}

// NewTestContextWithTimeout creates a new test context with a specific timeout
func NewTestContextWithTimeout(testName string, timeout time.Duration) *TestContext {
	return &TestContext{
		testName:  testName,
		asserter:  assertions.NewAsserter(),
		startTime: time.Now(),
		timeout:   timeout,
	}
}

// GetTestName returns the test name
func (tc *TestContext) GetTestName() string {
	return tc.testName
}

// GetStartTime returns the test start time
func (tc *TestContext) GetStartTime() time.Time {
	return tc.startTime
}

// GetTimeout returns the test timeout
func (tc *TestContext) GetTimeout() time.Duration {
	return tc.timeout
}

// AssertTrue asserts that a value is true
func (tc *TestContext) AssertTrue(value bool, message string) bool {
	return tc.asserter.True(value, message)
}

// AssertFalse asserts that a value is false
func (tc *TestContext) AssertFalse(value bool, message string) bool {
	return tc.asserter.False(value, message)
}

// AssertEqual asserts that two values are equal
func (tc *TestContext) AssertEqual(expected, actual interface{}, message string) bool {
	return tc.asserter.Equal(expected, actual, message)
}

// AssertNotEqual asserts that two values are not equal
func (tc *TestContext) AssertNotEqual(expected, actual interface{}, message string) bool {
	return tc.asserter.NotEqual(expected, actual, message)
}

// AssertNil asserts that a value is nil
func (tc *TestContext) AssertNil(value interface{}, message string) bool {
	return tc.asserter.Nil(value, message)
}

// AssertNotNil asserts that a value is not nil
func (tc *TestContext) AssertNotNil(value interface{}, message string) bool {
	return tc.asserter.NotNil(value, message)
}

// AssertContains asserts that a string contains a substring
func (tc *TestContext) AssertContains(haystack, needle string, message string) bool {
	return tc.asserter.Contains(haystack, needle, message)
}

// AssertNotContains asserts that a string does not contain a substring
func (tc *TestContext) AssertNotContains(haystack, needle string, message string) bool {
	return tc.asserter.NotContains(haystack, needle, message)
}

// GetSteps returns all assertion steps
func (tc *TestContext) GetSteps() []AssertionStep {
	return tc.asserter.GetSteps()
}

// HasFailures returns true if any assertions failed
func (tc *TestContext) HasFailures() bool {
	return tc.asserter.HasFailures()
}

// ToTestCaseResult converts the test context to a test case result
func (tc *TestContext) ToTestCaseResult() *TestCaseResult {
	return &TestCaseResult{
		Name:      tc.testName,
		StartTime: tc.startTime,
		EndTime:   time.Now(),
		Duration:  time.Since(tc.startTime),
		Status:    tc.getStatus(),
		Steps:     tc.asserter.GetSteps(),
	}
}

// Err returns an error if any assertions failed
func (tc *TestContext) Err() error {
	if tc.asserter.HasFailures() {
		return NewGowrightError(AssertionError, "one or more assertions failed", nil)
	}
	return nil
}

// AddCleanup adds a cleanup function to be called when the test finishes
func (tc *TestContext) AddCleanup(cleanup func() error) {
	// Store cleanup functions if needed
	// For now, this is a placeholder
}

// getStatus returns the test status based on assertion results
func (tc *TestContext) getStatus() TestStatus {
	if tc.asserter.HasFailures() {
		return TestStatusFailed
	}
	return TestStatusPassed
}

// Close performs cleanup operations
func (tc *TestContext) Close() {
	// Cleanup operations if needed
}
