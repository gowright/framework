package gowright

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestifyAssertion provides testify-compatible assertion methods with Gowright integration
type TestifyAssertion struct {
	*TestAssertion
	t testing.TB
}

// NewTestifyAssertion creates a new testify-compatible assertion instance
func NewTestifyAssertion(t testing.TB, testName string) *TestifyAssertion {
	return &TestifyAssertion{
		TestAssertion: NewTestAssertion(testName),
		t:             t,
	}
}

// Assert returns the underlying testify assert instance for direct access
func (ta *TestifyAssertion) Assert() *assert.Assertions {
	return assert.New(ta.t)
}

// Require returns the underlying testify require instance for direct access
func (ta *TestifyAssertion) Require() *require.Assertions {
	return require.New(ta.t)
}

// Equal asserts that two values are equal using testify and records the result
func (ta *TestifyAssertion) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Equal(expected, actual, msgAndArgs...)
	if !success {
		assert.Equal(ta.t, expected, actual, msgAndArgs...)
	}
	return success
}

// NotEqual asserts that two values are not equal using testify and records the result
func (ta *TestifyAssertion) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotEqual(expected, actual, msgAndArgs...)
	if !success {
		assert.NotEqual(ta.t, expected, actual, msgAndArgs...)
	}
	return success
}

// True asserts that the value is true using testify and records the result
func (ta *TestifyAssertion) True(value bool, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.True(value, msgAndArgs...)
	if !success {
		assert.True(ta.t, value, msgAndArgs...)
	}
	return success
}

// False asserts that the value is false using testify and records the result
func (ta *TestifyAssertion) False(value bool, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.False(value, msgAndArgs...)
	if !success {
		assert.False(ta.t, value, msgAndArgs...)
	}
	return success
}

// Nil asserts that the value is nil using testify and records the result
func (ta *TestifyAssertion) Nil(value interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Nil(value, msgAndArgs...)
	if !success {
		assert.Nil(ta.t, value, msgAndArgs...)
	}
	return success
}

// NotNil asserts that the value is not nil using testify and records the result
func (ta *TestifyAssertion) NotNil(value interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotNil(value, msgAndArgs...)
	if !success {
		assert.NotNil(ta.t, value, msgAndArgs...)
	}
	return success
}

// Contains asserts that the string contains the substring using testify and records the result
func (ta *TestifyAssertion) Contains(s, contains string, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Contains(s, contains, msgAndArgs...)
	if !success {
		assert.Contains(ta.t, s, contains, msgAndArgs...)
	}
	return success
}

// NotContains asserts that the string does not contain the substring using testify and records the result
func (ta *TestifyAssertion) NotContains(s, contains string, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotContains(s, contains, msgAndArgs...)
	if !success {
		assert.NotContains(ta.t, s, contains, msgAndArgs...)
	}
	return success
}

// Len asserts that the object has the expected length using testify and records the result
func (ta *TestifyAssertion) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Len(object, length, msgAndArgs...)
	if !success {
		assert.Len(ta.t, object, length, msgAndArgs...)
	}
	return success
}

// Empty asserts that the object is empty using testify and records the result
func (ta *TestifyAssertion) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Empty(object, msgAndArgs...)
	if !success {
		assert.Empty(ta.t, object, msgAndArgs...)
	}
	return success
}

// NotEmpty asserts that the object is not empty using testify and records the result
func (ta *TestifyAssertion) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotEmpty(object, msgAndArgs...)
	if !success {
		assert.NotEmpty(ta.t, object, msgAndArgs...)
	}
	return success
}

// Error asserts that the error is not nil using testify and records the result
func (ta *TestifyAssertion) Error(err error, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Error(err, msgAndArgs...)
	if !success {
		assert.Error(ta.t, err, msgAndArgs...)
	}
	return success
}

// NoError asserts that the error is nil using testify and records the result
func (ta *TestifyAssertion) NoError(err error, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NoError(err, msgAndArgs...)
	if !success {
		assert.NoError(ta.t, err, msgAndArgs...)
	}
	return success
}

// RequireEqual asserts that two values are equal and fails the test immediately if not
func (ta *TestifyAssertion) RequireEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	ta.Equal(expected, actual, msgAndArgs...)
	require.Equal(ta.t, expected, actual, msgAndArgs...)
}

