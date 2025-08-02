package gowright

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIntegrationTester is a mock implementation of IntegrationTester
type MockIntegrationTester struct {
	mock.Mock
}

func (m *MockIntegrationTester) Initialize(config interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockIntegrationTester) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockIntegrationTester) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIntegrationTester) ExecuteStep(step *IntegrationStep) error {
	args := m.Called(step)
	return args.Error(0)
}

func (m *MockIntegrationTester) ExecuteWorkflow(steps []IntegrationStep) error {
	args := m.Called(steps)
	return args.Error(0)
}

func (m *MockIntegrationTester) Rollback(steps []IntegrationStep) error {
	args := m.Called(steps)
	return args.Error(0)
}

func (m *MockIntegrationTester) ExecuteTest(test *IntegrationTest) *TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*TestCaseResult)
}

func TestNewIntegrationTestImpl(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Test Integration",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Navigate to page",
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	
	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	assert.NotNil(t, impl)
	assert.Equal(t, "Test Integration", impl.GetName())
	assert.Equal(t, integrationTest, impl.integrationTest)
	assert.Equal(t, mockTester, impl.tester)
	assert.NotNil(t, impl.setupData)
	assert.NotNil(t, impl.teardownData)
	assert.NotNil(t, impl.failureContext)
}

