package ui

import (
	"testing"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UITesterTestSuite defines the test suite for UITester
type UITesterTestSuite struct {
	suite.Suite
	tester *UITester
	config *config.BrowserConfig
}

// SetupTest runs before each test
func (suite *UITesterTestSuite) SetupTest() {
	suite.config = &config.BrowserConfig{
		Browser:    "chrome",
		Headless:   true,
		Timeout:    10 * time.Second,
		WindowSize: "1280x720",
	}
	suite.tester = NewUITester()
}

// TearDownTest runs after each test
func (suite *UITesterTestSuite) TearDownTest() {
	if suite.tester != nil {
		_ = suite.tester.Cleanup()
	}
}

// TestNewUITester tests the constructor
func (suite *UITesterTestSuite) TestNewUITester() {
	tester := NewUITester()
	suite.NotNil(tester)
	suite.Equal("UITester", tester.GetName())
	suite.False(tester.initialized)
}

// TestGetName tests the GetName method
func (suite *UITesterTestSuite) TestGetName() {
	name := suite.tester.GetName()
	suite.Equal("UITester", name)
}

// TestInitializeWithValidConfig tests initialization with valid config
func (suite *UITesterTestSuite) TestInitializeWithValidConfig() {
	err := suite.tester.Initialize(suite.config)
	// In a real implementation, this would set up browser automation
	// For now, we just test that the config is accepted
	suite.NoError(err)
	suite.True(suite.tester.initialized)
	suite.Equal(suite.config, suite.tester.config)
}

// TestInitializeWithInvalidConfig tests initialization with invalid config type
func (suite *UITesterTestSuite) TestInitializeWithInvalidConfig() {
	err := suite.tester.Initialize("invalid config")
	suite.Error(err)
	suite.False(suite.tester.initialized)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.ConfigurationError, gowrightErr.Type)
}

