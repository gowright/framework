package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// UICaptureManager handles screenshot and page source capture functionality
type UICaptureManager struct {
	tester    core.UITester
	outputDir string
}

// NewUICaptureManager creates a new UICaptureManager
func NewUICaptureManager(tester core.UITester, outputDir string) *UICaptureManager {
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

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create screenshot directory", err)
	}

	return c.tester.TakeScreenshot(filePath)
}

// CaptureScreenshotWithName captures a screenshot with a specific filename
func (c *UICaptureManager) CaptureScreenshotWithName(filename string) (string, error) {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(c.outputDir, "screenshots", filename)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create screenshot directory", err)
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
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create page source directory", err)
	}

	// Write page source to file
	if err := os.WriteFile(filePath, []byte(source), 0644); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to write page source", err)
	}

	return filePath, nil
}

// CapturePageSourceWithName captures page source with a specific filename
func (c *UICaptureManager) CapturePageSourceWithName(filename string) (string, error) {
	source, err := c.tester.GetPageSource()
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(filename) {
		filename = filepath.Join(c.outputDir, "page_sources", filename)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create page source directory", err)
	}

	// Write page source to file
	if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to write page source", err)
	}

	return filename, nil
}

// CaptureTestEvidence captures both screenshot and page source for a test
func (c *UICaptureManager) CaptureTestEvidence(testName string) (*TestEvidence, error) {
	evidence := &TestEvidence{
		TestName:  testName,
		Timestamp: time.Now(),
	}

	// Capture screenshot
	screenshotPath, err := c.CaptureScreenshot(testName)
	if err != nil {
		evidence.Errors = append(evidence.Errors, fmt.Sprintf("Screenshot capture failed: %v", err))
	} else {
		evidence.ScreenshotPath = screenshotPath
	}

	// Capture page source
	pageSourcePath, err := c.CapturePageSource(testName)
	if err != nil {
		evidence.Errors = append(evidence.Errors, fmt.Sprintf("Page source capture failed: %v", err))
	} else {
		evidence.PageSourcePath = pageSourcePath
	}

	return evidence, nil
}

// SetOutputDirectory sets the output directory for captures
func (c *UICaptureManager) SetOutputDirectory(outputDir string) {
	c.outputDir = outputDir
}

// GetOutputDirectory returns the current output directory
func (c *UICaptureManager) GetOutputDirectory() string {
	return c.outputDir
}

// TestEvidence holds captured evidence for a test
type TestEvidence struct {
	TestName       string    `json:"test_name"`
	Timestamp      time.Time `json:"timestamp"`
	ScreenshotPath string    `json:"screenshot_path,omitempty"`
	PageSourcePath string    `json:"page_source_path,omitempty"`
	Errors         []string  `json:"errors,omitempty"`
}

// HasScreenshot returns true if screenshot was captured successfully
func (te *TestEvidence) HasScreenshot() bool {
	return te.ScreenshotPath != ""
}

// HasPageSource returns true if page source was captured successfully
func (te *TestEvidence) HasPageSource() bool {
	return te.PageSourcePath != ""
}

// HasErrors returns true if there were any capture errors
func (te *TestEvidence) HasErrors() bool {
	return len(te.Errors) > 0
}

// IsComplete returns true if both screenshot and page source were captured
func (te *TestEvidence) IsComplete() bool {
	return te.HasScreenshot() && te.HasPageSource() && !te.HasErrors()
}

// CaptureOptions holds options for advanced capture operations
type CaptureOptions struct {
	IncludeTimestamp bool   `json:"include_timestamp"`
	CustomPrefix     string `json:"custom_prefix"`
	Format           string `json:"format"`
	Quality          int    `json:"quality"`
}

