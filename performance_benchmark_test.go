package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github/gowright/framework/pkg/gowright"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformanceBenchmarkSuite provides performance testing for large test suites
type PerformanceBenchmarkSuite struct {
	framework  *gowright.Gowright
	testServer *httptest.Server
	testDB     *sql.DB
	tempDir    string
}

// SetupPerformanceSuite initializes the performance test environment
func (suite *PerformanceBenchmarkSuite) SetupPerformanceSuite(b *testing.B) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gowright_perf_test_")
	require.NoError(b, err)
	suite.tempDir = tempDir
	
	// Setup test server with multiple endpoints
	suite.setupPerformanceTestServer()
	
	// Setup test database with larger dataset
	suite.setupPerformanceTestDatabase(b)
	
	// Initialize framework with performance-optimized configuration
	suite.setupPerformanceFramework(b)
}

// TearDownPerformanceSuite cleans up the performance test environment
func (suite *PerformanceBenchmarkSuite) TearDownPerformanceSuite(b *testing.B) {
	if suite.framework != nil {
		suite.framework.Cleanup()
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

// setupPerformanceTestServer creates a test server optimized for performance testing
func (suite *PerformanceBenchmarkSuite) setupPerformanceTestServer() {
	mux := http.NewServeMux()
	
	// Fast response endpoint
	mux.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	
	// Slow response endpoint (simulates network latency)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "slow"})
	})
	
	// Large response endpoint
	mux.HandleFunc("/large", func(w http.ResponseWriter, r *http.Request) {
		data := make([]map[string]interface{}, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = map[string]interface{}{
				"id":   i,
				"name": fmt.Sprintf("Item %d", i),
				"data": fmt.Sprintf("Large data payload for item %d with additional content", i),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})
	
	// Variable response time endpoint
	mux.HandleFunc("/variable", func(w http.ResponseWriter, r *http.Request) {
		// Random delay between 10-50ms
		delay := time.Duration(10+r.URL.Query().Get("delay")[0]%40) * time.Millisecond
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"delay": delay.String(),
			"timestamp": time.Now().Unix(),
		})
	})
	
	suite.testServer = httptest.NewServer(mux)
}

