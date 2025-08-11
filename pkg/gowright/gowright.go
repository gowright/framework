// Package gowright provides a comprehensive testing framework for Go
// that supports UI, API, database, mobile, and integration testing.
//
// This is the main entry point that brings together all the modular packages:
// - core: Main framework orchestrator and interfaces
// - config: Configuration management
// - ui: UI/browser testing
// - api: API/HTTP testing
// - database: Database testing
// - mobile: Mobile/Appium testing
// - integration: Integration testing that orchestrates all other modules
// - reporting: Test result reporting
// - assertions: Common assertion utilities
package gowright

import (
	"github.com/gowright/framework/pkg/api"
	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/gowright/framework/pkg/database"
	"github.com/gowright/framework/pkg/integration"
	"github.com/gowright/framework/pkg/mobile"
	"github.com/gowright/framework/pkg/reporting"
	"github.com/gowright/framework/pkg/ui"
)

// Re-export core types and interfaces for backward compatibility
type (
	// Core framework types
	Gowright         = core.Gowright
	GowrightOptions  = core.GowrightOptions
	TestSuite        = core.TestSuite
	Test             = core.Test
	TestStatus       = core.TestStatus
	TestCaseResult   = core.TestCaseResult
	TestResults      = core.TestResults
	GowrightError    = core.GowrightError
	ErrorType        = core.ErrorType
	AssertionStep    = core.AssertionStep
	TestSuiteManager = core.TestSuiteManager
	TestExecutor     = core.TestExecutor
	ParallelRunner   = core.ParallelRunner
	VersionInfo      = core.VersionInfo
	RetryConfig      = core.RetryConfig
	TestContext      = core.TestContext
	TestAssertion    = core.TestAssertion

	// Test types
	UITest              = core.UITest
	UIAction            = core.UIAction
	UIAssertion         = core.UIAssertion
	APITest             = core.APITest
	APIExpectation      = core.APIExpectation
	APIResponse         = core.APIResponse
	DatabaseTest        = core.DatabaseTest
	DatabaseExpectation = core.DatabaseExpectation
	DatabaseResult      = core.DatabaseResult
	IntegrationTest     = core.IntegrationTest
	IntegrationStep     = core.IntegrationStep
	IntegrationStepType = core.IntegrationStepType

	// Step actions and validations
	UIStepAction           = core.UIStepAction
	APIStepAction          = core.APIStepAction
	DatabaseStepAction     = core.DatabaseStepAction
	APIStepValidation      = core.APIStepValidation
	DatabaseStepValidation = core.DatabaseStepValidation

	// Interfaces
	Tester            = core.Tester
	UITester          = core.UITester
	APITester         = core.APITester
	DatabaseTester    = core.DatabaseTester
	IntegrationTester = core.IntegrationTester
	Reporter          = core.Reporter
	Transaction       = core.Transaction

	// Configuration types
	Config             = config.Config
	BrowserConfig      = config.BrowserConfig
	APIConfig          = config.APIConfig
	DatabaseConfig     = config.DatabaseConfig
	DatabaseConnection = config.DatabaseConnection
	ReportConfig       = config.ReportConfig
	MobileConfig       = config.MobileConfig
	AuthConfig         = config.AuthConfig
	TLSConfig          = config.TLSConfig
	ProxyConfig        = config.ProxyConfig
	AppiumServerConfig = config.AppiumServerConfig
	DeviceConfig       = config.DeviceConfig
)

// Re-export constants
const (
	TestStatusPassed  = core.TestStatusPassed
	TestStatusFailed  = core.TestStatusFailed
	TestStatusSkipped = core.TestStatusSkipped
	TestStatusError   = core.TestStatusError

	StepTypeUI       = core.StepTypeUI
	StepTypeAPI      = core.StepTypeAPI
	StepTypeDatabase = core.StepTypeDatabase
	StepTypeMobile   = core.StepTypeMobile

	ConfigurationError = core.ConfigurationError
	BrowserError       = core.BrowserError
	APIError           = core.APIError
	DatabaseError      = core.DatabaseError
	ReportingError     = core.ReportingError
	AssertionError     = core.AssertionError
)

// Factory functions for creating tester instances

// NewUITester creates a new UI tester instance
func NewUITester() UITester {
	return ui.NewUITester()
}

// NewAPITester creates a new API tester instance
func NewAPITester() APITester {
	return api.NewAPITester()
}

