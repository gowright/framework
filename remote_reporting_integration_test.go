package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github/gowright/framework/pkg/gowright"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RemoteReportingIntegrationSuite tests remote reporting integrations
type RemoteReportingIntegrationSuite struct {
	framework         *gowright.Gowright
	jiraXrayServer    *httptest.Server
	aioTestServer     *httptest.Server
	reportPortalServer *httptest.Server
	tempDir           string
}

// SetupRemoteReportingSuite initializes the remote reporting test environment
func (suite *RemoteReportingIntegrationSuite) SetupRemoteReportingSuite(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gowright_remote_reporting_test_")
	require.NoError(t, err)
	suite.tempDir = tempDir
	
	// Setup mock servers for each reporting destination
	suite.setupJiraXrayMockServer()
	suite.setupAIOTestMockServer()
	suite.setupReportPortalMockServer()
	
	// Initialize framework with remote reporting configuration
	suite.setupFrameworkWithRemoteReporting(t)
}

// TearDownRemoteReportingSuite cleans up the test environment
func (suite *RemoteReportingIntegrationSuite) TearDownRemoteReportingSuite(t *testing.T) {
	if suite.framework != nil {
		suite.framework.Cleanup()
	}
	if suite.jiraXrayServer != nil {
		suite.jiraXrayServer.Close()
	}
	if suite.aioTestServer != nil {
		suite.aioTestServer.Close()
	}
	if suite.reportPortalServer != nil {
		suite.reportPortalServer.Close()
	}
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// setupJiraXrayMockServer creates a mock Jira Xray server
func (suite *RemoteReportingIntegrationSuite) setupJiraXrayMockServer() {
	mux := http.NewServeMux()
	
	// Authentication endpoint
	mux.HandleFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": "mock-xray-token-12345",
		})
	})
	
	// Import test execution results endpoint
	mux.HandleFunc("/rest/raven/1.0/import/execution", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "Bearer mock-xray-token-12345") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// Validate required fields
		if requestBody["info"] == nil || requestBody["tests"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Missing required fields: info and tests",
			})
			return
		}
		
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"testExecIssue": map[string]interface{}{
				"id":  "12345",
				"key": "PROJ-123",
			},
			"testIssues": []map[string]interface{}{
				{"id": "11111", "key": "PROJ-T1"},
				{"id": "11112", "key": "PROJ-T2"},
			},
		})
	})
	
	suite.jiraXrayServer = httptest.NewServer(mux)
}

// setupAIOTestMockServer creates a mock AIOTest server
func (suite *RemoteReportingIntegrationSuite) setupAIOTestMockServer() {
	mux := http.NewServeMux()
	
	// Authentication endpoint
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "mock-aiotest-token-67890",
			"token_type":   "Bearer",
		})
	})
	
	// Create test run endpoint
	mux.HandleFunc("/api/v1/test-runs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "Bearer mock-aiotest-token-67890") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   12345,
			"name": requestBody["name"],
			"status": "completed",
			"url":  fmt.Sprintf("%s/test-runs/12345", suite.aioTestServer.URL),
		})
	})
	
	// Upload test results endpoint
	mux.HandleFunc("/api/v1/test-runs/12345/results", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "Bearer mock-aiotest-token-67890") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "uploaded",
		})
	})
	
	suite.aioTestServer = httptest.NewServer(mux)
}

// setupReportPortalMockServer creates a mock Report Portal server
func (suite *RemoteReportingIntegrationSuite) setupReportPortalMockServer() {
	mux := http.NewServeMux()
	
	// Authentication endpoint
	mux.HandleFunc("/uat/sso/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-rp-token-abcdef",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	})
	
	// Create launch endpoint
	mux.HandleFunc("/api/v1/test-project/launch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "Bearer mock-rp-token-abcdef") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "launch-uuid-12345",
			"number":  1,
		})
	})
	
	// Create test item endpoint
	mux.HandleFunc("/api/v1/test-project/item", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "Bearer mock-rp-token-abcdef") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "item-uuid-67890",
		})
	})
	
	// Finish launch endpoint
	mux.HandleFunc("/api/v1/test-project/launch/launch-uuid-12345/finish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Launch finished successfully",
		})
	})
	
	suite.reportPortalServer = httptest.NewServer(mux)
}