// RequireNotNil asserts that the value is not nil and fails the test immediately if it is
func (ta *TestifyAssertion) RequireNotNil(value interface{}, msgAndArgs ...interface{}) {
	ta.NotNil(value, msgAndArgs...)
	require.NotNil(ta.t, value, msgAndArgs...)
}

// RequireNoError asserts that the error is nil and fails the test immediately if not
func (ta *TestifyAssertion) RequireNoError(err error, msgAndArgs ...interface{}) {
	ta.NoError(err, msgAndArgs...)
	require.NoError(ta.t, err, msgAndArgs...)
}

// GowrightTestSuite provides a testify suite integration for Gowright tests
type GowrightTestSuite struct {
	suite.Suite
	gowright *Gowright
	config   *Config
}

// SetupSuite initializes the Gowright framework for the test suite
func (suite *GowrightTestSuite) SetupSuite() {
	config := &Config{
		LogLevel: "info",
		Parallel: false,
	}
	
	var err error
	suite.gowright, err = NewGowright(config)
	suite.Require().NoError(err, "Failed to initialize Gowright framework")
	suite.config = config
}

// TearDownSuite cleans up the Gowright framework after the test suite
func (suite *GowrightTestSuite) TearDownSuite() {
	if suite.gowright != nil {
		err := suite.gowright.Close()
		suite.Require().NoError(err, "Failed to close Gowright framework")
	}
}

// GetGowright returns the Gowright instance for use in tests
func (suite *GowrightTestSuite) GetGowright() *Gowright {
	return suite.gowright
}

// NewTestifyAssertion creates a new testify assertion instance for the suite
func (suite *GowrightTestSuite) NewTestifyAssertion(testName string) *TestifyAssertion {
	return NewTestifyAssertion(suite.T(), testName)
}

// MockInterface provides a base interface for creating mocks
type MockInterface interface {
	mock.TestingT
}

// GowrightMock extends testify mock with Gowright-specific functionality
type GowrightMock struct {
	mock.Mock
	testName string
	logs     []string
}

// TestMethod is a sample method for testing mock functionality
func (m *GowrightMock) TestMethod(arg1 string) string {
	args := m.Called(arg1)
	m.Log("TestMethod called with: " + arg1)
	return args.String(0)
}

// NewGowrightMock creates a new Gowright mock instance
func NewGowrightMock(testName string) *GowrightMock {
	return &GowrightMock{
		testName: testName,
		logs:     make([]string, 0),
	}
}

// Log adds a log entry to the mock
func (m *GowrightMock) Log(message string) {
	m.logs = append(m.logs, message)
}

// GetLogs returns all log entries from the mock
func (m *GowrightMock) GetLogs() []string {
	return m.logs
}

// GetTestName returns the test name associated with this mock
func (m *GowrightMock) GetTestName() string {
	return m.testName
}

// AssertExpectations asserts that all expectations were met and logs the result
func (m *GowrightMock) AssertExpectations(t mock.TestingT) bool {
	success := m.Mock.AssertExpectations(t)
	if success {
		m.Log("✓ All mock expectations were met")
	} else {
		m.Log("✗ Some mock expectations were not met")
	}
	return success
}

// UITesterMock provides a mock implementation of UITester for testing
type UITesterMock struct {
	*GowrightMock
}

// NewUITesterMock creates a new UITester mock
func NewUITesterMock(testName string) *UITesterMock {
	return &UITesterMock{
		GowrightMock: NewGowrightMock(testName),
	}
}

// Navigate mocks the Navigate method
func (m *UITesterMock) Navigate(url string) error {
	args := m.Called(url)
	m.Log("Navigate called with URL: " + url)
	return args.Error(0)
}

// Click mocks the Click method
func (m *UITesterMock) Click(selector string) error {
	args := m.Called(selector)
	m.Log("Click called with selector: " + selector)
	return args.Error(0)
}

// Type mocks the Type method
func (m *UITesterMock) Type(selector, text string) error {
	args := m.Called(selector, text)
	m.Log("Type called with selector: " + selector + ", text: " + text)
	return args.Error(0)
}

// GetText mocks the GetText method
func (m *UITesterMock) GetText(selector string) (string, error) {
	args := m.Called(selector)
	m.Log("GetText called with selector: " + selector)
	return args.String(0), args.Error(1)
}

// Initialize mocks the Initialize method
func (m *UITesterMock) Initialize(config interface{}) error {
	args := m.Called(config)
	m.Log("Initialize called")
	return args.Error(0)
}

