package core

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gowright/framework/pkg/assertions"
)

// TestAssertion provides assertion capabilities for tests
type TestAssertion struct {
	testName  string
	asserter  *assertions.Asserter
	startTime time.Time
}

// NewTestAssertion creates a new test assertion instance
func NewTestAssertion(testName string) *TestAssertion {
	return &TestAssertion{
		testName:  testName,
		asserter:  assertions.NewAsserter(),
		startTime: time.Now(),
	}
}

// GetTestName returns the test name
func (ta *TestAssertion) GetTestName() string {
	return ta.testName
}

// GetStartTime returns the test start time
func (ta *TestAssertion) GetStartTime() time.Time {
	return ta.startTime
}

// Equal asserts that two values are equal
func (ta *TestAssertion) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	message := "values should be equal"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.Equal(expected, actual, message)
}

// NotEqual asserts that two values are not equal
func (ta *TestAssertion) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	message := "values should not be equal"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.NotEqual(expected, actual, message)
}

// True asserts that a value is true
func (ta *TestAssertion) True(value bool, msgAndArgs ...interface{}) bool {
	message := "value should be true"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.True(value, message)
}

// False asserts that a value is false
func (ta *TestAssertion) False(value bool, msgAndArgs ...interface{}) bool {
	message := "value should be false"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.False(value, message)
}

// Nil asserts that a value is nil
func (ta *TestAssertion) Nil(value interface{}, msgAndArgs ...interface{}) bool {
	message := "value should be nil"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.Nil(value, message)
}

// NotNil asserts that a value is not nil
func (ta *TestAssertion) NotNil(value interface{}, msgAndArgs ...interface{}) bool {
	message := "value should not be nil"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.NotNil(value, message)
}

// Contains asserts that a string contains a substring
func (ta *TestAssertion) Contains(haystack, needle string, msgAndArgs ...interface{}) bool {
	message := "string should contain substring"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.Contains(haystack, needle, message)
}

// NotContains asserts that a string does not contain a substring
func (ta *TestAssertion) NotContains(haystack, needle string, msgAndArgs ...interface{}) bool {
	message := "string should not contain substring"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.NotContains(haystack, needle, message)
}

// Len asserts that an object has the expected length
func (ta *TestAssertion) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	message := "object should have expected length"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}

	actualLength := getLength(object)
	return ta.asserter.Equal(length, actualLength, message)
}

// Empty asserts that an object is empty
func (ta *TestAssertion) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	message := "object should be empty"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}

	isEmpty := getLength(object) == 0
	return ta.asserter.True(isEmpty, message)
}

// NotEmpty asserts that an object is not empty
func (ta *TestAssertion) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	message := "object should not be empty"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}

	isNotEmpty := getLength(object) > 0
	return ta.asserter.True(isNotEmpty, message)
}

// Error asserts that an error is not nil
func (ta *TestAssertion) Error(err error, msgAndArgs ...interface{}) bool {
	message := "error should not be nil"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.NotNil(err, message)
}

// NoError asserts that an error is nil
func (ta *TestAssertion) NoError(err error, msgAndArgs ...interface{}) bool {
	message := "error should be nil"
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			message = msg
		}
	}
	return ta.asserter.Nil(err, message)
}

// GetSteps returns all assertion steps
func (ta *TestAssertion) GetSteps() []AssertionStep {
	return ta.asserter.GetSteps()
}

// HasFailures returns true if any assertions failed
func (ta *TestAssertion) HasFailures() bool {
	return ta.asserter.HasFailures()
}

// Reset clears all assertion steps
func (ta *TestAssertion) Reset() {
	ta.asserter.Reset()
}

// GetLogs returns all log entries (placeholder implementation)
func (ta *TestAssertion) GetLogs() []string {
	// For now, return a simple log based on steps
	logs := make([]string, 0)
	for _, step := range ta.asserter.GetSteps() {
		if step.Status == TestStatusPassed {
			logs = append(logs, fmt.Sprintf("✓ %s: %s", step.Name, step.Description))
		} else {
			logs = append(logs, fmt.Sprintf("✗ %s: %s - %v", step.Name, step.Description, step.Error))
		}
	}
	return logs
}

// GetSummary returns the count of passed and failed assertions
func (ta *TestAssertion) GetSummary() (int, int) {
	passed := 0
	failed := 0
	for _, step := range ta.asserter.GetSteps() {
		if step.Status == TestStatusPassed {
			passed++
		} else {
			failed++
		}
	}
	return passed, failed
}

// getLength returns the length of an object
func getLength(object interface{}) int {
	if object == nil {
		return 0
	}

	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len()
	case reflect.String:
		return len(object.(string))
	default:
		return 0
	}
}
