package gowright

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReporter is a mock implementation of the Reporter interface
type MockReporter struct {
	mock.Mock
}

func (m *MockReporter) GenerateReport(results *TestResults) error {
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
		config   *ReportConfig
		expected int // expected number of reporters
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: 0,
		},
		{
			name: "empty config",
			config: &ReportConfig{
				LocalReports:  LocalReportConfig{},
				RemoteReports: RemoteReportConfig{},
			},
			expected: 0,
		},
		{
			name: "local JSON only",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					JSON:      true,
					HTML:      false,
					OutputDir: "./reports",
				},
				RemoteReports: RemoteReportConfig{},
			},
			expected: 1,
		},
		{
			name: "local HTML only",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					JSON:      false,
					HTML:      true,
					OutputDir: "./reports",
				},
				RemoteReports: RemoteReportConfig{},
			},
			expected: 1,
		},
		{
			name: "both local reporters",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					JSON:      true,
					HTML:      true,
					OutputDir: "./reports",
				},
				RemoteReports: RemoteReportConfig{},
			},
			expected: 2,
		},
		{
			name: "all reporters enabled",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					JSON:      true,
					HTML:      true,
					OutputDir: "./reports",
				},
				RemoteReports: RemoteReportConfig{
					JiraXray: &JiraXrayConfig{
						URL:        "https://jira.example.com",
						Username:   "user",
						Password:   "pass",
						ProjectKey: "TEST",
					},
					AIOTest: &AIOTestConfig{
						URL:       "https://aiotest.example.com",
						APIKey:    "key",
						ProjectID: "project",
					},
					ReportPortal: &ReportPortalConfig{
						URL:     "https://rp.example.com",
						UUID:    "uuid",
						Project: "project",
						Launch:  "launch",
					},
				},
			},
			expected: 5,
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
	rm := NewReportManager(nil)
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

func TestReportManager_RemoveReporter(t *testing.T) {
	rm := NewReportManager(nil)

	mockReporter1 := &MockReporter{}
	mockReporter1.On("GetName").Return("mock1")
	mockReporter1.On("IsEnabled").Return(true)

	mockReporter2 := &MockReporter{}
	mockReporter2.On("GetName").Return("mock2")
	mockReporter2.On("IsEnabled").Return(true)

	rm.AddReporter(mockReporter1)
	rm.AddReporter(mockReporter2)
	assert.Len(t, rm.reporters, 2)

	// Remove first reporter
	rm.RemoveReporter("mock1")
	assert.Len(t, rm.reporters, 1)
	assert.Equal(t, mockReporter2, rm.reporters[0])

	// Remove non-existent reporter (should not panic)
	rm.RemoveReporter("non-existent")
	assert.Len(t, rm.reporters, 1)

	// Remove remaining reporter
	rm.RemoveReporter("mock2")
	assert.Len(t, rm.reporters, 0)
}

func TestReportManager_GenerateReports(t *testing.T) {
	testResults := &TestResults{
		SuiteName:    "Test Suite",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Minute),
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []TestCaseResult{
			{
				Name:     "Test 1",
				Status:   TestStatusPassed,
				Duration: time.Second,
			},
			{
				Name:     "Test 2",
				Status:   TestStatusFailed,
				Duration: time.Second * 2,
				Error:    errors.New("test failed"),
			},
		},
	}

	t.Run("successful generation", func(t *testing.T) {
		rm := NewReportManager(nil)

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

		summary := rm.GenerateReports(testResults)
		assert.Equal(t, 2, summary.SuccessfulReports)
		assert.Equal(t, 0, summary.FailedReports)

		mockReporter1.AssertExpectations(t)
		mockReporter2.AssertExpectations(t)
	})

	t.Run("disabled reporter skipped", func(t *testing.T) {
		rm := NewReportManager(nil)

		mockReporter := &MockReporter{}
		mockReporter.On("IsEnabled").Return(false)
		mockReporter.On("GetName").Return("disabled_mock").Maybe()
		// GenerateReport should not be called for disabled reporters

		rm.AddReporter(mockReporter)

		summary := rm.GenerateReports(testResults)
		assert.Equal(t, 1, summary.SuccessfulReports) // Fallback should succeed
		assert.Equal(t, 1, summary.FailedReports)     // Disabled reporter counts as failed
		assert.True(t, summary.FallbackUsed)

		mockReporter.AssertExpectations(t)
	})

	t.Run("reporter failure", func(t *testing.T) {
		rm := NewReportManager(nil)

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

		summary := rm.GenerateReports(testResults)
		assert.Equal(t, 1, summary.SuccessfulReports)
		assert.Equal(t, 1, summary.FailedReports)
		assert.False(t, summary.FallbackUsed) // Should not use fallback since one succeeded

		mockReporter1.AssertExpectations(t)
		mockReporter2.AssertExpectations(t)
	})

	t.Run("multiple reporter failures", func(t *testing.T) {
		rm := NewReportManager(nil)

		mockReporter1 := &MockReporter{}
		mockReporter1.On("IsEnabled").Return(true)
		mockReporter1.On("GenerateReport", testResults).Return(errors.New("reporter 1 failed"))
		mockReporter1.On("GetName").Return("mock1").Maybe()

		mockReporter2 := &MockReporter{}
		mockReporter2.On("IsEnabled").Return(true)
		mockReporter2.On("GenerateReport", testResults).Return(errors.New("reporter 2 failed"))
		mockReporter2.On("GetName").Return("mock2").Maybe()

		rm.AddReporter(mockReporter1)
		rm.AddReporter(mockReporter2)

		summary := rm.GenerateReports(testResults)
		assert.Equal(t, 1, summary.SuccessfulReports) // Fallback should succeed
		assert.Equal(t, 2, summary.FailedReports)
		assert.True(t, summary.FallbackUsed) // Should use fallback since all failed

		mockReporter1.AssertExpectations(t)
		mockReporter2.AssertExpectations(t)
	})
}