// setupPerformanceTestDatabase creates a database with larger dataset for performance testing
func (suite *PerformanceBenchmarkSuite) setupPerformanceTestDatabase(b *testing.B) {
	dbPath := filepath.Join(suite.tempDir, "perf_test.db")
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(b, err)
	
	// Create performance test tables
	_, err = db.Exec(`
		CREATE TABLE performance_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER NOT NULL,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(b, err)
	
	// Insert large dataset for performance testing
	tx, err := db.Begin()
	require.NoError(b, err)
	
	stmt, err := tx.Prepare("INSERT INTO performance_test (name, value, data) VALUES (?, ?, ?)")
	require.NoError(b, err)
	
	for i := 0; i < 10000; i++ {
		_, err = stmt.Exec(
			fmt.Sprintf("Item %d", i),
			i,
			fmt.Sprintf("Performance test data for item %d with additional content", i),
		)
		require.NoError(b, err)
	}
	
	stmt.Close()
	err = tx.Commit()
	require.NoError(b, err)
	
	suite.testDB = db
}

// setupPerformanceFramework initializes the framework with performance-optimized settings
func (suite *PerformanceBenchmarkSuite) setupPerformanceFramework(b *testing.B) {
	config := &gowright.Config{
		BrowserConfig: &gowright.BrowserConfig{
			Headless:    true,
			Timeout:     5 * time.Second, // Shorter timeout for performance
			WindowSize:  &gowright.WindowSize{Width: 1280, Height: 720},
		},
		APIConfig: &gowright.APIConfig{
			BaseURL: suite.testServer.URL,
			Timeout: 5 * time.Second,
			Headers: map[string]string{
				"User-Agent": "Gowright-Performance-Test",
			},
		},
		DatabaseConfig: &gowright.DatabaseConfig{
			Connections: map[string]*gowright.DBConnection{
				"perf": {
					Driver:       "sqlite3",
					DSN:          filepath.Join(suite.tempDir, "perf_test.db"),
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
			},
		},
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      false, // Disable HTML for performance
				OutputDir: suite.tempDir,
			},
		},
	}
	
	framework := gowright.New(config)
	err := framework.Initialize()
	require.NoError(b, err)
	
	suite.framework = framework
}

// BenchmarkLargeTestSuiteExecution benchmarks execution of large test suites
func BenchmarkLargeTestSuiteExecution(b *testing.B) {
	suite := &PerformanceBenchmarkSuite{}
	suite.SetupPerformanceSuite(b)
	defer suite.TearDownPerformanceSuite(b)
	
	// Benchmark different test suite sizes
	testSizes := []int{10, 50, 100, 500}
	
	for _, size := range testSizes {
		b.Run(fmt.Sprintf("TestSuite_%d_Tests", size), func(b *testing.B) {
			suite.benchmarkTestSuiteSize(b, size)
		})
	}
}

// benchmarkTestSuiteSize benchmarks a specific test suite size
func (suite *PerformanceBenchmarkSuite) benchmarkTestSuiteSize(b *testing.B, testCount int) {
	// Create API tester for the tests
	apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
	apiTester.Initialize(suite.framework.GetConfig().APIConfig)
	
	// Create API tests
	tests := make([]gowright.Test, testCount)
	for i := 0; i < testCount; i++ {
		endpoint := "/fast"
		if i%10 == 0 {
			endpoint = "/slow" // 10% slow tests
		}
		
		tests[i] = gowright.NewAPITestBuilder(
			fmt.Sprintf("Performance Test %d", i),
			"GET",
			endpoint,
		).WithTester(apiTester).ExpectStatus(http.StatusOK).Build()
	}
	
	testSuite := &gowright.TestSuite{
		Name:  fmt.Sprintf("Performance Test Suite (%d tests)", testCount),
		Tests: tests,
	}
	
	suite.framework.SetTestSuite(testSuite)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := suite.framework.ExecuteTestSuite()
		if err != nil {
			b.Fatalf("Test suite execution failed: %v", err)
		}
		if results.TotalTests != testCount {
			b.Fatalf("Expected %d tests, got %d", testCount, results.TotalTests)
		}
	}
}

// BenchmarkConcurrentTestExecution benchmarks concurrent test execution
func BenchmarkConcurrentTestExecution(b *testing.B) {
	suite := &PerformanceBenchmarkSuite{}
	suite.SetupPerformanceSuite(b)
	defer suite.TearDownPerformanceSuite(b)
	
	concurrencyLevels := []int{1, 2, 4, 8, 16}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			suite.benchmarkConcurrentExecution(b, concurrency)
		})
	}
}

// benchmarkConcurrentExecution benchmarks concurrent test execution
func (suite *PerformanceBenchmarkSuite) benchmarkConcurrentExecution(b *testing.B, concurrency int) {
	apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
	err := apiTester.Initialize(suite.framework.GetConfig().APIConfig)
	require.NoError(b, err)
	defer apiTester.Cleanup()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		errors := make([]error, concurrency)
		
		for j := 0; j < concurrency; j++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				response, err := apiTester.Get("/fast", nil)
				if err != nil || response.StatusCode != http.StatusOK {
					errors[index] = err
				}
			}(j)
		}
		
		wg.Wait()
		
		// Verify all tests passed
		for j, err := range errors {
			if err != nil {
				b.Fatalf("Concurrent test %d failed: %v", j, err)
			}
		}
	}
}

// BenchmarkDatabasePerformance benchmarks database operations with large datasets
func BenchmarkDatabasePerformance(b *testing.B) {
	suite := &PerformanceBenchmarkSuite{}
	suite.SetupPerformanceSuite(b)
	defer suite.TearDownPerformanceSuite(b)
	
	dbTester := gowright.NewDatabaseTester()
	err := dbTester.Initialize(suite.framework.GetConfig().DatabaseConfig)
	require.NoError(b, err)
	defer dbTester.Cleanup()
	
	// Benchmark different query types
	b.Run("Simple_Select", func(b *testing.B) {
		query := "SELECT COUNT(*) as count FROM performance_test"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := dbTester.Execute("perf", query)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}
		}
	})
	
	b.Run("Complex_Select", func(b *testing.B) {
		query := "SELECT * FROM performance_test WHERE value > ? ORDER BY value LIMIT 100"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := dbTester.Execute("perf", query, i%1000)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}
		}
	})
	
	b.Run("Aggregation_Query", func(b *testing.B) {
		query := "SELECT AVG(value), MIN(value), MAX(value), COUNT(*) FROM performance_test WHERE value BETWEEN ? AND ?"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := i % 1000
			end := start + 100
			_, err := dbTester.Execute("perf", query, start, end)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}
		}
	})
}

// BenchmarkReportGeneration benchmarks report generation with large result sets
func BenchmarkReportGeneration(b *testing.B) {
	suite := &PerformanceBenchmarkSuite{}
	suite.SetupPerformanceSuite(b)
	defer suite.TearDownPerformanceSuite(b)
	
	// Create large test results
	resultSizes := []int{100, 500, 1000, 5000}
	
	for _, size := range resultSizes {
		b.Run(fmt.Sprintf("Results_%d", size), func(b *testing.B) {
			suite.benchmarkReportGeneration(b, size)
		})
	}
}

// benchmarkReportGeneration benchmarks report generation for different result sizes
func (suite *PerformanceBenchmarkSuite) benchmarkReportGeneration(b *testing.B, resultCount int) {
	testResults := &gowright.TestResults{
		SuiteName:    fmt.Sprintf("Performance Test Suite (%d results)", resultCount),
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		TotalTests:   resultCount,
		PassedTests:  int(float64(resultCount) * 0.9), // 90% pass rate
		FailedTests:  int(float64(resultCount) * 0.1), // 10% fail rate
		SkippedTests: 0,
		TestCases:    make([]gowright.TestCaseResult, resultCount),
	}
	
	// Generate test cases
	for i := 0; i < resultCount; i++ {
		status := gowright.TestStatusPassed
		var err error
		if i < testResults.FailedTests {
			status = gowright.TestStatusFailed
			err = fmt.Errorf("test assertion failed for test %d", i)
		}
		
		testResults.TestCases[i] = gowright.TestCaseResult{
			Name:      fmt.Sprintf("Performance Test %d", i),
			Status:    status,
			Duration:  time.Duration(50+i%200) * time.Millisecond,
			Error:     err,
			StartTime: time.Now().Add(-time.Hour).Add(time.Duration(i) * time.Second),
			EndTime:   time.Now().Add(-time.Hour).Add(time.Duration(i) * time.Second).Add(time.Duration(50+i%200) * time.Millisecond),
			Screenshots: []string{fmt.Sprintf("screenshot_%d.png", i)},
			Logs:        []string{fmt.Sprintf("Log entry for test %d", i)},
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

// BenchmarkMemoryUsage benchmarks memory usage during test execution
func BenchmarkMemoryUsage(b *testing.B) {
	suite := &PerformanceBenchmarkSuite{}
	suite.SetupPerformanceSuite(b)
	defer suite.TearDownPerformanceSuite(b)
	
	// Force garbage collection before starting
	runtime.GC()
	
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	// Create API tester for the tests
	apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
	apiTester.Initialize(suite.framework.GetConfig().APIConfig)
	
	// Create and execute a large test suite
	testCount := 1000
	tests := make([]gowright.Test, testCount)
	for i := 0; i < testCount; i++ {
		tests[i] = gowright.NewAPITestBuilder(
			fmt.Sprintf("Memory Test %d", i),
			"GET",
			"/fast",
		).WithTester(apiTester).ExpectStatus(http.StatusOK).Build()
	}
	
	testSuite := &gowright.TestSuite{
		Name:  "Memory Usage Test Suite",
		Tests: tests,
	}
	
	suite.framework.SetTestSuite(testSuite)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := suite.framework.ExecuteTestSuite()
		if err != nil {
			b.Fatalf("Test suite execution failed: %v", err)
		}
		if results.TotalTests != testCount {
			b.Fatalf("Expected %d tests, got %d", testCount, results.TotalTests)
		}
		
		// Force garbage collection between iterations
		runtime.GC()
	}
	
	runtime.ReadMemStats(&m2)
	
	// Report memory usage
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB/alloc")
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/1024/1024, "MB/total")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs), "mallocs")
}

// BenchmarkResourceCleanup benchmarks resource cleanup performance
func BenchmarkResourceCleanup(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "gowright_cleanup_test_")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create framework instance
		config := &gowright.Config{
			BrowserConfig: &gowright.BrowserConfig{
				Headless: true,
				Timeout:  5 * time.Second,
			},
			APIConfig: &gowright.APIConfig{
				BaseURL: "http://example.com",
				Timeout: 5 * time.Second,
			},
			DatabaseConfig: &gowright.DatabaseConfig{
				Connections: map[string]*gowright.DBConnection{
					"test": {
						Driver: "sqlite3",
						DSN:    ":memory:",
					},
				},
			},
			ReportConfig: &gowright.ReportConfig{
				LocalReports: gowright.LocalReportConfig{
					JSON:      true,
					OutputDir: tempDir,
				},
			},
		}
		
		framework := gowright.New(config)
		err := framework.Initialize()
		if err != nil {
			b.Fatalf("Framework initialization failed: %v", err)
		}
		
		// Measure cleanup time
		cleanupStart := time.Now()
		err = framework.Cleanup()
		cleanupDuration := time.Since(cleanupStart)
		
		if err != nil {
			b.Fatalf("Framework cleanup failed: %v", err)
		}
		
		// Report cleanup time
		b.ReportMetric(float64(cleanupDuration.Nanoseconds())/1000000, "ms/cleanup")
	}
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	suite := &PerformanceBenchmarkSuite{}
	suite.SetupPerformanceSuite(&testing.B{})
	defer suite.TearDownPerformanceSuite(&testing.B{})
	
	// Define performance thresholds
	thresholds := map[string]time.Duration{
		"api_test_execution":      100 * time.Millisecond,
		"database_test_execution": 50 * time.Millisecond,
		"report_generation":       500 * time.Millisecond,
		"framework_initialization": 1 * time.Second,
	}
	
	// Test API execution performance
	t.Run("API_Test_Performance", func(t *testing.T) {
		apiTester := gowright.NewAPITester(suite.framework.GetConfig().APIConfig)
		err := apiTester.Initialize(suite.framework.GetConfig().APIConfig)
		require.NoError(t, err)
		defer apiTester.Cleanup()
		
		start := time.Now()
		response, err := apiTester.Get("/fast", nil)
		duration := time.Since(start)
		
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Less(t, duration, thresholds["api_test_execution"],
			"API test execution took %v, expected less than %v", duration, thresholds["api_test_execution"])
	})
	
	// Test database execution performance
	t.Run("Database_Test_Performance", func(t *testing.T) {
		dbTester := gowright.NewDatabaseTester()
		err := dbTester.Initialize(suite.framework.GetConfig().DatabaseConfig)
		require.NoError(t, err)
		defer dbTester.Cleanup()
		
		dbTest := &gowright.DatabaseTest{
			Name:       "Performance Regression Test",
			Connection: "perf",
			Query:      "SELECT COUNT(*) as count FROM performance_test",
			Expected:   &gowright.DatabaseExpectation{RowCount: 1},
		}
		
		start := time.Now()
		result := dbTester.ExecuteTest(dbTest)
		duration := time.Since(start)
		
		assert.Equal(t, gowright.TestStatusPassed, result.Status)
		assert.Less(t, duration, thresholds["database_test_execution"],
			"Database test execution took %v, expected less than %v", duration, thresholds["database_test_execution"])
	})
	
	// Test framework initialization performance
	t.Run("Framework_Initialization_Performance", func(t *testing.T) {
		config := gowright.DefaultConfig()
		
		start := time.Now()
		framework := gowright.New(config)
		err := framework.Initialize()
		duration := time.Since(start)
		
		require.NoError(t, err)
		defer framework.Cleanup()
		
		assert.Less(t, duration, thresholds["framework_initialization"],
			"Framework initialization took %v, expected less than %v", duration, thresholds["framework_initialization"])
	})
}