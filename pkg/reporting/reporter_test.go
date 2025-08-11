package reporting

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReporter is a mock implementation of the Reporter interface
type MockReporter struct {
	mock.Mock
}

func (m *MockReporter) GenerateReport(results *core.TestResults) error {
	args := m.Called(results)
	return args.Error(0)
}

func (m *MockReporter) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockReporter) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestNewReportManager(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.ReportConfig
		expected int // expected number of reporters
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: 0,
		},
		{
			name: "empty formats",
			config: &config.ReportConfig{
				Enabled:   true,
				OutputDir: "./reports",
				Formats:   []string{},
			},
			expected: 0,
		},
		{
			name: "JSON only",
			config: &config.ReportConfig{
				Enabled:   true,
				OutputDir: "./reports",
				Formats:   []string{"json"},
			},
			expected: 1,
		},
		{
			name: "HTML only",
			config: &config.ReportConfig{
				Enabled:   true,
				OutputDir: "./reports",
				Formats:   []string{"html"},
			},
			expected: 1,
		},
		{
			name: "all formats",
			config: &config.ReportConfig{
				Enabled:   true,
				OutputDir: "./reports",
				Formats:   []string{"json", "html", "xml", "junit"},
			},
			expected: 4,
		},
		{
			name: "disabled config",
			config: &config.ReportConfig{
				Enabled:   false,
				OutputDir: "./reports",
				Formats:   []string{"json", "html"},
			},
			expected: 2, // Reporters are still created, just not used
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewReportManager(tt.config)

			assert.NotNil(t, rm)
			assert.Equal(t, tt.config, rm.config)
			assert.Len(t, rm.reporters, tt.expected)
		})
	}
}

func TestReportManager_AddReporter(t *testing.T) {
	config := &config.ReportConfig{
		Enabled:   true,
		OutputDir: "./reports",
		Formats:   []string{},
	}
	rm := NewReportManager(config)
	mockReporter := &MockReporter{}
	mockReporter.On("GetName").Return("mock")
	mockReporter.On("IsEnabled").Return(true)

	// Initially empty
	assert.Len(t, rm.reporters, 0)

	// Add reporter
	rm.AddReporter(mockReporter)
	assert.Len(t, rm.reporters, 1)
	assert.Equal(t, mockReporter, rm.reporters[0])

	// Add another reporter
	mockReporter2 := &MockReporter{}
	mockReporter2.On("GetName").Return("mock2")
	mockReporter2.On("IsEnabled").Return(true)

	rm.AddReporter(mockReporter2)
	assert.Len(t, rm.reporters, 2)
}

