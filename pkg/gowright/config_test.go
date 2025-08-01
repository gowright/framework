package gowright

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, "info", config.LogLevel)
	assert.False(t, config.Parallel)
	assert.Equal(t, 3, config.MaxRetries)
	
	// Test browser config
	require.NotNil(t, config.BrowserConfig)
	assert.True(t, config.BrowserConfig.Headless)
	assert.Equal(t, 30*time.Second, config.BrowserConfig.Timeout)
	assert.Equal(t, 1920, config.BrowserConfig.WindowSize.Width)
	assert.Equal(t, 1080, config.BrowserConfig.WindowSize.Height)
	
	// Test API config
	require.NotNil(t, config.APIConfig)
	assert.Equal(t, 30*time.Second, config.APIConfig.Timeout)
	assert.NotNil(t, config.APIConfig.Headers)
	
	// Test database config
	require.NotNil(t, config.DatabaseConfig)
	assert.NotNil(t, config.DatabaseConfig.Connections)
	
	// Test report config
	require.NotNil(t, config.ReportConfig)
	assert.True(t, config.ReportConfig.LocalReports.JSON)
	assert.True(t, config.ReportConfig.LocalReports.HTML)
	assert.Equal(t, "./reports", config.ReportConfig.LocalReports.OutputDir)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("GOWRIGHT_LOG_LEVEL", "debug")
	os.Setenv("GOWRIGHT_PARALLEL", "true")
	os.Setenv("GOWRIGHT_HEADLESS", "false")
	os.Setenv("GOWRIGHT_API_BASE_URL", "https://api.test.com")
	
	defer func() {
		os.Unsetenv("GOWRIGHT_LOG_LEVEL")
		os.Unsetenv("GOWRIGHT_PARALLEL")
		os.Unsetenv("GOWRIGHT_HEADLESS")
		os.Unsetenv("GOWRIGHT_API_BASE_URL")
	}()
	
	config := LoadConfigFromEnv()
	
	assert.Equal(t, "debug", config.LogLevel)
	assert.True(t, config.Parallel)
	assert.False(t, config.BrowserConfig.Headless)
	assert.Equal(t, "https://api.test.com", config.APIConfig.BaseURL)
}

func TestConfigSaveAndLoad(t *testing.T) {
	config := DefaultConfig()
	config.LogLevel = "debug"
	config.Parallel = true
	
	filename := "test-config.json"
	defer os.Remove(filename)
	
	// Test saving
	err := config.SaveToFile(filename)
	require.NoError(t, err)
	
	// Test loading
	loadedConfig, err := LoadConfigFromFile(filename)
	require.NoError(t, err)
	
	assert.Equal(t, "debug", loadedConfig.LogLevel)
	assert.True(t, loadedConfig.Parallel)
	assert.Equal(t, config.MaxRetries, loadedConfig.MaxRetries)
}

func TestLoadHierarchicalConfig(t *testing.T) {
	// Create a test config file
	filename := "test-hierarchical-config.json"
	defer os.Remove(filename)
	
	fileConfig := &Config{
		LogLevel:   "warn",
		MaxRetries: 5,
	}
	err := fileConfig.SaveToFile(filename)
	require.NoError(t, err)
	
	// Set environment variables
	os.Setenv("GOWRIGHT_LOG_LEVEL", "error")
	os.Setenv("GOWRIGHT_PARALLEL", "true")
	defer func() {
		os.Unsetenv("GOWRIGHT_LOG_LEVEL")
		os.Unsetenv("GOWRIGHT_PARALLEL")
	}()
	
	// Code config (highest priority)
	codeConfig := &Config{
		LogLevel: "debug",
	}
	
	// Load hierarchical config
	config, err := LoadHierarchicalConfig(filename, codeConfig)
	require.NoError(t, err)
	
	// Code config should take precedence
	assert.Equal(t, "debug", config.LogLevel)
	// Environment should override file
	assert.True(t, config.Parallel)
	// File should override defaults
	assert.Equal(t, 5, config.MaxRetries)
}

