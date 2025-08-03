package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github/gowright/framework/pkg/gowright"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGowrightFrameworkIntegration tests the complete framework end-to-end
func TestGowrightFrameworkIntegration(t *testing.T) {
	// Setup test environment
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	// Create test server for API testing
	server := createTestServer()
	defer server.Close()

	// Setup test database
	dbPath := setupTestDatabase(t, testDir)

	// Initialize Gowright framework
	framework := initializeFramework(t, server.URL, dbPath, testDir)

	// Run comprehensive integration tests
	t.Run("CompleteWorkflow", func(t *testing.T) {
		testCompleteWorkflow(t, framework)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, framework)
	})

	t.Run("ReportingSystem", func(t *testing.T) {
		testReportingSystem(t, framework, testDir)
	})

	t.Run("PerformanceBenchmarks", func(t *testing.T) {
		benchmarkFrameworkPerformance(t, framework)
	})

	t.Run("BackwardCompatibility", func(t *testing.T) {
		testBackwardCompatibility(t, framework)
	})
}

// setupTestEnvironment creates a temporary directory for test artifacts
func setupTestEnvironment(t *testing.T) string {
	testDir, err := os.MkdirTemp("", "gowright-integration-test-*")
	require.NoError(t, err, "Failed to create test directory")
	
	// Create subdirectories for different test artifacts
	dirs := []string{"reports", "screenshots", "logs", "data"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(testDir, dir), 0755)
		require.NoError(t, err, "Failed to create %s directory", dir)
	}
	
	return testDir
}

// cleanupTestEnvironment removes test artifacts
func cleanupTestEnvironment(testDir string) {
	os.RemoveAll(testDir)
}

// createTestServer creates a mock HTTP server for API testing
func createTestServer() *httptest.Server {
	mux := http.NewServeMux()
	
	// User endpoints
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":1,"name":"John Doe","email":"john@example.com"}]`))
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":2,"name":"Jane Smith","email":"jane@example.com"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	
	// Product endpoints
	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"Test Product","price":99.99}]`))
	})
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Error endpoint for testing error handling
	mux.HandleFunc("/api/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
	})
	
	return httptest.NewServer(mux)
}

// setupTestDatabase creates and initializes a test database
func setupTestDatabase(t *testing.T, testDir string) string {
	dbPath := filepath.Join(testDir, "test.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err, "Failed to open test database")
	defer db.Close()
	
	// Create test tables
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			total DECIMAL(10,2) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		);
	`
	
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create database schema")
	
	// Insert test data
	testData := `
		INSERT INTO users (name, email) VALUES 
			('John Doe', 'john@example.com'),
			('Jane Smith', 'jane@example.com');
		
		INSERT INTO products (name, price) VALUES 
			('Test Product 1', 99.99),
			('Test Product 2', 149.99);
	`
	
	_, err = db.Exec(testData)
	require.NoError(t, err, "Failed to insert test data")
	
	return dbPath
}

// initializeFramework creates and configures the Gowright framework
func initializeFramework(t *testing.T, serverURL, dbPath, testDir string) *gowright.Gowright {
	config := &gowright.Config{
		LogLevel: "DEBUG",
		Parallel: true,
		MaxRetries: 3,
		BrowserConfig: &gowright.BrowserConfig{
			Headless: true,
			Timeout:  30 * time.Second,
			WindowSize: &gowright.WindowSize{
				Width:  1920,
				Height: 1080,
			},
		},
		APIConfig: &gowright.APIConfig{
			BaseURL: serverURL,
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent":   "Gowright-Integration-Test/1.0",
				"Content-Type": "application/json",
			},
		},
		DatabaseConfig: &gowright.DatabaseConfig{
			Connections: map[string]*gowright.DBConnection{
				"test": {
					Driver:       "sqlite3",
					DSN:          dbPath,
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
			},
		},
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: filepath.Join(testDir, "reports"),
			},
		},
	}
	
	framework := gowright.New(config)
	require.NotNil(t, framework, "Failed to initialize Gowright framework")
	
	return framework
}