// Cleanup mocks the Cleanup method
func (m *UITesterMock) Cleanup() error {
	args := m.Called()
	m.Log("Cleanup called")
	return args.Error(0)
}

// GetName mocks the GetName method
func (m *UITesterMock) GetName() string {
	args := m.Called()
	m.Log("GetName called")
	return args.String(0)
}

// WaitForElement mocks the WaitForElement method
func (m *UITesterMock) WaitForElement(selector string, timeout time.Duration) error {
	args := m.Called(selector, timeout)
	m.Log("WaitForElement called with selector: " + selector)
	return args.Error(0)
}

// TakeScreenshot mocks the TakeScreenshot method
func (m *UITesterMock) TakeScreenshot(filename string) (string, error) {
	args := m.Called(filename)
	m.Log("TakeScreenshot called with filename: " + filename)
	return args.String(0), args.Error(1)
}

// GetPageSource mocks the GetPageSource method
func (m *UITesterMock) GetPageSource() (string, error) {
	args := m.Called()
	m.Log("GetPageSource called")
	return args.String(0), args.Error(1)
}

// ExecuteTest mocks the ExecuteTest method
func (m *UITesterMock) ExecuteTest(test *UITest) *TestCaseResult {
	args := m.Called(test)
	m.Log("ExecuteTest called with test: " + test.Name)
	return args.Get(0).(*TestCaseResult)
}

// APITesterMock provides a mock implementation of APITester for testing
type APITesterMock struct {
	*GowrightMock
}

// NewAPITesterMock creates a new APITester mock
func NewAPITesterMock(testName string) *APITesterMock {
	return &APITesterMock{
		GowrightMock: NewGowrightMock(testName),
	}
}

// Get mocks the Get method
func (m *APITesterMock) Get(endpoint string, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, headers)
	m.Log("Get called with endpoint: " + endpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*APIResponse), args.Error(1)
}

// Post mocks the Post method
func (m *APITesterMock) Post(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, body, headers)
	m.Log("Post called with endpoint: " + endpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*APIResponse), args.Error(1)
}

// Put mocks the Put method
func (m *APITesterMock) Put(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, body, headers)
	m.Log("Put called with endpoint: " + endpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*APIResponse), args.Error(1)
}

// Delete mocks the Delete method
func (m *APITesterMock) Delete(endpoint string, headers map[string]string) (*APIResponse, error) {
	args := m.Called(endpoint, headers)
	m.Log("Delete called with endpoint: " + endpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*APIResponse), args.Error(1)
}

// Initialize mocks the Initialize method
func (m *APITesterMock) Initialize(config interface{}) error {
	args := m.Called(config)
	m.Log("Initialize called")
	return args.Error(0)
}

// Cleanup mocks the Cleanup method
func (m *APITesterMock) Cleanup() error {
	args := m.Called()
	m.Log("Cleanup called")
	return args.Error(0)
}

// GetName mocks the GetName method
func (m *APITesterMock) GetName() string {
	args := m.Called()
	m.Log("GetName called")
	return args.String(0)
}

// SetAuth mocks the SetAuth method
func (m *APITesterMock) SetAuth(auth *AuthConfig) error {
	args := m.Called(auth)
	m.Log("SetAuth called")
	return args.Error(0)
}

// ExecuteTest mocks the ExecuteTest method
func (m *APITesterMock) ExecuteTest(test *APITest) *TestCaseResult {
	args := m.Called(test)
	m.Log("ExecuteTest called with test: " + test.Name)
	return args.Get(0).(*TestCaseResult)
}

// DatabaseTesterMock provides a mock implementation of DatabaseTester for testing
type DatabaseTesterMock struct {
	*GowrightMock
}

// NewDatabaseTesterMock creates a new DatabaseTester mock
func NewDatabaseTesterMock(testName string) *DatabaseTesterMock {
	return &DatabaseTesterMock{
		GowrightMock: NewGowrightMock(testName),
	}
}

// Query mocks the Query method
func (m *DatabaseTesterMock) Query(connection, query string, args ...interface{}) (*DatabaseResult, error) {
	mockArgs := m.Called(connection, query, args)
	m.Log("Query called with connection: " + connection + ", query: " + query)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*DatabaseResult), mockArgs.Error(1)
}

// Execute mocks the Execute method
func (m *DatabaseTesterMock) Execute(connection, query string, args ...interface{}) (*DatabaseResult, error) {
	mockArgs := m.Called(connection, query, args)
	m.Log("Execute called with connection: " + connection + ", query: " + query)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*DatabaseResult), mockArgs.Error(1)
}

