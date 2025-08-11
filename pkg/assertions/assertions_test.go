package assertions

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAsserter(t *testing.T) {
	asserter := NewAsserter()

	assert.NotNil(t, asserter)
	assert.Empty(t, asserter.steps)
}

func TestAsserter_Equal(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.Equal(5, 5, "numbers should be equal")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Equal", steps[0].Name)
	assert.Equal(t, "numbers should be equal", steps[0].Description)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	assert.Equal(t, 5, steps[0].Expected)
	assert.Equal(t, 5, steps[0].Actual)
	assert.Nil(t, steps[0].Error)

	// Test failed assertion
	result = asserter.Equal(5, 10, "numbers should be equal")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
	assert.NotNil(t, steps[1].Error)
	assert.Equal(t, 5, steps[1].Expected)
	assert.Equal(t, 10, steps[1].Actual)
}

func TestAsserter_NotEqual(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.NotEqual(5, 10, "numbers should not be equal")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotEqual", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.NotEqual(5, 5, "numbers should not be equal")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_True(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.True(true, "value should be true")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Equal", steps[0].Name) // True uses Equal internally
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	assert.Equal(t, true, steps[0].Expected)
	assert.Equal(t, true, steps[0].Actual)

	// Test failed assertion
	result = asserter.True(false, "value should be true")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_False(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.False(false, "value should be false")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Equal", steps[0].Name) // False uses Equal internally
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.False(true, "value should be false")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_Nil(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.Nil(nil, "value should be nil")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Nil", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.Nil("not nil", "value should be nil")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_NotNil(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.NotNil("not nil", "value should not be nil")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotNil", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.NotNil(nil, "value should not be nil")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_Contains(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.Contains("hello world", "world", "string should contain substring")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Contains", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.Contains("hello world", "foo", "string should contain substring")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_NotContains(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion
	result := asserter.NotContains("hello world", "foo", "string should not contain substring")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotContains", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.NotContains("hello world", "world", "string should not contain substring")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestAsserter_Greater(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion with int
	result := asserter.Greater(10, 5, "10 should be greater than 5")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Greater", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.Greater(5, 10, "5 should be greater than 10")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)

	// Test with float64
	result = asserter.Greater(10.5, 5.2, "10.5 should be greater than 5.2")
	assert.True(t, result)

	// Test with time.Duration
	result = asserter.Greater(10*time.Second, 5*time.Second, "10s should be greater than 5s")
	assert.True(t, result)
}

func TestAsserter_Less(t *testing.T) {
	asserter := NewAsserter()

	// Test successful assertion with int
	result := asserter.Less(5, 10, "5 should be less than 10")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Less", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	// Test failed assertion
	result = asserter.Less(10, 5, "10 should be less than 5")
	assert.False(t, result)

	steps = asserter.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)

	// Test with float64
	result = asserter.Less(5.2, 10.5, "5.2 should be less than 10.5")
	assert.True(t, result)

	// Test with time.Duration
	result = asserter.Less(5*time.Second, 10*time.Second, "5s should be less than 10s")
	assert.True(t, result)
}

func TestAsserter_GetSteps(t *testing.T) {
	asserter := NewAsserter()

	// Initially no steps
	steps := asserter.GetSteps()
	assert.Empty(t, steps)

	// Add some assertions
	asserter.Equal(5, 5, "test1")
	asserter.True(true, "test2")
	asserter.Contains("hello", "ell", "test3")

	steps = asserter.GetSteps()
	assert.Len(t, steps, 3)
	assert.Equal(t, "test1", steps[0].Description)
	assert.Equal(t, "test2", steps[1].Description)
	assert.Equal(t, "test3", steps[2].Description)
}

func TestAsserter_Reset(t *testing.T) {
	asserter := NewAsserter()

	// Add some assertions
	asserter.Equal(5, 5, "test1")
	asserter.True(true, "test2")

	steps := asserter.GetSteps()
	assert.Len(t, steps, 2)

	// Reset
	asserter.Reset()

	steps = asserter.GetSteps()
	assert.Empty(t, steps)
}

func TestAsserter_HasFailures(t *testing.T) {
	asserter := NewAsserter()

	// Initially no failures
	assert.False(t, asserter.HasFailures())

	// Add passing assertion
	asserter.Equal(5, 5, "should pass")
	assert.False(t, asserter.HasFailures())

	// Add failing assertion
	asserter.Equal(5, 10, "should fail")
	assert.True(t, asserter.HasFailures())
}

func TestAsserter_StepTiming(t *testing.T) {
	asserter := NewAsserter()

	startTime := time.Now()
	asserter.Equal(5, 5, "timing test")
	endTime := time.Now()

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)

	step := steps[0]
	assert.True(t, step.StartTime.After(startTime) || step.StartTime.Equal(startTime))
	assert.True(t, step.EndTime.Before(endTime) || step.EndTime.Equal(endTime))
	assert.True(t, step.Duration >= 0)
	assert.Equal(t, step.Duration, step.EndTime.Sub(step.StartTime))
}

