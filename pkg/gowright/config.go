package gowright

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds the complete framework configuration
type Config struct {
	// Global settings
	LogLevel   string `json:"log_level"`
	Parallel   bool   `json:"parallel"`
	MaxRetries int    `json:"max_retries"`

	// Module-specific configurations
	BrowserConfig  *BrowserConfig  `json:"browser_config,omitempty"`
	APIConfig      *APIConfig      `json:"api_config,omitempty"`
	DatabaseConfig *DatabaseConfig `json:"database_config,omitempty"`
	ReportConfig   *ReportConfig   `json:"report_config,omitempty"`
}

// BrowserConfig holds browser-specific settings for UI testing
type BrowserConfig struct {
	Headless   bool          `json:"headless"`
	Timeout    time.Duration `json:"timeout"`
	UserAgent  string        `json:"user_agent,omitempty"`
	WindowSize *WindowSize   `json:"window_size,omitempty"`
}

// WindowSize represents browser window dimensions
type WindowSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// APIConfig holds API testing configuration
type APIConfig struct {
	BaseURL    string            `json:"base_url,omitempty"`
	Timeout    time.Duration     `json:"timeout"`
	Headers    map[string]string `json:"headers,omitempty"`
	AuthConfig *AuthConfig       `json:"auth_config,omitempty"`
}

// AuthConfig holds authentication configuration for API testing
type AuthConfig struct {
	Type     string            `json:"type"` // bearer, basic, api_key, oauth2
	Token    string            `json:"token,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Connections map[string]*DBConnection `json:"connections"`
}

// DBConnection represents a database connection configuration
type DBConnection struct {
	Driver       string `json:"driver"`
	DSN          string `json:"dsn"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

// ReportConfig holds reporting configuration
type ReportConfig struct {
	LocalReports  LocalReportConfig  `json:"local_reports"`
	RemoteReports RemoteReportConfig `json:"remote_reports"`
}

// LocalReportConfig holds local reporting settings
type LocalReportConfig struct {
	JSON      bool   `json:"json"`
	HTML      bool   `json:"html"`
	OutputDir string `json:"output_dir"`
}

// RemoteReportConfig holds remote reporting settings
type RemoteReportConfig struct {
	JiraXray     *JiraXrayConfig     `json:"jira_xray,omitempty"`
	AIOTest      *AIOTestConfig      `json:"aio_test,omitempty"`
	ReportPortal *ReportPortalConfig `json:"report_portal,omitempty"`
}

// JiraXrayConfig holds Jira Xray integration settings
type JiraXrayConfig struct {
	URL       string `json:"url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	ProjectKey string `json:"project_key"`
}

// AIOTestConfig holds AIOTest integration settings
type AIOTestConfig struct {
	URL    string `json:"url"`
	APIKey string `json:"api_key"`
	ProjectID string `json:"project_id"`
}

// ReportPortalConfig holds Report Portal integration settings
type ReportPortalConfig struct {
	URL       string `json:"url"`
	UUID      string `json:"uuid"`
	Project   string `json:"project"`
	Launch    string `json:"launch"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LogLevel:   "info",
		Parallel:   false,
		MaxRetries: 3,
		BrowserConfig: &BrowserConfig{
			Headless: true,
			Timeout:  30 * time.Second,
			WindowSize: &WindowSize{
				Width:  1920,
				Height: 1080,
			},
		},
		APIConfig: &APIConfig{
			Timeout: 30 * time.Second,
			Headers: make(map[string]string),
		},
		DatabaseConfig: &DatabaseConfig{
			Connections: make(map[string]*DBConnection),
		},
		ReportConfig: &ReportConfig{
			LocalReports: LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: "./reports",
			},
			RemoteReports: RemoteReportConfig{},
		},
	}
}

