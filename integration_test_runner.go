package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github/gowright/framework/pkg/gowright"
	_ "github.com/mattn/go-sqlite3"
)

// Simple integration test runner
func main() {
	fmt.Println("Starting Gowright Framework Integration Tests...")
	
	// Create temporary directory for test artifacts
	tempDir, err := os.MkdirTemp("", "gowright_integration_test_")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)
	
	reportDir := filepath.Join(tempDir, "reports")
	err = os.MkdirAll(reportDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create reports directory: %v\n", err)
		os.Exit(1)
	}
	
	// Setup test HTTP server
	testServer := setupTestServer()
	defer testServer.Close()
	
	// Setup test database
	testDB := setupTestDatabase(tempDir)
	defer testDB.Close()
	
	// Initialize Gowright framework
	framework := setupFramework(testServer.URL, tempDir, reportDir)
	defer framework.Cleanup()
	
	// Run tests
	fmt.Println("✓ Framework initialization test passed")
	
	if testAPIModule(framework, testServer.URL) {
		fmt.Println("✓ API testing module test passed")
	} else {
		fmt.Println("✗ API testing module test failed")
		os.Exit(1)
	}
	
	if testDatabaseModule(framework, tempDir) {
		fmt.Println("✓ Database testing module test passed")
	} else {
		fmt.Println("✗ Database testing module test failed")
		os.Exit(1)
	}
	
	if testReportingSystem(framework, reportDir) {
		fmt.Println("✓ Reporting system test passed")
	} else {
		fmt.Println("✗ Reporting system test failed")
		os.Exit(1)
	}
	
	fmt.Println("All integration tests passed! ✓")
}