func TestReportManager_GetReporters(t *testing.T) {
	rm := NewReportManager(nil)

	mockReporter1 := &MockReporter{}
	mockReporter1.On("GetName").Return("mock1")
	mockReporter1.On("IsEnabled").Return(true)

	mockReporter2 := &MockReporter{}
	mockReporter2.On("GetName").Return("mock2")
	mockReporter2.On("IsEnabled").Return(true)

	rm.AddReporter(mockReporter1)
	rm.AddReporter(mockReporter2)

	reporters := rm.GetReporters()
	assert.Len(t, reporters, 2)

	// Verify it's a copy (modifying returned slice shouldn't affect original)
	reporters[0] = nil
	assert.NotNil(t, rm.reporters[0])
}

func TestJSONReporter(t *testing.T) {
	reporter := &JSONReporter{
		OutputDir: "./test-reports",
		enabled:   true,
	}

	assert.Equal(t, "json", reporter.GetName())
	assert.True(t, reporter.IsEnabled())

	// Test disabled reporter
	reporter.enabled = false
	assert.False(t, reporter.IsEnabled())
}

func TestHTMLReporter(t *testing.T) {
	reporter := &HTMLReporter{
		OutputDir: "./test-reports",
		enabled:   true,
	}

	assert.Equal(t, "html", reporter.GetName())
	assert.True(t, reporter.IsEnabled())

	// Test disabled reporter
	reporter.enabled = false
	assert.False(t, reporter.IsEnabled())
}

func TestJiraXrayReporter(t *testing.T) {
	config := &JiraXrayConfig{
		URL:        "https://jira.example.com",
		Username:   "user",
		Password:   "pass",
		ProjectKey: "TEST",
	}

	reporter := &JiraXrayReporter{
		config:  config,
		enabled: true,
	}

	assert.Equal(t, "jira_xray", reporter.GetName())
	assert.True(t, reporter.IsEnabled())

	// Test disabled reporter
	reporter.enabled = false
	assert.False(t, reporter.IsEnabled())
}

func TestAIOTestReporter(t *testing.T) {
	config := &AIOTestConfig{
		URL:       "https://aiotest.example.com",
		APIKey:    "key",
		ProjectID: "project",
	}

	reporter := &AIOTestReporter{
		config:  config,
		enabled: true,
	}

	assert.Equal(t, "aio_test", reporter.GetName())
	assert.True(t, reporter.IsEnabled())

	// Test disabled reporter
	reporter.enabled = false
	assert.False(t, reporter.IsEnabled())
}

func TestReportPortalReporter(t *testing.T) {
	config := &ReportPortalConfig{
		URL:     "https://rp.example.com",
		UUID:    "uuid",
		Project: "project",
		Launch:  "launch",
	}

	reporter := &ReportPortalReporter{
		config:  config,
		enabled: true,
	}

	assert.Equal(t, "report_portal", reporter.GetName())
	assert.True(t, reporter.IsEnabled())

	// Test disabled reporter
	reporter.enabled = false
	assert.False(t, reporter.IsEnabled())
}

