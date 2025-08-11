package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/gowright/framework/pkg/core"
)

// MockUITester for testing
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

func (m *MockUITester) ExecuteTest(test *core.UITest) *core.TestCaseResult {
	args := m.Called(test)
	return args.Get(0).(*core.TestCaseResult)
}

// UICaptureTestSuite defines the test suite for UI capture functionality
type UICaptureTestSuite struct {
	suite.Suite
	mockTester     *MockUITester
	captureManager *UICaptureManager
	tempDir        string
}

// SetupTest runs before each test
func (suite *UICaptureTestSuite) SetupTest() {
	suite.mockTester = new(MockUITester)

	// Create temporary directory for testing
	var err error
	suite.tempDir, err = os.MkdirTemp("", "gowright_capture_test")
	suite.Require().NoError(err)

	suite.captureManager = NewUICaptureManager(suite.mockTester, suite.tempDir)
}

// TearDownTest runs after each test
func (suite *UICaptureTestSuite) TearDownTest() {
	suite.mockTester.AssertExpectations(suite.T())

	// Clean up temporary directory
	_ = os.RemoveAll(suite.tempDir)
}

// TestNewUICaptureManager tests the constructor
func (suite *UICaptureTestSuite) TestNewUICaptureManager() {
	// Test with custom output directory
	manager := NewUICaptureManager(suite.mockTester, "/custom/path")
	suite.NotNil(manager)
	suite.Equal(suite.mockTester, manager.tester)
	suite.Equal("/custom/path", manager.outputDir)

	// Test with empty output directory (should use default)
	manager2 := NewUICaptureManager(suite.mockTester, "")
	suite.Equal("./captures", manager2.outputDir)
}

// TestCaptureScreenshot tests screenshot capture with automatic filename
func (suite *UICaptureTestSuite) TestCaptureScreenshot() {
	// Use a more flexible mock that accepts any path containing the test name
	suite.mockTester.On("TakeScreenshot", mock.MatchedBy(func(path string) bool {
		return filepath.Base(path) != "" &&
			filepath.Ext(path) == ".png" &&
			strings.Contains(path, "test_screenshot")
	})).Return("screenshot_path", nil)

	actualPath, err := suite.captureManager.CaptureScreenshot("test_screenshot")

	suite.NoError(err)
	suite.Equal("screenshot_path", actualPath)
}

// TestCaptureScreenshotWithName tests screenshot capture with specific filename
func (suite *UICaptureTestSuite) TestCaptureScreenshotWithName() {
	expectedPath := filepath.Join(suite.tempDir, "screenshots", "custom_screenshot.png")

	suite.mockTester.On("TakeScreenshot", expectedPath).Return(expectedPath, nil)

	actualPath, err := suite.captureManager.CaptureScreenshotWithName("custom_screenshot.png")

	suite.NoError(err)
	suite.Equal(expectedPath, actualPath)
}

// TestCaptureScreenshotWithAbsolutePath tests screenshot capture with absolute path
func (suite *UICaptureTestSuite) TestCaptureScreenshotWithAbsolutePath() {
	absolutePath := filepath.Join(suite.tempDir, "absolute_screenshot.png")

	suite.mockTester.On("TakeScreenshot", absolutePath).Return(absolutePath, nil)

	actualPath, err := suite.captureManager.CaptureScreenshotWithName(absolutePath)

	suite.NoError(err)
	suite.Equal(absolutePath, actualPath)
}

// TestCapturePageSource tests page source capture with automatic filename
func (suite *UICaptureTestSuite) TestCapturePageSource() {
	htmlContent := "<html><head><title>Test</title></head><body>Test content</body></html>"

	suite.mockTester.On("GetPageSource").Return(htmlContent, nil)

	actualPath, err := suite.captureManager.CapturePageSource("test_page")

	suite.NoError(err)
	suite.Contains(actualPath, "test_page")
	suite.Contains(actualPath, ".html")

	// Verify file was created and contains correct content
	content, err := os.ReadFile(actualPath)
	suite.NoError(err)
	suite.Equal(htmlContent, string(content))
}