// LoadConfigFromFile loads configuration from a JSON file
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	if logLevel := os.Getenv("GOWRIGHT_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if parallel := os.Getenv("GOWRIGHT_PARALLEL"); parallel == "true" {
		config.Parallel = true
	}

	if headless := os.Getenv("GOWRIGHT_HEADLESS"); headless == "false" {
		config.BrowserConfig.Headless = false
	}

	if baseURL := os.Getenv("GOWRIGHT_API_BASE_URL"); baseURL != "" {
		config.APIConfig.BaseURL = baseURL
	}

	// Additional environment variable support
	if userAgent := os.Getenv("GOWRIGHT_USER_AGENT"); userAgent != "" {
		config.BrowserConfig.UserAgent = userAgent
	}

	if outputDir := os.Getenv("GOWRIGHT_REPORT_OUTPUT_DIR"); outputDir != "" {
		config.ReportConfig.LocalReports.OutputDir = outputDir
	}

	return config
}

// LoadHierarchicalConfig loads configuration with hierarchical support
// Priority: code config > environment variables > file config > defaults
func LoadHierarchicalConfig(filePath string, codeConfig *Config) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()
	
	// Apply file configuration if provided
	if filePath != "" {
		fileConfig, err := LoadConfigFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load file config: %w", err)
		}
		config = mergeConfigs(config, fileConfig)
	}
	
	// Apply environment variables
	envConfig := LoadConfigFromEnv()
	config = mergeConfigs(config, envConfig)
	
	// Apply code configuration if provided
	if codeConfig != nil {
		config = mergeConfigs(config, codeConfig)
	}
	
	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	return config, nil
}

// mergeConfigs merges two configurations, with the second taking precedence
func mergeConfigs(base, override *Config) *Config {
	result := *base // Copy base config
	
	// Merge global settings
	if override.LogLevel != "" && override.LogLevel != "info" {
		result.LogLevel = override.LogLevel
	}
	if override.Parallel {
		result.Parallel = override.Parallel
	}
	if override.MaxRetries != 0 && override.MaxRetries != 3 {
		result.MaxRetries = override.MaxRetries
	}
	
	// Merge browser config
	if override.BrowserConfig != nil {
		if result.BrowserConfig == nil {
			result.BrowserConfig = &BrowserConfig{}
		}
		mergeBrowserConfig(result.BrowserConfig, override.BrowserConfig)
	}
	
	// Merge API config
	if override.APIConfig != nil {
		if result.APIConfig == nil {
			result.APIConfig = &APIConfig{}
		}
		mergeAPIConfig(result.APIConfig, override.APIConfig)
	}
	
	// Merge database config
	if override.DatabaseConfig != nil {
		if result.DatabaseConfig == nil {
			result.DatabaseConfig = &DatabaseConfig{}
		}
		mergeDatabaseConfig(result.DatabaseConfig, override.DatabaseConfig)
	}
	
	// Merge report config
	if override.ReportConfig != nil {
		if result.ReportConfig == nil {
			result.ReportConfig = &ReportConfig{}
		}
		mergeReportConfig(result.ReportConfig, override.ReportConfig)
	}
	
	return &result
}

func mergeBrowserConfig(base, override *BrowserConfig) {
	if override.Headless != base.Headless {
		base.Headless = override.Headless
	}
	if override.Timeout != 0 {
		base.Timeout = override.Timeout
	}
	if override.UserAgent != "" {
		base.UserAgent = override.UserAgent
	}
	if override.WindowSize != nil {
		if base.WindowSize == nil {
			base.WindowSize = &WindowSize{}
		}
		if override.WindowSize.Width != 0 {
			base.WindowSize.Width = override.WindowSize.Width
		}
		if override.WindowSize.Height != 0 {
			base.WindowSize.Height = override.WindowSize.Height
		}
	}
}

