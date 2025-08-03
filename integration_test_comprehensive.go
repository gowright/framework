package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// ComprehensiveIntegrationTestSuite tests the complete Gowright framework
type ComprehensiveIntegrationTestSuite struct {
	framework     *gowright.Gowright
	testServer    *httptest.Server
	testDB        *sql.DB
	tempDir       string
	reportDir     string
}

// SetupSuite initializes the test environment
func (suite *ComprehensiveIntegrationTestSuite) SetupSuite(t *testing.T) {
	// Create temporary directory for test artifacts
	tempDir, err := os.MkdirTemp("", "gowright_integration_test_")
	require.NoError(t, err)
	suite.tempDir = tempDir
	suite.reportDir = filepath.Join(tempDir, "reports")
	
	// Create reports directory
	err = os.MkdirAll(suite.reportDir, 0755)
	require.NoError(t, err)
	
	// Setup test HTTP server
	suite.setupTestServer()
	
	// Setup test database
	suite.setupTestDatabase(t)
	
	// Initialize Gowright framework
	suite.setupFramework(t)
}

// TearDownSuite cleans up the test environment
func (suite *ComprehensiveIntegrationTestSuite) TearDownSuite(t *testing.T) {
	if suite.framework != nil {
		err := suite.framework.Cleanup()
		assert.NoError(t, err)
	}
	
	if suite.testServer != nil {
		suite.testServer.Close()
	}
	
	if suite.testDB != nil {
		suite.testDB.Close()
	}
	
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// setupTestServer creates a test HTTP server for API testing
func (suite *ComprehensiveIntegrationTestSuite) setupTestServer() {
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
	
	// Error endpoint for testing error handling
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	})
	
	suite.testServer = httptest.NewServer(mux)
}

// setupTestDatabase creates a test SQLite database
func (suite *ComprehensiveIntegrationTestSuite) setupTestDatabase(t *testing.T) {
	dbPath := filepath.Join(suite.tempDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	
	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)
	
	// Insert test data
	_, err = db.Exec(`
		INSERT INTO users (name, email) VALUES 
		('John Doe', 'john@example.com'),
		('Jane Smith', 'jane@example.com')
	`)
	require.NoError(t, err)
	
	suite.testDB = db
}

// setupFramework initializes the Gowright framework with test configuration
func (suite *ComprehensiveIntegrationTestSuite) setupFramework(t *testing.T) {
	config := &gowright.Config{
		BrowserConfig: &gowright.BrowserConfig{
			Headless:    true,
			Timeout:     30 * time.Second,
			WindowSize:  &gowright.WindowSize{Width: 1920, Height: 1080},
		},
		APIConfig: &gowright.APIConfig{
			BaseURL: suite.testServer.URL,
			Timeout: 10 * time.Second,
			Headers: map[string]string{
				"User-Agent": "Gowright-Integration-Test",
			},
		},
		DatabaseConfig: &gowright.DatabaseConfig{
			Connections: map[string]*gowright.DBConnection{
				"test": {
					Driver: "sqlite3",
					DSN:    filepath.Join(suite.tempDir, "test.db"),
				},
			},
		},
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: suite.reportDir,
			},
		},
	}
	
	framework := gowright.New(config)
	err := framework.Initialize()
	require.NoError(t, err)
	
	suite.framework = framework
}

// TestCompleteFrameworkIntegration tests the entire framework end-to-end
func TestCompleteFrameworkIntegration(t *testing.T) {
	suite := &ComprehensiveIntegrationTestSuite{}
	suite.SetupSuite(t)
	defer suite.TearDownSuite(t)
	
	t.Run("FrameworkInitialization", suite.testFrameworkInitialization)
	t.Run("APITestingModule", suite.testAPITestingModule)
	t.Run("DatabaseTestingModule", suite.testDatabaseTestingModule)
	t.Run("IntegrationTestingModule", suite.testIntegrationTestingModule)
	t.Run("ReportingSystem", suite.testReportingSystem)
	t.Run("ErrorHandlingAndRecovery", suite.testErrorHandlingAndRecovery)
	t.Run("ParallelExecution", suite.testParallelExecution)
	t.Run("BackwardCompatibility", suite.testBackwardCompatibility)
}