// TestCapturePageSourceWithName tests page source capture with specific filename
func (suite *UICaptureTestSuite) TestCapturePageSourceWithName() {
	htmlContent := "<html><head><title>Test</title></head><body>Test content</body></html>"
	expectedPath := filepath.Join(suite.tempDir, "page_sources", "custom_page.html")

	suite.mockTester.On("GetPageSource").Return(htmlContent, nil)

	actualPath, err := suite.captureManager.CapturePageSourceWithName("custom_page.html")

	suite.NoError(err)
	suite.Equal(expectedPath, actualPath)

	// Verify file was created and contains correct content
	content, err := os.ReadFile(actualPath)
	suite.NoError(err)
	suite.Equal(htmlContent, string(content))
}

// TestCapturePageSourceWithAbsolutePath tests page source capture with absolute path
func (suite *UICaptureTestSuite) TestCapturePageSourceWithAbsolutePath() {
	htmlContent := "<html><head><title>Test</title></head><body>Test content</body></html>"
	absolutePath := filepath.Join(suite.tempDir, "absolute_page.html")

	suite.mockTester.On("GetPageSource").Return(htmlContent, nil)

	actualPath, err := suite.captureManager.CapturePageSourceWithName(absolutePath)

	suite.NoError(err)
	suite.Equal(absolutePath, actualPath)

	// Verify file was created and contains correct content
	content, err := os.ReadFile(actualPath)
	suite.NoError(err)
	suite.Equal(htmlContent, string(content))
}

// TestCaptureOnFailure tests capture on failure functionality
func (suite *UICaptureTestSuite) TestCaptureOnFailure() {
	htmlContent := "<html><head><title>Test</title></head><body>Test content</body></html>"
	testError := core.NewGowrightError(core.BrowserError, "test error", nil)

	// Mock both screenshot and page source capture
	suite.mockTester.On("TakeScreenshot", mock.AnythingOfType("string")).Return("screenshot_path", nil)
	suite.mockTester.On("GetPageSource").Return(htmlContent, nil)

	capture, err := suite.captureManager.CaptureOnFailure("test_failure", testError)

	suite.NoError(err)
	suite.NotNil(capture)
	suite.Equal("test_failure", capture.TestName)
	suite.Equal(testError, capture.Error)
	suite.NotEmpty(capture.ScreenshotPath)
	suite.NotEmpty(capture.PageSourcePath)
	suite.Empty(capture.CaptureErrors)
}

// TestCaptureOnFailureWithErrors tests capture on failure when captures fail
func (suite *UICaptureTestSuite) TestCaptureOnFailureWithErrors() {
	testError := core.NewGowrightError(core.BrowserError, "test error", nil)
	screenshotError := core.NewGowrightError(core.BrowserError, "screenshot failed", nil)
	pageSourceError := core.NewGowrightError(core.BrowserError, "page source failed", nil)

	// Mock both captures to fail
	suite.mockTester.On("TakeScreenshot", mock.AnythingOfType("string")).Return("", screenshotError)
	suite.mockTester.On("GetPageSource").Return("", pageSourceError)

	capture, err := suite.captureManager.CaptureOnFailure("test_failure", testError)

	suite.NoError(err)
	suite.NotNil(capture)
	suite.Equal("test_failure", capture.TestName)
	suite.Equal(testError, capture.Error)
	suite.Empty(capture.ScreenshotPath)
	suite.Empty(capture.PageSourcePath)
	suite.Len(capture.CaptureErrors, 2)
	suite.Contains(capture.CaptureErrors[0], "screenshot capture failed")
	suite.Contains(capture.CaptureErrors[1], "page source capture failed")
}

// TestSetAndGetOutputDirectory tests output directory management
func (suite *UICaptureTestSuite) TestSetAndGetOutputDirectory() {
	originalDir := suite.captureManager.GetOutputDirectory()
	suite.Equal(suite.tempDir, originalDir)

	newDir := "/new/output/dir"
	suite.captureManager.SetOutputDirectory(newDir)

	updatedDir := suite.captureManager.GetOutputDirectory()
	suite.Equal(newDir, updatedDir)
}

// TestAdvancedCaptureScreenshot tests advanced screenshot capture with options
func (suite *UICaptureTestSuite) TestAdvancedCaptureScreenshot() {
	options := &CaptureOptions{
		IncludeTimestamp: true,
		CustomPrefix:     "custom_prefix",
		Format:           "png",
	}

	suite.mockTester.On("TakeScreenshot", mock.AnythingOfType("string")).Return("screenshot_path", nil)

	actualPath, err := suite.captureManager.AdvancedCaptureScreenshot("test", options)

	suite.NoError(err)
	suite.NotEmpty(actualPath)
}