func mergeAPIConfig(base, override *APIConfig) {
	if override.BaseURL != "" {
		base.BaseURL = override.BaseURL
	}
	if override.Timeout != 0 {
		base.Timeout = override.Timeout
	}
	if override.Headers != nil {
		if base.Headers == nil {
			base.Headers = make(map[string]string)
		}
		for k, v := range override.Headers {
			base.Headers[k] = v
		}
	}
	if override.AuthConfig != nil {
		base.AuthConfig = override.AuthConfig
	}
}

func mergeDatabaseConfig(base, override *DatabaseConfig) {
	if override.Connections != nil {
		if base.Connections == nil {
			base.Connections = make(map[string]*DBConnection)
		}
		for k, v := range override.Connections {
			base.Connections[k] = v
		}
	}
}

func mergeReportConfig(base, override *ReportConfig) {
	// Merge local reports
	if override.LocalReports.JSON != base.LocalReports.JSON {
		base.LocalReports.JSON = override.LocalReports.JSON
	}
	if override.LocalReports.HTML != base.LocalReports.HTML {
		base.LocalReports.HTML = override.LocalReports.HTML
	}
	if override.LocalReports.OutputDir != "" && override.LocalReports.OutputDir != "./reports" {
		base.LocalReports.OutputDir = override.LocalReports.OutputDir
	}
	
	// Merge remote reports
	if override.RemoteReports.JiraXray != nil {
		base.RemoteReports.JiraXray = override.RemoteReports.JiraXray
	}
	if override.RemoteReports.AIOTest != nil {
		base.RemoteReports.AIOTest = override.RemoteReports.AIOTest
	}
	if override.RemoteReports.ReportPortal != nil {
		base.RemoteReports.ReportPortal = override.RemoteReports.ReportPortal
	}
}

// SaveToFile saves the configuration to a JSON file
func (c *Config) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration parameters
func (c *Config) Validate() error {
	var errors []string
	
	// Validate global settings
	if c.LogLevel != "" {
		validLogLevels := []string{"debug", "info", "warn", "error"}
		valid := false
		for _, level := range validLogLevels {
			if c.LogLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, fmt.Sprintf("invalid log level: %s (must be one of: %v)", c.LogLevel, validLogLevels))
		}
	}
	
	if c.MaxRetries < 0 {
		errors = append(errors, "max_retries cannot be negative")
	}
	
	// Validate browser config
	if c.BrowserConfig != nil {
		if err := c.BrowserConfig.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("browser config validation failed: %v", err))
		}
	}
	
	// Validate API config
	if c.APIConfig != nil {
		if err := c.APIConfig.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("API config validation failed: %v", err))
		}
	}
	
	// Validate database config
	if c.DatabaseConfig != nil {
		if err := c.DatabaseConfig.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("database config validation failed: %v", err))
		}
	}
	
	// Validate report config
	if c.ReportConfig != nil {
		if err := c.ReportConfig.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("report config validation failed: %v", err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation errors: %v", errors)
	}
	
	return nil
}