// testCompleteWorkflow tests a complete end-to-end workflow
func testCompleteWorkflow(t *testing.T, framework *gowright.Gowright) {
	// Create a test suite using the framework's test suite manager
	testSuite := &gowright.TestSuite{
		Name:  "CompleteWorkflowSuite",
		Tests: make([]gowright.Test, 0),
		SetupFunc: func() error {
			log.Println("Setting up complete workflow test suite")
			return nil
		},
		TeardownFunc: func() error {
			log.Println("Tearing down complete workflow test suite")
			return nil
		},
	}
	
	// Set the test suite in the framework
	framework.SetTestSuite(testSuite)
	
	// Create test suite manager
	tsm := framework.CreateTestSuiteManager()
	
	// Create API tester for the tests
	apiTester := &gowright.APITesterImpl{}
	err := apiTester.Initialize(framework.GetConfig().APIConfig)
	require.NoError(t, err, "Failed to initialize API tester")
	
	// Create database tester for the tests
	dbTester := &gowright.DatabaseTesterImpl{}
	err = dbTester.Initialize(framework.GetConfig().DatabaseConfig)
	require.NoError(t, err, "Failed to initialize database tester")
	
	// Test 1: API operations
	apiTest := gowright.NewAPITestBuilder("GetUsers", "GET", "/api/users").
		WithTester(apiTester).
		ExpectStatus(200).
		ExpectHeader("Content-Type", "application/json").
		ExpectJSONPath("$[0].name", "John Doe").
		ExpectJSONPath("$[0].email", "john@example.com").
		Build()
	
	tsm.RegisterTest(apiTest)
	
	// Test 2: Database operations
	dbTest := gowright.NewDatabaseTestBuilder("VerifyUserData", "test").
		WithTester(dbTester).
		WithQuery("SELECT COUNT(*) as count FROM users").
		ExpectRowCount(1).
		Build()
	
	tsm.RegisterTest(dbTest)
	
	// Execute the test suite
	results, err := tsm.ExecuteTestSuite()
	require.NoError(t, err, "Failed to execute complete workflow test suite")
	
	// Verify results
	assert.NotNil(t, results, "Test results should not be nil")
	assert.Greater(t, results.TotalTests, 0, "Should have executed tests")
	
	// Log results for debugging
	t.Logf("Complete workflow test results: Total=%d, Passed=%d, Failed=%d", 
		results.TotalTests, results.PassedTests, results.FailedTests)
}

// testErrorHandling tests the framework's error handling capabilities
func testErrorHandling(t *testing.T, framework *gowright.Gowright) {
	suite := framework.NewTestSuite("ErrorHandlingSuite")
	
	// Test API error handling
	suite.AddTest(&gowright.APITest{
		Name:     "TestAPIError",
		Method:   "GET",
		Endpoint: "/api/error",
		Expected: &gowright.APIExpectation{
			StatusCode: 500,
			JSONPath: map[string]interface{}{
				"$.error": "Internal server error",
			},
		},
	})
	
	// Test database error handling
	suite.AddTest(&gowright.DatabaseTest{
		Name:       "TestInvalidQuery",
		Connection: "test",
		Query:      "SELECT * FROM nonexistent_table",
		Expected: &gowright.DatabaseExpectation{
			ShouldFail: true,
		},
	})
	
	// Execute error handling tests
	results, err := suite.Execute(context.Background())
	require.NoError(t, err, "Error handling test suite should execute without framework errors")
	
	// Verify that we handled errors gracefully
	assert.NotNil(t, results, "Error handling test results should not be nil")
	t.Logf("Error handling test results: Total=%d, Passed=%d, Failed=%d", 
		results.TotalTests, results.PassedTests, results.FailedTests)
}

