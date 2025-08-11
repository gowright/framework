package testify

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/gowright/framework/pkg/gowright"
)

// TestifyAssertion provides testify-compatible assertion methods with Gowright integration
type TestifyAssertion struct {
	*core.TestAssertion
	t testing.TB
}

// NewTestifyAssertion creates a new testify-compatible assertion instance
func NewTestifyAssertion(t testing.TB, testName string) *TestifyAssertion {
	return &TestifyAssertion{
		TestAssertion: core.NewTestAssertion(testName),
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
	return success
}

// NotEqual asserts that two values are not equal using testify and records the result
func (ta *TestifyAssertion) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotEqual(expected, actual, msgAndArgs...)
	return success
}

// True asserts that the value is true using testify and records the result
func (ta *TestifyAssertion) True(value bool, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.True(value, msgAndArgs...)
	return success
}

// False asserts that the value is false using testify and records the result
func (ta *TestifyAssertion) False(value bool, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.False(value, msgAndArgs...)
	return success
}

// Nil asserts that the value is nil using testify and records the result
func (ta *TestifyAssertion) Nil(value interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Nil(value, msgAndArgs...)
	return success
}

// NotNil asserts that the value is not nil using testify and records the result
func (ta *TestifyAssertion) NotNil(value interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotNil(value, msgAndArgs...)
	return success
}

// Contains asserts that the string contains the substring using testify and records the result
func (ta *TestifyAssertion) Contains(s, contains string, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Contains(s, contains, msgAndArgs...)
	return success
}

// NotContains asserts that the string does not contain the substring using testify and records the result
func (ta *TestifyAssertion) NotContains(s, contains string, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotContains(s, contains, msgAndArgs...)
	return success
}

// Len asserts that the object has the expected length using testify and records the result
func (ta *TestifyAssertion) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Len(object, length, msgAndArgs...)
	return success
}

// Empty asserts that the object is empty using testify and records the result
func (ta *TestifyAssertion) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Empty(object, msgAndArgs...)
	return success
}

// NotEmpty asserts that the object is not empty using testify and records the result
func (ta *TestifyAssertion) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NotEmpty(object, msgAndArgs...)
	return success
}

// Error asserts that the error is not nil using testify and records the result
func (ta *TestifyAssertion) Error(err error, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.Error(err, msgAndArgs...)
	return success
}

// NoError asserts that the error is nil using testify and records the result
func (ta *TestifyAssertion) NoError(err error, msgAndArgs ...interface{}) bool {
	success := ta.TestAssertion.NoError(err, msgAndArgs...)
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
	gowright *core.Gowright
	config   *config.Config
}

// SetupSuite initializes the Gowright framework for the test suite
func (suite *GowrightTestSuite) SetupSuite() {
	config := config.DefaultConfig()

	suite.gowright = gowright.NewGowrightWithAllTesters(config)
	err := suite.gowright.Initialize()
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
func (suite *GowrightTestSuite) GetGowright() *core.Gowright {
	return suite.gowright
}

// NewTestifyAssertion creates a new testify assertion instance for the suite
func (suite *GowrightTestSuite) NewTestifyAssertion(testName string) *TestifyAssertion {
	return NewTestifyAssertion(suite.T(), testName)
}

// GowrightMock extends testify mock with Gowright-specific functionality
type GowrightMock struct {
	mock.Mock
	testName string
	logs     []string
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
func (m *UITesterMock) ExecuteTest(test *core.UITest) *core.TestCaseResult {
	args := m.Called(test)
	m.Log("ExecuteTest called with test: " + test.Name)
	return args.Get(0).(*core.TestCaseResult)
}

// TestifyIntegrationHelper provides helper methods for integrating Gowright with Go's testing package
type TestifyIntegrationHelper struct {
	t        *testing.T
	gowright *core.Gowright
}

// NewTestifyIntegrationHelper creates a new integration helper
func NewTestifyIntegrationHelper(t *testing.T, config *config.Config) (*TestifyIntegrationHelper, error) {
	gowrightInstance := gowright.NewGowrightWithAllTesters(config)
	err := gowrightInstance.Initialize()
	if err != nil {
		return nil, err
	}

	return &TestifyIntegrationHelper{
		t:        t,
		gowright: gowrightInstance,
	}, nil
}

// RunUITest runs a UI test with testify integration
func (h *TestifyIntegrationHelper) RunUITest(test *core.UITest) *core.TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running UI test: %s", test.Name)

	// Execute the test using Gowright
	result := h.gowright.ExecuteUITest(test)

	// Log the result
	if result.Status == core.TestStatusPassed {
		h.t.Logf("✓ UI test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ UI test failed: %s - %v", test.Name, result.Error)
	}

	return result
}

// RunAPITest runs an API test with testify integration
func (h *TestifyIntegrationHelper) RunAPITest(test *core.APITest) *core.TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running API test: %s", test.Name)

	// Execute the test using Gowright
	result := h.gowright.ExecuteAPITest(test)

	// Log the result
	if result.Status == core.TestStatusPassed {
		h.t.Logf("✓ API test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ API test failed: %s - %v", test.Name, result.Error)
	}

	return result
}

// RunDatabaseTest runs a database test with testify integration
func (h *TestifyIntegrationHelper) RunDatabaseTest(test *core.DatabaseTest) *core.TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running database test: %s", test.Name)

	// Execute the test using Gowright
	result := h.gowright.ExecuteDatabaseTest(test)

	// Log the result
	if result.Status == core.TestStatusPassed {
		h.t.Logf("✓ Database test passed: %s", test.Name)
	} else {
		h.t.Errorf("✗ Database test failed: %s - %v", test.Name, result.Error)
	}

	return result
}

// RunIntegrationTest runs an integration test with testify integration
func (h *TestifyIntegrationHelper) RunIntegrationTest(test *core.IntegrationTest) *core.TestCaseResult {
	h.t.Helper()
	h.t.Logf("Running integration test: %s", test.Name)

	// Execute the test using Gowright
	result := h.gowright.ExecuteIntegrationTest(test)

	// Log the result
	if result.Status == core.TestStatusPassed {
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
