// Package gowright provides a comprehensive testing framework for Go
// that supports UI, API, database, and integration testing.
package gowright

import (
	"fmt"
	"sync"
)

// Gowright is the main framework struct that orchestrates all testing activities
type Gowright struct {
	config    *Config
	reporter  *ReportManager
	testSuite *TestSuite
	
	// Dependency injection support
	uiTester          UITester
	apiTester         APITester
	databaseTester    DatabaseTester
	integrationTester IntegrationTester
	
	// Internal state management
	initialized bool
	mutex       sync.RWMutex
}

// GowrightOptions provides options for creating a Gowright instance
type GowrightOptions struct {
	Config            *Config
	UITester          UITester
	APITester         APITester
	DatabaseTester    DatabaseTester
	IntegrationTester IntegrationTester
	ReportManager     *ReportManager
}

// New creates a new Gowright instance with the provided configuration
func New(config *Config) *Gowright {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Gowright{
		config:      config,
		reporter:    NewReportManager(config.ReportConfig),
		initialized: false,
	}
}

// NewWithOptions creates a new Gowright instance with dependency injection support
func NewWithOptions(options *GowrightOptions) *Gowright {
	if options == nil {
		return NewWithDefaults()
	}
	
	config := options.Config
	if config == nil {
		config = DefaultConfig()
	}
	
	reporter := options.ReportManager
	if reporter == nil {
		reporter = NewReportManager(config.ReportConfig)
	}
	
	return &Gowright{
		config:            config,
		reporter:          reporter,
		uiTester:          options.UITester,
		apiTester:         options.APITester,
		databaseTester:    options.DatabaseTester,
		integrationTester: options.IntegrationTester,
		initialized:       false,
	}
}

// NewWithDefaults creates a new Gowright instance with default configuration
func NewWithDefaults() *Gowright {
	return New(DefaultConfig())
}

// NewGowright creates a new Gowright instance with the provided configuration
// This is an alias for New() to maintain compatibility
func NewGowright(config *Config) (*Gowright, error) {
	gowright := New(config)
	if err := gowright.Initialize(); err != nil {
		return nil, err
	}
	return gowright, nil
}

// Initialize initializes the framework and all its components
func (g *Gowright) Initialize() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if g.initialized {
		return nil
	}
	
	// Initialize testers if they are provided
	if g.uiTester != nil {
		if err := g.uiTester.Initialize(g.config.BrowserConfig); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize UI tester", err)
		}
	}
	
	if g.apiTester != nil {
		if err := g.apiTester.Initialize(g.config.APIConfig); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize API tester", err)
		}
	}
	
	if g.databaseTester != nil {
		if err := g.databaseTester.Initialize(g.config.DatabaseConfig); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize database tester", err)
		}
	}
	
	if g.integrationTester != nil {
		if err := g.integrationTester.Initialize(g.config); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize integration tester", err)
		}
	}
	
	g.initialized = true
	return nil
}

// Cleanup performs cleanup operations for all components
func (g *Gowright) Cleanup() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	var errors []error
	
	// Cleanup testers
	if g.uiTester != nil {
		if err := g.uiTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("UI tester cleanup failed: %w", err))
		}
	}
	
	if g.apiTester != nil {
		if err := g.apiTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("API tester cleanup failed: %w", err))
		}
	}
	
	if g.databaseTester != nil {
		if err := g.databaseTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("database tester cleanup failed: %w", err))
		}
	}
	
	if g.integrationTester != nil {
		if err := g.integrationTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("integration tester cleanup failed: %w", err))
		}
	}
	
	g.initialized = false
	
	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors occurred: %v", errors)
	}
	
	return nil
}

// IsInitialized returns whether the framework has been initialized
func (g *Gowright) IsInitialized() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.initialized
}

// SetTestSuite sets the test suite for this Gowright instance
func (g *Gowright) SetTestSuite(suite *TestSuite) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.testSuite = suite
}