// testReportingSystem tests the reporting system with all formats
func testReportingSystem(t *testing.T, framework *gowright.Gowright, testDir string) {
	suite := framework.NewTestSuite("ReportingSystemSuite")
	
	// Add a simple test to generate reports
	suite.AddTest(&gowright.APITest{
		Name:     "HealthCheck",
		Method:   "GET",
		Endpoint: "/health",
		Expected: &gowright.APIExpectation{
			StatusCode: 200,
		},
	})
	
	// Execute the test suite
	results, err := suite.Execute(context.Background())
	require.NoError(t, err, "Reporting system test suite should execute successfully")
	
	// Generate reports
	err = framework.GenerateReports(results)
	require.NoError(t, err, "Report generation should succeed")
	
	// Verify report files were created
	reportsDir := filepath.Join(testDir, "reports")
	
	// Check for JSON report
	jsonFiles, err := filepath.Glob(filepath.Join(reportsDir, "*.json"))
	require.NoError(t, err, "Should be able to search for JSON reports")
	assert.Greater(t, len(jsonFiles), 0, "Should have generated JSON reports")
	
	// Check for HTML report
	htmlFiles, err := filepath.Glob(filepath.Join(reportsDir, "*.html"))
	require.NoError(t, err, "Should be able to search for HTML reports")
	assert.Greater(t, len(htmlFiles), 0, "Should have generated HTML reports")
	
	t.Logf("Generated %d JSON reports and %d HTML reports", len(jsonFiles), len(htmlFiles))
}

// benchmarkFrameworkPerformance runs performance benchmarks
func benchmarkFrameworkPerformance(t *testing.T, framework *gowright.Gowright) {
	// Create a large test suite for performance testing
	suite := framework.NewTestSuite("PerformanceBenchmarkSuite")
	
	// Add multiple API tests
	for i := 0; i < 50; i++ {
		suite.AddTest(&gowright.APITest{
			Name:     fmt.Sprintf("PerformanceTest_%d", i),
			Method:   "GET",
			Endpoint: "/health",
			Expected: &gowright.APIExpectation{
				StatusCode: 200,
			},
		})
	}
	
	// Measure execution time
	startTime := time.Now()
	results, err := suite.Execute(context.Background())
	executionTime := time.Since(startTime)
	
	require.NoError(t, err, "Performance benchmark should execute successfully")
	assert.NotNil(t, results, "Performance benchmark results should not be nil")
	
	// Performance assertions
	assert.Less(t, executionTime, 60*time.Second, "Large test suite should complete within 60 seconds")
	assert.Equal(t, 50, results.TotalTests, "Should have executed all 50 tests")
	
	// Calculate performance metrics
	avgTestTime := executionTime / time.Duration(results.TotalTests)
	testsPerSecond := float64(results.TotalTests) / executionTime.Seconds()
	
	t.Logf("Performance metrics: Total time=%v, Avg per test=%v, Tests/sec=%.2f", 
		executionTime, avgTestTime, testsPerSecond)
	
	// Performance thresholds
	assert.Less(t, avgTestTime, 2*time.Second, "Average test time should be under 2 seconds")
	assert.Greater(t, testsPerSecond, 1.0, "Should execute at least 1 test per second")
}

// testBackwardCompatibility tests API backward compatibility
func testBackwardCompatibility(t *testing.T, framework *gowright.Gowright) {
	// Test that core interfaces haven't changed
	t.Run("CoreInterfaces", func(t *testing.T) {
		// Test that we can still create framework instances
		config := &gowright.Config{
			LogLevel: "INFO",
		}
		
		_, err := gowright.New(config)
		assert.NoError(t, err, "Should be able to create framework with minimal config")
	})
	
	t.Run("TestSuiteAPI", func(t *testing.T) {
		// Test that test suite API is stable
		suite := framework.NewTestSuite("BackwardCompatibilityTest")
		assert.NotNil(t, suite, "Should be able to create test suite")
		assert.Equal(t, "BackwardCompatibilityTest", suite.Name, "Test suite name should be preserved")
	})
	
	t.Run("ConfigurationAPI", func(t *testing.T) {
		// Test that configuration structure is backward compatible
		config := &gowright.Config{
			LogLevel:   "DEBUG",
			Parallel:   true,
			MaxRetries: 3,
		}
		
		assert.Equal(t, "DEBUG", config.LogLevel, "LogLevel field should be accessible")
		assert.True(t, config.Parallel, "Parallel field should be accessible")
		assert.Equal(t, 3, config.MaxRetries, "MaxRetries field should be accessible")
	})
}

// TestMain runs the integration tests
func TestMain(m *testing.M) {
	// Setup global test environment
	log.Println("Starting Gowright Framework Integration Tests")
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	log.Println("Completed Gowright Framework Integration Tests")
	
	os.Exit(code)
}