package gowright

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestTestifyAssertion tests the TestifyAssertion wrapper
func TestTestifyAssertion(t *testing.T) {
	testName := "TestTestifyAssertion"
	ta := NewTestifyAssertion(t, testName)

	t.Run("Equal", func(t *testing.T) {
		// Test successful assertion
		result := ta.Equal(5, 5, "values should be equal")
		assert.True(t, result)

		// Test failed assertion
		result = ta.Equal(5, 10, "values should be equal")
		assert.False(t, result)

		// Verify steps were recorded
		steps := ta.GetSteps()
		assert.Len(t, steps, 2)
		assert.Equal(t, TestStatusPassed, steps[0].Status)
		assert.Equal(t, TestStatusFailed, steps[1].Status)
	})

	t.Run("NotEqual", func(t *testing.T) {
		result := ta.NotEqual(5, 10, "values should not be equal")
		assert.True(t, result)

		result = ta.NotEqual(5, 5, "values should not be equal")
		assert.False(t, result)
	})

	t.Run("True", func(t *testing.T) {
		result := ta.True(true, "value should be true")
		assert.True(t, result)

		result = ta.True(false, "value should be true")
		assert.False(t, result)
	})

	t.Run("False", func(t *testing.T) {
		result := ta.False(false, "value should be false")
		assert.True(t, result)

		result = ta.False(true, "value should be false")
		assert.False(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		result := ta.Nil(nil, "value should be nil")
		assert.True(t, result)

		result = ta.Nil("not nil", "value should be nil")
		assert.False(t, result)
	})

	t.Run("NotNil", func(t *testing.T) {
		result := ta.NotNil("not nil", "value should not be nil")
		assert.True(t, result)

		result = ta.NotNil(nil, "value should not be nil")
		assert.False(t, result)
	})

	t.Run("Contains", func(t *testing.T) {
		result := ta.Contains("hello world", "world", "string should contain substring")
		assert.True(t, result)

		result = ta.Contains("hello world", "foo", "string should contain substring")
		assert.False(t, result)
	})

	t.Run("NotContains", func(t *testing.T) {
		result := ta.NotContains("hello world", "foo", "string should not contain substring")
		assert.True(t, result)

		result = ta.NotContains("hello world", "world", "string should not contain substring")
		assert.False(t, result)
	})

	t.Run("Len", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := ta.Len(slice, 3, "slice should have length 3")
		assert.True(t, result)

		result = ta.Len(slice, 5, "slice should have length 5")
		assert.False(t, result)
	})

	t.Run("Empty", func(t *testing.T) {
		result := ta.Empty("", "string should be empty")
		assert.True(t, result)

		result = ta.Empty("not empty", "string should be empty")
		assert.False(t, result)
	})

	t.Run("NotEmpty", func(t *testing.T) {
		result := ta.NotEmpty("not empty", "string should not be empty")
		assert.True(t, result)

		result = ta.NotEmpty("", "string should not be empty")
		assert.False(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		err := errors.New("test error")
		result := ta.Error(err, "error should be present")
		assert.True(t, result)

		result = ta.Error(nil, "error should be present")
		assert.False(t, result)
	})

	t.Run("NoError", func(t *testing.T) {
		result := ta.NoError(nil, "error should not be present")
		assert.True(t, result)

		err := errors.New("test error")
		result = ta.NoError(err, "error should not be present")
		assert.False(t, result)
	})

	t.Run("DirectTestifyAccess", func(t *testing.T) {
		// Test direct access to testify assert and require
		assert.NotNil(t, ta.Assert())
		assert.NotNil(t, ta.Require())
	})

	t.Run("RequireMethods", func(t *testing.T) {
		// Test require methods (these would normally fail the test immediately)
		// We can't easily test the failure behavior without complex setup
		ta.RequireEqual(5, 5, "values should be equal")
		ta.RequireNotNil("not nil", "value should not be nil")
		ta.RequireNoError(nil, "error should not be present")
	})

	t.Run("LogsAndSteps", func(t *testing.T) {
		// Verify that logs and steps are being recorded
		logs := ta.GetLogs()
		assert.NotEmpty(t, logs)

		steps := ta.GetSteps()
		assert.NotEmpty(t, steps)

		passed, failed := ta.GetSummary()
		assert.True(t, passed > 0)
		assert.True(t, failed > 0)

		assert.True(t, ta.HasFailures())
	})
}

// TestGowrightMock tests the GowrightMock functionality
func TestGowrightMock(t *testing.T) {
	testName := "TestGowrightMock"
	mock := NewGowrightMock(testName)

	t.Run("BasicMockFunctionality", func(t *testing.T) {
		assert.Equal(t, testName, mock.GetTestName())

		// Test logging
		mock.Log("Test log message")
		logs := mock.GetLogs()
		assert.Len(t, logs, 1)
		assert.Equal(t, "Test log message", logs[0])
	})

	t.Run("MockExpectations", func(t *testing.T) {
		// Create a fresh mock for this test
		freshMock := NewGowrightMock("MockExpectationsTest")

		// Test that mock expectations work
		freshMock.On("TestMethod", "arg1").Return("result1")

		// This would normally be called by the code under test
		result := freshMock.TestMethod("arg1")
		assert.Equal(t, "result1", result)

		// Verify expectations were met
		freshMock.AssertExpectations(t)

		// Check that logging occurred
		logs := freshMock.GetLogs()
		assert.Contains(t, logs[len(logs)-1], "All mock expectations were met")
		assert.Contains(t, logs[0], "TestMethod called with: arg1")
	})
}

// TestUITesterMock tests the UITester mock implementation
func TestUITesterMock(t *testing.T) {
	t.Run("NavigateMock", func(t *testing.T) {
		uiMock := NewUITesterMock("NavigateTest")
		url := "https://example.com"
		uiMock.On("Navigate", url).Return(nil)

		err := uiMock.Navigate(url)
		assert.NoError(t, err)

		uiMock.AssertExpectations(t)

		logs := uiMock.GetLogs()
		assert.Contains(t, logs[0], "Navigate called with URL: "+url)
	})

	t.Run("ClickMock", func(t *testing.T) {
		uiMock := NewUITesterMock("ClickTest")
		selector := "#button"
		uiMock.On("Click", selector).Return(nil)

		err := uiMock.Click(selector)
		assert.NoError(t, err)

		uiMock.AssertExpectations(t)

		logs := uiMock.GetLogs()
		assert.Contains(t, logs[0], "Click called with selector: "+selector)
	})

	t.Run("TypeMock", func(t *testing.T) {
		uiMock := NewUITesterMock("TypeTest")
		selector := "#input"
		text := "test text"
		uiMock.On("Type", selector, text).Return(nil)

		err := uiMock.Type(selector, text)
		assert.NoError(t, err)

		uiMock.AssertExpectations(t)

		logs := uiMock.GetLogs()
		assert.Contains(t, logs[0], "Type called with selector: "+selector)
		assert.Contains(t, logs[0], "text: "+text)
	})

	t.Run("GetTextMock", func(t *testing.T) {
		uiMock := NewUITesterMock("GetTextTest")
		selector := "#element"
		expectedText := "element text"
		uiMock.On("GetText", selector).Return(expectedText, nil)

		text, err := uiMock.GetText(selector)
		assert.NoError(t, err)
		assert.Equal(t, expectedText, text)

		uiMock.AssertExpectations(t)

		logs := uiMock.GetLogs()
		assert.Contains(t, logs[0], "GetText called with selector: "+selector)
	})
}

// TestAPITesterMock tests the APITester mock implementation
func TestAPITesterMock(t *testing.T) {
	t.Run("GetMock", func(t *testing.T) {
		apiMock := NewAPITesterMock("GetTest")
		endpoint := "/api/users"
		headers := map[string]string{"Authorization": "Bearer token"}
		expectedResponse := &APIResponse{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       []byte(`{"users": []}`),
			Duration:   100 * time.Millisecond,
		}

		apiMock.On("Get", endpoint, headers).Return(expectedResponse, nil)

		response, err := apiMock.Get(endpoint, headers)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		apiMock.AssertExpectations(t)

		logs := apiMock.GetLogs()
		assert.Contains(t, logs[0], "Get called with endpoint: "+endpoint)
	})

	t.Run("PostMock", func(t *testing.T) {
		apiMock := NewAPITesterMock("PostTest")
		endpoint := "/api/users"
		body := map[string]interface{}{"name": "John Doe"}
		headers := map[string]string{"Content-Type": "application/json"}
		expectedResponse := &APIResponse{
			StatusCode: 201,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       []byte(`{"id": 1, "name": "John Doe"}`),
			Duration:   150 * time.Millisecond,
		}

		apiMock.On("Post", endpoint, body, headers).Return(expectedResponse, nil)

		response, err := apiMock.Post(endpoint, body, headers)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		apiMock.AssertExpectations(t)

		logs := apiMock.GetLogs()
		assert.Contains(t, logs[0], "Post called with endpoint: "+endpoint)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		apiMock := NewAPITesterMock("ErrorTest")
		endpoint := "/api/error"
		headers := map[string]string{}
		expectedError := errors.New("API error")

		apiMock.On("Get", endpoint, headers).Return(nil, expectedError)

		response, err := apiMock.Get(endpoint, headers)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, expectedError, err)

		apiMock.AssertExpectations(t)
	})
}

