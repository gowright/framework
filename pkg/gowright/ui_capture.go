package gowright

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UICaptureManager handles screenshot and page source capture functionality
type UICaptureManager struct {
	tester    UITester
	outputDir string
}

// NewUICaptureManager creates a new UICaptureManager
func NewUICaptureManager(tester UITester, outputDir string) *UICaptureManager {
	if outputDir == "" {
		outputDir = "./captures"
	}

	return &UICaptureManager{
		tester:    tester,
		outputDir: outputDir,
	}
}

// CaptureScreenshot captures a screenshot with automatic filename generation
func (c *UICaptureManager) CaptureScreenshot(testName string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.png", testName, timestamp)
	filePath := filepath.Join(c.outputDir, "screenshots", filename)

	return c.tester.TakeScreenshot(filePath)
}

// CaptureScreenshotWithName captures a screenshot with a specific filename
func (c *UICaptureManager) CaptureScreenshotWithName(filename string) (string, error) {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(c.outputDir, "screenshots", filename)
	}

	return c.tester.TakeScreenshot(filename)
}

// CapturePageSource captures the current page source
func (c *UICaptureManager) CapturePageSource(testName string) (string, error) {
	source, err := c.tester.GetPageSource()
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.html", testName, timestamp)
	filePath := filepath.Join(c.outputDir, "page_sources", filename)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", NewGowrightError(BrowserError, "failed to create page source directory", err)
	}

	// Write page source to file
	err = os.WriteFile(filePath, []byte(source), 0644)
	if err != nil {
		return "", NewGowrightError(BrowserError, "failed to save page source", err)
	}

	return filePath, nil
}

// CapturePageSourceWithName captures the page source with a specific filename
func (c *UICaptureManager) CapturePageSourceWithName(filename string) (string, error) {
	source, err := c.tester.GetPageSource()
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(filename) {
		filename = filepath.Join(c.outputDir, "page_sources", filename)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", NewGowrightError(BrowserError, "failed to create page source directory", err)
	}

	// Write page source to file
	err = os.WriteFile(filename, []byte(source), 0644)
	if err != nil {
		return "", NewGowrightError(BrowserError, "failed to save page source", err)
	}

	return filename, nil
}

// CaptureOnFailure captures both screenshot and page source when a test fails
func (c *UICaptureManager) CaptureOnFailure(testName string, err error) (*FailureCapture, error) {
	capture := &FailureCapture{
		TestName:  testName,
		Error:     err,
		Timestamp: time.Now(),
	}

	// Capture screenshot
	screenshotPath, screenshotErr := c.CaptureScreenshot(fmt.Sprintf("%s_failure", testName))
	if screenshotErr == nil {
		capture.ScreenshotPath = screenshotPath
	} else {
		capture.CaptureErrors = append(capture.CaptureErrors, fmt.Sprintf("screenshot capture failed: %v", screenshotErr))
	}

	// Capture page source
	pageSourcePath, pageSourceErr := c.CapturePageSource(fmt.Sprintf("%s_failure", testName))
	if pageSourceErr == nil {
		capture.PageSourcePath = pageSourcePath
	} else {
		capture.CaptureErrors = append(capture.CaptureErrors, fmt.Sprintf("page source capture failed: %v", pageSourceErr))
	}

	// Get current URL if possible
	if rodTester, ok := c.tester.(*RodUITester); ok {
		if currentURL, urlErr := rodTester.GetCurrentURL(); urlErr == nil {
			capture.CurrentURL = currentURL
		}
	}

	return capture, nil
}

// FailureCapture holds information about captures taken during test failures
type FailureCapture struct {
	TestName       string    `json:"test_name"`
	Error          error     `json:"error"`
	Timestamp      time.Time `json:"timestamp"`
	ScreenshotPath string    `json:"screenshot_path,omitempty"`
	PageSourcePath string    `json:"page_source_path,omitempty"`
	CurrentURL     string    `json:"current_url,omitempty"`
	CaptureErrors  []string  `json:"capture_errors,omitempty"`
}