// CreateTestSuiteManager creates a new test suite manager for the current test suite
func (g *Gowright) CreateTestSuiteManager() *TestSuiteManager {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if g.testSuite == nil {
		g.testSuite = &TestSuite{
			Name:  "Default Test Suite",
			Tests: make([]Test, 0),
		}
	}
	
	return NewTestSuiteManager(g.testSuite, g.config)
}

// ExecuteTestSuite executes the current test suite and returns results
func (g *Gowright) ExecuteTestSuite() (*TestResults, error) {
	tsm := g.CreateTestSuiteManager()
	return tsm.ExecuteTestSuite()
}

// GetTestSuite returns the current test suite
func (g *Gowright) GetTestSuite() *TestSuite {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.testSuite
}

// GetConfig returns the current configuration
func (g *Gowright) GetConfig() *Config {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.config
}

// GetReporter returns the report manager
func (g *Gowright) GetReporter() *ReportManager {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.reporter
}

// GetUITester returns the UI tester instance
func (g *Gowright) GetUITester() UITester {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.uiTester
}

// GetAPITester returns the API tester instance
func (g *Gowright) GetAPITester() APITester {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.apiTester
}

// GetDatabaseTester returns the database tester instance
func (g *Gowright) GetDatabaseTester() DatabaseTester {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.databaseTester
}

// GetIntegrationTester returns the integration tester instance
func (g *Gowright) GetIntegrationTester() IntegrationTester {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.integrationTester
}

// SetUITester sets the UI tester instance
func (g *Gowright) SetUITester(tester UITester) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.uiTester = tester
}

// SetAPITester sets the API tester instance
func (g *Gowright) SetAPITester(tester APITester) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.apiTester = tester
}

// SetDatabaseTester sets the database tester instance
func (g *Gowright) SetDatabaseTester(tester DatabaseTester) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.databaseTester = tester
}

// SetIntegrationTester sets the integration tester instance
func (g *Gowright) SetIntegrationTester(tester IntegrationTester) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.integrationTester = tester
}

// ExecuteUITest executes a single UI test and returns the result
func (g *Gowright) ExecuteUITest(test *UITest) *TestCaseResult {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if g.uiTester == nil {
		return &TestCaseResult{
			Name:   test.Name,
			Status: TestStatusError,
			Error:  NewGowrightError(ConfigurationError, "UI tester not configured", nil),
		}
	}
	
	return g.uiTester.ExecuteTest(test)
}

// ExecuteAPITest executes a single API test and returns the result
func (g *Gowright) ExecuteAPITest(test *APITest) *TestCaseResult {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if g.apiTester == nil {
		return &TestCaseResult{
			Name:   test.Name,
			Status: TestStatusError,
			Error:  NewGowrightError(ConfigurationError, "API tester not configured", nil),
		}
	}
	
	return g.apiTester.ExecuteTest(test)
}

// ExecuteDatabaseTest executes a single database test and returns the result
func (g *Gowright) ExecuteDatabaseTest(test *DatabaseTest) *TestCaseResult {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if g.databaseTester == nil {
		return &TestCaseResult{
			Name:   test.Name,
			Status: TestStatusError,
			Error:  NewGowrightError(ConfigurationError, "Database tester not configured", nil),
		}
	}
	
	return g.databaseTester.ExecuteTest(test)
}

// ExecuteIntegrationTest executes a single integration test and returns the result
func (g *Gowright) ExecuteIntegrationTest(test *IntegrationTest) *TestCaseResult {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if g.integrationTester == nil {
		return &TestCaseResult{
			Name:   test.Name,
			Status: TestStatusError,
			Error:  NewGowrightError(ConfigurationError, "Integration tester not configured", nil),
		}
	}
	
	return g.integrationTester.ExecuteTest(test)
}

// Close performs cleanup and closes the Gowright instance
func (g *Gowright) Close() error {
	return g.Cleanup()
}