func TestLoadHierarchicalConfigWithInvalidFile(t *testing.T) {
	config, err := LoadHierarchicalConfig("nonexistent.json", nil)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid default config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				LogLevel: "invalid",
			},
			expectError: true,
		},
		{
			name: "negative max retries",
			config: &Config{
				MaxRetries: -1,
			},
			expectError: true,
		},
		{
			name: "invalid browser config",
			config: &Config{
				BrowserConfig: &BrowserConfig{
					Timeout: -1 * time.Second,
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBrowserConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *BrowserConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &BrowserConfig{
				Timeout: 30 * time.Second,
				WindowSize: &WindowSize{
					Width:  1920,
					Height: 1080,
				},
			},
			expectError: false,
		},
		{
			name: "negative timeout",
			config: &BrowserConfig{
				Timeout: -1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "invalid window size",
			config: &BrowserConfig{
				Timeout: 30 * time.Second,
				WindowSize: &WindowSize{
					Width:  -1,
					Height: 1080,
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *APIConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &APIConfig{
				Timeout: 30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "negative timeout",
			config: &APIConfig{
				Timeout: -1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "invalid auth config",
			config: &APIConfig{
				Timeout: 30 * time.Second,
				AuthConfig: &AuthConfig{
					Type: "invalid",
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *AuthConfig
		expectError bool
	}{
		{
			name: "valid bearer auth",
			config: &AuthConfig{
				Type:  "bearer",
				Token: "test-token",
			},
			expectError: false,
		},
		{
			name: "valid basic auth",
			config: &AuthConfig{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			expectError: false,
		},
		{
			name: "invalid auth type",
			config: &AuthConfig{
				Type: "invalid",
			},
			expectError: true,
		},
		{
			name: "bearer auth without token",
			config: &AuthConfig{
				Type: "bearer",
			},
			expectError: true,
		},
		{
			name: "basic auth without credentials",
			config: &AuthConfig{
				Type: "basic",
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *DatabaseConfig
		expectError bool
	}{
		{
			name: "empty connections",
			config: &DatabaseConfig{
				Connections: make(map[string]*DBConnection),
			},
			expectError: false,
		},
		{
			name: "valid connection",
			config: &DatabaseConfig{
				Connections: map[string]*DBConnection{
					"test": {
						Driver: "postgres",
						DSN:    "postgres://user:pass@localhost/db",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid connection",
			config: &DatabaseConfig{
				Connections: map[string]*DBConnection{
					"test": {
						Driver: "",
						DSN:    "",
					},
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBConnectionValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *DBConnection
		expectError bool
	}{
		{
			name: "valid connection",
			config: &DBConnection{
				Driver:       "postgres",
				DSN:          "postgres://user:pass@localhost/db",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			expectError: false,
		},
		{
			name: "missing driver",
			config: &DBConnection{
				DSN: "postgres://user:pass@localhost/db",
			},
			expectError: true,
		},
		{
			name: "missing DSN",
			config: &DBConnection{
				Driver: "postgres",
			},
			expectError: true,
		},
		{
			name: "negative max connections",
			config: &DBConnection{
				Driver:       "postgres",
				DSN:          "postgres://user:pass@localhost/db",
				MaxOpenConns: -1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReportConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *ReportConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					OutputDir: "./reports",
				},
			},
			expectError: false,
		},
		{
			name: "empty output dir",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					OutputDir: "",
				},
			},
			expectError: true,
		},
		{
			name: "invalid jira xray config",
			config: &ReportConfig{
				LocalReports: LocalReportConfig{
					OutputDir: "./reports",
				},
				RemoteReports: RemoteReportConfig{
					JiraXray: &JiraXrayConfig{
						URL: "", // Invalid - empty URL
					},
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	base := &Config{
		LogLevel:   "info",
		MaxRetries: 3,
		BrowserConfig: &BrowserConfig{
			Headless: true,
			Timeout:  30 * time.Second,
		},
	}
	
	override := &Config{
		LogLevel: "debug",
		BrowserConfig: &BrowserConfig{
			Headless: false,
		},
	}
	
	result := mergeConfigs(base, override)
	
	assert.Equal(t, "debug", result.LogLevel)
	assert.Equal(t, 3, result.MaxRetries) // Should keep base value
	assert.False(t, result.BrowserConfig.Headless)
	assert.Equal(t, 30*time.Second, result.BrowserConfig.Timeout) // Should keep base value
}