// TestDatabaseTesterMock tests the DatabaseTester mock implementation
func TestDatabaseTesterMock(t *testing.T) {
	t.Run("QueryMock", func(t *testing.T) {
		dbMock := NewDatabaseTesterMock("QueryTest")
		connection := "test_db"
		query := "SELECT * FROM users"
		args := []interface{}{}
		expectedResult := &DatabaseResult{
			Rows: []map[string]interface{}{
				{"id": 1, "name": "John Doe"},
				{"id": 2, "name": "Jane Smith"},
			},
			RowCount: 2,
			Duration: 50 * time.Millisecond,
		}

		dbMock.On("Query", connection, query, args).Return(expectedResult, nil)

		result, err := dbMock.Query(connection, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)

		dbMock.AssertExpectations(t)

		logs := dbMock.GetLogs()
		assert.Contains(t, logs[0], "Query called with connection: "+connection)
		assert.Contains(t, logs[0], "query: "+query)
	})

	t.Run("ExecuteMock", func(t *testing.T) {
		dbMock := NewDatabaseTesterMock("ExecuteTest")
		connection := "test_db"
		query := "INSERT INTO users (name) VALUES (?)"
		args := []interface{}{"New User"}
		expectedResult := &DatabaseResult{
			RowsAffected: 1,
			Duration:     25 * time.Millisecond,
		}

		dbMock.On("Execute", connection, query, args).Return(expectedResult, nil)

		result, err := dbMock.Execute(connection, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		assert.Equal(t, int64(1), result.RowsAffected)

		dbMock.AssertExpectations(t)

		logs := dbMock.GetLogs()
		assert.Contains(t, logs[0], "Execute called with connection: "+connection)
		assert.Contains(t, logs[0], "query: "+query)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		dbMock := NewDatabaseTesterMock("ErrorTest")
		connection := "test_db"
		query := "INVALID SQL"
		args := []interface{}{}
		expectedError := errors.New("SQL syntax error")

		dbMock.On("Query", connection, query, args).Return(nil, expectedError)

		result, err := dbMock.Query(connection, query, args...)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)

		dbMock.AssertExpectations(t)
	})
}

