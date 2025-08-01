package gowright

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTestExecutor(t *testing.T) {
	executor := NewTestExecutor("Test Case")
	
	assert.NotNil(t, executor)
	assert.NotNil(t, executor.assertion)
	assert.Equal(t, "Test Case", executor.assertion.testName)
	assert.True(t, time.Since(executor.startTime) < time.Second)
}

func TestTestExecutor_Assert(t *testing.T) {
	executor := NewTestExecutor("Test Case")
	
	assertion := executor.Assert()
	assert.Equal(t, executor.assertion, assertion)
}

func TestTestExecutor_Log(t *testing.T) {
	executor := NewTestExecutor("Test Case")
	
	executor.Log("Test log message")
	executor.Logf("Formatted log: %d", 42)
	
	logs := executor.assertion.GetLogs()
	assert.Len(t, logs, 2)
	assert.Contains(t, logs[0], "Test log message")
	assert.Contains(t, logs[1], "Formatted log: 42")
}

func TestTestExecutor_Complete(t *testing.T) {
	executor := NewTestExecutor("Test Case")
	
	// Add some assertions
	executor.Assert().Equal(5, 5, "should be equal")
	executor.Assert().True(true, "should be true")
	
	result := executor.Complete("Test Case")
	
	assert.Equal(t, "Test Case", result.Name)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.Nil(t, result.Error)
	assert.Len(t, result.Steps, 2)
	assert.True(t, result.Duration > 0)
	assert.True(t, result.EndTime.After(result.StartTime))
}

func TestTestExecutor_CompleteWithFailures(t *testing.T) {
	executor := NewTestExecutor("Test Case")
	
	// Add passing and failing assertions
	executor.Assert().Equal(5, 5, "should be equal")
	executor.Assert().Equal(5, 10, "should fail")
	
	result := executor.Complete("Test Case")
	
	assert.Equal(t, "Test Case", result.Name)
	assert.Equal(t, TestStatusFailed, result.Status)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "1 assertion failures")
	assert.Len(t, result.Steps, 2)
}

func TestExecuteTestWithAssertions(t *testing.T) {
	result := ExecuteTestWithAssertions("Sample Test", func(assert *TestAssertion) {
		assert.Log("Starting test")
		assert.Equal(10, 10, "numbers should be equal")
		assert.True(true, "should be true")
		assert.Log("Test completed")
	})
	
	assert.Equal(t, "Sample Test", result.Name)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.Nil(t, result.Error)
	assert.Len(t, result.Steps, 2)
	assert.Len(t, result.Logs, 4) // 2 manual logs + 2 assertion logs
}

func TestExecuteTestWithAssertions_WithFailures(t *testing.T) {
	result := ExecuteTestWithAssertions("Failing Test", func(assert *TestAssertion) {
		assert.Log("Starting test")
		assert.Equal(10, 20, "this will fail")
		assert.True(false, "this will also fail")
		assert.Log("Test completed")
	})
	
	assert.Equal(t, "Failing Test", result.Name)
	assert.Equal(t, TestStatusFailed, result.Status)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "2 assertion failures")
	assert.Len(t, result.Steps, 2)
}

func TestNewTestSuiteExecutor(t *testing.T) {
	suite := NewTestSuiteExecutor("Test Suite")
	
	assert.NotNil(t, suite)
	assert.Equal(t, "Test Suite", suite.suiteName)
	assert.Empty(t, suite.results)
	assert.True(t, time.Since(suite.startTime) < time.Second)
}

func TestTestSuiteExecutor_AddTest(t *testing.T) {
	suite := NewTestSuiteExecutor("Test Suite")
	
	result := &TestCaseResult{
		Name:   "Test 1",
		Status: TestStatusPassed,
	}
	
	suite.AddTest(result)
	assert.Len(t, suite.results, 1)
	assert.Equal(t, "Test 1", suite.results[0].Name)
}

func TestTestSuiteExecutor_ExecuteTest(t *testing.T) {
	suite := NewTestSuiteExecutor("Test Suite")
	
	suite.ExecuteTest("Test 1", func(assert *TestAssertion) {
		assert.Equal(5, 5, "should be equal")
		assert.True(true, "should be true")
	})
	
	suite.ExecuteTest("Test 2", func(assert *TestAssertion) {
		assert.Equal(10, 15, "this will fail")
	})
	
	assert.Len(t, suite.results, 2)
	assert.Equal(t, "Test 1", suite.results[0].Name)
	assert.Equal(t, TestStatusPassed, suite.results[0].Status)
	assert.Equal(t, "Test 2", suite.results[1].Name)
	assert.Equal(t, TestStatusFailed, suite.results[1].Status)
}

func TestTestSuiteExecutor_GetResults(t *testing.T) {
	suite := NewTestSuiteExecutor("Test Suite")
	
	// Add various test results
	suite.AddTest(&TestCaseResult{Name: "Test 1", Status: TestStatusPassed})
	suite.AddTest(&TestCaseResult{Name: "Test 2", Status: TestStatusFailed})
	suite.AddTest(&TestCaseResult{Name: "Test 3", Status: TestStatusPassed})
	suite.AddTest(&TestCaseResult{Name: "Test 4", Status: TestStatusSkipped})
	suite.AddTest(&TestCaseResult{Name: "Test 5", Status: TestStatusError})
	
	results := suite.GetResults()
	
	assert.Equal(t, "Test Suite", results.SuiteName)
	assert.Equal(t, 5, results.TotalTests)
	assert.Equal(t, 2, results.PassedTests)
	assert.Equal(t, 1, results.FailedTests)
	assert.Equal(t, 1, results.SkippedTests)
	assert.Equal(t, 1, results.ErrorTests)
	assert.Len(t, results.TestCases, 5)
	assert.True(t, results.EndTime.After(results.StartTime))
}

func TestTestSuiteExecutor_Integration(t *testing.T) {
	suite := NewTestSuiteExecutor("Integration Test Suite")
	
	// Execute multiple tests
	suite.ExecuteTest("Authentication Test", func(assert *TestAssertion) {
		assert.Log("Testing authentication")
		assert.Equal("success", "success", "login should succeed")
		assert.NotNil("token", "token should be present")
	})
	
	suite.ExecuteTest("Data Validation Test", func(assert *TestAssertion) {
		assert.Log("Testing data validation")
		assert.Len([]int{1, 2, 3}, 3, "should have 3 items")
		assert.True(true, "validation should pass")
	})
	
	suite.ExecuteTest("Error Handling Test", func(assert *TestAssertion) {
		assert.Log("Testing error handling")
		assert.Equal("expected", "actual", "this will fail")
	})
	
	results := suite.GetResults()
	
	assert.Equal(t, "Integration Test Suite", results.SuiteName)
	assert.Equal(t, 3, results.TotalTests)
	assert.Equal(t, 2, results.PassedTests)
	assert.Equal(t, 1, results.FailedTests)
	assert.Equal(t, 0, results.SkippedTests)
	assert.Equal(t, 0, results.ErrorTests)
	
	// Verify test details
	authTest := results.TestCases[0]
	assert.Equal(t, "Authentication Test", authTest.Name)
	assert.Equal(t, TestStatusPassed, authTest.Status)
	assert.Len(t, authTest.Steps, 2)
	
	errorTest := results.TestCases[2]
	assert.Equal(t, "Error Handling Test", errorTest.Name)
	assert.Equal(t, TestStatusFailed, errorTest.Status)
	assert.NotNil(t, errorTest.Error)
}