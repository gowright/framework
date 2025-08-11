package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUITester is a mock implementation for testing
type MockUITester struct {
	mock.Mock
}

func (m *MockUITester) Initialize(cfg interface{}) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockUITester) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUITester) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUITester) Navigate(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockUITester) Click(selector string) error {
	args := m.Called(selector)
	return args.Error(0)
}

func (m *MockUITester) Type(selector, text string) error {
	args := m.Called(selector, text)
	return args.Error(0)
}

func (m *MockUITester) GetText(selector string) (string, error) {
	args := m.Called(selector)
	return args.String(0), args.Error(1)
}

func (m *MockUITester) WaitForElement(selector string, timeout time.Duration) error {
	args := m.Called(selector, timeout)
	return args.Error(0)
}

func (m *MockUITester) TakeScreenshot(filename string) (string, error) {
	args := m.Called(filename)
	return args.String(0), args.Error(1)
}

func (m *MockUITester) GetPageSource() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockUITester) ExecuteTest(test *core.UITest) *core.TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*core.TestCaseResult)
}

// MockAPITester is a mock implementation for testing
type MockAPITester struct {
	mock.Mock
}

func (m *MockAPITester) Initialize(cfg interface{}) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockAPITester) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAPITester) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAPITester) Get(endpoint string, headers map[string]string) (*core.APIResponse, error) {
	args := m.Called(endpoint, headers)
	return args.Get(0).(*core.APIResponse), args.Error(1)
}

func (m *MockAPITester) Post(endpoint string, body interface{}, headers map[string]string) (*core.APIResponse, error) {
	args := m.Called(endpoint, body, headers)
	return args.Get(0).(*core.APIResponse), args.Error(1)
}

func (m *MockAPITester) Put(endpoint string, body interface{}, headers map[string]string) (*core.APIResponse, error) {
	args := m.Called(endpoint, body, headers)
	return args.Get(0).(*core.APIResponse), args.Error(1)
}

func (m *MockAPITester) Delete(endpoint string, headers map[string]string) (*core.APIResponse, error) {
	args := m.Called(endpoint, headers)
	return args.Get(0).(*core.APIResponse), args.Error(1)
}

func (m *MockAPITester) SetAuth(auth *config.AuthConfig) error {
	args := m.Called(auth)
	return args.Error(0)
}

func (m *MockAPITester) ExecuteTest(test *core.APITest) *core.TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*core.TestCaseResult)
}

// MockDatabaseTester is a mock implementation for testing
type MockDatabaseTester struct {
	mock.Mock
}

func (m *MockDatabaseTester) Initialize(cfg interface{}) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockDatabaseTester) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabaseTester) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDatabaseTester) Connect(connectionName string) error {
	args := m.Called(connectionName)
	return args.Error(0)
}

func (m *MockDatabaseTester) Execute(connectionName, query string, args ...interface{}) (*core.DatabaseResult, error) {
	mockArgs := m.Called(connectionName, query, args)
	return mockArgs.Get(0).(*core.DatabaseResult), mockArgs.Error(1)
}

func (m *MockDatabaseTester) BeginTransaction(connectionName string) (core.Transaction, error) {
	args := m.Called(connectionName)
	return args.Get(0).(core.Transaction), args.Error(1)
}

func (m *MockDatabaseTester) ValidateData(connectionName, query string, expected interface{}) error {
	args := m.Called(connectionName, query, expected)
	return args.Error(0)
}

func (m *MockDatabaseTester) ExecuteTest(test *core.DatabaseTest) *core.TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*core.TestCaseResult)
}

func TestNewIntegrationTester(t *testing.T) {
	tester := NewIntegrationTester()

	assert.NotNil(t, tester)
	assert.Equal(t, "IntegrationTester", tester.GetName())
	assert.False(t, tester.initialized)
}