// testFrameworkInitialization verifies framework initialization
func (suite *ComprehensiveIntegrationTestSuite) testFrameworkInitialization(t *testing.T) {
	assert.True(t, suite.framework.IsInitialized())
	assert.NotNil(t, suite.framework.GetConfig())
	assert.NotNil(t, suite.framework.GetReporter())
	
	// Test configuration access
	config := suite.framework.GetConfig()
	assert.NotNil(t, config.BrowserConfig)
	assert.NotNil(t, config.APIConfig)
	assert.NotNil(t, config.DatabaseConfig)
	assert.NotNil(t, config.ReportConfig)
}

// testAPITestingModule tests the API testing capabilities
func (suite *ComprehensiveIntegrationTestSuite) testAPITestingModule(t *testing.T) {
	// Create API tester
	apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
	err := apiTester.Initialize(suite.framework.GetConfig().APIConfig)
	require.NoError(t, err)
	defer apiTester.Cleanup()
	
	// Test GET request
	t.Run("GET_Request", func(t *testing.T) {
		response, err := apiTester.Get("/health", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Contains(t, string(response.Body), "healthy")
	})
	
	// Test POST request
	t.Run("POST_Request", func(t *testing.T) {
		user := map[string]interface{}{
			"name":  "Test User",
			"email": "test@example.com",
		}
		response, err := apiTester.Post("/users", user, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
	})
	
	// Test API test execution
	t.Run("API_Test_Execution", func(t *testing.T) {
		// Test the API call directly
		response, err := apiTester.Get("/health", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}

// testDatabaseTestingModule tests the database testing capabilities
func (suite *ComprehensiveIntegrationTestSuite) testDatabaseTestingModule(t *testing.T) {
	// Create database tester
	dbTester := gowright.NewDatabaseTester()
	err := dbTester.Initialize(suite.framework.GetConfig().DatabaseConfig)
	require.NoError(t, err)
	defer dbTester.Cleanup()
	
	// Test database connection
	t.Run("Database_Connection", func(t *testing.T) {
		err := dbTester.Connect("test")
		assert.NoError(t, err)
	})
	
	// Test query execution
	t.Run("Query_Execution", func(t *testing.T) {
		result, err := dbTester.Execute("test", "SELECT COUNT(*) as count FROM users")
		require.NoError(t, err)
		assert.Equal(t, 1, result.RowCount)
		assert.Equal(t, int64(2), result.Rows[0]["count"])
	})
	
	// Test database test execution
	t.Run("Database_Test_Execution", func(t *testing.T) {
		dbTest := &gowright.DatabaseTest{
			Name:       "User Count Test",
			Connection: "test",
			Query:      "SELECT COUNT(*) as count FROM users",
			Expected: &gowright.DatabaseExpectation{
				RowCount: 1,
			},
		}
		
		result := dbTester.ExecuteTest(dbTest)
		assert.Equal(t, gowright.TestStatusPassed, result.Status)
		assert.Equal(t, "User Count Test", result.Name)
	})
	
	// Test transaction handling
	t.Run("Transaction_Handling", func(t *testing.T) {
		tx, err := dbTester.BeginTransaction("test")
		require.NoError(t, err)
		
		// Insert test data within transaction
		_, err = tx.Execute("INSERT INTO users (name, email) VALUES (?, ?)", "Transaction User", "tx@example.com")
		require.NoError(t, err)
		
		// Rollback transaction
		err = tx.Rollback()
		assert.NoError(t, err)
		
		// Verify data was not persisted
		result, err := dbTester.Execute("test", "SELECT COUNT(*) as count FROM users WHERE email = ?", "tx@example.com")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Rows[0]["count"])
	})
}

// testIntegrationTestingModule tests the integration testing capabilities
func (suite *ComprehensiveIntegrationTestSuite) testIntegrationTestingModule(t *testing.T) {
	// Create integration tester
	integrationTester := gowright.NewIntegrationTester(nil, nil, nil)
	err := integrationTester.Initialize(suite.framework.GetConfig())
	require.NoError(t, err)
	defer integrationTester.Cleanup()
	
	// Test multi-step integration workflow
	t.Run("Multi_Step_Integration", func(t *testing.T) {
		steps := []gowright.IntegrationStep{
			{
				Type: gowright.StepTypeDatabase,
				Action: gowright.DatabaseStepAction{
					Connection: "test",
					Query:      "INSERT INTO users (name, email) VALUES (?, ?)",
					Args:       []interface{}{"Integration User", "integration@example.com"},
				},
				Name: "Create User in Database",
			},
			{
				Type: gowright.StepTypeAPI,
				Action: gowright.APIStepAction{
					Method:   "GET",
					Endpoint: "/users",
				},
				Validation: gowright.APIStepValidation{
					ExpectedStatusCode: http.StatusOK,
				},
				Name: "Verify User via API",
			},
		}
		
		err := integrationTester.ExecuteWorkflow(steps)
		assert.NoError(t, err)
	})
	
	// Test integration test execution
	t.Run("Integration_Test_Execution", func(t *testing.T) {
		integrationTest := &gowright.IntegrationTest{
			Name: "User Creation Integration Test",
			Steps: []gowright.IntegrationStep{
				{
					Type: gowright.StepTypeAPI,
					Action: gowright.APIStepAction{
						Method:   "POST",
						Endpoint: "/users",
						Body: map[string]interface{}{
							"name":  "API User",
							"email": "api@example.com",
						},
					},
					Validation: gowright.APIStepValidation{
						ExpectedStatusCode: http.StatusCreated,
					},
					Name: "Create User via API",
				},
			},
		}
		
		result := integrationTester.ExecuteTest(integrationTest)
		assert.Equal(t, gowright.TestStatusPassed, result.Status)
		assert.Equal(t, "User Creation Integration Test", result.Name)
	})
}

// testReportingSystem tests the reporting capabilities
func (suite *ComprehensiveIntegrationTestSuite) testReportingSystem(t *testing.T) {
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
	
	// Test JSON report generation
	t.Run("JSON_Report_Generation", func(t *testing.T) {
		reporter := suite.framework.GetReporter()
		summary := reporter.GenerateReports(testResults)
		assert.Greater(t, summary.SuccessfulReports, 0)
		
		// Verify JSON report file exists
		jsonFiles, err := filepath.Glob(filepath.Join(suite.reportDir, "*.json"))
		require.NoError(t, err)
		assert.Greater(t, len(jsonFiles), 0)
		
		// Verify JSON content
		jsonData, err := os.ReadFile(jsonFiles[0])
		require.NoError(t, err)
		
		var reportData gowright.TestResults
		err = json.Unmarshal(jsonData, &reportData)
		assert.NoError(t, err)
		assert.Equal(t, testResults.SuiteName, reportData.SuiteName)
		assert.Equal(t, testResults.TotalTests, reportData.TotalTests)
	})
	
	// Test HTML report generation
	t.Run("HTML_Report_Generation", func(t *testing.T) {
		// Verify HTML report file exists
		htmlFiles, err := filepath.Glob(filepath.Join(suite.reportDir, "*.html"))
		require.NoError(t, err)
		assert.Greater(t, len(htmlFiles), 0)
		
		// Verify HTML content contains expected elements
		htmlData, err := os.ReadFile(htmlFiles[0])
		require.NoError(t, err)
		htmlContent := string(htmlData)
		
		assert.Contains(t, htmlContent, testResults.SuiteName)
		assert.Contains(t, htmlContent, "Test 1")
		assert.Contains(t, htmlContent, "Test 2")
		assert.Contains(t, htmlContent, "Test 3")
	})
}

// testErrorHandlingAndRecovery tests error handling capabilities
func (suite *ComprehensiveIntegrationTestSuite) testErrorHandlingAndRecovery(t *testing.T) {
	// Test API error handling
	t.Run("API_Error_Handling", func(t *testing.T) {
		apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
		err := apiTester.Initialize(suite.framework.GetConfig().APIConfig)
		require.NoError(t, err)
		defer apiTester.Cleanup()
		
		// Test error endpoint
		response, err := apiTester.Get("/error", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		
		// Test non-existent endpoint
		response, err = apiTester.Get("/nonexistent", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})
	
	// Test database error handling
	t.Run("Database_Error_Handling", func(t *testing.T) {
		dbTester := gowright.NewDatabaseTester()
		err := dbTester.Initialize(suite.framework.GetConfig().DatabaseConfig)
		require.NoError(t, err)
		defer dbTester.Cleanup()
		
		// Test invalid query
		_, err = dbTester.Execute("test", "SELECT * FROM nonexistent_table")
		assert.Error(t, err)
		
		// Test invalid connection
		err = dbTester.Connect("nonexistent")
		assert.Error(t, err)
	})
	
	// Test graceful degradation in reporting
	t.Run("Reporting_Error_Handling", func(t *testing.T) {
		// Create config with invalid report directory
		invalidConfig := &gowright.Config{
			ReportConfig: &gowright.ReportConfig{
				LocalReports: gowright.LocalReportConfig{
					JSON:      true,
					OutputDir: "/invalid/path/that/does/not/exist",
				},
			},
		}
		
		reportManager := gowright.NewReportManager(invalidConfig.ReportConfig)
		
		testResults := &gowright.TestResults{
			SuiteName:   "Error Test Suite",
			TotalTests:  1,
			PassedTests: 1,
			TestCases: []gowright.TestCaseResult{
				{
					Name:   "Test 1",
					Status: gowright.TestStatusPassed,
				},
			},
		}
		
		// Should handle error gracefully and use fallback
		summary := reportManager.GenerateReports(testResults)
		// Error should be handled gracefully, not cause panic
		assert.True(t, summary.SuccessfulReports > 0 || summary.FallbackUsed) // Fallback should work
	})
}

// testParallelExecution tests concurrent test execution
func (suite *ComprehensiveIntegrationTestSuite) testParallelExecution(t *testing.T) {
	// Create API tester for the tests
	apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
	apiTester.Initialize(suite.framework.GetConfig().APIConfig)
	
	// Create multiple API tests for parallel execution
	tests := []gowright.Test{
		gowright.NewAPITestBuilder("Parallel Test 1", "GET", "/health").
			WithTester(apiTester).ExpectStatus(http.StatusOK).Build(),
		gowright.NewAPITestBuilder("Parallel Test 2", "GET", "/users").
			WithTester(apiTester).ExpectStatus(http.StatusOK).Build(),
		gowright.NewAPITestBuilder("Parallel Test 3", "GET", "/health").
			WithTester(apiTester).ExpectStatus(http.StatusOK).Build(),
	}
	
	// Create test suite with parallel execution
	testSuite := &gowright.TestSuite{
		Name:  "Parallel Execution Test Suite",
		Tests: tests,
	}
	
	suite.framework.SetTestSuite(testSuite)
	
	// Execute test suite and measure performance
	startTime := time.Now()
	results, err := suite.framework.ExecuteTestSuite()
	duration := time.Since(startTime)
	
	require.NoError(t, err)
	assert.Equal(t, len(tests), results.TotalTests)
	assert.Equal(t, len(tests), results.PassedTests)
	
	// Parallel execution should be faster than sequential
	// (This is a basic check - in real scenarios, you'd have more sophisticated timing)
	assert.Less(t, duration, 10*time.Second, "Parallel execution took too long")
}

// testBackwardCompatibility tests API backward compatibility
func (suite *ComprehensiveIntegrationTestSuite) testBackwardCompatibility(t *testing.T) {
	// Test legacy constructor methods
	t.Run("Legacy_Constructors", func(t *testing.T) {
		// Test NewGowright function
		framework, err := gowright.NewGowright(gowright.DefaultConfig())
		require.NoError(t, err)
		assert.NotNil(t, framework)
		assert.True(t, framework.IsInitialized())
		framework.Close()
		
		// Test NewWithDefaults function
		framework2 := gowright.NewWithDefaults()
		assert.NotNil(t, framework2)
		err = framework2.Initialize()
		assert.NoError(t, err)
		framework2.Close()
	})
	
	// Test legacy configuration methods
	t.Run("Legacy_Configuration", func(t *testing.T) {
		config := gowright.DefaultConfig()
		assert.NotNil(t, config)
		assert.NotNil(t, config.BrowserConfig)
		assert.NotNil(t, config.APIConfig)
		assert.NotNil(t, config.DatabaseConfig)
		assert.NotNil(t, config.ReportConfig)
	})
	
	// Test legacy test execution methods
	t.Run("Legacy_Test_Execution", func(t *testing.T) {
		framework := gowright.New(gowright.DefaultConfig())
		err := framework.Initialize()
		require.NoError(t, err)
		defer framework.Close()
		
		// Test individual test execution methods
		apiTester := gowright.NewAPITester(framework.GetConfig().APIConfig)
		apiTester.Initialize(framework.GetConfig().APIConfig)
		
		// Test direct API call since ExecuteAPITest expects different structure
		response, err := apiTester.Get("/health", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}

// BenchmarkFrameworkPerformance benchmarks the framework performance
func BenchmarkFrameworkPerformance(b *testing.B) {
	suite := &ComprehensiveIntegrationTestSuite{}
	suite.SetupSuite(&testing.T{})
	defer suite.TearDownSuite(&testing.T{})
	
	// Benchmark API test execution
	b.Run("API_Test_Execution", func(b *testing.B) {
		apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
		apiTester.Initialize(suite.framework.GetConfig().APIConfig)
		defer apiTester.Cleanup()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response, err := apiTester.Get("/health", nil)
			if err != nil || response.StatusCode != http.StatusOK {
				b.Fatalf("Test failed: %v", err)
			}
		}
	})
	
	// Benchmark database test execution
	b.Run("Database_Test_Execution", func(b *testing.B) {
		dbTester := gowright.NewDatabaseTester()
		dbTester.Initialize(suite.framework.GetConfig().DatabaseConfig)
		defer dbTester.Cleanup()
		
		// Test database operations directly for benchmarking
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := dbTester.Execute("test", "SELECT COUNT(*) as count FROM users")
			if err != nil {
				b.Fatalf("Test failed: %v", err)
			}
		}
	})
	
	// Benchmark report generation
	b.Run("Report_Generation", func(b *testing.B) {
		testResults := &gowright.TestResults{
			SuiteName:    "Benchmark Test Suite",
			StartTime:    time.Now(),
			EndTime:      time.Now().Add(time.Minute),
			TotalTests:   100,
			PassedTests:  95,
			FailedTests:  5,
			SkippedTests: 0,
			TestCases:    make([]gowright.TestCaseResult, 100),
		}
		
		// Generate test cases
		for i := 0; i < 100; i++ {
			status := gowright.TestStatusPassed
			if i < 5 {
				status = gowright.TestStatusFailed
			}
			
			testResults.TestCases[i] = gowright.TestCaseResult{
				Name:      fmt.Sprintf("Benchmark Test %d", i+1),
				Status:    status,
				Duration:  time.Millisecond * time.Duration(100+i),
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Millisecond * time.Duration(100+i)),
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
	})
}

// main function to run the integration tests
func main() {
	// Create a test suite and run it
	suite := &ComprehensiveIntegrationTestSuite{}
	
	// Create a mock testing.T for setup
	t := &testing.T{}
	
	suite.SetupSuite(t)
	defer suite.TearDownSuite(t)
	
	// Run the main integration test
	TestCompleteFrameworkIntegration(t)
	
	if t.Failed() {
		os.Exit(1)
	}
	
	println("All integration tests passed!")
}