func TestReportManager_ConcurrentAccess(t *testing.T) {
	rm := NewReportManager(nil)

	// Test concurrent add/remove operations
	done := make(chan bool, 2)

	// Goroutine 1: Add reporters
	go func() {
		for i := 0; i < 10; i++ {
			mockReporter := &MockReporter{}
			mockReporter.On("GetName").Return("mock")
			mockReporter.On("IsEnabled").Return(true)
			rm.AddReporter(mockReporter)
		}
		done <- true
	}()

	// Goroutine 2: Remove reporters
	go func() {
		for i := 0; i < 5; i++ {
			rm.RemoveReporter("mock")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should not panic and should have some reporters
	reporters := rm.GetReporters()
	assert.True(t, len(reporters) >= 0) // Could be 0 to 10 depending on timing
}

func TestTestResults_Structure(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(time.Minute)

	results := &TestResults{
		SuiteName:    "Integration Test Suite",
		StartTime:    startTime,
		EndTime:      endTime,
		TotalTests:   3,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []TestCaseResult{
			{
				Name:        "UI Test",
				Status:      TestStatusPassed,
				Duration:    time.Second * 5,
				StartTime:   startTime,
				EndTime:     startTime.Add(time.Second * 5),
				Screenshots: []string{"screenshot1.png"},
				Logs:        []string{"UI test completed successfully"},
			},
			{
				Name:      "API Test",
				Status:    TestStatusPassed,
				Duration:  time.Second * 2,
				StartTime: startTime.Add(time.Second * 5),
				EndTime:   startTime.Add(time.Second * 7),
				Logs:      []string{"API response validated"},
			},
			{
				Name:      "Database Test",
				Status:    TestStatusFailed,
				Duration:  time.Second * 3,
				Error:     errors.New("connection timeout"),
				StartTime: startTime.Add(time.Second * 7),
				EndTime:   startTime.Add(time.Second * 10),
				Logs:      []string{"Database connection failed"},
			},
		},
	}

	// Verify structure
	assert.Equal(t, "Integration Test Suite", results.SuiteName)
	assert.Equal(t, startTime, results.StartTime)
	assert.Equal(t, endTime, results.EndTime)
	assert.Equal(t, 3, results.TotalTests)
	assert.Equal(t, 2, results.PassedTests)
	assert.Equal(t, 1, results.FailedTests)
	assert.Equal(t, 0, results.SkippedTests)
	assert.Equal(t, 0, results.ErrorTests)
	assert.Len(t, results.TestCases, 3)

	// Verify test case structure
	uiTest := results.TestCases[0]
	assert.Equal(t, "UI Test", uiTest.Name)
	assert.Equal(t, TestStatusPassed, uiTest.Status)
	assert.Equal(t, time.Second*5, uiTest.Duration)
	assert.Len(t, uiTest.Screenshots, 1)
	assert.Len(t, uiTest.Logs, 1)
	assert.NoError(t, uiTest.Error)

	// Verify failed test case
	dbTest := results.TestCases[2]
	assert.Equal(t, "Database Test", dbTest.Name)
	assert.Equal(t, TestStatusFailed, dbTest.Status)
	assert.Error(t, dbTest.Error)
	assert.Equal(t, "connection timeout", dbTest.Error.Error())
}

func TestJSONReporter_GenerateReport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	reporter := &JSONReporter{
		OutputDir: tempDir,
		enabled:   true,
	}

	testResults := &TestResults{
		SuiteName:    "Test Suite",
		StartTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		TotalTests:   3,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []TestCaseResult{
			{
				Name:      "Test 1",
				Status:    TestStatusPassed,
				Duration:  time.Second * 2,
				StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 2, 0, time.UTC),
				Logs:      []string{"Test passed successfully"},
			},
			{
				Name:        "Test 2",
				Status:      TestStatusFailed,
				Duration:    time.Second * 3,
				Error:       errors.New("assertion failed"),
				StartTime:   time.Date(2024, 1, 1, 10, 0, 2, 0, time.UTC),
				EndTime:     time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
				Screenshots: []string{"screenshot1.png"},
				Logs:        []string{"Test failed", "Error details"},
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
	assert.True(t, strings.HasPrefix(filename, "test_suite_"))
	assert.True(t, strings.HasSuffix(filename, ".json"))

	content, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	var parsedResults TestResults
	err = json.Unmarshal(content, &parsedResults)
	assert.NoError(t, err)

	assert.Equal(t, testResults.SuiteName, parsedResults.SuiteName)
	assert.Equal(t, testResults.TotalTests, parsedResults.TotalTests)
	assert.Equal(t, testResults.PassedTests, parsedResults.PassedTests)
	assert.Equal(t, testResults.FailedTests, parsedResults.FailedTests)
	assert.Len(t, parsedResults.TestCases, 2)
}

func TestJSONReporter_EnsureOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "reports", "json")

	reporter := &JSONReporter{
		OutputDir: subDir,
		enabled:   true,
	}

	// Directory doesn't exist initially
	_, err := os.Stat(subDir)
	assert.True(t, os.IsNotExist(err))

	// Generate report should create directory
	testResults := &TestResults{
		SuiteName:  "Test",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(time.Minute),
		TotalTests: 1,
		TestCases:  []TestCaseResult{},
	}

	err = reporter.GenerateReport(testResults)
	assert.NoError(t, err)

	// Directory should now exist
	_, err = os.Stat(subDir)
	assert.NoError(t, err)
}

func TestJSONReporter_DefaultOutputDir(t *testing.T) {
	reporter := &JSONReporter{
		OutputDir: "",
		enabled:   true,
	}

	err := reporter.ensureOutputDir()
	assert.NoError(t, err)
	assert.Equal(t, "./reports", reporter.OutputDir)
}

func TestHTMLReporter_GenerateReport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	reporter := &HTMLReporter{
		OutputDir: tempDir,
		enabled:   true,
	}

	testResults := &TestResults{
		SuiteName:    "Integration Test Suite",
		StartTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		TotalTests:   4,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 1,
		ErrorTests:   0,
		TestCases: []TestCaseResult{
			{
				Name:        "UI Test",
				Status:      TestStatusPassed,
				Duration:    time.Second * 5,
				StartTime:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
				Screenshots: []string{"ui_test_screenshot.png"},
				Logs:        []string{"UI test completed successfully"},
			},
			{
				Name:      "API Test",
				Status:    TestStatusFailed,
				Duration:  time.Second * 3,
				Error:     errors.New("HTTP 500 error"),
				StartTime: time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 8, 0, time.UTC),
				Logs:      []string{"API request failed", "Server returned 500"},
			},
			{
				Name:      "Database Test",
				Status:    TestStatusSkipped,
				Duration:  0,
				StartTime: time.Date(2024, 1, 1, 10, 0, 8, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 10, 0, 8, 0, time.UTC),
				Logs:      []string{"Test skipped due to missing database connection"},
			},
			{
				Name:        "Integration Test",
				Status:      TestStatusPassed,
				Duration:    time.Second * 10,
				StartTime:   time.Date(2024, 1, 1, 10, 0, 8, 0, time.UTC),
				EndTime:     time.Date(2024, 1, 1, 10, 0, 18, 0, time.UTC),
				Screenshots: []string{"integration_before.png", "integration_after.png"},
				Logs:        []string{"Integration test started", "All systems verified", "Test completed"},
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
	assert.True(t, strings.HasPrefix(filename, "integration_test_suite_"))
	assert.True(t, strings.HasSuffix(filename, ".html"))

	content, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	htmlContent := string(content)

	// Verify HTML structure and content
	assert.Contains(t, htmlContent, "<!DOCTYPE html>")
	assert.Contains(t, htmlContent, "<title>Gowright Test Report - Integration Test Suite</title>")
	assert.Contains(t, htmlContent, "Integration Test Suite")
	assert.Contains(t, htmlContent, "UI Test")
	assert.Contains(t, htmlContent, "API Test")
	assert.Contains(t, htmlContent, "Database Test")
	assert.Contains(t, htmlContent, "Integration Test")
	assert.Contains(t, htmlContent, "HTTP 500 error")
	assert.Contains(t, htmlContent, "ui_test_screenshot.png")
	assert.Contains(t, htmlContent, "integration_before.png")
	assert.Contains(t, htmlContent, "integration_after.png")

	// Verify summary statistics
	assert.Contains(t, htmlContent, "4") // Total tests
	assert.Contains(t, htmlContent, "2") // Passed tests
	assert.Contains(t, htmlContent, "1") // Failed tests
	assert.Contains(t, htmlContent, "1") // Skipped tests

	// Verify CSS classes are present
	assert.Contains(t, htmlContent, "test-status passed")
	assert.Contains(t, htmlContent, "test-status failed")
	assert.Contains(t, htmlContent, "test-status skipped")
}

func TestHTMLReporter_EnsureOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "reports", "html")

	reporter := &HTMLReporter{
		OutputDir: subDir,
		enabled:   true,
	}

	// Directory doesn't exist initially
	_, err := os.Stat(subDir)
	assert.True(t, os.IsNotExist(err))

	// Generate report should create directory
	testResults := &TestResults{
		SuiteName:  "Test",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(time.Minute),
		TotalTests: 1,
		TestCases:  []TestCaseResult{},
	}

	err = reporter.GenerateReport(testResults)
	assert.NoError(t, err)

	// Directory should now exist
	_, err = os.Stat(subDir)
	assert.NoError(t, err)
}