// TestNavigateWithoutInitialization tests navigation without initialization
func (suite *UITesterTestSuite) TestNavigateWithoutInitialization() {
	err := suite.tester.Navigate("https://example.com")
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestClickWithoutInitialization tests clicking without initialization
func (suite *UITesterTestSuite) TestClickWithoutInitialization() {
	err := suite.tester.Click("#button")
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestTypeWithoutInitialization tests typing without initialization
func (suite *UITesterTestSuite) TestTypeWithoutInitialization() {
	err := suite.tester.Type("#input", "test text")
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestGetTextWithoutInitialization tests getting text without initialization
func (suite *UITesterTestSuite) TestGetTextWithoutInitialization() {
	text, err := suite.tester.GetText("#element")
	suite.Empty(text)
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestWaitForElementWithoutInitialization tests waiting for element without initialization
func (suite *UITesterTestSuite) TestWaitForElementWithoutInitialization() {
	err := suite.tester.WaitForElement("#element", 5*time.Second)
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestTakeScreenshotWithoutInitialization tests taking screenshot without initialization
func (suite *UITesterTestSuite) TestTakeScreenshotWithoutInitialization() {
	filename, err := suite.tester.TakeScreenshot("test.png")
	suite.Empty(filename)
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestGetPageSourceWithoutInitialization tests getting page source without initialization
func (suite *UITesterTestSuite) TestGetPageSourceWithoutInitialization() {
	source, err := suite.tester.GetPageSource()
	suite.Empty(source)
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "UI tester not initialized")
}

// TestCleanupWithoutInitialization tests cleanup without initialization
func (suite *UITesterTestSuite) TestCleanupWithoutInitialization() {
	err := suite.tester.Cleanup()
	suite.NoError(err) // Cleanup should not error when nothing is initialized
}

// TestExecuteTest tests the ExecuteTest method
func (suite *UITesterTestSuite) TestExecuteTest() {
	// Initialize the tester first
	err := suite.tester.Initialize(suite.config)
	suite.NoError(err)

	// Create a test with a data URL containing test elements
	testHTML := `data:text/html,<html><body>
		<input id="username" type="text" />
		<input id="password" type="password" />
		<button id="login-btn">Login</button>
		<div class="welcome">Welcome to the application</div>
	</body></html>`

	test := &core.UITest{
		Name: "Test Login",
		URL:  testHTML,
		Actions: []core.UIAction{
			{Type: "type", Selector: "#username", Value: "testuser"},
			{Type: "type", Selector: "#password", Value: "testpass"},
			{Type: "click", Selector: "#login-btn"},
		},
		Assertions: []core.UIAssertion{
			{Type: "text_contains", Selector: ".welcome", Expected: "Welcome"},
		},
	}

	// Execute the test
	result := suite.tester.ExecuteTest(test)
	suite.NotNil(result)
	suite.Equal("Test Login", result.Name)
	suite.NotZero(result.Duration)

	// Since this is a mock implementation, we expect it to pass
	// In a real implementation, this would depend on actual browser interactions
	suite.Equal(core.TestStatusPassed, result.Status)
}

// TestExecuteAction tests individual action execution
func (suite *UITesterTestSuite) TestExecuteAction() {
	// Initialize the tester first
	err := suite.tester.Initialize(suite.config)
	suite.NoError(err)

	// Navigate to a test HTML page with the elements we need
	testHTML := `data:text/html,<html><body>
		<button id="button">Click me</button>
		<input id="input" type="text" />
	</body></html>`

	err = suite.tester.Navigate(testHTML)
	suite.NoError(err)

	// Test click action
	clickAction := &core.UIAction{Type: "click", Selector: "#button"}
	err = suite.tester.executeAction(clickAction)
	suite.NoError(err)

	// Test type action
	typeAction := &core.UIAction{Type: "type", Selector: "#input", Value: "test"}
	err = suite.tester.executeAction(typeAction)
	suite.NoError(err)

	// Test navigate action
	navAction := &core.UIAction{Type: "navigate", Value: "https://example.com"}
	err = suite.tester.executeAction(navAction)
	suite.NoError(err)

	// Test unsupported action
	unsupportedAction := &core.UIAction{Type: "unsupported"}
	err = suite.tester.executeAction(unsupportedAction)
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
}

// TestExecuteAssertion tests individual assertion execution
func (suite *UITesterTestSuite) TestExecuteAssertion() {
	// Initialize the tester first
	err := suite.tester.Initialize(suite.config)
	suite.NoError(err)

	// Navigate to a test HTML page with the elements we need
	// Using data URL to create a test page without external dependencies
	testHTML := `data:text/html,<html><body>
		<div id="element">expected text</div>
		<input id="input" value="test value" />
		<div class="visible-element" style="display: block;">Visible</div>
	</body></html>`

	err = suite.tester.Navigate(testHTML)
	suite.NoError(err)

	// Test text_equals assertion
	textEqualsAssertion := &core.UIAssertion{
		Type:     "text_equals",
		Selector: "#element",
		Expected: "expected text",
	}
	err = suite.tester.executeAssertion(textEqualsAssertion)
	suite.NoError(err)

	// Test text_contains assertion
	textContainsAssertion := &core.UIAssertion{
		Type:     "text_contains",
		Selector: "#element",
		Expected: "expected",
	}
	err = suite.tester.executeAssertion(textContainsAssertion)
	suite.NoError(err)

	// Test element_exists assertion
	elementExistsAssertion := &core.UIAssertion{
		Type:     "element_exists",
		Selector: "#element",
	}
	err = suite.tester.executeAssertion(elementExistsAssertion)
	suite.NoError(err)

	// Test unsupported assertion
	unsupportedAssertion := &core.UIAssertion{Type: "unsupported"}
	err = suite.tester.executeAssertion(unsupportedAssertion)
	suite.Error(err)

	gowrightErr, ok := err.(*core.GowrightError)
	suite.True(ok)
	suite.Equal(core.BrowserError, gowrightErr.Type)
}

// TestUITesterTestSuite runs the test suite
func TestUITesterTestSuite(t *testing.T) {
	suite.Run(t, new(UITesterTestSuite))
}

// Additional unit tests for specific functionality

// TestErrorHandling tests error handling and error types
func TestErrorHandling(t *testing.T) {
	tester := NewUITester()

	// Test that all methods return GowrightError when not initialized
	testCases := []struct {
		name string
		fn   func() error
	}{
		{"Navigate", func() error { return tester.Navigate("https://example.com") }},
		{"Click", func() error { return tester.Click("#button") }},
		{"Type", func() error { return tester.Type("#input", "text") }},
		{"WaitForElement", func() error { return tester.WaitForElement("#element", 5*time.Second) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			assert.Error(t, err)

			gowrightErr, ok := err.(*core.GowrightError)
			assert.True(t, ok, "Error should be of type GowrightError")
			assert.Equal(t, core.BrowserError, gowrightErr.Type)
		})
	}

	// Test methods that return values
	t.Run("GetText", func(t *testing.T) {
		text, err := tester.GetText("#element")
		assert.Empty(t, text)
		assert.Error(t, err)

		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.BrowserError, gowrightErr.Type)
	})

	t.Run("TakeScreenshot", func(t *testing.T) {
		filename, err := tester.TakeScreenshot("test.png")
		assert.Empty(t, filename)
		assert.Error(t, err)

		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.BrowserError, gowrightErr.Type)
	})

	t.Run("GetPageSource", func(t *testing.T) {
		source, err := tester.GetPageSource()
		assert.Empty(t, source)
		assert.Error(t, err)

		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.BrowserError, gowrightErr.Type)
	})
}

func TestUITesterInitialization(t *testing.T) {
	tester := NewUITester()

	// Test initialization with valid config
	config := &config.BrowserConfig{
		Browser:  "chrome",
		Headless: true,
		Timeout:  30 * time.Second,
	}

	err := tester.Initialize(config)
	assert.NoError(t, err)
	assert.True(t, tester.initialized)
	assert.Equal(t, config, tester.config)

	// Test cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
	assert.False(t, tester.initialized)
}