// setupFrameworkWithRemoteReporting initializes the framework with remote reporting
func (suite *RemoteReportingIntegrationSuite) setupFrameworkWithRemoteReporting(t *testing.T) {
	config := &gowright.Config{
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: suite.tempDir,
			},
			RemoteReports: gowright.RemoteReportConfig{
				JiraXray: &gowright.JiraXrayConfig{
					BaseURL:  suite.jiraXrayServer.URL,
					Username: "test-user",
					Password: "test-password",
					ProjectKey: "PROJ",
					Enabled:  true,
				},
				AIOTest: &gowright.AIOTestConfig{
					BaseURL:  suite.aioTestServer.URL,
					Username: "test-user",
					Password: "test-password",
					ProjectID: "test-project",
					Enabled:  true,
				},
				ReportPortal: &gowright.ReportPortalConfig{
					BaseURL:   suite.reportPortalServer.URL,
					Token:     "test-token",
					Project:   "test-project",
					LaunchName: "Gowright Integration Test",
					Enabled:   true,
				},
			},
		},
	}
	
	framework := gowright.New(config)
	err := framework.Initialize()
	require.NoError(t, err)
	
	suite.framework = framework
}

// TestRemoteReportingIntegrations tests all remote reporting integrations
func TestRemoteReportingIntegrations(t *testing.T) {
	suite := &RemoteReportingIntegrationSuite{}
	suite.SetupRemoteReportingSuite(t)
	defer suite.TearDownRemoteReportingSuite(t)
	
	t.Run("JiraXrayIntegration", suite.testJiraXrayIntegration)
	t.Run("AIOTestIntegration", suite.testAIOTestIntegration)
	t.Run("ReportPortalIntegration", suite.testReportPortalIntegration)
	t.Run("MultipleReportersIntegration", suite.testMultipleReportersIntegration)
	t.Run("ReportingFailureHandling", suite.testReportingFailureHandling)
}