// Validate validates browser configuration
func (bc *BrowserConfig) Validate() error {
	var errors []string
	
	if bc.Timeout <= 0 {
		errors = append(errors, "timeout must be positive")
	}
	
	if bc.WindowSize != nil {
		if bc.WindowSize.Width <= 0 {
			errors = append(errors, "window width must be positive")
		}
		if bc.WindowSize.Height <= 0 {
			errors = append(errors, "window height must be positive")
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("browser config errors: %v", errors)
	}
	
	return nil
}

// Validate validates API configuration
func (ac *APIConfig) Validate() error {
	var errors []string
	
	if ac.Timeout <= 0 {
		errors = append(errors, "timeout must be positive")
	}
	
	if ac.AuthConfig != nil {
		if err := ac.AuthConfig.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("auth config validation failed: %v", err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("API config errors: %v", errors)
	}
	
	return nil
}

// Validate validates authentication configuration
func (auth *AuthConfig) Validate() error {
	var errors []string
	
	validTypes := []string{"bearer", "basic", "api_key", "oauth2"}
	valid := false
	for _, t := range validTypes {
		if auth.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		errors = append(errors, fmt.Sprintf("invalid auth type: %s (must be one of: %v)", auth.Type, validTypes))
	}
	
	switch auth.Type {
	case "bearer":
		if auth.Token == "" {
			errors = append(errors, "token is required for bearer authentication")
		}
	case "basic":
		if auth.Username == "" || auth.Password == "" {
			errors = append(errors, "username and password are required for basic authentication")
		}
	case "api_key":
		if auth.Token == "" {
			errors = append(errors, "token is required for API key authentication")
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("auth config errors: %v", errors)
	}
	
	return nil
}

// Validate validates database configuration
func (dc *DatabaseConfig) Validate() error {
	var errors []string
	
	if dc.Connections == nil || len(dc.Connections) == 0 {
		// Empty connections is valid - no validation needed
		return nil
	}
	
	for name, conn := range dc.Connections {
		if err := conn.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("connection '%s' validation failed: %v", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("database config errors: %v", errors)
	}
	
	return nil
}

// Validate validates database connection configuration
func (dbc *DBConnection) Validate() error {
	var errors []string
	
	if dbc.Driver == "" {
		errors = append(errors, "driver is required")
	}
	
	if dbc.DSN == "" {
		errors = append(errors, "DSN is required")
	}
	
	if dbc.MaxOpenConns < 0 {
		errors = append(errors, "max_open_conns cannot be negative")
	}
	
	if dbc.MaxIdleConns < 0 {
		errors = append(errors, "max_idle_conns cannot be negative")
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("DB connection errors: %v", errors)
	}
	
	return nil
}

// Validate validates report configuration
func (rc *ReportConfig) Validate() error {
	var errors []string
	
	if rc.LocalReports.OutputDir == "" {
		errors = append(errors, "output directory cannot be empty")
	}
	
	// Validate remote report configs
	if rc.RemoteReports.JiraXray != nil {
		if err := rc.RemoteReports.JiraXray.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("Jira Xray config validation failed: %v", err))
		}
	}
	
	if rc.RemoteReports.AIOTest != nil {
		if err := rc.RemoteReports.AIOTest.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("AIOTest config validation failed: %v", err))
		}
	}
	
	if rc.RemoteReports.ReportPortal != nil {
		if err := rc.RemoteReports.ReportPortal.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("Report Portal config validation failed: %v", err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("report config errors: %v", errors)
	}
	
	return nil
}

// Validate validates Jira Xray configuration
func (jx *JiraXrayConfig) Validate() error {
	var errors []string
	
	if jx.URL == "" {
		errors = append(errors, "URL is required")
	}
	if jx.Username == "" {
		errors = append(errors, "username is required")
	}
	if jx.Password == "" {
		errors = append(errors, "password is required")
	}
	if jx.ProjectKey == "" {
		errors = append(errors, "project key is required")
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("Jira Xray config errors: %v", errors)
	}
	
	return nil
}

// Validate validates AIOTest configuration
func (at *AIOTestConfig) Validate() error {
	var errors []string
	
	if at.URL == "" {
		errors = append(errors, "URL is required")
	}
	if at.APIKey == "" {
		errors = append(errors, "API key is required")
	}
	if at.ProjectID == "" {
		errors = append(errors, "project ID is required")
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("AIOTest config errors: %v", errors)
	}
	
	return nil
}

// Validate validates Report Portal configuration
func (rp *ReportPortalConfig) Validate() error {
	var errors []string
	
	if rp.URL == "" {
		errors = append(errors, "URL is required")
	}
	if rp.UUID == "" {
		errors = append(errors, "UUID is required")
	}
	if rp.Project == "" {
		errors = append(errors, "project is required")
	}
	if rp.Launch == "" {
		errors = append(errors, "launch is required")
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("Report Portal config errors: %v", errors)
	}
	
	return nil
}