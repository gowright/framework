package testify

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
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
		assert.Equal(t, core.TestStatusPassed, steps[0].Status)
		assert.Equal(t, core.TestStatusFailed, steps[1].Status)
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

	t.Run("StepsAndFailures", func(t *testing.T) {
		// Verify that steps are being recorded
		steps := ta.GetSteps()
		assert.NotEmpty(t, steps)

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

		// Actually call the mocked method using MethodCalled
		result := freshMock.MethodCalled("TestMethod", "arg1")
		assert.Equal(t, "result1", result.String(0))

		// Verify expectations were met
		freshMock.AssertExpectations(t)

		// Check that logging occurred
		logs := freshMock.GetLogs()
		assert.Contains(t, logs[len(logs)-1], "All mock expectations were met")
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

// TestTestifyIntegrationHelper tests the integration helper functionality
func TestTestifyIntegrationHelper(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Run("NewTestifyIntegrationHelper", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, helper)

		// Clean up
		err = helper.Close()
		assert.NoError(t, err)
	})

	t.Run("RunUITestWithMock", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, cfg)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		// Create a mock UI tester
		uiMock := NewUITesterMock("test")
		uiMock.On("ExecuteTest", mock.AnythingOfType("*core.UITest")).Return(&core.TestCaseResult{
			Name:   "Test UI",
			Status: core.TestStatusPassed,
		})
		uiMock.On("Cleanup").Return(nil)

		test := &core.UITest{
			Name: "Test UI",
			URL:  "https://example.com",
			Actions: []core.UIAction{
				{
					Type: "navigate",
				},
			},
		}

		result := helper.RunUITest(test)
		assert.Equal(t, core.TestStatusPassed, result.Status)
		assert.Equal(t, "Test UI", result.Name)
	})

	t.Run("RunAPITestWithMock", func(t *testing.T) {
		// Create a config with proper API base URL
		apiConfig := config.DefaultConfig()
		apiConfig.APIConfig.BaseURL = "https://httpbin.org"

		helper, err := NewTestifyIntegrationHelper(t, apiConfig)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		test := &core.APITest{
			Name:     "Test API",
			Method:   "GET",
			Endpoint: "/get",
		}

		result := helper.RunAPITest(test)
		assert.Equal(t, "Test API", result.Name)
	})

	t.Run("RunDatabaseTestWithMock", func(t *testing.T) {
		// Create a config with a test database connection
		dbConfig := config.DefaultConfig()
		dbConfig.DatabaseConfig.Connections = map[string]*config.DatabaseConnection{
			"test_db": {
				Driver:   "sqlite3",
				Database: ":memory:",
			},
		}

		helper, err := NewTestifyIntegrationHelper(t, dbConfig)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		test := &core.DatabaseTest{
			Name:       "Test Database",
			Connection: "test_db",
			Query:      "SELECT 1",
		}

		result := helper.RunDatabaseTest(test)
		assert.Equal(t, "Test Database", result.Name)
	})

	t.Run("RunIntegrationTestWithMock", func(t *testing.T) {
		helper, err := NewTestifyIntegrationHelper(t, cfg)
		require.NoError(t, err)
		defer func() { _ = helper.Close() }()

		test := &core.IntegrationTest{
			Name: "Test Integration",
			Steps: []core.IntegrationStep{
				{
					Type: core.StepTypeUI,
					Name: "Navigate",
					Action: &core.UIStepAction{
						Type: "navigate",
						Parameters: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
		}

		result := helper.RunIntegrationTest(test)
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
	suite.Assert().Equal(core.TestStatusPassed, steps[0].Status)
	suite.Assert().Equal(core.TestStatusPassed, steps[1].Status)
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
