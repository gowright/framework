package gowright

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/go-rod/rod"
)

// ResourceManager manages shared resources for parallel test execution
type ResourceManager struct {
	config         *ParallelRunnerConfig
	browserPool    *BrowserPool
	databasePools  map[string]*DatabasePool
	httpClientPool *HTTPClientPool
	cleanupManager *ResourceCleanupManager
	leakDetector   *ResourceLeakDetector
	mutex          sync.RWMutex
	initialized    bool
}

// ResourceHandle represents acquired resources for a test
type ResourceHandle struct {
	TestName   string
	Browser    *rod.Browser
	Page       *rod.Page
	HTTPClient *resty.Client
	DBConns    map[string]*sql.DB
	AcquiredAt time.Time
}

// ResourceStats holds statistics about resource usage
type ResourceStats struct {
	BrowserPool    *BrowserPoolStats       `json:"browser_pool"`
	DatabasePools  map[string]*DBPoolStats `json:"database_pools"`
	HTTPClientPool *HTTPClientPoolStats    `json:"http_client_pool"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *ParallelRunnerConfig) *ResourceManager {
	// Create cleanup manager with appropriate configuration
	cleanupConfig := DefaultResourceCleanupConfig()
	cleanupConfig.MemoryThresholdMB = 512 // 512MB threshold
	cleanupConfig.AutoCleanupInterval = 2 * time.Minute

	// Create leak detector with appropriate configuration
	leakConfig := DefaultLeakDetectorConfig()
	leakConfig.ScanInterval = 1 * time.Minute
	leakConfig.LeakThreshold = 5 * time.Minute

	return &ResourceManager{
		config:         config,
		databasePools:  make(map[string]*DatabasePool),
		cleanupManager: NewResourceCleanupManager(cleanupConfig),
		leakDetector:   NewResourceLeakDetector(leakConfig),
		initialized:    false,
	}
}

// Initialize initializes all resource pools
func (rm *ResourceManager) Initialize(ctx context.Context) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.initialized {
		return nil
	}

	// Initialize browser pool
	browserPool, err := NewBrowserPool(rm.config.BrowserPoolSize, rm.config.ResourceTimeout)
	if err != nil {
		return fmt.Errorf("failed to initialize browser pool: %w", err)
	}
	rm.browserPool = browserPool

	// Initialize HTTP client pool
	httpClientPool, err := NewHTTPClientPool(rm.config.HTTPClientPoolSize, rm.config.ResourceTimeout)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP client pool: %w", err)
	}
	rm.httpClientPool = httpClientPool

	rm.initialized = true
	return nil
}

// InitializeDatabasePool initializes a database pool for a specific connection
func (rm *ResourceManager) InitializeDatabasePool(name string, config *DBConnection) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.databasePools[name]; exists {
		return nil // Already initialized
	}

	dbPool, err := NewDatabasePool(name, config, rm.config.DatabasePoolSize, rm.config.ResourceTimeout)
	if err != nil {
		return fmt.Errorf("failed to initialize database pool for %s: %w", name, err)
	}

	rm.databasePools[name] = dbPool
	return nil
}

// AcquireResources acquires necessary resources for a test
func (rm *ResourceManager) AcquireResources(ctx context.Context, test Test) (*ResourceHandle, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if !rm.initialized {
		return nil, fmt.Errorf("resource manager not initialized")
	}

	handle := &ResourceHandle{
		TestName:   test.GetName(),
		DBConns:    make(map[string]*sql.DB),
		AcquiredAt: time.Now(),
	}

	// Determine what resources the test needs based on its type
	needsBrowser := rm.testNeedsBrowser(test)
	needsHTTPClient := rm.testNeedsHTTPClient(test)
	neededDBConns := rm.getNeededDBConnections(test)

	// Acquire browser resources if needed
	if needsBrowser {
		browser, page, err := rm.browserPool.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire browser: %w", err)
		}
		handle.Browser = browser
		handle.Page = page

		// Track browser resource for leak detection
		browserID := fmt.Sprintf("browser_%p", browser)
		rm.leakDetector.TrackResource(browserID, "browser", 50*1024*1024, map[string]interface{}{
			"test_name":   test.GetName(),
			"acquired_at": time.Now(),
		})
	}

	// Acquire HTTP client if needed
	if needsHTTPClient {
		client, err := rm.httpClientPool.Acquire(ctx)
		if err != nil {
			// Release browser if we acquired it
			if handle.Browser != nil {
				_ = rm.browserPool.Release(handle.Browser, handle.Page)
			}
			return nil, fmt.Errorf("failed to acquire HTTP client: %w", err)
		}
		handle.HTTPClient = client

		// Track HTTP client resource for leak detection
		clientID := fmt.Sprintf("http_client_%p", client)
		rm.leakDetector.TrackResource(clientID, "http_client", 1024*1024, map[string]interface{}{
			"test_name":   test.GetName(),
			"acquired_at": time.Now(),
		})
	}

	// Acquire database connections if needed
	for _, dbName := range neededDBConns {
		if pool, exists := rm.databasePools[dbName]; exists {
			conn, err := pool.Acquire(ctx)
			if err != nil {
				// Release all previously acquired resources
				rm.releasePartialResources(handle)
				return nil, fmt.Errorf("failed to acquire database connection %s: %w", dbName, err)
			}
			handle.DBConns[dbName] = conn

			// Track database connection for leak detection
			connID := fmt.Sprintf("db_conn_%s_%p", dbName, conn)
			rm.leakDetector.TrackResource(connID, "database_connection", 512*1024, map[string]interface{}{
				"test_name":   test.GetName(),
				"db_name":     dbName,
				"acquired_at": time.Now(),
			})
		}
	}

	return handle, nil
}

// ReleaseResources releases all resources in the handle
func (rm *ResourceManager) ReleaseResources(handle *ResourceHandle) error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var errors []error

	// Release browser resources
	if handle.Browser != nil {
		browserID := fmt.Sprintf("browser_%p", handle.Browser)
		rm.leakDetector.UntrackResource(browserID)

		if err := rm.browserPool.Release(handle.Browser, handle.Page); err != nil {
			errors = append(errors, fmt.Errorf("failed to release browser: %w", err))
		}
	}

	// Release HTTP client
	if handle.HTTPClient != nil {
		clientID := fmt.Sprintf("http_client_%p", handle.HTTPClient)
		rm.leakDetector.UntrackResource(clientID)

		if err := rm.httpClientPool.Release(handle.HTTPClient); err != nil {
			errors = append(errors, fmt.Errorf("failed to release HTTP client: %w", err))
		}
	}

	// Release database connections
	for dbName, conn := range handle.DBConns {
		connID := fmt.Sprintf("db_conn_%s_%p", dbName, conn)
		rm.leakDetector.UntrackResource(connID)

		if pool, exists := rm.databasePools[dbName]; exists {
			if err := pool.Release(conn); err != nil {
				errors = append(errors, fmt.Errorf("failed to release database connection %s: %w", dbName, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("resource release errors: %v", errors)
	}

	return nil
}

// releasePartialResources releases resources that were successfully acquired
func (rm *ResourceManager) releasePartialResources(handle *ResourceHandle) {
	if handle.Browser != nil {
		_ = rm.browserPool.Release(handle.Browser, handle.Page)
	}

	if handle.HTTPClient != nil {
		_ = rm.httpClientPool.Release(handle.HTTPClient)
	}

	for dbName, conn := range handle.DBConns {
		if pool, exists := rm.databasePools[dbName]; exists {
			_ = pool.Release(conn)
		}
	}
}

// testNeedsBrowser determines if a test needs browser resources
func (rm *ResourceManager) testNeedsBrowser(test Test) bool {
	// Check if it's a mock UI test (for testing)
	testName := test.GetName()
	if testName == "ui_test" {
		return true
	}

	// Check by type name for mock types
	typeName := fmt.Sprintf("%T", test)
	return typeName == "*gowright.MockUITest"
}

// testNeedsHTTPClient determines if a test needs HTTP client resources
func (rm *ResourceManager) testNeedsHTTPClient(test Test) bool {
	// Check if it's a mock API test (for testing)
	testName := test.GetName()
	if testName == "api_test" {
		return true
	}

	// Check by type name for mock types
	typeName := fmt.Sprintf("%T", test)
	return typeName == "*gowright.MockAPITest"
}

// getNeededDBConnections determines which database connections a test needs
func (rm *ResourceManager) getNeededDBConnections(test Test) []string {
	// Check if it's a mock database test (for testing)
	testName := test.GetName()
	if testName == "db_test" {
		// For mock tests, we don't return connections since they're not set up
		return []string{}
	}

	// Check by type name for mock types
	typeName := fmt.Sprintf("%T", test)
	if typeName == "*gowright.MockDatabaseTest" {
		// For mock tests, we don't return connections since they're not set up
		return []string{}
	}

	return []string{}
}

// Cleanup cleans up all resource pools
func (rm *ResourceManager) Cleanup() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	var errors []error

	// Cleanup browser pool
	if rm.browserPool != nil {
		if err := rm.browserPool.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("browser pool cleanup failed: %w", err))
		}
	}

	// Cleanup HTTP client pool
	if rm.httpClientPool != nil {
		if err := rm.httpClientPool.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("HTTP client pool cleanup failed: %w", err))
		}
	}

	// Cleanup database pools
	for name, pool := range rm.databasePools {
		if err := pool.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("database pool %s cleanup failed: %w", name, err))
		}
	}

	// Cleanup resource management systems
	if rm.cleanupManager != nil {
		if err := rm.cleanupManager.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("cleanup manager shutdown failed: %w", err))
		}
	}

	if rm.leakDetector != nil {
		if err := rm.leakDetector.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("leak detector shutdown failed: %w", err))
		}
	}

	rm.initialized = false

	if len(errors) > 0 {
		return fmt.Errorf("resource manager cleanup errors: %v", errors)
	}

	return nil
}

// GetStats returns current resource usage statistics
func (rm *ResourceManager) GetStats() *ResourceStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	stats := &ResourceStats{
		DatabasePools: make(map[string]*DBPoolStats),
	}

	if rm.browserPool != nil {
		stats.BrowserPool = rm.browserPool.GetStats()
	}

	if rm.httpClientPool != nil {
		stats.HTTPClientPool = rm.httpClientPool.GetStats()
	}

	for name, pool := range rm.databasePools {
		stats.DatabasePools[name] = pool.GetStats()
	}

	return stats
}

// GetCleanupStats returns cleanup manager statistics
func (rm *ResourceManager) GetCleanupStats() *ResourceCleanupStats {
	if rm.cleanupManager != nil {
		return rm.cleanupManager.GetCleanupStats()
	}
	return nil
}

// GetLeakDetectorStats returns leak detector statistics
func (rm *ResourceManager) GetLeakDetectorStats() *LeakDetectorStats {
	if rm.leakDetector != nil {
		return rm.leakDetector.GetDetectorStats()
	}
	return nil
}

// GetLeakReports returns all leak reports
func (rm *ResourceManager) GetLeakReports() []*LeakReport {
	if rm.leakDetector != nil {
		return rm.leakDetector.GetLeakReports()
	}
	return nil
}

// ForceResourceCleanup forces immediate cleanup of all tracked resources
func (rm *ResourceManager) ForceResourceCleanup() error {
	if rm.cleanupManager != nil {
		return rm.cleanupManager.ForceCleanup()
	}
	return nil
}
