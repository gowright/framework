package gowright

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewIntegrationTester(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	assert.NotNil(t, tester)
	assert.Equal(t, "IntegrationTester", tester.GetName())
	assert.Equal(t, mockUI, tester.uiTester)
	assert.Equal(t, mockAPI, tester.apiTester)
	assert.Equal(t, mockDB, tester.dbTester)
	assert.NotNil(t, tester.config)
	assert.Equal(t, 3, tester.config.MaxRetries)
	assert.Equal(t, 1*time.Second, tester.config.RetryDelay)
	assert.True(t, tester.config.RollbackOnError)
	assert.False(t, tester.config.ParallelSteps)
}

func TestIntegrationTester_Initialize(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Initialize", mock.Anything).Return(nil)
	mockAPI.On("Initialize", mock.Anything).Return(nil)
	mockDB.On("Initialize", mock.Anything).Return(nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	err := tester.Initialize(nil)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestIntegrationTester_Initialize_WithConfig(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Initialize", mock.Anything).Return(nil)
	mockAPI.On("Initialize", mock.Anything).Return(nil)
	mockDB.On("Initialize", mock.Anything).Return(nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	config := &IntegrationConfig{
		MaxRetries:      5,
		RetryDelay:      2 * time.Second,
		RollbackOnError: false,
		ParallelSteps:   true,
	}

	err := tester.Initialize(config)
	assert.NoError(t, err)
	assert.Equal(t, 5, tester.config.MaxRetries)
	assert.Equal(t, 2*time.Second, tester.config.RetryDelay)
	assert.False(t, tester.config.RollbackOnError)
	assert.True(t, tester.config.ParallelSteps)
}

func TestIntegrationTester_Initialize_UITesterError(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Initialize", mock.Anything).Return(errors.New("UI initialization failed"))

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	err := tester.Initialize(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize UI tester")

	mockUI.AssertExpectations(t)
}

func TestIntegrationTester_Cleanup(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Cleanup").Return(nil)
	mockAPI.On("Cleanup").Return(nil)
	mockDB.On("Cleanup").Return(nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	err := tester.Cleanup()
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestIntegrationTester_Cleanup_WithErrors(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Cleanup").Return(errors.New("UI cleanup failed"))
	mockAPI.On("Cleanup").Return(nil)
	mockDB.On("Cleanup").Return(errors.New("DB cleanup failed"))

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	err := tester.Cleanup()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup failed for one or more testers")

	mockUI.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_UIStep(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Navigate", "https://example.com").Return(nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	step := &IntegrationStep{
		Type: StepTypeUI,
		Name: "Navigate to homepage",
		Action: &UIStepAction{
			Type: "navigate",
			Parameters: map[string]interface{}{
				"url": "https://example.com",
			},
		},
	}

	err := tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_APIStep(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	expectedResponse := &APIResponse{
		StatusCode: 200,
		Body:       []byte(`{"status": "ok"}`),
	}

	mockAPI.On("Get", "/api/users", map[string]string{"Authorization": "Bearer token"}).Return(expectedResponse, nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	step := &IntegrationStep{
		Type: StepTypeAPI,
		Name: "Get users",
		Action: &APIStepAction{
			Method:   "GET",
			Endpoint: "/api/users",
			Headers:  map[string]string{"Authorization": "Bearer token"},
		},
	}

	err := tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockAPI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_DatabaseStep(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	expectedResult := &DatabaseResult{
		Rows:         []map[string]interface{}{{"id": 1, "name": "John"}},
		RowsAffected: 1,
	}

	mockDB.On("Execute", "main", "SELECT * FROM users WHERE id = ?", []interface{}{1}).Return(expectedResult, nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	step := &IntegrationStep{
		Type: StepTypeDatabase,
		Name: "Get user by ID",
		Action: &DatabaseStepAction{
			Connection: "main",
			Query:      "SELECT * FROM users WHERE id = ?",
			Args:       []interface{}{1},
		},
	}

	err := tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_NilStep(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	err := tester.ExecuteStep(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "integration step cannot be nil")
}

func TestIntegrationTester_ExecuteStep_UnsupportedStepType(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	step := &IntegrationStep{
		Type: StepType(999), // Invalid step type
		Name: "Invalid step",
	}

	err := tester.ExecuteStep(step)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported step type")
}

func TestIntegrationTester_ExecuteStep_WithRetries(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	// First call fails, second call succeeds
	mockUI.On("Navigate", "https://example.com").Return(errors.New("network error")).Once()
	mockUI.On("Navigate", "https://example.com").Return(nil).Once()

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)
	tester.config.RetryDelay = 10 * time.Millisecond // Speed up test

	step := &IntegrationStep{
		Type: StepTypeUI,
		Name: "Navigate to homepage",
		Action: &UIStepAction{
			Type: "navigate",
			Parameters: map[string]interface{}{
				"url": "https://example.com",
			},
		},
	}

	err := tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteWorkflow(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Navigate", "https://example.com").Return(nil)
	mockAPI.On("Get", "/api/status", map[string]string(nil)).Return(&APIResponse{StatusCode: 200}, nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	steps := []IntegrationStep{
		{
			Type: StepTypeUI,
			Name: "Navigate to homepage",
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
	}

	err := tester.ExecuteWorkflow(steps)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteWorkflow_EmptySteps(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	err := tester.ExecuteWorkflow([]IntegrationStep{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow steps cannot be empty")
}

func TestIntegrationTester_ExecuteWorkflow_WithRollback(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("Navigate", "https://example.com").Return(nil)
	mockAPI.On("Get", "/api/status", map[string]string(nil)).Return((*APIResponse)(nil), errors.New("API error"))
	mockUI.On("TakeScreenshot", mock.AnythingOfType("string")).Return("screenshot.png", nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	steps := []IntegrationStep{
		{
			Type: StepTypeUI,
			Name: "Navigate to homepage",
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
	}

	err := tester.ExecuteWorkflow(steps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow failed at step 2")

	mockUI.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
}

func TestIntegrationTester_Rollback(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	mockUI.On("TakeScreenshot", mock.AnythingOfType("string")).Return("rollback_screenshot.png", nil)

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	steps := []IntegrationStep{
		{
			Type: StepTypeUI,
			Name: "Navigate to homepage",
			Action: &UIStepAction{
				Type: "navigate",
				Parameters: map[string]interface{}{
					"url": "https://example.com",
				},
			},
		},
	}

	err := tester.Rollback(steps)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
}

func TestIntegrationTester_validateAPIResponse(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	response := &APIResponse{
		StatusCode: 200,
		Body:       []byte(`{"status": "ok"}`),
	}

	validation := &APIStepValidation{
		ExpectedStatusCode: 200,
	}

	err := tester.validateAPIResponse(response, validation)
	assert.NoError(t, err)
}

func TestIntegrationTester_validateAPIResponse_StatusCodeMismatch(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	response := &APIResponse{
		StatusCode: 404,
		Body:       []byte(`{"error": "not found"}`),
	}

	validation := &APIStepValidation{
		ExpectedStatusCode: 200,
	}

	err := tester.validateAPIResponse(response, validation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected status code 200, got 404")
}

func TestIntegrationTester_validateDatabaseResult(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	result := &DatabaseResult{
		Rows: []map[string]interface{}{
			{"id": 1, "name": "John"},
			{"id": 2, "name": "Jane"},
		},
	}

	expectedRowCount := 2
	validation := &DatabaseStepValidation{
		ExpectedRowCount: &expectedRowCount,
	}

	err := tester.validateDatabaseResult(result, validation)
	assert.NoError(t, err)
}

func TestIntegrationTester_validateDatabaseResult_RowCountMismatch(t *testing.T) {
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester := NewIntegrationTester(mockUI, mockAPI, mockDB)

	result := &DatabaseResult{
		Rows: []map[string]interface{}{
			{"id": 1, "name": "John"},
		},
	}

	expectedRowCount := 2
	validation := &DatabaseStepValidation{
		ExpectedRowCount: &expectedRowCount,
	}

	err := tester.validateDatabaseResult(result, validation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 2 rows, got 1")
}