// CaptureOptions holds options for capture operations
type CaptureOptions struct {
	IncludeTimestamp bool   `json:"include_timestamp"`
	CustomPrefix     string `json:"custom_prefix,omitempty"`
	Format           string `json:"format,omitempty"`  // png, jpg for screenshots
	Quality          int    `json:"quality,omitempty"` // for jpg screenshots
}

// AdvancedCaptureScreenshot captures a screenshot with advanced options
func (c *UICaptureManager) AdvancedCaptureScreenshot(testName string, options *CaptureOptions) (string, error) {
	if options == nil {
		return c.CaptureScreenshot(testName)
	}

	var filename string
	if options.CustomPrefix != "" {
		filename = options.CustomPrefix
	} else {
		filename = testName
	}

	if options.IncludeTimestamp {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("%s_%s", filename, timestamp)
	}

	format := "png"
	if options.Format != "" {
		format = options.Format
	}

	filename = fmt.Sprintf("%s.%s", filename, format)
	filePath := filepath.Join(c.outputDir, "screenshots", filename)

	// For now, we only support PNG format as that's what rod provides
	// In a more advanced implementation, we could add format conversion
	return c.tester.TakeScreenshot(filePath)
}

// CleanupOldCaptures removes capture files older than the specified duration
func (c *UICaptureManager) CleanupOldCaptures(maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)

	// Clean screenshots
	screenshotDir := filepath.Join(c.outputDir, "screenshots")
	if err := c.cleanupDirectory(screenshotDir, cutoffTime); err != nil {
		return fmt.Errorf("failed to cleanup screenshots: %w", err)
	}

	// Clean page sources
	pageSourceDir := filepath.Join(c.outputDir, "page_sources")
	if err := c.cleanupDirectory(pageSourceDir, cutoffTime); err != nil {
		return fmt.Errorf("failed to cleanup page sources: %w", err)
	}

	return nil
}

// cleanupDirectory removes files older than cutoffTime from the specified directory
func (c *UICaptureManager) cleanupDirectory(dir string, cutoffTime time.Time) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				// Log error but continue with other files
				fmt.Printf("Warning: failed to remove old capture file %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}

// GetCaptureStats returns statistics about captured files
func (c *UICaptureManager) GetCaptureStats() (*CaptureStats, error) {
	stats := &CaptureStats{}

	// Count screenshots
	screenshotDir := filepath.Join(c.outputDir, "screenshots")
	screenshotCount, screenshotSize, err := c.getDirectoryStats(screenshotDir)
	if err == nil {
		stats.ScreenshotCount = screenshotCount
		stats.ScreenshotTotalSize = screenshotSize
	}

	// Count page sources
	pageSourceDir := filepath.Join(c.outputDir, "page_sources")
	pageSourceCount, pageSourceSize, err := c.getDirectoryStats(pageSourceDir)
	if err == nil {
		stats.PageSourceCount = pageSourceCount
		stats.PageSourceTotalSize = pageSourceSize
	}

	stats.TotalFiles = stats.ScreenshotCount + stats.PageSourceCount
	stats.TotalSize = stats.ScreenshotTotalSize + stats.PageSourceTotalSize

	return stats, nil
}

// getDirectoryStats returns file count and total size for a directory
func (c *UICaptureManager) getDirectoryStats(dir string) (int, int64, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, 0, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}

	var count int
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		count++
		totalSize += info.Size()
	}

	return count, totalSize, nil
}

// CaptureStats holds statistics about captured files
type CaptureStats struct {
	ScreenshotCount     int   `json:"screenshot_count"`
	ScreenshotTotalSize int64 `json:"screenshot_total_size"`
	PageSourceCount     int   `json:"page_source_count"`
	PageSourceTotalSize int64 `json:"page_source_total_size"`
	TotalFiles          int   `json:"total_files"`
	TotalSize           int64 `json:"total_size"`
}

// SetOutputDirectory changes the output directory for captures
func (c *UICaptureManager) SetOutputDirectory(dir string) {
	c.outputDir = dir
}

// GetOutputDirectory returns the current output directory
func (c *UICaptureManager) GetOutputDirectory() string {
	return c.outputDir
}