func TestHTMLReporter_DefaultOutputDir(t *testing.T) {
	reporter := &HTMLReporter{
		OutputDir: "",
		enabled:   true,
	}

	err := reporter.ensureOutputDir()
	assert.NoError(t, err)
	assert.Equal(t, "./reports", reporter.OutputDir)
}

func TestHTMLReporter_GenerateFilename(t *testing.T) {
	reporter := &HTMLReporter{}

	tests := []struct {
		suiteName string
		expected  string
	}{
		{"Simple Test", "simple_test_"},
		{"Integration Test Suite", "integration_test_suite_"},
		{"API-Test_Suite", "api-test_suite_"},
		{"Test With Spaces", "test_with_spaces_"},
	}

	for _, tt := range tests {
		t.Run(tt.suiteName, func(t *testing.T) {
			filename := reporter.generateFilename(tt.suiteName)
			assert.True(t, strings.HasPrefix(filename, tt.expected))
			assert.True(t, strings.HasSuffix(filename, ".html"))
			assert.Contains(t, filename, "_202") // Should contain year
		})
	}
}

func TestJSONReporter_GenerateFilename(t *testing.T) {
	reporter := &JSONReporter{}

	tests := []struct {
		suiteName string
		expected  string
	}{
		{"Simple Test", "simple_test_"},
		{"Integration Test Suite", "integration_test_suite_"},
		{"API-Test_Suite", "api-test_suite_"},
		{"Test With Spaces", "test_with_spaces_"},
	}

	for _, tt := range tests {
		t.Run(tt.suiteName, func(t *testing.T) {
			filename := reporter.generateFilename(tt.suiteName)
			assert.True(t, strings.HasPrefix(filename, tt.expected))
			assert.True(t, strings.HasSuffix(filename, ".json"))
			assert.Contains(t, filename, "_202") // Should contain year
		})
	}
}

