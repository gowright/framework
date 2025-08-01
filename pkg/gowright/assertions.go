package gowright

import (
	"fmt"
	"reflect"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAssertion provides assertion capabilities with logging and reporting
type TestAssertion struct {
	testName string
	steps    []AssertionStep
	logs     []string
}

// NewTestAssertion creates a new test assertion instance
func NewTestAssertion(testName string) *TestAssertion {
	return &TestAssertion{
		testName: testName,
		steps:    make([]AssertionStep, 0),
		logs:     make([]string, 0),
	}
}

// Log adds a log entry to the test
func (ta *TestAssertion) Log(message string) {
	ta.logs = append(ta.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), message))
}

// Logf adds a formatted log entry to the test
func (ta *TestAssertion) Logf(format string, args ...interface{}) {
	ta.Log(fmt.Sprintf(format, args...))
}

// executeAssertion executes an assertion and records the result
func (ta *TestAssertion) executeAssertion(name, description string, assertionFunc func() bool, expected, actual interface{}) bool {
	startTime := time.Now()
	
	step := AssertionStep{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		Expected:    expected,
		Actual:      actual,
	}

	// Execute the assertion
	success := assertionFunc()
	
	endTime := time.Now()
	step.EndTime = endTime
	step.Duration = endTime.Sub(startTime)

	if success {
		step.Status = TestStatusPassed
		ta.Log(fmt.Sprintf("✓ %s: %s", name, description))
	} else {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("assertion failed: expected %v, got %v", expected, actual)
		ta.Log(fmt.Sprintf("✗ %s: %s - Expected: %v, Actual: %v", name, description, expected, actual))
	}

	ta.steps = append(ta.steps, step)
	return success
}

// Equal asserts that two values are equal
func (ta *TestAssertion) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("values should be equal")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"Equal",
		description,
		func() bool { return assert.Equal(&mockT{}, expected, actual) },
		expected,
		actual,
	)
}

// NotEqual asserts that two values are not equal
func (ta *TestAssertion) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("values should not be equal")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"NotEqual",
		description,
		func() bool { return assert.NotEqual(&mockT{}, expected, actual) },
		expected,
		actual,
	)
}

// True asserts that the value is true
func (ta *TestAssertion) True(value bool, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("value should be true")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"True",
		description,
		func() bool { return assert.True(&mockT{}, value) },
		true,
		value,
	)
}

// False asserts that the value is false
func (ta *TestAssertion) False(value bool, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("value should be false")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"False",
		description,
		func() bool { return assert.False(&mockT{}, value) },
		false,
		value,
	)
}

// Nil asserts that the value is nil
func (ta *TestAssertion) Nil(value interface{}, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("value should be nil")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"Nil",
		description,
		func() bool { return assert.Nil(&mockT{}, value) },
		nil,
		value,
	)
}

// NotNil asserts that the value is not nil
func (ta *TestAssertion) NotNil(value interface{}, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("value should not be nil")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"NotNil",
		description,
		func() bool { return assert.NotNil(&mockT{}, value) },
		"not nil",
		value,
	)
}

// Contains asserts that the string contains the substring
func (ta *TestAssertion) Contains(s, contains string, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("string should contain substring")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"Contains",
		description,
		func() bool { return assert.Contains(&mockT{}, s, contains) },
		fmt.Sprintf("string containing '%s'", contains),
		s,
	)
}

// NotContains asserts that the string does not contain the substring
func (ta *TestAssertion) NotContains(s, contains string, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("string should not contain substring")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"NotContains",
		description,
		func() bool { return assert.NotContains(&mockT{}, s, contains) },
		fmt.Sprintf("string not containing '%s'", contains),
		s,
	)
}

// Len asserts that the object has the expected length
func (ta *TestAssertion) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("object should have expected length")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	actualLength := getLength(object)
	return ta.executeAssertion(
		"Len",
		description,
		func() bool { return assert.Len(&mockT{}, object, length) },
		length,
		actualLength,
	)
}

// Empty asserts that the object is empty
func (ta *TestAssertion) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("object should be empty")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"Empty",
		description,
		func() bool { return assert.Empty(&mockT{}, object) },
		"empty",
		object,
	)
}

// NotEmpty asserts that the object is not empty
func (ta *TestAssertion) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("object should not be empty")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"NotEmpty",
		description,
		func() bool { return assert.NotEmpty(&mockT{}, object) },
		"not empty",
		object,
	)
}

// Error asserts that the error is not nil
func (ta *TestAssertion) Error(err error, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("error should be present")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"Error",
		description,
		func() bool { return assert.Error(&mockT{}, err) },
		"error present",
		err,
	)
}

// NoError asserts that the error is nil
func (ta *TestAssertion) NoError(err error, msgAndArgs ...interface{}) bool {
	description := fmt.Sprintf("error should not be present")
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				description = fmt.Sprintf(msg, msgAndArgs[1:]...)
			} else {
				description = msg
			}
		}
	}

	return ta.executeAssertion(
		"NoError",
		description,
		func() bool { return assert.NoError(&mockT{}, err) },
		nil,
		err,
	)
}

// GetSteps returns all assertion steps
func (ta *TestAssertion) GetSteps() []AssertionStep {
	return ta.steps
}

// GetLogs returns all log entries
func (ta *TestAssertion) GetLogs() []string {
	return ta.logs
}

// GetSummary returns a summary of the test assertions
func (ta *TestAssertion) GetSummary() (passed, failed int) {
	for _, step := range ta.steps {
		if step.Status == TestStatusPassed {
			passed++
		} else if step.Status == TestStatusFailed {
			failed++
		}
	}
	return passed, failed
}

// HasFailures returns true if any assertion failed
func (ta *TestAssertion) HasFailures() bool {
	for _, step := range ta.steps {
		if step.Status == TestStatusFailed {
			return true
		}
	}
	return false
}

// mockT is a minimal implementation of testing.T for testify
type mockT struct{}

func (m *mockT) Errorf(format string, args ...interface{}) {}
func (m *mockT) FailNow()                                  {}
func (m *mockT) Helper()                                   {}

// getLength returns the length of an object
func getLength(object interface{}) int {
	if object == nil {
		return 0
	}

	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return objValue.Len()
	default:
		return 0
	}
}