// Initialize mocks the Initialize method
func (m *DatabaseTesterMock) Initialize(config interface{}) error {
	args := m.Called(config)
	m.Log("Initialize called")
	return args.Error(0)
}

// Cleanup mocks the Cleanup method
func (m *DatabaseTesterMock) Cleanup() error {
	args := m.Called()
	m.Log("Cleanup called")
	return args.Error(0)
}

// GetName mocks the GetName method
func (m *DatabaseTesterMock) GetName() string {
	args := m.Called()
	m.Log("GetName called")
	return args.String(0)
}

// Connect mocks the Connect method
func (m *DatabaseTesterMock) Connect(connectionName string) error {
	args := m.Called(connectionName)
	m.Log("Connect called with connection: " + connectionName)
	return args.Error(0)
}

// BeginTransaction mocks the BeginTransaction method
func (m *DatabaseTesterMock) BeginTransaction(connectionName string) (Transaction, error) {
	args := m.Called(connectionName)
	m.Log("BeginTransaction called with connection: " + connectionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Transaction), args.Error(1)
}

// ValidateData mocks the ValidateData method
func (m *DatabaseTesterMock) ValidateData(connectionName, query string, expected interface{}) error {
	args := m.Called(connectionName, query, expected)
	m.Log("ValidateData called with connection: " + connectionName + ", query: " + query)
	return args.Error(0)
}

// ExecuteTest mocks the ExecuteTest method
func (m *DatabaseTesterMock) ExecuteTest(test *DatabaseTest) *TestCaseResult {
	args := m.Called(test)
	m.Log("ExecuteTest called with test: " + test.Name)
	return args.Get(0).(*TestCaseResult)
}

// TestifyIntegrationHelper provides helper methods for integrating Gowright with Go's testing package
type TestifyIntegrationHelper struct {
	t        *testing.T
	gowright *Gowright
}

// NewTestifyIntegrationHelper creates a new integration helper
func NewTestifyIntegrationHelper(t *testing.T, config *Config) (*TestifyIntegrationHelper, error) {
	gowright, err := NewGowright(config)
	if err != nil {
		return nil, err
	}
	
	return &TestifyIntegrationHelper{
		t:        t,
		gowright: gowright,
	}, nil
}

// RunUITest runs a UI test with testify integration
func (h *TestifyIntegrationHelper) RunUITest(test *UITest) *TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running UI test: %s", test.Name)
	
	// Execute the test using Gowright
	result := h.gowright.ExecuteUITest(test)
	
	// Log the result
	if result.Status == TestStatusPassed {
		h.t.Logf("✓ UI test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ UI test failed: %s - %v", test.Name, result.Error)
	}
	
	return result
}

// RunAPITest runs an API test with testify integration
func (h *TestifyIntegrationHelper) RunAPITest(test *APITest) *TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running API test: %s", test.Name)
	
	// Execute the test using Gowright
	result := h.gowright.ExecuteAPITest(test)
	
	// Log the result
	if result.Status == TestStatusPassed {
		h.t.Logf("✓ API test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ API test failed: %s - %v", test.Name, result.Error)
	}
	
	return result
}

// RunDatabaseTest runs a database test with testify integration
func (h *TestifyIntegrationHelper) RunDatabaseTest(test *DatabaseTest) *TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running database test: %s", test.Name)
	
	// Execute the test using Gowright
	result := h.gowright.ExecuteDatabaseTest(test)
	
	// Log the result
	if result.Status == TestStatusPassed {
		h.t.Logf("✓ Database test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ Database test failed: %s - %v", test.Name, result.Error)
	}
	
	return result
}

// RunIntegrationTest runs an integration test with testify integration
func (h *TestifyIntegrationHelper) RunIntegrationTest(test *IntegrationTest) *TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running integration test: %s", test.Name)
	
	// Execute the test using Gowright
	result := h.gowright.ExecuteIntegrationTest(test)
	
	// Log the result
	if result.Status == TestStatusPassed {
		h.t.Logf("✓ Integration test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ Integration test failed: %s - %v", test.Name, result.Error)
	}
	
	return result
}

// Close cleans up the integration helper
func (h *TestifyIntegrationHelper) Close() error {
	if h.gowright != nil {
		return h.gowright.Close()
	}
	return nil
}