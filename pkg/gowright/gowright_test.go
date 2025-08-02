package gowright

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type MockUITester struct {
	mock.Mock
}

func (m *MockUITester) Initialize(config interface{}) error {
	args := m.Called(config)
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

func (m *MockUITester) ExecuteTest(test *UITest) *TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*TestCaseResult)
}

type MockAPITester struct {
	mock.Mock
}

func (m *MockAPITester) Initialize(config interface{}) error {
	args := m.Called(config)
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

func (m *MockAPITester) Get(endpoint string, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, headers)
	return args.Get(0).(*APIResponse), args.Error(1)
}

func (m *MockAPITester) Post(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, body, headers)
	return args.Get(0).(*APIResponse), args.Error(1)
}

func (m *MockAPITester) Put(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, body, headers)
	return args.Get(0).(*APIResponse), args.Error(1)
}

func (m *MockAPITester) Delete(endpoint string, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, headers)
	return args.Get(0).(*APIResponse), args.Error(1)
}

func (m *MockAPITester) SetAuth(auth *AuthConfig) error {
	args := m.Called(auth)
	return args.Error(0)
}

func (m *MockAPITester) ExecuteTest(test *APITest) *TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*TestCaseResult)
}

// MockDatabaseTester is a mock implementation of DatabaseTester
type MockDatabaseTester struct {
	mock.Mock
}

func (m *MockDatabaseTester) Initialize(config interface{}) error {
	args := m.Called(config)
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

func (m *MockDatabaseTester) Execute(connectionName, query string, args ...interface{}) (*DatabaseResult, error) {
	mockArgs := m.Called(connectionName, query, args)
	return mockArgs.Get(0).(*DatabaseResult), mockArgs.Error(1)
}

func (m *MockDatabaseTester) BeginTransaction(connectionName string) (Transaction, error) {
	args := m.Called(connectionName)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockDatabaseTester) ValidateData(connectionName, query string, expected interface{}) error {
	args := m.Called(connectionName, query, expected)
	return args.Error(0)
}

func (m *MockDatabaseTester) ExecuteTest(test *DatabaseTest) *TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*TestCaseResult)
}

// SimpleTest implements the Test interface for testing in gowright_test.go
type SimpleTest struct {
	name   string
	result *TestCaseResult
}

func NewSimpleTest(name string, status TestStatus, err error) *SimpleTest {
	return &SimpleTest{
		name: name,
		result: &TestCaseResult{
			Name:     name,
			Status:   status,
			Duration: 100 * time.Millisecond,
			Error:    err,
		},
	}
}

func (st *SimpleTest) GetName() string {
	return st.name
}

func (st *SimpleTest) Execute() *TestCaseResult {
	return st.result
}

func TestNewWithDefaults(t *testing.T) {
	gw := NewWithDefaults()
	
	require.NotNil(t, gw)
	assert.NotNil(t, gw.GetConfig())
	assert.NotNil(t, gw.GetReporter())
	assert.False(t, gw.IsInitialized())
	
	config := gw.GetConfig()
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, false, config.Parallel)
	assert.Equal(t, 3, config.MaxRetries)
	
	// Test browser config defaults
	assert.NotNil(t, config.BrowserConfig)
	assert.True(t, config.BrowserConfig.Headless)
	assert.NotNil(t, config.BrowserConfig.WindowSize)
	assert.Equal(t, 1920, config.BrowserConfig.WindowSize.Width)
	assert.Equal(t, 1080, config.BrowserConfig.WindowSize.Height)
	
	// Test API config defaults
	assert.NotNil(t, config.APIConfig)
	assert.NotNil(t, config.APIConfig.Headers)
	
	// Test database config defaults
	assert.NotNil(t, config.DatabaseConfig)
	assert.NotNil(t, config.DatabaseConfig.Connections)
	
	// Test report config defaults
	assert.NotNil(t, config.ReportConfig)
	assert.True(t, config.ReportConfig.LocalReports.JSON)
	assert.True(t, config.ReportConfig.LocalReports.HTML)
	assert.Equal(t, "./reports", config.ReportConfig.LocalReports.OutputDir)
}