func TestReportManager_GenerateReports(t *testing.T) {
	testResults := &core.TestResults{
		SuiteName:    "Test Suite",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Minute),
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []core.TestCaseResult{
			{
				Name:     "Test 1",
				Status:   core.TestStatusPassed,
				Duration: time.Second,
			},
			{
				Name:     "Test 2",
				Status:   core.TestStatusFailed,
				Duration: time.Second * 2,
				Error:    errors.New("test failed"),
			},
		},
	}

	t.Run("successful generation", func(t *testing.T) {
		config := &config.ReportConfig{
			Enabled:   true,
			OutputDir: "./reports",
			Formats:   []string{},
		}
		rm := NewReportManager(config)

		mockReporter1 := &MockReporter{}
		mockReporter1.On("IsEnabled").Return(true)
		mockReporter1.On("GenerateReport", testResults).Return(nil)
		mockReporter1.On("GetName").Return("mock1").Maybe()

		mockReporter2 := &MockReporter{}
		mockReporter2.On("IsEnabled").Return(true)
		mockReporter2.On("GenerateReport", testResults).Return(nil)
		mockReporter2.On("GetName").Return("mock2").Maybe()

		rm.AddReporter(mockReporter1)
		rm.AddReporter(mockReporter2)

		err := rm.GenerateReports(testResults)
		assert.NoError(t, err)

		mockReporter1.AssertExpectations(t)
		mockReporter2.AssertExpectations(t)
	})

	t.Run("disabled reporter skipped", func(t *testing.T) {
		config := &config.ReportConfig{
			Enabled:   true,
			OutputDir: "./reports",
			Formats:   []string{},
		}
		rm := NewReportManager(config)

		mockReporter := &MockReporter{}
		mockReporter.On("IsEnabled").Return(false)
		mockReporter.On("GetName").Return("disabled_mock").Maybe()
		// GenerateReport should not be called for disabled reporters

		rm.AddReporter(mockReporter)

		err := rm.GenerateReports(testResults)
		assert.NoError(t, err)

		mockReporter.AssertExpectations(t)
	})

	t.Run("reporter failure", func(t *testing.T) {
		config := &config.ReportConfig{
			Enabled:   true,
			OutputDir: "./reports",
			Formats:   []string{},
		}
		rm := NewReportManager(config)

		mockReporter1 := &MockReporter{}
		mockReporter1.On("IsEnabled").Return(true)
		mockReporter1.On("GenerateReport", testResults).Return(errors.New("reporter 1 failed"))
		mockReporter1.On("GetName").Return("mock1").Maybe()

		mockReporter2 := &MockReporter{}
		mockReporter2.On("IsEnabled").Return(true)
		mockReporter2.On("GenerateReport", testResults).Return(nil)
		mockReporter2.On("GetName").Return("mock2").Maybe()

		rm.AddReporter(mockReporter1)
		rm.AddReporter(mockReporter2)

		err := rm.GenerateReports(testResults)
		assert.Error(t, err)

		mockReporter1.AssertExpectations(t)
		mockReporter2.AssertExpectations(t)
	})

	t.Run("disabled config", func(t *testing.T) {
		config := &config.ReportConfig{
			Enabled:   false,
			OutputDir: "./reports",
			Formats:   []string{},
		}
		rm := NewReportManager(config)

		mockReporter := &MockReporter{}
		rm.AddReporter(mockReporter)

		err := rm.GenerateReports(testResults)
		assert.NoError(t, err)

		// No expectations set because GenerateReport should not be called
		mockReporter.AssertExpectations(t)
	})

	t.Run("nil config", func(t *testing.T) {
		rm := NewReportManager(nil)

		mockReporter := &MockReporter{}
		rm.AddReporter(mockReporter)

		err := rm.GenerateReports(testResults)
		assert.NoError(t, err)

		// No expectations set because GenerateReport should not be called
		mockReporter.AssertExpectations(t)
	})
}

func TestJSONReporter(t *testing.T) {
	config := &config.ReportConfig{
		OutputDir: "./test-reports",
	}
	reporter := NewJSONReporter(config)

	assert.Equal(t, "JSONReporter", reporter.GetName())
	assert.True(t, reporter.IsEnabled())
}

func TestHTMLReporter(t *testing.T) {
	config := &config.ReportConfig{
		OutputDir: "./test-reports",
	}
	reporter := NewHTMLReporter(config)

	assert.Equal(t, "HTMLReporter", reporter.GetName())
	assert.True(t, reporter.IsEnabled())
}

func TestXMLReporter(t *testing.T) {
	config := &config.ReportConfig{
		OutputDir: "./test-reports",
	}
	reporter := NewXMLReporter(config)

	assert.Equal(t, "XMLReporter", reporter.GetName())
	assert.True(t, reporter.IsEnabled())
}

func TestJUnitReporter(t *testing.T) {
	config := &config.ReportConfig{
		OutputDir: "./test-reports",
	}
	reporter := NewJUnitReporter(config)

	assert.Equal(t, "JUnitReporter", reporter.GetName())
	assert.True(t, reporter.IsEnabled())
}

