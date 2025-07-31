// Package gowright provides a comprehensive testing framework for Go
// that supports UI (browser, mobile), API, database, and integration testing.
package gowright

import (
	"fmt"
)

// New creates a new Gowright instance with the provided configuration
func New(config *Config) (*Gowright, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	gw := &Gowright{
		config: config,
	}

	// Initialize report manager
	reportManager, err := NewReportManager(config.Reporting)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report manager: %w", err)
	}
	gw.reporter = reportManager

	return gw, nil
}

// NewWithDefaults creates a new Gowright instance with default configuration
func NewWithDefaults() (*Gowright, error) {
	return New(DefaultConfig())
}

// NewFromFile creates a new Gowright instance with configuration loaded from file
func NewFromFile(configPath string) (*Gowright, error) {
	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from file: %w", err)
	}

	return New(config)
}

// NewFromEnv creates a new Gowright instance with configuration loaded from environment
func NewFromEnv() (*Gowright, error) {
	config := LoadConfigFromEnv()
	return New(config)
}

// GetConfig returns the current configuration
func (gw *Gowright) GetConfig() *Config {
	return gw.config
}

// SetTestSuite sets the test suite for execution
func (gw *Gowright) SetTestSuite(suite *TestSuite) {
	gw.testSuite = suite
}

// GetTestSuite returns the current test suite
func (gw *Gowright) GetTestSuite() *TestSuite {
	return gw.testSuite
}

// GetReportManager returns the report manager
func (gw *Gowright) GetReportManager() *ReportManager {
	return gw.reporter
}

// NewReportManager creates a new report manager with the given configuration
func NewReportManager(config *ReportConfig) (*ReportManager, error) {
	if config == nil {
		return nil, fmt.Errorf("report config cannot be nil")
	}

	rm := &ReportManager{
		config:    config,
		reporters: make([]Reporter, 0),
		results:   &TestResults{},
	}

	return rm, nil
}

// Version returns the framework version
func Version() string {
	return "0.1.0"
}