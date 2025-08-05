package gowright

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTest implements the Test interface for testing
type MockTest struct {
	name        string
	result      *TestCaseResult
	executed    bool
	executeFunc func() *TestCaseResult
}

func NewMockTest(name string, status TestStatus, err error) *MockTest {
	return &MockTest{
		name: name,
		result: &TestCaseResult{
			Name:     name,
			Status:   status,
			Duration: 100 * time.Millisecond,
			Error:    err,
		},
	}
}

func (mt *MockTest) GetName() string {
	return mt.name
}

func (mt *MockTest) Execute() *TestCaseResult {
	mt.executed = true
	if mt.executeFunc != nil {
		return mt.executeFunc()
	}
	return mt.result
}

func (mt *MockTest) WasExecuted() bool {
	return mt.executed
}

func TestNewTestSuiteManager(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()

	tsm := NewTestSuiteManager(suite, config)

	require.NotNil(t, tsm)
	assert.Equal(t, suite, tsm.GetTestSuite())
	assert.Equal(t, 0, tsm.GetTestCount())
	assert.Equal(t, "Test Suite", tsm.GetResults().SuiteName)
}

func TestRegisterTest(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	test2 := NewMockTest("test2", TestStatusPassed, nil)

	tsm.RegisterTest(test1)
	assert.Equal(t, 1, tsm.GetTestCount())

	tsm.RegisterTest(test2)
	assert.Equal(t, 2, tsm.GetTestCount())

	names := tsm.GetTestNames()
	assert.Contains(t, names, "test1")
	assert.Contains(t, names, "test2")
}

func TestRegisterTests(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	tests := []Test{
		NewMockTest("test1", TestStatusPassed, nil),
		NewMockTest("test2", TestStatusPassed, nil),
		NewMockTest("test3", TestStatusPassed, nil),
	}

	tsm.RegisterTests(tests)
	assert.Equal(t, 3, tsm.GetTestCount())

	names := tsm.GetTestNames()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "test1")
	assert.Contains(t, names, "test2")
	assert.Contains(t, names, "test3")
}

func TestExecuteTestSuiteSequential(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	config.Parallel = false
	tsm := NewTestSuiteManager(suite, config)

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	test2 := NewMockTest("test2", TestStatusFailed, errors.New("test failed"))
	test3 := NewMockTest("test3", TestStatusPassed, nil)

	tsm.RegisterTests([]Test{test1, test2, test3})

	results, err := tsm.ExecuteTestSuite()
	require.NoError(t, err)
	require.NotNil(t, results)

	assert.Equal(t, "Test Suite", results.SuiteName)
	assert.Equal(t, 3, results.TotalTests)
	assert.Equal(t, 2, results.PassedTests)
	assert.Equal(t, 1, results.FailedTests)
	assert.Equal(t, 0, results.SkippedTests)
	assert.Equal(t, 0, results.ErrorTests)

	// Verify all tests were executed
	assert.True(t, test1.WasExecuted())
	assert.True(t, test2.WasExecuted())
	assert.True(t, test3.WasExecuted())

	// Verify timing
	assert.True(t, results.StartTime.Before(results.EndTime))
}

func TestExecuteTestSuiteParallel(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	config.Parallel = true
	tsm := NewTestSuiteManager(suite, config)

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	test2 := NewMockTest("test2", TestStatusPassed, nil)
	test3 := NewMockTest("test3", TestStatusPassed, nil)

	tsm.RegisterTests([]Test{test1, test2, test3})

	results, err := tsm.ExecuteTestSuite()
	require.NoError(t, err)
	require.NotNil(t, results)

	assert.Equal(t, 3, results.TotalTests)
	assert.Equal(t, 3, results.PassedTests)

	// Verify all tests were executed
	assert.True(t, test1.WasExecuted())
	assert.True(t, test2.WasExecuted())
	assert.True(t, test3.WasExecuted())
}