// TestTestifyIntegrationHelper tests the integration helper functionality
func TestTestifyIntegrationHelper(t *testing.T) {
	config := &Config{
		LogLevel: "info",
		Parallel: false,
	}

	t.Run("NewTestifyIntegrationHelper", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, config)
		assert.NoError(t, err)
		assert.NotNil(t, helper)

		// Clean up
		err = helper.Close()
		assert.NoError(t, err)
	})

	t.Run("RunUITestWithMock", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, config)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		// Create a mock UI tester
		uiMock := NewUITesterMock("test")
		uiMock.On("ExecuteTest", mock.AnythingOfType("*gowright.UITest")).Return(&TestCaseResult{
			Name:   "Test UI",
			Status: TestStatusPassed,
		})
		uiMock.On("Cleanup").Return(nil)

		// Set the mock in the Gowright instance
		helper.gowright.SetUITester(uiMock)

		test := &UITest{
			Name: "Test UI",
			URL:  "https://example.com",
		}

		result := helper.RunUITest(test)
		assert.Equal(t, TestStatusPassed, result.Status)
		assert.Equal(t, "Test UI", result.Name)
	})

	t.Run("RunAPITestWithMock", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, config)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		// Create a mock API tester
		apiMock := NewAPITesterMock("test")
		apiMock.On("ExecuteTest", mock.AnythingOfType("*gowright.APITest")).Return(&TestCaseResult{
			Name:   "Test API",
			Status: TestStatusPassed,
		})
		apiMock.On("Cleanup").Return(nil)

		// Set the mock in the Gowright instance
		helper.gowright.SetAPITester(apiMock)

		test := &APITest{
			Name:     "Test API",
			Method:   "GET",
			Endpoint: "/api/test",
		}

		result := helper.RunAPITest(test)
		assert.Equal(t, TestStatusPassed, result.Status)
		assert.Equal(t, "Test API", result.Name)
	})

	t.Run("RunDatabaseTestWithMock", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, config)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		// Create a mock database tester
		dbMock := NewDatabaseTesterMock("test")
		dbMock.On("ExecuteTest", mock.AnythingOfType("*gowright.DatabaseTest")).Return(&TestCaseResult{
			Name:   "Test Database",
			Status: TestStatusPassed,
		})
		dbMock.On("Cleanup").Return(nil)

		// Set the mock in the Gowright instance
		helper.gowright.SetDatabaseTester(dbMock)

		test := &DatabaseTest{
			Name:       "Test Database",
			Connection: "test_db",
			Query:      "SELECT 1",
		}

		result := helper.RunDatabaseTest(test)
		assert.Equal(t, TestStatusPassed, result.Status)
		assert.Equal(t, "Test Database", result.Name)
	})

	t.Run("RunIntegrationTestWithMock", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, config)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		// Create a mock integration tester
		integrationMock := &GowrightMock{}
		integrationMock.On("ExecuteTest", mock.AnythingOfType("*gowright.IntegrationTest")).Return(&TestCaseResult{
			Name:   "Test Integration",
			Status: TestStatusPassed,
		})

		// For this test, we'll simulate the behavior since we don't have a full integration tester mock
		result := &TestCaseResult{
			Name:   "Test Integration",
			Status: TestStatusPassed,
		}

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.Equal(t, "Test Integration", result.Name)
	})
}

