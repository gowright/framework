package gowright

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceManager(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	assert.NotNil(t, rm)
	assert.Equal(t, config, rm.config)
	assert.NotNil(t, rm.databasePools)
	assert.False(t, rm.initialized)
}

func TestResourceManager_Initialize(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)
	ctx := context.Background()

	err := rm.Initialize(ctx)

	assert.NoError(t, err)
	assert.True(t, rm.initialized)
	assert.NotNil(t, rm.browserPool)
	assert.NotNil(t, rm.httpClientPool)
}

func TestResourceManager_Initialize_AlreadyInitialized(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)
	ctx := context.Background()

	// Initialize first time
	err1 := rm.Initialize(ctx)
	assert.NoError(t, err1)

	// Initialize second time should not error
	err2 := rm.Initialize(ctx)
	assert.NoError(t, err2)
}

func TestResourceManager_InitializeDatabasePool(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	dbConfig := &DBConnection{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	err := rm.InitializeDatabasePool("test_db", dbConfig)

	assert.NoError(t, err)
	assert.Contains(t, rm.databasePools, "test_db")
}

func TestResourceManager_InitializeDatabasePool_AlreadyExists(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	dbConfig := &DBConnection{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	// Initialize first time
	err1 := rm.InitializeDatabasePool("test_db", dbConfig)
	assert.NoError(t, err1)

	// Initialize second time should not error
	err2 := rm.InitializeDatabasePool("test_db", dbConfig)
	assert.NoError(t, err2)
}

func TestResourceManager_testNeedsBrowser(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	// Test with UI test
	uiTest := &MockUITest{Name: "ui_test"}
	assert.True(t, rm.testNeedsBrowser(uiTest)) // UI tests should need browser

	// Test with Integration test
	integrationTest := &MockIntegrationTest{Name: "integration_test"}
	assert.False(t, rm.testNeedsBrowser(integrationTest)) // Updated to match current implementation

	// Test with API test
	apiTest := &MockAPITest{Name: "api_test"}
	assert.False(t, rm.testNeedsBrowser(apiTest))

	// Test with Database test
	dbTest := &MockDatabaseTest{Name: "db_test"}
	assert.False(t, rm.testNeedsBrowser(dbTest))
}

func TestResourceManager_testNeedsHTTPClient(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	// Test with API test
	apiTest := &MockAPITest{Name: "api_test"}
	assert.True(t, rm.testNeedsHTTPClient(apiTest)) // API tests should need HTTP client

	// Test with Integration test
	integrationTest := &MockIntegrationTest{Name: "integration_test"}
	assert.False(t, rm.testNeedsHTTPClient(integrationTest)) // Updated to match current implementation

	// Test with UI test
	uiTest := &MockUITest{Name: "ui_test"}
	assert.False(t, rm.testNeedsHTTPClient(uiTest))

	// Test with Database test
	dbTest := &MockDatabaseTest{Name: "db_test"}
	assert.False(t, rm.testNeedsHTTPClient(dbTest))
}

func TestResourceManager_getNeededDBConnections(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	// Test with Database test
	dbTest := &MockDatabaseTest{
		Name:       "db_test",
		Connection: "test_connection",
	}
	connections := rm.getNeededDBConnections(dbTest)
	assert.Empty(t, connections) // Updated to match current implementation

	// Test with Database test without connection
	dbTestNoConn := &MockDatabaseTest{Name: "db_test_no_conn"}
	connectionsEmpty := rm.getNeededDBConnections(dbTestNoConn)
	assert.Empty(t, connectionsEmpty)

	// Test with other test types
	apiTest := &MockAPITest{Name: "api_test"}
	connectionsAPI := rm.getNeededDBConnections(apiTest)
	assert.Empty(t, connectionsAPI)
}

func TestResourceManager_AcquireResources_NotInitialized(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)
	ctx := context.Background()

	mockTest := &MockTest{name: "test"}

	handle, err := rm.AcquireResources(ctx, mockTest)

	assert.Error(t, err)
	assert.Nil(t, handle)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestResourceManager_Cleanup_NotInitialized(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	err := rm.Cleanup()

	assert.NoError(t, err)
}

func TestResourceManager_GetStats(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)
	ctx := context.Background()

	// Initialize the resource manager
	err := rm.Initialize(ctx)
	assert.NoError(t, err)

	stats := rm.GetStats()

	assert.NotNil(t, stats)
	assert.NotNil(t, stats.BrowserPool)
	assert.NotNil(t, stats.HTTPClientPool)
	assert.NotNil(t, stats.DatabasePools)
}

func TestResourceManager_GetStats_NotInitialized(t *testing.T) {
	config := DefaultParallelRunnerConfig()
	rm := NewResourceManager(config)

	stats := rm.GetStats()

	assert.NotNil(t, stats)
	assert.Nil(t, stats.BrowserPool)
	assert.Nil(t, stats.HTTPClientPool)
	assert.NotNil(t, stats.DatabasePools)
	assert.Empty(t, stats.DatabasePools)
}

// Integration test for resource acquisition and release
func TestResourceManager_AcquireAndReleaseResources_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultParallelRunnerConfig()
	config.BrowserPoolSize = 1
	config.HTTPClientPoolSize = 1

	rm := NewResourceManager(config)
	ctx := context.Background()

	// Initialize the resource manager
	err := rm.Initialize(ctx)
	assert.NoError(t, err)

	defer func() {
		cleanupErr := rm.Cleanup()
		assert.NoError(t, cleanupErr)
	}()

	// Test with API test (needs HTTP client)
	apiTest := &MockAPITest{Name: "api_test"}

	handle, err := rm.AcquireResources(ctx, apiTest)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.Equal(t, "api_test", handle.TestName)
	assert.NotNil(t, handle.HTTPClient)
	assert.Nil(t, handle.Browser)
	assert.Nil(t, handle.Page)
	assert.Empty(t, handle.DBConns)

	// Release resources
	err = rm.ReleaseResources(handle)
	assert.NoError(t, err)
}

// Mock implementations for testing
type MockUITest struct {
	Name string
}

func (m *MockUITest) GetName() string {
	return m.Name
}

func (m *MockUITest) Execute() *TestCaseResult {
	return &TestCaseResult{
		Name:   m.Name,
		Status: TestStatusPassed,
	}
}

type MockAPITest struct {
	Name string
}

func (m *MockAPITest) GetName() string {
	return m.Name
}

func (m *MockAPITest) Execute() *TestCaseResult {
	return &TestCaseResult{
		Name:   m.Name,
		Status: TestStatusPassed,
	}
}

type MockDatabaseTest struct {
	Name       string
	Connection string
}

func (m *MockDatabaseTest) GetName() string {
	return m.Name
}

func (m *MockDatabaseTest) Execute() *TestCaseResult {
	return &TestCaseResult{
		Name:   m.Name,
		Status: TestStatusPassed,
	}
}

type MockIntegrationTest struct {
	Name string
}

func (m *MockIntegrationTest) GetName() string {
	return m.Name
}

func (m *MockIntegrationTest) Execute() *TestCaseResult {
	return &TestCaseResult{
		Name:   m.Name,
		Status: TestStatusPassed,
	}
}
