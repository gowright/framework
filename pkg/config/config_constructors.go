package config

import (
	"time"
)

// ConfigBuilder provides a fluent interface for building configurations
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// WithBrowser configures browser settings
func (cb *ConfigBuilder) WithBrowser(browser string, headless bool) *ConfigBuilder {
	if cb.config.BrowserConfig == nil {
		cb.config.BrowserConfig = &BrowserConfig{}
	}
	cb.config.BrowserConfig.Browser = browser
	cb.config.BrowserConfig.Headless = headless
	return cb
}

// WithBrowserTimeout sets browser timeout
func (cb *ConfigBuilder) WithBrowserTimeout(timeout time.Duration) *ConfigBuilder {
	if cb.config.BrowserConfig == nil {
		cb.config.BrowserConfig = &BrowserConfig{}
	}
	cb.config.BrowserConfig.Timeout = timeout
	return cb
}

// WithAPIConfig configures API settings
func (cb *ConfigBuilder) WithAPIConfig(baseURL string, timeout time.Duration) *ConfigBuilder {
	if cb.config.APIConfig == nil {
		cb.config.APIConfig = &APIConfig{}
	}
	cb.config.APIConfig.BaseURL = baseURL
	cb.config.APIConfig.Timeout = timeout
	return cb
}

// WithDatabaseConfig configures database settings
func (cb *ConfigBuilder) WithDatabaseConfig(maxOpen, maxIdle int) *ConfigBuilder {
	if cb.config.DatabaseConfig == nil {
		cb.config.DatabaseConfig = &DatabaseConfig{}
	}
	cb.config.DatabaseConfig.MaxOpenConns = maxOpen
	cb.config.DatabaseConfig.MaxIdleConns = maxIdle
	return cb
}

// WithParallel configures parallel execution
func (cb *ConfigBuilder) WithParallel(enabled bool, maxWorkers int) *ConfigBuilder {
	cb.config.Parallel = enabled
	cb.config.MaxWorkers = maxWorkers
	return cb
}

// WithTimeout sets global timeout
func (cb *ConfigBuilder) WithTimeout(timeout time.Duration) *ConfigBuilder {
	cb.config.Timeout = timeout
	return cb
}

// Build returns the built configuration
func (cb *ConfigBuilder) Build() *Config {
	return cb.config
}
