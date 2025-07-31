package gowright

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds framework configuration
type Config struct {
	// Global settings
	LogLevel    string `json:"log_level"`
	Parallel    bool   `json:"parallel"`
	MaxRetries  int    `json:"max_retries"`
	
	// Module-specific configs
	Browser   *BrowserConfig   `json:"browser"`
	API       *APIConfig       `json:"api"`
	Database  *DatabaseConfig  `json:"database"`
	Reporting *ReportConfig    `json:"reporting"`
}

// BrowserConfig holds browser-specific settings
type BrowserConfig struct {
	Headless    bool         `json:"headless"`
	Timeout     time.Duration `json:"timeout"`
	UserAgent   string       `json:"user_agent"`
	WindowSize  *WindowSize  `json:"window_size"`
}

// APIConfig holds API testing configuration
type APIConfig struct {
	BaseURL     string            `json:"base_url"`
	Timeout     time.Duration     `json:"timeout"`
	Headers     map[string]string `json:"headers"`
	AuthConfig  *AuthConfig       `json:"auth_config"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Connections map[string]*DBConnection `json:"connections"`
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
	JiraXray     *JiraXrayConfig     `json:"jira_xray"`
	AIOTest      *AIOTestConfig      `json:"aio_test"`
	ReportPortal *ReportPortalConfig `json:"report_portal"`
}

// JiraXrayConfig holds Jira Xray configuration
type JiraXrayConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Project  string `json:"project"`
}

// AIOTestConfig holds AIOTest configuration
type AIOTestConfig struct {
	URL    string `json:"url"`
	APIKey string `json:"api_key"`
	Project string `json:"project"`
}

// ReportPortalConfig holds Report Portal configuration
type ReportPortalConfig struct {
	URL     string `json:"url"`
	UUID    string `json:"uuid"`
	Project string `json:"project"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		LogLevel:   "INFO",
		Parallel:   false,
		MaxRetries: 3,
		Browser: &BrowserConfig{
			Headless:  true,
			Timeout:   30 * time.Second,
			UserAgent: "Gowright Testing Framework",
			WindowSize: &WindowSize{
				Width:  1920,
				Height: 1080,
			},
		},
		API: &APIConfig{
			Timeout: 30 * time.Second,
			Headers: make(map[string]string),
		},
		Database: &DatabaseConfig{
			Connections: make(map[string]*DBConnection),
		},
		Reporting: &ReportConfig{
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
func LoadConfigFromFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
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

	if baseURL := os.Getenv("GOWRIGHT_API_BASE_URL"); baseURL != "" {
		config.API.BaseURL = baseURL
	}

	if headless := os.Getenv("GOWRIGHT_BROWSER_HEADLESS"); headless == "false" {
		config.Browser.Headless = false
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	if c.Browser != nil {
		if c.Browser.Timeout <= 0 {
			return fmt.Errorf("browser timeout must be positive")
		}
		if c.Browser.WindowSize != nil {
			if c.Browser.WindowSize.Width <= 0 || c.Browser.WindowSize.Height <= 0 {
				return fmt.Errorf("browser window size must be positive")
			}
		}
	}

	if c.API != nil && c.API.Timeout <= 0 {
		return fmt.Errorf("API timeout must be positive")
	}

	return nil
}