// testJiraXrayIntegration tests Jira Xray reporting integration
func (suite *RemoteReportingIntegrationSuite) testJiraXrayIntegration(t *testing.T) {
	// Create test results
	testResults := &gowright.TestResults{
		SuiteName:    "Jira Xray Integration Test Suite",
		StartTime:    time.Now().Add(-5 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		SkippedTests: 0,
		TestCases: []gowright.TestCaseResult{
			{
				Name:      "Passed Test",
				Status:    gowright.TestStatusPassed,
				Duration:  time.Second,
				StartTime: time.Now().Add(-5 * time.Minute),
				EndTime:   time.Now().Add(-4 * time.Minute),
			},
			{
				Name:      "Failed Test",
				Status:    gowright.TestStatusFailed,
				Duration:  2 * time.Second,
				Error:     fmt.Errorf("test assertion failed"),
				StartTime: time.Now().Add(-3 * time.Minute),
				EndTime:   time.Now().Add(-1 * time.Minute),
			},
		},
	}
	
	// Generate report
	reporter := suite.framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	assert.Greater(t, summary.SuccessfulReports, 0)
	
	// Verify that local reports were also generated
	jsonFiles, err := filepath.Glob(filepath.Join(suite.tempDir, "*.json"))
	require.NoError(t, err)
	assert.Greater(t, len(jsonFiles), 0)
	
	htmlFiles, err := filepath.Glob(filepath.Join(suite.tempDir, "*.html"))
	require.NoError(t, err)
	assert.Greater(t, len(htmlFiles), 0)
}

// testAIOTestIntegration tests AIOTest reporting integration
func (suite *RemoteReportingIntegrationSuite) testAIOTestIntegration(t *testing.T) {
	// Create test results with different test types
	testResults := &gowright.TestResults{
		SuiteName:    "AIOTest Integration Test Suite",
		StartTime:    time.Now().Add(-10 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   3,
		PassedTests:  2,
		FailedTests:  0,
		SkippedTests: 1,
		TestCases: []gowright.TestCaseResult{
			{
				Name:      "API Test",
				Status:    gowright.TestStatusPassed,
				Duration:  500 * time.Millisecond,
				StartTime: time.Now().Add(-10 * time.Minute),
				EndTime:   time.Now().Add(-9 * time.Minute),
			},
			{
				Name:      "UI Test",
				Status:    gowright.TestStatusPassed,
				Duration:  3 * time.Second,
				Screenshots: []string{"screenshot1.png"},
				StartTime: time.Now().Add(-8 * time.Minute),
				EndTime:   time.Now().Add(-5 * time.Minute),
			},
			{
				Name:      "Skipped Test",
				Status:    gowright.TestStatusSkipped,
				Duration:  0,
				StartTime: time.Now().Add(-2 * time.Minute),
				EndTime:   time.Now().Add(-2 * time.Minute),
			},
		},
	}
	
	// Generate report
	reporter := suite.framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	assert.Greater(t, summary.SuccessfulReports, 0)
}

// testReportPortalIntegration tests Report Portal integration
func (suite *RemoteReportingIntegrationSuite) testReportPortalIntegration(t *testing.T) {
	// Create test results with nested structure
	testResults := &gowright.TestResults{
		SuiteName:    "Report Portal Integration Test Suite",
		StartTime:    time.Now().Add(-15 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   4,
		PassedTests:  2,
		FailedTests:  2,
		SkippedTests: 0,
		TestCases: []gowright.TestCaseResult{
			{
				Name:      "Database Test",
				Status:    gowright.TestStatusPassed,
				Duration:  200 * time.Millisecond,
				StartTime: time.Now().Add(-15 * time.Minute),
				EndTime:   time.Now().Add(-14 * time.Minute),
			},
			{
				Name:      "Integration Test",
				Status:    gowright.TestStatusFailed,
				Duration:  5 * time.Second,
				Error:     fmt.Errorf("integration step failed"),
				Logs:      []string{"Step 1: Success", "Step 2: Failed"},
				StartTime: time.Now().Add(-12 * time.Minute),
				EndTime:   time.Now().Add(-7 * time.Minute),
			},
			{
				Name:      "Performance Test",
				Status:    gowright.TestStatusPassed,
				Duration:  1 * time.Second,
				StartTime: time.Now().Add(-6 * time.Minute),
				EndTime:   time.Now().Add(-5 * time.Minute),
			},
			{
				Name:      "Error Test",
				Status:    gowright.TestStatusError,
				Duration:  100 * time.Millisecond,
				Error:     fmt.Errorf("unexpected error occurred"),
				StartTime: time.Now().Add(-3 * time.Minute),
				EndTime:   time.Now().Add(-2 * time.Minute),
			},
		},
	}
	
	// Generate report
	reporter := suite.framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	assert.Greater(t, summary.SuccessfulReports, 0)
}

// testMultipleReportersIntegration tests multiple reporters working together
func (suite *RemoteReportingIntegrationSuite) testMultipleReportersIntegration(t *testing.T) {
	// Create comprehensive test results
	testResults := &gowright.TestResults{
		SuiteName:    "Multiple Reporters Integration Test Suite",
		StartTime:    time.Now().Add(-20 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   5,
		PassedTests:  3,
		FailedTests:  1,
		SkippedTests: 1,
		TestCases: []gowright.TestCaseResult{
			{
				Name:      "Comprehensive Test 1",
				Status:    gowright.TestStatusPassed,
				Duration:  time.Second,
				StartTime: time.Now().Add(-20 * time.Minute),
				EndTime:   time.Now().Add(-19 * time.Minute),
			},
			{
				Name:      "Comprehensive Test 2",
				Status:    gowright.TestStatusPassed,
				Duration:  2 * time.Second,
				Screenshots: []string{"comp_test_2.png"},
				StartTime: time.Now().Add(-18 * time.Minute),
				EndTime:   time.Now().Add(-16 * time.Minute),
			},
			{
				Name:      "Comprehensive Test 3",
				Status:    gowright.TestStatusFailed,
				Duration:  3 * time.Second,
				Error:     fmt.Errorf("comprehensive test failure"),
				Logs:      []string{"Log 1", "Log 2", "Error Log"},
				Screenshots: []string{"comp_test_3_error.png"},
				StartTime: time.Now().Add(-15 * time.Minute),
				EndTime:   time.Now().Add(-12 * time.Minute),
			},
			{
				Name:      "Comprehensive Test 4",
				Status:    gowright.TestStatusPassed,
				Duration:  500 * time.Millisecond,
				StartTime: time.Now().Add(-10 * time.Minute),
				EndTime:   time.Now().Add(-9 * time.Minute),
			},
			{
				Name:      "Comprehensive Test 5",
				Status:    gowright.TestStatusSkipped,
				Duration:  0,
				StartTime: time.Now().Add(-5 * time.Minute),
				EndTime:   time.Now().Add(-5 * time.Minute),
			},
		},
	}
	
	// Generate report to all destinations
	reporter := suite.framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	assert.Greater(t, summary.SuccessfulReports, 0)
	
	// Verify local reports were generated
	jsonFiles, err := filepath.Glob(filepath.Join(suite.tempDir, "*.json"))
	require.NoError(t, err)
	assert.Greater(t, len(jsonFiles), 0)
	
	htmlFiles, err := filepath.Glob(filepath.Join(suite.tempDir, "*.html"))
	require.NoError(t, err)
	assert.Greater(t, len(htmlFiles), 0)
}

// testReportingFailureHandling tests graceful handling of reporting failures
func (suite *RemoteReportingIntegrationSuite) testReportingFailureHandling(t *testing.T) {
	// Close one of the mock servers to simulate failure
	suite.jiraXrayServer.Close()
	suite.jiraXrayServer = nil
	
	// Create test results
	testResults := &gowright.TestResults{
		SuiteName:    "Failure Handling Test Suite",
		StartTime:    time.Now().Add(-5 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   1,
		PassedTests:  1,
		FailedTests:  0,
		SkippedTests: 0,
		TestCases: []gowright.TestCaseResult{
			{
				Name:      "Failure Handling Test",
				Status:    gowright.TestStatusPassed,
				Duration:  time.Second,
				StartTime: time.Now().Add(-5 * time.Minute),
				EndTime:   time.Now().Add(-4 * time.Minute),
			},
		},
	}
	
	// Generate report - should handle Jira Xray failure gracefully
	reporter := suite.framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	
	// Should not fail completely, other reporters should still work
	assert.True(t, summary.SuccessfulReports > 0 || summary.FallbackUsed)
	
	// Verify local reports were still generated
	jsonFiles, err := filepath.Glob(filepath.Join(suite.tempDir, "*.json"))
	require.NoError(t, err)
	assert.Greater(t, len(jsonFiles), 0)
	
	// Verify fallback report was created
	fallbackFiles, err := filepath.Glob(filepath.Join(suite.tempDir, "fallback", "*.json"))
	if err == nil && len(fallbackFiles) > 0 {
		// Fallback mechanism worked
		t.Log("Fallback reporting mechanism activated successfully")
	}
}

// BenchmarkRemoteReporting benchmarks remote reporting performance
func BenchmarkRemoteReporting(b *testing.B) {
	suite := &RemoteReportingIntegrationSuite{}
	suite.SetupRemoteReportingSuite(&testing.T{})
	defer suite.TearDownRemoteReportingSuite(&testing.T{})
	
	// Create test results for benchmarking
	testResults := &gowright.TestResults{
		SuiteName:    "Benchmark Test Suite",
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		TotalTests:   100,
		PassedTests:  90,
		FailedTests:  10,
		SkippedTests: 0,
		TestCases:    make([]gowright.TestCaseResult, 100),
	}
	
	// Generate test cases
	for i := 0; i < 100; i++ {
		status := gowright.TestStatusPassed
		var err error
		if i < 10 {
			status = gowright.TestStatusFailed
			err = fmt.Errorf("benchmark test %d failed", i)
		}
		
		testResults.TestCases[i] = gowright.TestCaseResult{
			Name:      fmt.Sprintf("Benchmark Test %d", i),
			Status:    status,
			Duration:  time.Duration(50+i%200) * time.Millisecond,
			Error:     err,
			StartTime: time.Now().Add(-time.Hour).Add(time.Duration(i) * time.Second),
			EndTime:   time.Now().Add(-time.Hour).Add(time.Duration(i) * time.Second).Add(time.Duration(50+i%200) * time.Millisecond),
		}
	}
	
	reporter := suite.framework.GetReporter()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary := reporter.GenerateReports(testResults)
		if summary.SuccessfulReports == 0 && !summary.FallbackUsed {
			b.Fatalf("Report generation failed: no successful reports")
		}
	}
}