func TestIntegrationTestImpl_Execute_Success(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Successful Integration Test",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Navigate to page",
				Action: &UIStepAction{
					Type: "navigate",
					Parameters: map[string]interface{}{
						"url": "https://example.com",
					},
				},
			},
			{
				Type: StepTypeAPI,
				Name: "Check API status",
				Action: &APIStepAction{
					Method:   "GET",
					Endpoint: "/api/status",
				},
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(nil)
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.Execute()
	
	assert.NotNil(t, result)
	assert.Equal(t, "Successful Integration Test", result.Name)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.NoError(t, result.Error)
	assert.True(t, result.Duration > 0)
	assert.False(t, result.StartTime.IsZero())
	assert.False(t, result.EndTime.IsZero())

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_Execute_SetupFailure(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name:  "Setup Failure Test",
		Steps: []IntegrationStep{},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(errors.New("setup failed"))

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.Execute()
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusError, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "setup phase failed")

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_Execute_StepFailure(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Step Failure Test",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Failing step",
				Action: &UIStepAction{
					Type: "navigate",
					Parameters: map[string]interface{}{
						"url": "https://example.com",
					},
				},
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(errors.New("step failed"))
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.Execute()
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusFailed, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "integration test failed at step 1")

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_Execute_TeardownFailure(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Teardown Failure Test",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Successful step",
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(nil)
	mockTester.On("Cleanup").Return(errors.New("cleanup failed"))

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.Execute()
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusError, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "teardown phase failed")

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_Execute_WithRollback(t *testing.T) {
	rollbackSteps := []IntegrationStep{
		{
			Type: StepTypeDatabase,
			Name: "Rollback database changes",
		},
	}

	integrationTest := &IntegrationTest{
		Name: "Test with Rollback",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Successful step",
			},
		},
		Rollback: rollbackSteps,
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(nil)
	mockTester.On("Rollback", rollbackSteps).Return(nil)
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.Execute()
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.NoError(t, result.Error)

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_ExecuteWithSetupAndTeardown_Success(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Test with Custom Setup/Teardown",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Main test step",
			},
		},
	}

	setup := &TestDataSetup{
		UISetup: []UISetupAction{
			{
				Type: "navigate",
				Parameters: map[string]interface{}{
					"url": "https://setup.example.com",
				},
			},
		},
		APISetup: []APISetupAction{
			{
				Method:   "POST",
				Endpoint: "/api/setup",
				Body:     map[string]string{"action": "initialize"},
			},
		},
		DatabaseSetup: []DatabaseSetupAction{
			{
				Connection: "test_db",
				Query:      "INSERT INTO test_data (name) VALUES (?)",
				Args:       []interface{}{"test_value"},
			},
		},
	}

	teardown := &TestDataTeardown{
		DatabaseTeardown: []DatabaseTeardownAction{
			{
				Connection: "test_db",
				Query:      "DELETE FROM test_data WHERE name = ?",
				Args:       []interface{}{"test_value"},
			},
		},
		APITeardown: []APITeardownAction{
			{
				Method:   "POST",
				Endpoint: "/api/cleanup",
				Body:     map[string]string{"action": "cleanup"},
			},
		},
		UITeardown: []UITeardownAction{
			{
				Type: "navigate",
				Parameters: map[string]interface{}{
					"url": "about:blank",
				},
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(nil)
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.ExecuteWithSetupAndTeardown(setup, teardown)
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.NoError(t, result.Error)

	// Verify that setup, main steps, and teardown were all executed
	// The mock should have been called for:
	// - 1 UI setup step
	// - 1 API setup step  
	// - 1 Database setup step
	// - 1 main test step
	// - 1 Database teardown step
	// - 1 API teardown step
	// - 1 UI teardown step
	// Total: 7 ExecuteStep calls
	mockTester.AssertNumberOfCalls(t, "ExecuteStep", 7)
	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_ExecuteWithSetupAndTeardown_SetupFailure(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name:  "Test with Setup Failure",
		Steps: []IntegrationStep{},
	}

	setup := &TestDataSetup{
		UISetup: []UISetupAction{
			{
				Type: "navigate",
				Parameters: map[string]interface{}{
					"url": "https://setup.example.com",
				},
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(errors.New("setup step failed"))

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.ExecuteWithSetupAndTeardown(setup, nil)
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusError, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "custom setup phase failed")

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_SetupAndTeardownData(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name:  "Data Management Test",
		Steps: []IntegrationStep{},
	}

	mockTester := &MockIntegrationTester{}
	impl := NewIntegrationTestImpl(integrationTest, mockTester)

	// Test setup data management
	impl.SetSetupData("user_id", 12345)
	impl.SetSetupData("session_token", "abc123")

	userID, exists := impl.GetSetupData("user_id")
	assert.True(t, exists)
	assert.Equal(t, 12345, userID)

	token, exists := impl.GetSetupData("session_token")
	assert.True(t, exists)
	assert.Equal(t, "abc123", token)

	_, exists = impl.GetSetupData("nonexistent")
	assert.False(t, exists)

	// Test teardown data management
	impl.SetTeardownData("cleanup_id", "cleanup_123")
	
	cleanupID, exists := impl.GetTeardownData("cleanup_id")
	assert.True(t, exists)
	assert.Equal(t, "cleanup_123", cleanupID)

	_, exists = impl.GetTeardownData("nonexistent")
	assert.False(t, exists)
}

func TestIntegrationTestImpl_FailureContext(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Failure Context Test",
		Steps: []IntegrationStep{
			{
				Type: StepTypeAPI,
				Name: "API step that fails",
				Action: &APIStepAction{
					Method:   "GET",
					Endpoint: "/api/test",
				},
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(errors.New("API call failed"))
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.Execute()
	
	assert.Equal(t, TestStatusFailed, result.Status)
	
	failureContext := impl.GetFailureContext()
	assert.NotNil(t, failureContext)
	assert.NotNil(t, failureContext.FailedStep)
	assert.Equal(t, "API step that fails", failureContext.FailedStep.Name)
	assert.Equal(t, 0, failureContext.StepIndex)
	assert.Error(t, failureContext.Error)
	assert.NotNil(t, failureContext.SystemStates)
	assert.False(t, failureContext.Timestamp.IsZero())
	assert.NotEmpty(t, failureContext.Logs)

	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_CollectSystemStates(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name:  "System State Collection Test",
		Steps: []IntegrationStep{},
	}

	mockTester := &MockIntegrationTester{}
	impl := NewIntegrationTestImpl(integrationTest, mockTester)

	// Test API state collection
	apiStep := &IntegrationStep{
		Type: StepTypeAPI,
		Name: "API step",
		Action: &APIStepAction{
			Method:   "POST",
			Endpoint: "/api/test",
			Headers:  map[string]string{"Authorization": "Bearer token"},
			Body:     map[string]string{"test": "data"},
		},
	}

	impl.collectSystemStates(apiStep)
	
	failureContext := impl.GetFailureContext()
	apiAction, exists := failureContext.SystemStates["failed_api_action"]
	assert.True(t, exists)
	
	apiActionMap := apiAction.(map[string]interface{})
	assert.Equal(t, "POST", apiActionMap["method"])
	assert.Equal(t, "/api/test", apiActionMap["endpoint"])

	// Test Database state collection
	dbStep := &IntegrationStep{
		Type: StepTypeDatabase,
		Name: "Database step",
		Action: &DatabaseStepAction{
			Connection: "test_db",
			Query:      "SELECT * FROM users WHERE id = ?",
			Args:       []interface{}{123},
		},
	}

	impl.collectSystemStates(dbStep)
	
	dbAction, exists := failureContext.SystemStates["failed_database_action"]
	assert.True(t, exists)
	
	dbActionMap := dbAction.(map[string]interface{})
	assert.Equal(t, "test_db", dbActionMap["connection"])
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", dbActionMap["query"])
}

func TestIntegrationTestImpl_AddLog(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name:  "Log Test",
		Steps: []IntegrationStep{},
	}

	mockTester := &MockIntegrationTester{}
	impl := NewIntegrationTestImpl(integrationTest, mockTester)

	impl.addLog("Test log message 1")
	impl.addLog("Test log message 2")

	failureContext := impl.GetFailureContext()
	assert.Len(t, failureContext.Logs, 2)
	assert.Contains(t, failureContext.Logs[0], "Test log message 1")
	assert.Contains(t, failureContext.Logs[1], "Test log message 2")
	
	// Check that timestamps are included
	assert.Contains(t, failureContext.Logs[0], "[")
	assert.Contains(t, failureContext.Logs[0], "]")
}

func TestIntegrationTestImpl_ExecuteCustomSetupTeardown_EmptyOperations(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Empty Operations Test",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Main step",
			},
		},
	}

	setup := &TestDataSetup{
		UISetup:       []UISetupAction{},
		APISetup:      []APISetupAction{},
		DatabaseSetup: []DatabaseSetupAction{},
	}

	teardown := &TestDataTeardown{
		UITeardown:       []UITeardownAction{},
		APITeardown:      []APITeardownAction{},
		DatabaseTeardown: []DatabaseTeardownAction{},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", mock.Anything).Return(nil)
	mockTester.On("ExecuteStep", mock.AnythingOfType("*gowright.IntegrationStep")).Return(nil)
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.ExecuteWithSetupAndTeardown(setup, teardown)
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.NoError(t, result.Error)

	// Should only execute the main test step (1 call)
	mockTester.AssertNumberOfCalls(t, "ExecuteStep", 1)
	mockTester.AssertExpectations(t)
}

func TestIntegrationTestImpl_ExecuteWithSetupAndTeardown_TeardownFailure(t *testing.T) {
	integrationTest := &IntegrationTest{
		Name: "Teardown Failure Test",
		Steps: []IntegrationStep{
			{
				Type: StepTypeUI,
				Name: "Main step",
			},
		},
	}

	teardown := &TestDataTeardown{
		DatabaseTeardown: []DatabaseTeardownAction{
			{
				Connection: "test_db",
				Query:      "DELETE FROM test_data",
			},
		},
	}

	mockTester := &MockIntegrationTester{}
	mockTester.On("Initialize", nil).Return(nil)
	// Main step succeeds
	mockTester.On("ExecuteStep", mock.MatchedBy(func(step *IntegrationStep) bool {
		return step.Name == "Main step"
	})).Return(nil)
	// Teardown step fails
	mockTester.On("ExecuteStep", mock.MatchedBy(func(step *IntegrationStep) bool {
		return step.Name == "Database Teardown: test_db"
	})).Return(errors.New("teardown failed"))
	// Cleanup should still be called even if teardown fails
	mockTester.On("Cleanup").Return(nil)

	impl := NewIntegrationTestImpl(integrationTest, mockTester)
	
	result := impl.ExecuteWithSetupAndTeardown(nil, teardown)
	
	assert.NotNil(t, result)
	assert.Equal(t, TestStatusError, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "custom teardown phase failed")

	mockTester.AssertExpectations(t)
}