func TestJSONReporter_GenerateReport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &config.ReportConfig{
		OutputDir: tempDir,
	}
	reporter := NewJSONReporter(config)

	testResults := &core.TestResults{
		SuiteName:    "Test Suite",
		StartTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		TotalTests:   3,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []core.TestCaseResult{
			{
				Name:      "Test 1",
				Status:    core.TestStatusPassed,
				Duration:  time.Second * 2,
				StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 2, 0, time.UTC),
			},
			{
				Name:      "Test 2",
				Status:    core.TestStatusFailed,
				Duration:  time.Second * 3,
				Error:     errors.New("assertion failed"),
				StartTime: time.Date(2024, 1, 1, 10, 0, 2, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
			},
		},
	}

	err := reporter.GenerateReport(testResults)
	assert.NoError(t, err)

	// Verify file was created
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	// Verify file content
	filename := files[0].Name()
	assert.True(t, strings.HasPrefix(filename, "test-results-"))
	assert.True(t, strings.HasSuffix(filename, ".json"))

	content, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	var parsedResults core.TestResults
	err = json.Unmarshal(content, &parsedResults)
	assert.NoError(t, err)

	assert.Equal(t, testResults.SuiteName, parsedResults.SuiteName)
	assert.Equal(t, testResults.TotalTests, parsedResults.TotalTests)
	assert.Equal(t, testResults.PassedTests, parsedResults.PassedTests)
	assert.Equal(t, testResults.FailedTests, parsedResults.FailedTests)
	assert.Len(t, parsedResults.TestCases, 2)
}

func TestHTMLReporter_GenerateReport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &config.ReportConfig{
		OutputDir: tempDir,
	}
	reporter := NewHTMLReporter(config)

	testResults := &core.TestResults{
		SuiteName:    "Integration Test Suite",
		StartTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		TotalTests:   4,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 1,
		ErrorTests:   0,
		TestCases: []core.TestCaseResult{
			{
				Name:      "UI Test",
				Status:    core.TestStatusPassed,
				Duration:  time.Second * 5,
				StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
			},
			{
				Name:      "API Test",
				Status:    core.TestStatusFailed,
				Duration:  time.Second * 3,
				Error:     errors.New("HTTP 500 error"),
				StartTime: time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 8, 0, time.UTC),
			},
		},
	}

	err := reporter.GenerateReport(testResults)
	assert.NoError(t, err)

	// Verify file was created
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	// Verify file content
	filename := files[0].Name()
	assert.True(t, strings.HasPrefix(filename, "test-results-"))
	assert.True(t, strings.HasSuffix(filename, ".html"))

	content, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	htmlContent := string(content)

	// Verify HTML structure and content
	assert.Contains(t, htmlContent, "<!DOCTYPE html>")
	assert.Contains(t, htmlContent, "<title>Test Results - Integration Test Suite</title>")
	assert.Contains(t, htmlContent, "Integration Test Suite")
	assert.Contains(t, htmlContent, "UI Test")
	assert.Contains(t, htmlContent, "API Test")
	assert.Contains(t, htmlContent, "HTTP 500 error")

	// Verify summary statistics
	assert.Contains(t, htmlContent, "4") // Total tests
	assert.Contains(t, htmlContent, "2") // Passed tests
	assert.Contains(t, htmlContent, "1") // Failed tests
	assert.Contains(t, htmlContent, "1") // Skipped tests
}

func TestXMLReporter_GenerateReport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &config.ReportConfig{
		OutputDir: tempDir,
	}
	reporter := NewXMLReporter(config)

	testResults := &core.TestResults{
		SuiteName:    "Test Suite",
		StartTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []core.TestCaseResult{
			{
				Name:     "Test 1",
				Status:   core.TestStatusPassed,
				Duration: time.Second * 2,
			},
			{
				Name:     "Test 2",
				Status:   core.TestStatusFailed,
				Duration: time.Second * 3,
				Error:    errors.New("test failed"),
			},
		},
	}

	err := reporter.GenerateReport(testResults)
	assert.NoError(t, err)

	// Verify file was created
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	// Verify file content
	filename := files[0].Name()
	assert.True(t, strings.HasPrefix(filename, "test-results-"))
	assert.True(t, strings.HasSuffix(filename, ".xml"))

	content, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	xmlContent := string(content)

	// Verify XML structure and content
	assert.Contains(t, xmlContent, `<?xml version="1.0" encoding="UTF-8"?>`)
	assert.Contains(t, xmlContent, "<testsuites>")
	assert.Contains(t, xmlContent, `<testsuite name="Test Suite"`)
	assert.Contains(t, xmlContent, `tests="2"`)
	assert.Contains(t, xmlContent, `failures="1"`)
	assert.Contains(t, xmlContent, `<testcase name="Test 1"`)
	assert.Contains(t, xmlContent, `<testcase name="Test 2"`)
	assert.Contains(t, xmlContent, `<failure message="test failed">`)
}

