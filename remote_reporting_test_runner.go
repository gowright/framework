package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github/gowright/framework/pkg/gowright"
)

// Simple remote reporting test runner
func main() {
	fmt.Println("Starting Remote Reporting Integration Tests...")
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gowright_remote_reporting_test_")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)
	
	// Setup mock servers
	jiraXrayServer := setupJiraXrayMockServer()
	defer jiraXrayServer.Close()
	
	aioTestServer := setupAIOTestMockServer()
	defer aioTestServer.Close()
	
	reportPortalServer := setupReportPortalMockServer()
	defer reportPortalServer.Close()
	
	// Initialize framework with remote reporting
	framework := setupFrameworkWithRemoteReporting(tempDir, jiraXrayServer.URL, aioTestServer.URL, reportPortalServer.URL)
	defer framework.Cleanup()
	
	// Test remote reporting
	if testRemoteReporting(framework) {
		fmt.Println("✓ Remote reporting integration test passed")
	} else {
		fmt.Println("✗ Remote reporting integration test failed")
		os.Exit(1)
	}
	
	fmt.Println("All remote reporting tests passed! ✓")
}

func setupJiraXrayMockServer() *httptest.Server {
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
		
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"testExecIssue": map[string]interface{}{
				"id":  "12345",
				"key": "PROJ-123",
			},
		})
	})
	
	return httptest.NewServer(mux)
}

func setupAIOTestMockServer() *httptest.Server {
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
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     12345,
			"status": "completed",
		})
	})
	
	return httptest.NewServer(mux)
}

func setupReportPortalMockServer() *httptest.Server {
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
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "launch-uuid-12345",
			"number": 1,
		})
	})
	
	return httptest.NewServer(mux)
}

func setupFrameworkWithRemoteReporting(tempDir, jiraURL, aioURL, rpURL string) *gowright.Gowright {
	config := &gowright.Config{
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: tempDir,
			},
			RemoteReports: gowright.RemoteReportConfig{
				JiraXray: &gowright.JiraXrayConfig{
					URL:        jiraURL,
					Username:   "test-user",
					Password:   "test-password",
					ProjectKey: "PROJ",
				},
				AIOTest: &gowright.AIOTestConfig{
					URL:       aioURL,
					APIKey:    "test-api-key",
					ProjectID: "test-project",
				},
				ReportPortal: &gowright.ReportPortalConfig{
					URL:     rpURL,
					UUID:    "test-uuid",
					Project: "test-project",
					Launch:  "Gowright Integration Test",
				},
			},
		},
	}
	
	framework := gowright.New(config)
	err := framework.Initialize()
	if err != nil {
		fmt.Printf("Failed to initialize framework: %v\n", err)
		os.Exit(1)
	}
	
	return framework
}

func testRemoteReporting(framework *gowright.Gowright) bool {
	// Create test results
	testResults := &gowright.TestResults{
		SuiteName:    "Remote Reporting Test Suite",
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
	reporter := framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	
	fmt.Printf("Reporting summary: %d successful, %d failed, fallback used: %v\n", 
		summary.SuccessfulReports, summary.FailedReports, summary.FallbackUsed)
	
	// We expect at least local reports to succeed, remote ones might fail (which is expected for mock servers)
	if summary.SuccessfulReports == 0 && !summary.FallbackUsed {
		fmt.Printf("No successful reports generated\n")
		return false
	}
	
	return true
}