func TestNew(t *testing.T) {
	config := &Config{
		LogLevel:   "debug",
		Parallel:   true,
		MaxRetries: 5,
	}
	
	gw := New(config)
	
	require.NotNil(t, gw)
	assert.Equal(t, config, gw.GetConfig())
	assert.NotNil(t, gw.GetReporter())
	assert.False(t, gw.IsInitialized())
}

func TestNewWithNilConfig(t *testing.T) {
	gw := New(nil)
	
	require.NotNil(t, gw)
	assert.NotNil(t, gw.GetConfig())
	assert.NotNil(t, gw.GetReporter())
	assert.False(t, gw.IsInitialized())
}

func TestNewWithOptions(t *testing.T) {
	config := &Config{
		LogLevel:   "debug",
		Parallel:   true,
		MaxRetries: 5,
	}
	
	mockUITester := &MockUITester{}
	mockAPITester := &MockAPITester{}
	
	options := &GowrightOptions{
		Config:    config,
		UITester:  mockUITester,
		APITester: mockAPITester,
	}
	
	gw := NewWithOptions(options)
	
	require.NotNil(t, gw)
	assert.Equal(t, config, gw.GetConfig())
	assert.Equal(t, mockUITester, gw.GetUITester())
	assert.Equal(t, mockAPITester, gw.GetAPITester())
	assert.False(t, gw.IsInitialized())
}

func TestNewWithOptionsNil(t *testing.T) {
	gw := NewWithOptions(nil)
	
	require.NotNil(t, gw)
	assert.NotNil(t, gw.GetConfig())
	assert.NotNil(t, gw.GetReporter())
	assert.False(t, gw.IsInitialized())
}

func TestInitialize(t *testing.T) {
	mockUITester := &MockUITester{}
	mockAPITester := &MockAPITester{}
	
	mockUITester.On("Initialize", mock.Anything).Return(nil)
	mockAPITester.On("Initialize", mock.Anything).Return(nil)
	
	options := &GowrightOptions{
		UITester:  mockUITester,
		APITester: mockAPITester,
	}
	
	gw := NewWithOptions(options)
	
	err := gw.Initialize()
	assert.NoError(t, err)
	assert.True(t, gw.IsInitialized())
	
	// Test that calling Initialize again doesn't cause issues
	err = gw.Initialize()
	assert.NoError(t, err)
	assert.True(t, gw.IsInitialized())
	
	mockUITester.AssertExpectations(t)
	mockAPITester.AssertExpectations(t)
}

func TestInitializeWithError(t *testing.T) {
	mockUITester := &MockUITester{}
	expectedError := errors.New("initialization failed")
	
	mockUITester.On("Initialize", mock.Anything).Return(expectedError)
	
	options := &GowrightOptions{
		UITester: mockUITester,
	}
	
	gw := NewWithOptions(options)
	
	err := gw.Initialize()
	assert.Error(t, err)
	assert.False(t, gw.IsInitialized())
	
	// Check that it's a GowrightError
	var gowrightErr *GowrightError
	assert.True(t, errors.As(err, &gowrightErr))
	assert.Equal(t, ConfigurationError, gowrightErr.Type)
	
	mockUITester.AssertExpectations(t)
}

func TestCleanup(t *testing.T) {
	mockUITester := &MockUITester{}
	mockAPITester := &MockAPITester{}
	
	mockUITester.On("Initialize", mock.Anything).Return(nil)
	mockAPITester.On("Initialize", mock.Anything).Return(nil)
	mockUITester.On("Cleanup").Return(nil)
	mockAPITester.On("Cleanup").Return(nil)
	
	options := &GowrightOptions{
		UITester:  mockUITester,
		APITester: mockAPITester,
	}
	
	gw := NewWithOptions(options)
	
	// Initialize first
	err := gw.Initialize()
	require.NoError(t, err)
	assert.True(t, gw.IsInitialized())
	
	// Then cleanup
	err = gw.Cleanup()
	assert.NoError(t, err)
	assert.False(t, gw.IsInitialized())
	
	mockUITester.AssertExpectations(t)
	mockAPITester.AssertExpectations(t)
}

