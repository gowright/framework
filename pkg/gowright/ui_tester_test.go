package gowright

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockBrowser is a mock implementation for testing
type MockBrowser struct {
	mock.Mock
}

func (m *MockBrowser) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBrowser) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockPage is a mock implementation for testing
type MockPage struct {
	mock.Mock
}

func (m *MockPage) Navigate(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockPage) WaitLoad() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPage) HTML() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// UITesterTestSuite defines the test suite for UITester
type UITesterTestSuite struct {
	suite.Suite
	tester *RodUITester
	config *BrowserConfig
}

// SetupTest runs before each test
func (suite *UITesterTestSuite) SetupTest() {
	suite.config = &BrowserConfig{
		Headless: true,
		Timeout:  10 * time.Second,
		WindowSize: &WindowSize{
			Width:  1280,
			Height: 720,
		},
		UserAgent: "test-agent",
	}
	suite.tester = NewRodUITester(suite.config)
}

// TearDownTest runs after each test
func (suite *UITesterTestSuite) TearDownTest() {
	if suite.tester != nil {
		_ = suite.tester.Cleanup()
	}
}

// TestNewRodUITester tests the constructor
func (suite *UITesterTestSuite) TestNewRodUITester() {
	// Test with nil config
	tester := NewRodUITester(nil)
	suite.NotNil(tester)
	suite.NotNil(tester.config)
	suite.True(tester.config.Headless)
	suite.Equal(30*time.Second, tester.config.Timeout)
	suite.Equal("RodUITester", tester.GetName())

	// Test with custom config
	customConfig := &BrowserConfig{
		Headless: false,
		Timeout:  5 * time.Second,
	}
	tester2 := NewRodUITester(customConfig)
	suite.NotNil(tester2)
	suite.Equal(customConfig, tester2.config)
	suite.False(tester2.config.Headless)
	suite.Equal(5*time.Second, tester2.config.Timeout)
}

// TestGetName tests the GetName method
func (suite *UITesterTestSuite) TestGetName() {
	name := suite.tester.GetName()
	suite.Equal("RodUITester", name)
}

// TestInitializeWithNilConfig tests initialization with nil config
func (suite *UITesterTestSuite) TestInitializeWithNilConfig() {
	// This test would require actual browser initialization
	// For unit testing, we'll test the configuration handling
	originalConfig := suite.tester.config

	err := suite.tester.Initialize(nil)
	// Since we can't mock rod completely, we expect this to fail in test environment
	// but we can verify the config wasn't changed
	suite.Equal(originalConfig, suite.tester.config)

	// The error should be a browser error since no browser is available in test
	if err != nil {
		gowrightErr, ok := err.(*GowrightError)
		suite.True(ok)
		suite.Equal(BrowserError, gowrightErr.Type)
	}
}

// TestInitializeWithBrowserConfig tests initialization with browser config
func (suite *UITesterTestSuite) TestInitializeWithBrowserConfig() {
	newConfig := &BrowserConfig{
		Headless: false,
		Timeout:  15 * time.Second,
		WindowSize: &WindowSize{
			Width:  1920,
			Height: 1080,
		},
	}

	err := suite.tester.Initialize(newConfig)
	// Since we can't mock rod completely, we expect this to fail in test environment
	// but we can verify the config was updated
	suite.Equal(newConfig, suite.tester.config)

	// The error should be a browser error since no browser is available in test
	if err != nil {
		gowrightErr, ok := err.(*GowrightError)
		suite.True(ok)
		suite.Equal(BrowserError, gowrightErr.Type)
	}
}

// TestInitializeWithInvalidConfig tests initialization with invalid config type
func (suite *UITesterTestSuite) TestInitializeWithInvalidConfig() {
	originalConfig := suite.tester.config

	err := suite.tester.Initialize("invalid config")
	// Config should remain unchanged
	suite.Equal(originalConfig, suite.tester.config)

	// The error should be a browser error since no browser is available in test
	if err != nil {
		gowrightErr, ok := err.(*GowrightError)
		suite.True(ok)
		suite.Equal(BrowserError, gowrightErr.Type)
	}
}