func TestJUnitReporter_GenerateReport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &config.ReportConfig{
		OutputDir: tempDir,
	}
	reporter := NewJUnitReporter(config)

	testResults := &core.TestResults{
		SuiteName:    "Test Suite",
		StartTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		TotalTests:   3,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 1,
		ErrorTests:   0,
		TestCases: []core.TestCaseResult{
			{
				Name:     "Test 1",
				Status:   core.TestStatusPassed,
				Duration: time.Second * 2,
			},
			{
				Name:     "Test 2",
				Status:   core.TestStatusFailed,
				Duration: time.Second * 3,
				Error:    errors.New("test failed"),
			},
			{
				Name:     "Test 3",
				Status:   core.TestStatusSkipped,
				Duration: 0,
			},
		},
	}

	err := reporter.GenerateReport(testResults)
	assert.NoError(t, err)

	// Verify file was created
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	// Verify file content
	filename := files[0].Name()
	assert.True(t, strings.HasPrefix(filename, "junit-results-"))
	assert.True(t, strings.HasSuffix(filename, ".xml"))

	content, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	xmlContent := string(content)

	// Verify JUnit XML structure and content
	assert.Contains(t, xmlContent, `<?xml version="1.0" encoding="UTF-8"?>`)
	assert.Contains(t, xmlContent, `<testsuite name="Test Suite"`)
	assert.Contains(t, xmlContent, `tests="3"`)
	assert.Contains(t, xmlContent, `failures="1"`)
	assert.Contains(t, xmlContent, `skipped="1"`)
	assert.Contains(t, xmlContent, `timestamp="2024-01-01T10:00:00Z"`)
	assert.Contains(t, xmlContent, `<testcase classname="Test Suite" name="Test 1"`)
	assert.Contains(t, xmlContent, `<failure message="test failed" type="AssertionError">`)
	assert.Contains(t, xmlContent, `<skipped />`)
}

func TestReportManager_Integration(t *testing.T) {
	tempDir := t.TempDir()

	config := &config.ReportConfig{
		Enabled:     true,
		OutputDir:   tempDir,
		Formats:     []string{"json", "html"},
		IncludeLogs: true,
		Compress:    false,
	}

	rm := NewReportManager(config)

	testResults := &core.TestResults{
		SuiteName:    "Integration Test",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Minute),
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []core.TestCaseResult{
			{
				Name:     "Test 1",
				Status:   core.TestStatusPassed,
				Duration: time.Second,
			},
			{
				Name:     "Test 2",
				Status:   core.TestStatusFailed,
				Duration: time.Second * 2,
				Error:    errors.New("test failed"),
			},
		},
	}

	err := rm.GenerateReports(testResults)
	assert.NoError(t, err)

	// Verify both JSON and HTML files were created
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var jsonFile, htmlFile string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			jsonFile = file.Name()
		} else if strings.HasSuffix(file.Name(), ".html") {
			htmlFile = file.Name()
		}
	}

	assert.NotEmpty(t, jsonFile)
	assert.NotEmpty(t, htmlFile)

	// Verify JSON content
	jsonContent, err := os.ReadFile(filepath.Join(tempDir, jsonFile))
	assert.NoError(t, err)

	var parsedResults core.TestResults
	err = json.Unmarshal(jsonContent, &parsedResults)
	assert.NoError(t, err)
	assert.Equal(t, "Integration Test", parsedResults.SuiteName)

	// Verify HTML content
	htmlContent, err := os.ReadFile(filepath.Join(tempDir, htmlFile))
	assert.NoError(t, err)
	assert.Contains(t, string(htmlContent), "Integration Test")
	assert.Contains(t, string(htmlContent), "Test 1")
	assert.Contains(t, string(htmlContent), "Test 2")
}