func TestCleanupWithError(t *testing.T) {
	mockUITester := &MockUITester{}
	expectedError := errors.New("cleanup failed")
	
	mockUITester.On("Initialize", mock.Anything).Return(nil)
	mockUITester.On("Cleanup").Return(expectedError)
	
	options := &GowrightOptions{
		UITester: mockUITester,
	}
	
	gw := NewWithOptions(options)
	
	// Initialize first
	err := gw.Initialize()
	require.NoError(t, err)
	
	// Then cleanup with error
	err = gw.Cleanup()
	assert.Error(t, err)
	assert.False(t, gw.IsInitialized())
	
	mockUITester.AssertExpectations(t)
}

func TestSetTestSuite(t *testing.T) {
	gw := NewWithDefaults()
	
	testSuite := &TestSuite{
		Name:  "Test Suite",
		Tests: make([]Test, 0),
	}
	
	gw.SetTestSuite(testSuite)
	assert.Equal(t, testSuite, gw.GetTestSuite())
}

func TestTesterSettersAndGetters(t *testing.T) {
	gw := NewWithDefaults()
	
	mockUITester := &MockUITester{}
	mockAPITester := &MockAPITester{}
	
	// Test setters
	gw.SetUITester(mockUITester)
	gw.SetAPITester(mockAPITester)
	
	// Test getters
	assert.Equal(t, mockUITester, gw.GetUITester())
	assert.Equal(t, mockAPITester, gw.GetAPITester())
	assert.Nil(t, gw.GetDatabaseTester())
	assert.Nil(t, gw.GetIntegrationTester())
}

func TestCreateTestSuiteManager(t *testing.T) {
	gw := NewWithDefaults()
	
	// Test with no test suite set
	tsm := gw.CreateTestSuiteManager()
	require.NotNil(t, tsm)
	assert.Equal(t, "Default Test Suite", tsm.GetTestSuite().Name)
	
	// Test with existing test suite
	customSuite := &TestSuite{
		Name:  "Custom Suite",
		Tests: make([]Test, 0),
	}
	gw.SetTestSuite(customSuite)
	
	tsm2 := gw.CreateTestSuiteManager()
	require.NotNil(t, tsm2)
	assert.Equal(t, "Custom Suite", tsm2.GetTestSuite().Name)
}

func TestExecuteTestSuiteIntegration(t *testing.T) {
	gw := NewWithDefaults()
	
	// Create a test suite with mock tests
	suite := &TestSuite{
		Name:  "Integration Test Suite",
		Tests: []Test{
			NewSimpleTest("test1", TestStatusPassed, nil),
			NewSimpleTest("test2", TestStatusPassed, nil),
		},
	}
	gw.SetTestSuite(suite)
	
	results, err := gw.ExecuteTestSuite()
	require.NoError(t, err)
	require.NotNil(t, results)
	
	assert.Equal(t, "Integration Test Suite", results.SuiteName)
	assert.Equal(t, 2, results.TotalTests)
	assert.Equal(t, 2, results.PassedTests)
}

func TestConcurrentAccess(t *testing.T) {
	gw := NewWithDefaults()
	
	// Test concurrent access to getters and setters
	done := make(chan bool, 2)
	
	go func() {
		for i := 0; i < 100; i++ {
			gw.SetUITester(&MockUITester{})
			_ = gw.GetUITester()
		}
		done <- true
	}()
	
	go func() {
		for i := 0; i < 100; i++ {
			_ = gw.IsInitialized()
			_ = gw.GetConfig()
		}
		done <- true
	}()
	
	// Wait for both goroutines to complete
	<-done
	<-done
	
	// If we get here without deadlock, the test passes
	assert.True(t, true)
}