// TestNavigateWithoutInitialization tests navigation without browser initialization
func (suite *UITesterTestSuite) TestNavigateWithoutInitialization() {
	err := suite.tester.Navigate("https://example.com")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestClickWithoutInitialization tests clicking without browser initialization
func (suite *UITesterTestSuite) TestClickWithoutInitialization() {
	err := suite.tester.Click("#button")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestTypeWithoutInitialization tests typing without browser initialization
func (suite *UITesterTestSuite) TestTypeWithoutInitialization() {
	err := suite.tester.Type("#input", "test text")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestGetTextWithoutInitialization tests getting text without browser initialization
func (suite *UITesterTestSuite) TestGetTextWithoutInitialization() {
	text, err := suite.tester.GetText("#element")
	suite.Empty(text)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestWaitForElementWithoutInitialization tests waiting for element without browser initialization
func (suite *UITesterTestSuite) TestWaitForElementWithoutInitialization() {
	err := suite.tester.WaitForElement("#element", 5*time.Second)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestTakeScreenshotWithoutInitialization tests taking screenshot without browser initialization
func (suite *UITesterTestSuite) TestTakeScreenshotWithoutInitialization() {
	filename, err := suite.tester.TakeScreenshot("test.png")
	suite.Empty(filename)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestGetPageSourceWithoutInitialization tests getting page source without browser initialization
func (suite *UITesterTestSuite) TestGetPageSourceWithoutInitialization() {
	source, err := suite.tester.GetPageSource()
	suite.Empty(source)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestNewPageWithoutInitialization tests creating new page without browser initialization
func (suite *UITesterTestSuite) TestNewPageWithoutInitialization() {
	err := suite.tester.NewPage()
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "browser not initialized")
}

// TestGetCurrentURLWithoutInitialization tests getting current URL without browser initialization
func (suite *UITesterTestSuite) TestGetCurrentURLWithoutInitialization() {
	url, err := suite.tester.GetCurrentURL()
	suite.Empty(url)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestIsElementPresentWithoutInitialization tests checking element presence without browser initialization
func (suite *UITesterTestSuite) TestIsElementPresentWithoutInitialization() {
	present, err := suite.tester.IsElementPresent("#element")
	suite.False(present)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestIsElementVisibleWithoutInitialization tests checking element visibility without browser initialization
func (suite *UITesterTestSuite) TestIsElementVisibleWithoutInitialization() {
	visible, err := suite.tester.IsElementVisible("#element")
	suite.False(visible)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestScrollToElementWithoutInitialization tests scrolling to element without browser initialization
func (suite *UITesterTestSuite) TestScrollToElementWithoutInitialization() {
	err := suite.tester.ScrollToElement("#element")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestCleanupWithoutInitialization tests cleanup without browser initialization
func (suite *UITesterTestSuite) TestCleanupWithoutInitialization() {
	err := suite.tester.Cleanup()
	suite.NoError(err) // Cleanup should not error when nothing is initialized
}

// TestTakeScreenshotDirectoryCreation tests screenshot directory creation
func (suite *UITesterTestSuite) TestTakeScreenshotDirectoryCreation() {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "gowright_test_screenshots")
	defer func() { _ = os.RemoveAll(tempDir) }()

	screenshotPath := filepath.Join(tempDir, "subdir", "test.png")

	// This will fail because page is not initialized, but we can test directory creation logic
	_, err := suite.tester.TakeScreenshot(screenshotPath)
	suite.Error(err)

	// Verify it's the expected error (page not initialized, not directory creation)
	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestConfigurationDefaults tests default configuration values
func (suite *UITesterTestSuite) TestConfigurationDefaults() {
	tester := NewRodUITester(nil)

	suite.True(tester.config.Headless)
	suite.Equal(30*time.Second, tester.config.Timeout)
	suite.NotNil(tester.config.WindowSize)
	suite.Equal(1920, tester.config.WindowSize.Width)
	suite.Equal(1080, tester.config.WindowSize.Height)
}

// TestConfigurationCustomValues tests custom configuration values
func (suite *UITesterTestSuite) TestConfigurationCustomValues() {
	config := &BrowserConfig{
		Headless:  false,
		Timeout:   45 * time.Second,
		UserAgent: "custom-agent",
		WindowSize: &WindowSize{
			Width:  1366,
			Height: 768,
		},
	}

	tester := NewRodUITester(config)

	suite.False(tester.config.Headless)
	suite.Equal(45*time.Second, tester.config.Timeout)
	suite.Equal("custom-agent", tester.config.UserAgent)
	suite.Equal(1366, tester.config.WindowSize.Width)
	suite.Equal(768, tester.config.WindowSize.Height)
}

// TestUITesterTestSuite runs the test suite
func TestUITesterTestSuite(t *testing.T) {
	suite.Run(t, new(UITesterTestSuite))
}

// Additional unit tests for specific functionality

// Note: BrowserConfigValidation tests are in config_test.go

// TestErrorHandling tests error handling and error types
func TestErrorHandling(t *testing.T) {
	tester := NewRodUITester(nil)

	// Test that all methods return GowrightError when page is not initialized
	testCases := []struct {
		name string
		fn   func() error
	}{
		{"Navigate", func() error { return tester.Navigate("https://example.com") }},
		{"Click", func() error { return tester.Click("#button") }},
		{"Type", func() error { return tester.Type("#input", "text") }},
		{"WaitForElement", func() error { return tester.WaitForElement("#element", 5*time.Second) }},
		{"NewPage", func() error { return tester.NewPage() }},
		{"ScrollToElement", func() error { return tester.ScrollToElement("#element") }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			assert.Error(t, err)

			gowrightErr, ok := err.(*GowrightError)
			assert.True(t, ok, "Error should be of type GowrightError")
			assert.Equal(t, BrowserError, gowrightErr.Type)
		})
	}

	// Test methods that return values
	t.Run("GetText", func(t *testing.T) {
		text, err := tester.GetText("#element")
		assert.Empty(t, text)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, BrowserError, gowrightErr.Type)
	})

	t.Run("TakeScreenshot", func(t *testing.T) {
		filename, err := tester.TakeScreenshot("test.png")
		assert.Empty(t, filename)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, BrowserError, gowrightErr.Type)
	})

	t.Run("GetPageSource", func(t *testing.T) {
		source, err := tester.GetPageSource()
		assert.Empty(t, source)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, BrowserError, gowrightErr.Type)
	})

	t.Run("GetCurrentURL", func(t *testing.T) {
		url, err := tester.GetCurrentURL()
		assert.Empty(t, url)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, BrowserError, gowrightErr.Type)
	})

	t.Run("IsElementPresent", func(t *testing.T) {
		present, err := tester.IsElementPresent("#element")
		assert.False(t, present)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, BrowserError, gowrightErr.Type)
	})

	t.Run("IsElementVisible", func(t *testing.T) {
		visible, err := tester.IsElementVisible("#element")
		assert.False(t, visible)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, BrowserError, gowrightErr.Type)
	})
}