// TestAdvancedCaptureScreenshotWithNilOptions tests advanced screenshot capture with nil options
func (suite *UICaptureTestSuite) TestAdvancedCaptureScreenshotWithNilOptions() {
	suite.mockTester.On("TakeScreenshot", mock.AnythingOfType("string")).Return("screenshot_path", nil)

	actualPath, err := suite.captureManager.AdvancedCaptureScreenshot("test", nil)

	suite.NoError(err)
	suite.NotEmpty(actualPath)
}

// TestUICaptureTestSuite runs the test suite
func TestUICaptureTestSuite(t *testing.T) {
	suite.Run(t, new(UICaptureTestSuite))
}

// TestCleanupOldCaptures tests cleanup functionality
func TestCleanupOldCaptures(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gowright_cleanup_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	mockTester := new(MockUITester)
	manager := NewUICaptureManager(mockTester, tempDir)

	// Create test directories
	screenshotDir := filepath.Join(tempDir, "screenshots")
	pageSourceDir := filepath.Join(tempDir, "page_sources")
	assert.NoError(t, os.MkdirAll(screenshotDir, 0755))
	assert.NoError(t, os.MkdirAll(pageSourceDir, 0755))

	// Create old and new files
	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now().Add(-30 * time.Minute)

	oldScreenshot := filepath.Join(screenshotDir, "old_screenshot.png")
	newScreenshot := filepath.Join(screenshotDir, "new_screenshot.png")
	oldPageSource := filepath.Join(pageSourceDir, "old_page.html")
	newPageSource := filepath.Join(pageSourceDir, "new_page.html")

	// Create files
	assert.NoError(t, os.WriteFile(oldScreenshot, []byte("old screenshot"), 0644))
	assert.NoError(t, os.WriteFile(newScreenshot, []byte("new screenshot"), 0644))
	assert.NoError(t, os.WriteFile(oldPageSource, []byte("old page"), 0644))
	assert.NoError(t, os.WriteFile(newPageSource, []byte("new page"), 0644))

	// Set file modification times
	assert.NoError(t, os.Chtimes(oldScreenshot, oldTime, oldTime))
	assert.NoError(t, os.Chtimes(oldPageSource, oldTime, oldTime))
	assert.NoError(t, os.Chtimes(newScreenshot, newTime, newTime))
	assert.NoError(t, os.Chtimes(newPageSource, newTime, newTime))

	// Cleanup files older than 1 hour
	err = manager.CleanupOldCaptures(1 * time.Hour)
	assert.NoError(t, err)

	// Check that old files are removed and new files remain
	_, err = os.Stat(oldScreenshot)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(oldPageSource)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(newScreenshot)
	assert.NoError(t, err)

	_, err = os.Stat(newPageSource)
	assert.NoError(t, err)
}

// TestGetCaptureStats tests capture statistics functionality
func TestGetCaptureStats(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gowright_stats_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	mockTester := new(MockUITester)
	manager := NewUICaptureManager(mockTester, tempDir)

	// Create test directories
	screenshotDir := filepath.Join(tempDir, "screenshots")
	pageSourceDir := filepath.Join(tempDir, "page_sources")
	assert.NoError(t, os.MkdirAll(screenshotDir, 0755))
	assert.NoError(t, os.MkdirAll(pageSourceDir, 0755))

	// Create test files
	screenshot1 := filepath.Join(screenshotDir, "screenshot1.png")
	screenshot2 := filepath.Join(screenshotDir, "screenshot2.png")
	pageSource1 := filepath.Join(pageSourceDir, "page1.html")

	assert.NoError(t, os.WriteFile(screenshot1, []byte("screenshot1 content"), 0644))
	assert.NoError(t, os.WriteFile(screenshot2, []byte("screenshot2 content longer"), 0644))
	assert.NoError(t, os.WriteFile(pageSource1, []byte("page source content"), 0644))

	// Get stats
	stats, err := manager.GetCaptureStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, 2, stats.ScreenshotCount)
	assert.Equal(t, 1, stats.PageSourceCount)
	assert.Equal(t, 3, stats.TotalFiles)
	assert.Greater(t, stats.ScreenshotTotalSize, int64(0))
	assert.Greater(t, stats.PageSourceTotalSize, int64(0))
	assert.Equal(t, stats.ScreenshotTotalSize+stats.PageSourceTotalSize, stats.TotalSize)
}
