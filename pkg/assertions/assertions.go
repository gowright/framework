// Package assertions provides common assertion utilities for all testing modules
package assertions

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Asserter provides assertion capabilities for tests
type Asserter struct {
	steps []AssertionStep
}

// NewAsserter creates a new asserter instance
func NewAsserter() *Asserter {
	return &Asserter{
		steps: make([]AssertionStep, 0),
	}
}

// Equal asserts that two values are equal
func (a *Asserter) Equal(expected, actual interface{}, message string) bool {
	step := AssertionStep{
		Name:        "Equal",
		Description: message,
		Expected:    expected,
		Actual:      actual,
		StartTime:   time.Now(),
	}

	success := reflect.DeepEqual(expected, actual)
	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected %v, got %v", expected, actual)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// NotEqual asserts that two values are not equal
func (a *Asserter) NotEqual(expected, actual interface{}, message string) bool {
	step := AssertionStep{
		Name:        "NotEqual",
		Description: message,
		Expected:    expected,
		Actual:      actual,
		StartTime:   time.Now(),
	}

	success := !reflect.DeepEqual(expected, actual)
	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected %v to not equal %v", expected, actual)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// True asserts that a value is true
func (a *Asserter) True(value bool, message string) bool {
	return a.Equal(true, value, message)
}

// False asserts that a value is false
func (a *Asserter) False(value bool, message string) bool {
	return a.Equal(false, value, message)
}

// Nil asserts that a value is nil
func (a *Asserter) Nil(value interface{}, message string) bool {
	step := AssertionStep{
		Name:        "Nil",
		Description: message,
		Expected:    nil,
		Actual:      value,
		StartTime:   time.Now(),
	}

	success := value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected nil, got %v", value)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// NotNil asserts that a value is not nil
func (a *Asserter) NotNil(value interface{}, message string) bool {
	step := AssertionStep{
		Name:        "NotNil",
		Description: message,
		Expected:    "not nil",
		Actual:      value,
		StartTime:   time.Now(),
	}

	success := value != nil && (reflect.ValueOf(value).Kind() != reflect.Ptr || !reflect.ValueOf(value).IsNil())
	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected not nil, got nil")
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// Contains asserts that a string contains a substring
func (a *Asserter) Contains(haystack, needle string, message string) bool {
	step := AssertionStep{
		Name:        "Contains",
		Description: message,
		Expected:    needle,
		Actual:      haystack,
		StartTime:   time.Now(),
	}

	success := strings.Contains(haystack, needle)
	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected '%s' to contain '%s'", haystack, needle)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// NotContains asserts that a string does not contain a substring
func (a *Asserter) NotContains(haystack, needle string, message string) bool {
	step := AssertionStep{
		Name:        "NotContains",
		Description: message,
		Expected:    fmt.Sprintf("not containing '%s'", needle),
		Actual:      haystack,
		StartTime:   time.Now(),
	}

	success := !strings.Contains(haystack, needle)
	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected '%s' to not contain '%s'", haystack, needle)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// Greater asserts that a value is greater than another
func (a *Asserter) Greater(actual, expected interface{}, message string) bool {
	step := AssertionStep{
		Name:        "Greater",
		Description: message,
		Expected:    expected,
		Actual:      actual,
		StartTime:   time.Now(),
	}

	success := false
	switch actualVal := actual.(type) {
	case int:
		if expectedVal, ok := expected.(int); ok {
			success = actualVal > expectedVal
		}
	case int64:
		if expectedVal, ok := expected.(int64); ok {
			success = actualVal > expectedVal
		}
	case float64:
		if expectedVal, ok := expected.(float64); ok {
			success = actualVal > expectedVal
		}
	case time.Duration:
		if expectedVal, ok := expected.(time.Duration); ok {
			success = actualVal > expectedVal
		}
	}

	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected %v to be greater than %v", actual, expected)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// Less asserts that a value is less than another
func (a *Asserter) Less(actual, expected interface{}, message string) bool {
	step := AssertionStep{
		Name:        "Less",
		Description: message,
		Expected:    expected,
		Actual:      actual,
		StartTime:   time.Now(),
	}

	success := false
	switch actualVal := actual.(type) {
	case int:
		if expectedVal, ok := expected.(int); ok {
			success = actualVal < expectedVal
		}
	case int64:
		if expectedVal, ok := expected.(int64); ok {
			success = actualVal < expectedVal
		}
	case float64:
		if expectedVal, ok := expected.(float64); ok {
			success = actualVal < expectedVal
		}
	case time.Duration:
		if expectedVal, ok := expected.(time.Duration); ok {
			success = actualVal < expectedVal
		}
	}

	step.Status = TestStatusPassed
	if !success {
		step.Status = TestStatusFailed
		step.Error = fmt.Errorf("expected %v to be less than %v", actual, expected)
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	a.steps = append(a.steps, step)

	return success
}

// GetSteps returns all assertion steps
func (a *Asserter) GetSteps() []AssertionStep {
	return a.steps
}

// Reset clears all assertion steps
func (a *Asserter) Reset() {
	a.steps = make([]AssertionStep, 0)
}

// HasFailures returns true if any assertions failed
func (a *Asserter) HasFailures() bool {
	for _, step := range a.steps {
		if step.Status == TestStatusFailed || step.Status == TestStatusError {
			return true
		}
	}
	return false
}