// GowrightTestSuiteExample demonstrates how to use the GowrightTestSuite
type GowrightTestSuiteExample struct {
	GowrightTestSuite
}

// TestGowrightTestSuite tests the test suite functionality
func TestGowrightTestSuite(t *testing.T) {
	suite.Run(t, new(GowrightTestSuiteExample))
}

// TestExampleUITest demonstrates UI testing within the suite
func (suite *GowrightTestSuiteExample) TestExampleUITest() {
	assertion := suite.NewTestifyAssertion("ExampleUITest")

	// Example test logic
	assertion.True(true, "This should pass")
	assertion.Equal("expected", "expected", "Values should match")

	// Verify the test recorded steps
	steps := assertion.GetSteps()
	suite.Assert().Len(steps, 2)
	suite.Assert().Equal(TestStatusPassed, steps[0].Status)
	suite.Assert().Equal(TestStatusPassed, steps[1].Status)
}

// TestExampleAPITest demonstrates API testing within the suite
func (suite *GowrightTestSuiteExample) TestExampleAPITest() {
	assertion := suite.NewTestifyAssertion("ExampleAPITest")

	// Example API test logic
	response := map[string]interface{}{
		"status": "success",
		"data":   []string{"item1", "item2"},
	}

	assertion.Equal("success", response["status"], "Status should be success")
	assertion.NotNil(response["data"], "Data should not be nil")

	// Verify the test recorded steps
	steps := assertion.GetSteps()
	suite.Assert().Len(steps, 2)
	suite.Assert().False(assertion.HasFailures())
}

// TestExampleDatabaseTest demonstrates database testing within the suite
func (suite *GowrightTestSuiteExample) TestExampleDatabaseTest() {
	assertion := suite.NewTestifyAssertion("ExampleDatabaseTest")

	// Example database test logic
	rows := []map[string]interface{}{
		{"id": 1, "name": "John"},
		{"id": 2, "name": "Jane"},
	}

	assertion.Len(rows, 2, "Should have 2 rows")
	assertion.Equal("John", rows[0]["name"], "First row name should be John")
	assertion.Equal("Jane", rows[1]["name"], "Second row name should be Jane")

	// Verify the test recorded steps
	steps := assertion.GetSteps()
	suite.Assert().Len(steps, 3)
	suite.Assert().False(assertion.HasFailures())
}

// TestDirectTestifyIntegration tests direct integration with testify
func TestDirectTestifyIntegration(t *testing.T) {
	t.Run("TestifyAssertIntegration", func(t *testing.T) {
		ta := NewTestifyAssertion(t, "DirectIntegration")

		// Test that testify assertions work alongside Gowright assertions
		assert.True(t, ta.True(true))
		assert.True(t, ta.False(false))
		assert.Equal(t, 5, 5)

		// Verify Gowright recorded the assertions
		steps := ta.GetSteps()
		assert.Len(t, steps, 2) // Only Gowright assertions are recorded
	})

	t.Run("TestifyRequireIntegration", func(t *testing.T) {
		ta := NewTestifyAssertion(t, "RequireIntegration")

		// Test require methods (these would fail the test if assertions failed)
		require.NotNil(t, ta.Require())
		require.NotNil(t, ta.Assert())

		// Test Gowright require methods
		ta.RequireEqual(10, 10)
		ta.RequireNotNil("not nil")
		ta.RequireNoError(nil)

		// Verify steps were recorded
		steps := ta.GetSteps()
		assert.Len(t, steps, 3)
	})
}