func setupTestServer() *httptest.Server {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	
	// User management endpoints
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			users := []map[string]interface{}{
				{"id": 1, "name": "John Doe", "email": "john@example.com"},
				{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		case http.MethodPost:
			var user map[string]interface{}
			json.NewDecoder(r.Body).Decode(&user)
			user["id"] = 3
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	
	return httptest.NewServer(mux)
}

func setupTestDatabase(tempDir string) *sql.DB {
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	
	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("Failed to create table: %v\n", err)
		os.Exit(1)
	}
	
	// Insert test data
	_, err = db.Exec(`
		INSERT INTO users (name, email) VALUES 
		('John Doe', 'john@example.com'),
		('Jane Smith', 'jane@example.com')
	`)
	if err != nil {
		fmt.Printf("Failed to insert test data: %v\n", err)
		os.Exit(1)
	}
	
	return db
}

func setupFramework(serverURL, tempDir, reportDir string) *gowright.Gowright {
	config := &gowright.Config{
		BrowserConfig: &gowright.BrowserConfig{
			Headless:    true,
			Timeout:     30 * time.Second,
			WindowSize:  &gowright.WindowSize{Width: 1920, Height: 1080},
		},
		APIConfig: &gowright.APIConfig{
			BaseURL: serverURL,
			Timeout: 10 * time.Second,
			Headers: map[string]string{
				"User-Agent": "Gowright-Integration-Test",
			},
		},
		DatabaseConfig: &gowright.DatabaseConfig{
			Connections: map[string]*gowright.DBConnection{
				"test": {
					Driver: "sqlite3",
					DSN:    filepath.Join(tempDir, "test.db"),
				},
			},
		},
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: reportDir,
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

func testAPIModule(framework *gowright.Gowright, serverURL string) bool {
	// Create API tester
	apiTester := gowright.NewAPITester(framework.GetConfig().APIConfig)
	err := apiTester.Initialize(framework.GetConfig().APIConfig)
	if err != nil {
		fmt.Printf("Failed to initialize API tester: %v\n", err)
		return false
	}
	defer apiTester.Cleanup()
	
	// Test GET request
	response, err := apiTester.Get("/health", nil)
	if err != nil {
		fmt.Printf("GET request failed: %v\n", err)
		return false
	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Expected status 200, got %d\n", response.StatusCode)
		return false
	}
	
	// Test POST request
	user := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}
	response, err = apiTester.Post("/users", user, nil)
	if err != nil {
		fmt.Printf("POST request failed: %v\n", err)
		return false
	}
	if response.StatusCode != http.StatusCreated {
		fmt.Printf("Expected status 201, got %d\n", response.StatusCode)
		return false
	}
	
	return true
}

func testDatabaseModule(framework *gowright.Gowright, tempDir string) bool {
	// Create database tester
	dbTester := gowright.NewDatabaseTester()
	err := dbTester.Initialize(framework.GetConfig().DatabaseConfig)
	if err != nil {
		fmt.Printf("Failed to initialize database tester: %v\n", err)
		return false
	}
	defer dbTester.Cleanup()
	
	// Test database connection
	err = dbTester.Connect("test")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return false
	}
	
	// Test query execution
	result, err := dbTester.Execute("test", "SELECT COUNT(*) as count FROM users")
	if err != nil {
		fmt.Printf("Failed to execute query: %v\n", err)
		return false
	}
	if len(result.Rows) != 1 {
		fmt.Printf("Expected 1 row, got %d\n", len(result.Rows))
		return false
	}
	
	count, ok := result.Rows[0]["count"].(int64)
	if !ok {
		fmt.Printf("Failed to get count from result\n")
		return false
	}
	if count != 2 {
		fmt.Printf("Expected count 2, got %d\n", count)
		return false
	}
	
	// Test database test execution
	dbTest := &gowright.DatabaseTest{
		Name:       "User Count Test",
		Connection: "test",
		Query:      "SELECT COUNT(*) as count FROM users",
		Expected: &gowright.DatabaseExpectation{
			RowCount: 1,
		},
	}
	
	testResult := dbTester.ExecuteTest(dbTest)
	if testResult.Status != gowright.TestStatusPassed {
		fmt.Printf("Database test failed: %v\n", testResult.Error)
		return false
	}
	
	return true
}

func testReportingSystem(framework *gowright.Gowright, reportDir string) bool {
	// Create test results
	testResults := &gowright.TestResults{
		SuiteName:    "Integration Test Suite",
		StartTime:    time.Now().Add(-5 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   3,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 0,
		TestCases: []gowright.TestCaseResult{
			{
				Name:      "Test 1",
				Status:    gowright.TestStatusPassed,
				Duration:  time.Second,
				StartTime: time.Now().Add(-5 * time.Minute),
				EndTime:   time.Now().Add(-4 * time.Minute),
			},
			{
				Name:      "Test 2",
				Status:    gowright.TestStatusPassed,
				Duration:  2 * time.Second,
				StartTime: time.Now().Add(-4 * time.Minute),
				EndTime:   time.Now().Add(-2 * time.Minute),
			},
			{
				Name:      "Test 3",
				Status:    gowright.TestStatusFailed,
				Duration:  time.Second,
				Error:     fmt.Errorf("test assertion failed"),
				StartTime: time.Now().Add(-2 * time.Minute),
				EndTime:   time.Now(),
			},
		},
	}
	
	// Generate report
	reporter := framework.GetReporter()
	summary := reporter.GenerateReports(testResults)
	if summary.SuccessfulReports == 0 {
		fmt.Printf("No successful reports generated\n")
		return false
	}
	
	// Verify JSON report file exists
	jsonFiles, err := filepath.Glob(filepath.Join(reportDir, "*.json"))
	if err != nil {
		fmt.Printf("Failed to check for JSON files: %v\n", err)
		return false
	}
	if len(jsonFiles) == 0 {
		fmt.Printf("No JSON report files found\n")
		return false
	}
	
	// Verify HTML report file exists
	htmlFiles, err := filepath.Glob(filepath.Join(reportDir, "*.html"))
	if err != nil {
		fmt.Printf("Failed to check for HTML files: %v\n", err)
		return false
	}
	if len(htmlFiles) == 0 {
		fmt.Printf("No HTML report files found\n")
		return false
	}
	
	// Verify JSON content
	jsonData, err := os.ReadFile(jsonFiles[0])
	if err != nil {
		fmt.Printf("Failed to read JSON file: %v\n", err)
		return false
	}
	
	var reportData gowright.TestResults
	err = json.Unmarshal(jsonData, &reportData)
	if err != nil {
		fmt.Printf("Failed to parse JSON report: %v\n", err)
		return false
	}
	if reportData.SuiteName != testResults.SuiteName {
		fmt.Printf("Suite name mismatch in JSON report\n")
		return false
	}
	if reportData.TotalTests != testResults.TotalTests {
		fmt.Printf("Total tests mismatch in JSON report\n")
		return false
	}
	
	return true
}