func TestIntegrationTester_Initialize(t *testing.T) {
	tester := NewIntegrationTester()

	t.Run("with valid config", func(t *testing.T) {
		config := &config.Config{
			Parallel:   true,
			MaxWorkers: 4,
		}

		err := tester.Initialize(config)
		assert.NoError(t, err)
		assert.True(t, tester.initialized)
		assert.Equal(t, config, tester.config)
	})

	t.Run("with invalid config type", func(t *testing.T) {
		err := tester.Initialize("invalid")
		assert.Error(t, err)
		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.ConfigurationError, gowrightErr.Type)
	})
}

func TestIntegrationTester_SetTesters(t *testing.T) {
	tester := NewIntegrationTester()
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	mockDB := &MockDatabaseTester{}

	tester.SetUITester(mockUI)
	tester.SetAPITester(mockAPI)
	tester.SetDatabaseTester(mockDB)

	assert.Equal(t, mockUI, tester.uiTester)
	assert.Equal(t, mockAPI, tester.apiTester)
	assert.Equal(t, mockDB, tester.dbTester)
}

func TestIntegrationTester_ExecuteStep_UIStep(t *testing.T) {
	tester := NewIntegrationTester()
	mockUI := &MockUITester{}
	tester.SetUITester(mockUI)

	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	mockUI.On("Navigate", "https://example.com").Return(nil)

	step := &core.IntegrationStep{
		Type: core.StepTypeUI,
		Name: "Navigate to homepage",
		Action: &core.UIStepAction{
			Type: "navigate",
			Parameters: map[string]interface{}{
				"url": "https://example.com",
			},
		},
	}

	err = tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_APIStep(t *testing.T) {
	tester := NewIntegrationTester()
	mockAPI := &MockAPITester{}
	tester.SetAPITester(mockAPI)

	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	expectedResponse := &core.APIResponse{
		StatusCode: 200,
		Body:       []byte(`{"status": "ok"}`),
	}

	mockAPI.On("Get", "/api/users", map[string]string{"Authorization": "Bearer token"}).Return(expectedResponse, nil)

	step := &core.IntegrationStep{
		Type: core.StepTypeAPI,
		Name: "Get users",
		Action: &core.APIStepAction{
			Method:   "GET",
			Endpoint: "/api/users",
			Headers:  map[string]string{"Authorization": "Bearer token"},
		},
	}

	err = tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockAPI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_DatabaseStep(t *testing.T) {
	tester := NewIntegrationTester()
	mockDB := &MockDatabaseTester{}
	tester.SetDatabaseTester(mockDB)

	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	expectedResult := &core.DatabaseResult{
		Rows:         []map[string]interface{}{{"id": 1, "name": "John"}},
		RowsAffected: 1,
	}

	mockDB.On("Execute", "main", "SELECT * FROM users WHERE id = ?", []interface{}{1}).Return(expectedResult, nil)

	step := &core.IntegrationStep{
		Type: core.StepTypeDatabase,
		Name: "Get user by ID",
		Action: &core.DatabaseStepAction{
			Connection: "main",
			Query:      "SELECT * FROM users WHERE id = ?",
			Args:       []interface{}{1},
		},
	}

	err = tester.ExecuteStep(step)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteStep_NotInitialized(t *testing.T) {
	tester := NewIntegrationTester()

	step := &core.IntegrationStep{
		Type: core.StepTypeUI,
		Name: "Test step",
	}

	err := tester.ExecuteStep(step)
	assert.Error(t, err)
	gowrightErr, ok := err.(*core.GowrightError)
	assert.True(t, ok)
	assert.Equal(t, core.ConfigurationError, gowrightErr.Type)
}

func TestIntegrationTester_ExecuteStep_UnsupportedStepType(t *testing.T) {
	tester := NewIntegrationTester()
	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	step := &core.IntegrationStep{
		Type: core.IntegrationStepType(999), // Invalid step type
		Name: "Invalid step",
	}

	err = tester.ExecuteStep(step)
	assert.Error(t, err)
	gowrightErr, ok := err.(*core.GowrightError)
	assert.True(t, ok)
	assert.Equal(t, core.ConfigurationError, gowrightErr.Type)
}

func TestIntegrationTester_ExecuteWorkflow(t *testing.T) {
	tester := NewIntegrationTester()
	mockUI := &MockUITester{}
	mockAPI := &MockAPITester{}
	tester.SetUITester(mockUI)
	tester.SetAPITester(mockAPI)

	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	mockUI.On("Navigate", "https://example.com").Return(nil)
	mockAPI.On("Get", "/api/status", map[string]string(nil)).Return(&core.APIResponse{StatusCode: 200}, nil)

	steps := []core.IntegrationStep{
		{
			Type: core.StepTypeUI,
			Name: "Navigate to homepage",
			Action: &core.UIStepAction{
				Type: "navigate",
				Parameters: map[string]interface{}{
					"url": "https://example.com",
				},
			},
		},
		{
			Type: core.StepTypeAPI,
			Name: "Check API status",
			Action: &core.APIStepAction{
				Method:   "GET",
				Endpoint: "/api/status",
			},
		},
	}

	err = tester.ExecuteWorkflow(steps)
	assert.NoError(t, err)

	mockUI.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
}

func TestIntegrationTester_ExecuteTest(t *testing.T) {
	tester := NewIntegrationTester()
	mockUI := &MockUITester{}
	tester.SetUITester(mockUI)

	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	mockUI.On("Navigate", "https://example.com").Return(nil)

	test := &core.IntegrationTest{
		Name: "User Registration Flow",
		Steps: []core.IntegrationStep{
			{
				Name: "Navigate to registration page",
				Type: core.StepTypeUI,
				Action: &core.UIStepAction{
					Type: "navigate",
					Parameters: map[string]interface{}{
						"url": "https://example.com",
					},
				},
			},
		},
	}

	result := tester.ExecuteTest(test)

	assert.NotNil(t, result)
	assert.Equal(t, "User Registration Flow", result.Name)
	assert.Equal(t, core.TestStatusPassed, result.Status)
	assert.NotZero(t, result.Duration)

	mockUI.AssertExpectations(t)
}

func TestIntegrationTester_ValidateAPIResponse(t *testing.T) {
	tester := NewIntegrationTester()
	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	response := &core.APIResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"status": "ok"}`),
	}

	validation := &core.APIStepValidation{
		ExpectedStatusCode: 200,
		ExpectedHeaders: map[string]string{
			"Content-Type": "application/json",
		},
	}

	err = tester.validateAPIResponse(response, validation)
	assert.NoError(t, err)
	assert.False(t, tester.asserter.HasFailures())
}

func TestIntegrationTester_ValidateDatabaseResult(t *testing.T) {
	tester := NewIntegrationTester()
	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	result := &core.DatabaseResult{
		Rows: []map[string]interface{}{
			{"id": 1, "name": "John"},
			{"id": 2, "name": "Jane"},
		},
		RowCount:     2,
		RowsAffected: 0,
	}

	expectedRowCount := 2
	validation := &core.DatabaseStepValidation{
		ExpectedRowCount: &expectedRowCount,
	}

	err = tester.validateDatabaseResult(result, validation)
	assert.NoError(t, err)
	assert.False(t, tester.asserter.HasFailures())
}

func TestIntegrationTester_Cleanup(t *testing.T) {
	tester := NewIntegrationTester()
	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)
	assert.True(t, tester.initialized)

	err = tester.Cleanup()
	assert.NoError(t, err)
	assert.False(t, tester.initialized)
}

func TestIntegrationTester_ErrorHandling(t *testing.T) {
	tester := NewIntegrationTester()
	mockUI := &MockUITester{}
	tester.SetUITester(mockUI)

	config := &config.Config{}
	err := tester.Initialize(config)
	assert.NoError(t, err)

	// Test step execution failure
	mockUI.On("Navigate", "https://example.com").Return(errors.New("navigation failed"))

	step := &core.IntegrationStep{
		Type: core.StepTypeUI,
		Name: "Navigate to homepage",
		Action: &core.UIStepAction{
			Type: "navigate",
			Parameters: map[string]interface{}{
				"url": "https://example.com",
			},
		},
	}

	err = tester.ExecuteStep(step)
	assert.Error(t, err)

	mockUI.AssertExpectations(t)
}