func TestAsserter_Integration(t *testing.T) {
	asserter := NewAsserter()

	// Simulate a test scenario
	// Test API response
	apiResponse := map[string]interface{}{
		"status": "success",
		"data":   []string{"item1", "item2", "item3"},
		"count":  3,
	}

	asserter.Equal("success", apiResponse["status"], "API should return success status")
	asserter.NotNil(apiResponse["data"], "API should return data")
	asserter.Equal(3, apiResponse["count"], "Count should match data length")

	// Test database operation
	dbError := error(nil) // Simulate successful DB operation
	asserter.Nil(dbError, "Database operation should succeed")

	// Verify results
	steps := asserter.GetSteps()
	assert.Len(t, steps, 4)

	passed := 0
	failed := 0
	for _, step := range steps {
		switch step.Status {
		case TestStatusPassed:
			passed++
		case TestStatusFailed:
			failed++
		}
	}

	assert.Equal(t, 4, passed)
	assert.Equal(t, 0, failed)
	assert.False(t, asserter.HasFailures())

	// Verify step details
	assert.Equal(t, "Equal", steps[0].Name)
	assert.Equal(t, "API should return success status", steps[0].Description)
	assert.Equal(t, TestStatusPassed, steps[0].Status)

	assert.Equal(t, "NotNil", steps[1].Name)
	assert.Equal(t, "Equal", steps[2].Name)
	assert.Equal(t, "Nil", steps[3].Name)
}

func TestAsserter_ErrorHandling(t *testing.T) {
	asserter := NewAsserter()

	// Test with actual error
	testError := errors.New("test error")
	result := asserter.Nil(testError, "error should be nil")
	assert.False(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, TestStatusFailed, steps[0].Status)
	assert.NotNil(t, steps[0].Error)
	assert.Contains(t, steps[0].Error.Error(), "expected nil, got")
}

func TestAsserter_TypeComparisons(t *testing.T) {
	asserter := NewAsserter()

	// Test different numeric types
	asserter.Greater(int64(10), int64(5), "int64 comparison")
	asserter.Less(float64(5.5), float64(10.5), "float64 comparison")
	asserter.Greater(10*time.Millisecond, 5*time.Millisecond, "duration comparison")

	steps := asserter.GetSteps()
	assert.Len(t, steps, 3)

	for _, step := range steps {
		assert.Equal(t, TestStatusPassed, step.Status)
	}
}

func TestAsserter_NilPointerHandling(t *testing.T) {
	asserter := NewAsserter()

	// Test with nil pointer
	var nilPtr *string
	result := asserter.Nil(nilPtr, "nil pointer should be nil")
	assert.True(t, result)

	// Test with non-nil pointer
	str := "test"
	ptr := &str
	result = asserter.NotNil(ptr, "pointer should not be nil")
	assert.True(t, result)

	steps := asserter.GetSteps()
	assert.Len(t, steps, 2)

	for _, step := range steps {
		assert.Equal(t, TestStatusPassed, step.Status)
	}
}
