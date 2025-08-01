package gowright

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTestAssertion(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	assert.Equal(t, "Test Case", ta.testName)
	assert.Empty(t, ta.steps)
	assert.Empty(t, ta.logs)
}

func TestTestAssertion_Log(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	ta.Log("First log entry")
	ta.Logf("Second log entry with value: %d", 42)
	
	logs := ta.GetLogs()
	assert.Len(t, logs, 2)
	assert.Contains(t, logs[0], "First log entry")
	assert.Contains(t, logs[1], "Second log entry with value: 42")
}

func TestTestAssertion_Equal(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.Equal(5, 5, "numbers should be equal")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Equal", steps[0].Name)
	assert.Equal(t, "numbers should be equal", steps[0].Description)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	assert.Equal(t, 5, steps[0].Expected)
	assert.Equal(t, 5, steps[0].Actual)
	assert.Nil(t, steps[0].Error)
	
	// Test failed assertion
	result = ta.Equal(5, 10, "numbers should be equal")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
	assert.NotNil(t, steps[1].Error)
	assert.Equal(t, 5, steps[1].Expected)
	assert.Equal(t, 10, steps[1].Actual)
}

func TestTestAssertion_NotEqual(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.NotEqual(5, 10, "numbers should not be equal")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotEqual", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.NotEqual(5, 5, "numbers should not be equal")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_True(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.True(true, "value should be true")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "True", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	assert.Equal(t, true, steps[0].Expected)
	assert.Equal(t, true, steps[0].Actual)
	
	// Test failed assertion
	result = ta.True(false, "value should be true")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_False(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.False(false, "value should be false")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "False", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.False(true, "value should be false")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_Nil(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.Nil(nil, "value should be nil")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Nil", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.Nil("not nil", "value should be nil")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_NotNil(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.NotNil("not nil", "value should not be nil")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotNil", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.NotNil(nil, "value should not be nil")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_Contains(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.Contains("hello world", "world", "string should contain substring")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Contains", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.Contains("hello world", "foo", "string should contain substring")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_NotContains(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.NotContains("hello world", "foo", "string should not contain substring")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotContains", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.NotContains("hello world", "world", "string should not contain substring")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_Len(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion with slice
	slice := []int{1, 2, 3}
	result := ta.Len(slice, 3, "slice should have length 3")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Len", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	assert.Equal(t, 3, steps[0].Expected)
	assert.Equal(t, 3, steps[0].Actual)
	
	// Test failed assertion
	result = ta.Len(slice, 5, "slice should have length 5")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_Empty(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.Empty([]int{}, "slice should be empty")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Empty", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.Empty([]int{1, 2, 3}, "slice should be empty")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_NotEmpty(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.NotEmpty([]int{1, 2, 3}, "slice should not be empty")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NotEmpty", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.NotEmpty([]int{}, "slice should not be empty")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_Error(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	err := errors.New("test error")
	result := ta.Error(err, "error should be present")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "Error", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	result = ta.Error(nil, "error should be present")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_NoError(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test successful assertion
	result := ta.NoError(nil, "error should not be present")
	assert.True(t, result)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "NoError", steps[0].Name)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	// Test failed assertion
	err := errors.New("test error")
	result = ta.NoError(err, "error should not be present")
	assert.False(t, result)
	
	steps = ta.GetSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, TestStatusFailed, steps[1].Status)
}

func TestTestAssertion_GetSummary(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Add some passing and failing assertions
	ta.Equal(5, 5)
	ta.Equal(10, 15) // This will fail
	ta.True(true)
	ta.False(true) // This will fail
	
	passed, failed := ta.GetSummary()
	assert.Equal(t, 2, passed)
	assert.Equal(t, 2, failed)
}

func TestTestAssertion_HasFailures(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Initially no failures
	assert.False(t, ta.HasFailures())
	
	// Add passing assertion
	ta.Equal(5, 5)
	assert.False(t, ta.HasFailures())
	
	// Add failing assertion
	ta.Equal(5, 10)
	assert.True(t, ta.HasFailures())
}

func TestTestAssertion_DefaultMessages(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	// Test assertions without custom messages
	ta.Equal(5, 5)
	ta.True(true)
	ta.Nil(nil)
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 3)
	
	assert.Equal(t, "values should be equal", steps[0].Description)
	assert.Equal(t, "value should be true", steps[1].Description)
	assert.Equal(t, "value should be nil", steps[2].Description)
}

func TestTestAssertion_StepTiming(t *testing.T) {
	ta := NewTestAssertion("Test Case")
	
	startTime := time.Now()
	ta.Equal(5, 5)
	endTime := time.Now()
	
	steps := ta.GetSteps()
	assert.Len(t, steps, 1)
	
	step := steps[0]
	assert.True(t, step.StartTime.After(startTime) || step.StartTime.Equal(startTime))
	assert.True(t, step.EndTime.Before(endTime) || step.EndTime.Equal(endTime))
	assert.True(t, step.Duration >= 0)
	assert.Equal(t, step.Duration, step.EndTime.Sub(step.StartTime))
}

func TestGetLength(t *testing.T) {
	tests := []struct {
		name     string
		object   interface{}
		expected int
	}{
		{"nil", nil, 0},
		{"empty slice", []int{}, 0},
		{"slice with elements", []int{1, 2, 3}, 3},
		{"empty string", "", 0},
		{"string with content", "hello", 5},
		{"empty map", map[string]int{}, 0},
		{"map with elements", map[string]int{"a": 1, "b": 2}, 2},
		{"array", [3]int{1, 2, 3}, 3},
		{"non-collection type", 42, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLength(tt.object)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTestAssertion_Integration(t *testing.T) {
	ta := NewTestAssertion("Integration Test")
	
	// Simulate a test scenario
	ta.Log("Starting integration test")
	
	// Test API response
	apiResponse := map[string]interface{}{
		"status": "success",
		"data":   []string{"item1", "item2", "item3"},
		"count":  3,
	}
	
	ta.Equal("success", apiResponse["status"], "API should return success status")
	ta.NotNil(apiResponse["data"], "API should return data")
	ta.Len(apiResponse["data"], 3, "API should return 3 items")
	ta.Equal(3, apiResponse["count"], "Count should match data length")
	
	ta.Log("API response validation completed")
	
	// Test database operation
	dbError := error(nil) // Simulate successful DB operation
	ta.NoError(dbError, "Database operation should succeed")
	
	ta.Log("Database operation completed")
	
	// Verify results
	steps := ta.GetSteps()
	logs := ta.GetLogs()
	passed, failed := ta.GetSummary()
	
	assert.Len(t, steps, 5)
	assert.Len(t, logs, 3)
	assert.Equal(t, 5, passed)
	assert.Equal(t, 0, failed)
	assert.False(t, ta.HasFailures())
	
	// Verify step details
	assert.Equal(t, "Equal", steps[0].Name)
	assert.Equal(t, "API should return success status", steps[0].Description)
	assert.Equal(t, TestStatusPassed, steps[0].Status)
	
	assert.Equal(t, "NotNil", steps[1].Name)
	assert.Equal(t, "Len", steps[2].Name)
	assert.Equal(t, "Equal", steps[3].Name)
	assert.Equal(t, "NoError", steps[4].Name)
}