// NewDatabaseTester creates a new database tester instance
func NewDatabaseTester() DatabaseTester {
	return database.NewDatabaseTester()
}

// NewMobileTester creates a new mobile tester instance
func NewMobileTester() *mobile.MobileTester {
	return mobile.NewMobileTester()
}

// NewIntegrationTester creates a new integration tester instance
func NewIntegrationTester() IntegrationTester {
	return integration.NewIntegrationTester()
}

// NewReportManager creates a new report manager
func NewReportManager(cfg *ReportConfig) *reporting.ReportManager {
	return reporting.NewReportManager(cfg)
}

// Convenience functions for creating Gowright instances

// New creates a new Gowright instance with the provided configuration
func New(cfg *Config) *Gowright {
	return core.New(cfg)
}

// NewWithOptions creates a new Gowright instance with dependency injection support
func NewWithOptions(options *GowrightOptions) *Gowright {
	return core.NewWithOptions(options)
}

// NewWithDefaults creates a new Gowright instance with default configuration
func NewWithDefaults() *Gowright {
	return core.NewWithDefaults()
}

// NewGowright creates a new Gowright instance with the provided configuration and initializes it
// This is an alias for backward compatibility
func NewGowright(cfg *Config) (*Gowright, error) {
	gowright := New(cfg)
	if err := gowright.Initialize(); err != nil {
		return nil, err
	}
	return gowright, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return config.DefaultConfig()
}

// NewGowrightError creates a new GowrightError
func NewGowrightError(errorType ErrorType, message string, cause error) *GowrightError {
	return core.NewGowrightError(errorType, message, cause)
}

// NewGowrightWithAllTesters creates a Gowright instance with all tester types configured
func NewGowrightWithAllTesters(cfg *Config) *Gowright {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	options := &GowrightOptions{
		Config:            cfg,
		UITester:          NewUITester(),
		APITester:         NewAPITester(),
		DatabaseTester:    NewDatabaseTester(),
		IntegrationTester: NewIntegrationTester(),
	}

	gowright := NewWithOptions(options)

	// Set up integration tester with other testers
	if integrationTester, ok := gowright.GetIntegrationTester().(*integration.IntegrationTester); ok {
		integrationTester.SetUITester(gowright.GetUITester())
		integrationTester.SetAPITester(gowright.GetAPITester())
		integrationTester.SetDatabaseTester(gowright.GetDatabaseTester())
	}

	return gowright
}

// Additional utility functions

// GetVersion returns the current framework version
func GetVersion() string {
	return core.GetVersion()
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() *VersionInfo {
	return core.GetVersionInfo()
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return core.GetVersionString()
}

// NewTestExecutor creates a new test executor
func NewTestExecutor(testName string) *TestExecutor {
	return core.NewTestExecutor(testName)
}

// NewTestAssertion creates a new test assertion instance
func NewTestAssertion(testName string) *TestAssertion {
	return core.NewTestAssertion(testName)
}

// NewTestSuite creates a new test suite
func NewTestSuite(name string) *TestSuite {
	return core.NewTestSuite(name)
}

// NewFunctionTest creates a new function-based test
func NewFunctionTest(name string, testFunc func(*TestContext)) Test {
	return core.NewFunctionTest(name, testFunc)
}

// NewMockUITester creates a new mock UI tester
func NewMockUITester() *core.MockUITester {
	return core.NewMockUITester()
}

// NewMockAPITester creates a new mock API tester
func NewMockAPITester() *core.MockAPITester {
	return core.NewMockAPITester()
}

// NewMockDatabaseTester creates a new mock database tester
func NewMockDatabaseTester() *core.MockDatabaseTester {
	return core.NewMockDatabaseTester()
}

// NewMockIntegrationTester creates a new mock integration tester
func NewMockIntegrationTester() *core.MockIntegrationTester {
	return core.NewMockIntegrationTester()
}

// NewTestSuiteManager creates a new test suite manager
func NewTestSuiteManager(suite *TestSuite, cfg *Config) *TestSuiteManager {
	return core.NewTestSuiteManager(suite, cfg)
}

// NewParallelRunner creates a new parallel test runner
func NewParallelRunner(cfg *Config, runnerConfig *core.ParallelRunnerConfig) *ParallelRunner {
	return core.NewParallelRunner(cfg, runnerConfig)
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return core.DefaultRetryConfig()
}