// FailureCapture holds information about a test failure capture
type FailureCapture struct {
	TestName       string    `json:"test_name"`
	Error          error     `json:"error"`
	ScreenshotPath string    `json:"screenshot_path,omitempty"`
	PageSourcePath string    `json:"page_source_path,omitempty"`
	CaptureErrors  []string  `json:"capture_errors,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// CaptureStats holds statistics about captured files
type CaptureStats struct {
	ScreenshotCount     int       `json:"screenshot_count"`
	PageSourceCount     int       `json:"page_source_count"`
	TotalFiles          int       `json:"total_files"`
	ScreenshotTotalSize int64     `json:"screenshot_total_size"`
	PageSourceTotalSize int64     `json:"page_source_total_size"`
	TotalSize           int64     `json:"total_size"`
	OldestCaptureTime   time.Time `json:"oldest_capture_time,omitempty"`
	NewestCaptureTime   time.Time `json:"newest_capture_time,omitempty"`
}

// CaptureOnFailure captures both screenshot and page source when a test fails
func (c *UICaptureManager) CaptureOnFailure(testName string, testError error) (*FailureCapture, error) {
	capture := &FailureCapture{
		TestName:  testName,
		Error:     testError,
		Timestamp: time.Now(),
	}

	// Capture screenshot
	screenshotPath, err := c.CaptureScreenshot(testName + "_failure")
	if err != nil {
		capture.CaptureErrors = append(capture.CaptureErrors, fmt.Sprintf("screenshot capture failed: %v", err))
	} else {
		capture.ScreenshotPath = screenshotPath
	}

	// Capture page source
	pageSourcePath, err := c.CapturePageSource(testName + "_failure")
	if err != nil {
		capture.CaptureErrors = append(capture.CaptureErrors, fmt.Sprintf("page source capture failed: %v", err))
	} else {
		capture.PageSourcePath = pageSourcePath
	}

	return capture, nil
}

// AdvancedCaptureScreenshot captures a screenshot with advanced options
func (c *UICaptureManager) AdvancedCaptureScreenshot(testName string, options *CaptureOptions) (string, error) {
	if options == nil {
		return c.CaptureScreenshot(testName)
	}

	var filename string
	if options.IncludeTimestamp {
		timestamp := time.Now().Format("20060102_150405")
		if options.CustomPrefix != "" {
			filename = fmt.Sprintf("%s_%s_%s", options.CustomPrefix, testName, timestamp)
		} else {
			filename = fmt.Sprintf("%s_%s", testName, timestamp)
		}
	} else {
		if options.CustomPrefix != "" {
			filename = fmt.Sprintf("%s_%s", options.CustomPrefix, testName)
		} else {
			filename = testName
		}
	}

	format := options.Format
	if format == "" {
		format = "png"
	}
	filename = fmt.Sprintf("%s.%s", filename, format)

	filePath := filepath.Join(c.outputDir, "screenshots", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to create screenshot directory", err)
	}

	return c.tester.TakeScreenshot(filePath)
}

// CleanupOldCaptures removes capture files older than the specified duration
func (c *UICaptureManager) CleanupOldCaptures(maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)

	// Clean up screenshots
	screenshotDir := filepath.Join(c.outputDir, "screenshots")
	if err := c.cleanupDirectory(screenshotDir, cutoffTime); err != nil {
		return fmt.Errorf("failed to cleanup screenshots: %w", err)
	}

	// Clean up page sources
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

		filePath := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't get info for
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				// Log error but continue with other files
				continue
			}
		}
	}

	return nil
}

// GetCaptureStats returns statistics about captured files
func (c *UICaptureManager) GetCaptureStats() (*CaptureStats, error) {
	stats := &CaptureStats{}

	// Get screenshot stats
	screenshotDir := filepath.Join(c.outputDir, "screenshots")
	screenshotCount, screenshotSize, screenshotOldest, screenshotNewest, err := c.getDirectoryStats(screenshotDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to get screenshot stats: %w", err)
	}
	stats.ScreenshotCount = screenshotCount
	stats.ScreenshotTotalSize = screenshotSize

	// Get page source stats
	pageSourceDir := filepath.Join(c.outputDir, "page_sources")
	pageSourceCount, pageSourceSize, pageSourceOldest, pageSourceNewest, err := c.getDirectoryStats(pageSourceDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to get page source stats: %w", err)
	}
	stats.PageSourceCount = pageSourceCount
	stats.PageSourceTotalSize = pageSourceSize

	// Calculate totals
	stats.TotalFiles = stats.ScreenshotCount + stats.PageSourceCount
	stats.TotalSize = stats.ScreenshotTotalSize + stats.PageSourceTotalSize

	// Determine oldest and newest times
	if !screenshotOldest.IsZero() && !pageSourceOldest.IsZero() {
		if screenshotOldest.Before(pageSourceOldest) {
			stats.OldestCaptureTime = screenshotOldest
		} else {
			stats.OldestCaptureTime = pageSourceOldest
		}
	} else if !screenshotOldest.IsZero() {
		stats.OldestCaptureTime = screenshotOldest
	} else if !pageSourceOldest.IsZero() {
		stats.OldestCaptureTime = pageSourceOldest
	}

	if !screenshotNewest.IsZero() && !pageSourceNewest.IsZero() {
		if screenshotNewest.After(pageSourceNewest) {
			stats.NewestCaptureTime = screenshotNewest
		} else {
			stats.NewestCaptureTime = pageSourceNewest
		}
	} else if !screenshotNewest.IsZero() {
		stats.NewestCaptureTime = screenshotNewest
	} else if !pageSourceNewest.IsZero() {
		stats.NewestCaptureTime = pageSourceNewest
	}

	return stats, nil
}

// getDirectoryStats returns count, total size, oldest and newest modification times for files in a directory
func (c *UICaptureManager) getDirectoryStats(dir string) (int, int64, time.Time, time.Time, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, 0, time.Time{}, time.Time{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, time.Time{}, time.Time{}, err
	}

	var count int
	var totalSize int64
	var oldest, newest time.Time

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

		modTime := info.ModTime()
		if oldest.IsZero() || modTime.Before(oldest) {
			oldest = modTime
		}
		if newest.IsZero() || modTime.After(newest) {
			newest = modTime
		}
	}

	return count, totalSize, oldest, newest, nil
}