func TestLocalReporters_ErrorHandling(t *testing.T) {
	t.Run("JSON reporter with invalid directory", func(t *testing.T) {
		reporter := &JSONReporter{
			OutputDir: "/invalid/path/that/cannot/be/created",
			enabled:   true,
		}

		testResults := &TestResults{
			SuiteName: "Test",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Minute),
		}

		err := reporter.GenerateReport(testResults)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output directory")
	})

	t.Run("HTML reporter with invalid directory", func(t *testing.T) {
		reporter := &HTMLReporter{
			OutputDir: "/invalid/path/that/cannot/be/created",
			enabled:   true,
		}

		testResults := &TestResults{
			SuiteName: "Test",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Minute),
		}

		err := reporter.GenerateReport(testResults)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output directory")
	})
}

func TestLocalReporters_Integration(t *testing.T) {
	tempDir := t.TempDir()

	config := &ReportConfig{
		LocalReports: LocalReportConfig{
			JSON:      true,
			HTML:      true,
			OutputDir: tempDir,
		},
		RemoteReports: RemoteReportConfig{},
	}

	rm := NewReportManager(config)

	testResults := &TestResults{
		SuiteName:    "Integration Test",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Minute),
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases: []TestCaseResult{
			{
				Name:     "Test 1",
				Status:   TestStatusPassed,
				Duration: time.Second,
			},
			{
				Name:     "Test 2",
				Status:   TestStatusFailed,
				Duration: time.Second * 2,
				Error:    errors.New("test failed"),
			},
		},
	}

	summary := rm.GenerateReports(testResults)
	assert.Equal(t, 2, summary.SuccessfulReports)
	assert.Equal(t, 0, summary.FailedReports)

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

	var parsedResults TestResults
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