func TestExecuteTestSuiteWithSetupAndTeardown(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	setupCalled := false
	teardownCalled := false

	tsm.SetSetupFunc(func() error {
		setupCalled = true
		return nil
	})

	tsm.SetTeardownFunc(func() error {
		teardownCalled = true
		return nil
	})

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	tsm.RegisterTest(test1)

	results, err := tsm.ExecuteTestSuite()
	require.NoError(t, err)
	require.NotNil(t, results)

	assert.True(t, setupCalled)
	assert.True(t, teardownCalled)
	assert.True(t, test1.WasExecuted())
}

func TestExecuteTestSuiteWithSetupFailure(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	expectedError := errors.New("setup failed")
	tsm.SetSetupFunc(func() error {
		return expectedError
	})

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	tsm.RegisterTest(test1)

	results, err := tsm.ExecuteTestSuite()
	assert.Error(t, err)
	assert.Nil(t, results)

	// Check that it's a GowrightError
	var gowrightErr *GowrightError
	assert.True(t, errors.As(err, &gowrightErr))
	assert.Equal(t, ConfigurationError, gowrightErr.Type)

	// Test should not have been executed
	assert.False(t, test1.WasExecuted())
}

func TestExecuteTestSuiteWithTeardownFailure(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	teardownCalled := false
	tsm.SetTeardownFunc(func() error {
		teardownCalled = true
		return errors.New("teardown failed")
	})

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	tsm.RegisterTest(test1)

	// Teardown failure should not fail the suite execution
	results, err := tsm.ExecuteTestSuite()
	require.NoError(t, err)
	require.NotNil(t, results)

	assert.True(t, teardownCalled)
	assert.True(t, test1.WasExecuted())
	assert.Equal(t, 1, results.PassedTests)
}

func TestExecuteTestWithRetries(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	config.MaxRetries = 2
	tsm := NewTestSuiteManager(suite, config)

	// Create a test that fails initially but succeeds on retry
	executionCount := 0
	failingTest := &MockTest{
		name: "flaky_test",
		executeFunc: func() *TestCaseResult {
			executionCount++
			if executionCount < 2 {
				return &TestCaseResult{
					Name:     "flaky_test",
					Status:   TestStatusError,
					Duration: 50 * time.Millisecond,
					Error:    errors.New("temporary failure"),
				}
			}
			return &TestCaseResult{
				Name:     "flaky_test",
				Status:   TestStatusPassed,
				Duration: 50 * time.Millisecond,
				Error:    nil,
			}
		},
	}

	tsm.RegisterTest(failingTest)

	results, err := tsm.ExecuteTestSuite()
	require.NoError(t, err)
	require.NotNil(t, results)

	assert.Equal(t, 1, results.TotalTests)
	assert.Equal(t, 1, results.PassedTests)
	assert.Equal(t, 0, results.ErrorTests)
	assert.Equal(t, 2, executionCount) // Should have been executed twice
}

func TestExecuteTestByName(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	test2 := NewMockTest("test2", TestStatusPassed, nil)

	tsm.RegisterTests([]Test{test1, test2})

	// Execute specific test
	result, err := tsm.ExecuteTestByName("test1")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test1", result.Name)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.True(t, test1.WasExecuted())
	assert.False(t, test2.WasExecuted()) // Should not have been executed
}

func TestExecuteTestByNameNotFound(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	tsm.RegisterTest(test1)

	result, err := tsm.ExecuteTestByName("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestClearTests(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	test1 := NewMockTest("test1", TestStatusPassed, nil)
	test2 := NewMockTest("test2", TestStatusPassed, nil)

	tsm.RegisterTests([]Test{test1, test2})
	assert.Equal(t, 2, tsm.GetTestCount())

	tsm.ClearTests()
	assert.Equal(t, 0, tsm.GetTestCount())
	assert.Empty(t, tsm.GetTestNames())
}

func TestTestSuiteConcurrentAccess(t *testing.T) {
	suite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	config := DefaultConfig()
	tsm := NewTestSuiteManager(suite, config)

	// Test concurrent access to test registration and execution
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			test := NewMockTest(fmt.Sprintf("test_%d", i), TestStatusPassed, nil)
			tsm.RegisterTest(test)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			_ = tsm.GetTestCount()
			_ = tsm.GetTestNames()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we get here without deadlock, the test passes
	assert.True(t, tsm.GetTestCount() >= 0)
}
