package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.False(t, config.Parallel)
	assert.Equal(t, 4, config.MaxWorkers)

	// Test browser config
	require.NotNil(t, config.BrowserConfig)
	assert.True(t, config.BrowserConfig.Headless)
	assert.Equal(t, 30*time.Second, config.BrowserConfig.Timeout)
	assert.Equal(t, "1920x1080", config.BrowserConfig.WindowSize)

	// Test API config
	require.NotNil(t, config.APIConfig)
	assert.Equal(t, 30*time.Second, config.APIConfig.Timeout)
	assert.Equal(t, 3, config.APIConfig.RetryCount)

	// Test database config
	require.NotNil(t, config.DatabaseConfig)
	assert.NotNil(t, config.DatabaseConfig.Connections)
	assert.Equal(t, 10, config.DatabaseConfig.MaxOpenConns)

	// Test report config
	require.NotNil(t, config.ReportConfig)
	assert.True(t, config.ReportConfig.Enabled)
	assert.Equal(t, "./reports", config.ReportConfig.OutputDir)
	assert.Contains(t, config.ReportConfig.Formats, "html")
	assert.Contains(t, config.ReportConfig.Formats, "json")

	// Test mobile config
	require.NotNil(t, config.MobileConfig)
	assert.Equal(t, "localhost", config.MobileConfig.AppiumServer.Host)
	assert.Equal(t, 4723, config.MobileConfig.AppiumServer.Port)
}

func TestBrowserConfigDefaults(t *testing.T) {
	config := DefaultConfig()
	browserConfig := config.BrowserConfig

	assert.Equal(t, "chrome", browserConfig.Browser)
	assert.True(t, browserConfig.Headless)
	assert.Equal(t, "1920x1080", browserConfig.WindowSize)
	assert.Equal(t, 30*time.Second, browserConfig.Timeout)
	assert.Equal(t, "./screenshots", browserConfig.ScreenshotPath)
	assert.Equal(t, 5, browserConfig.MaxInstances)
	assert.True(t, browserConfig.ReuseInstances)
}

func TestAPIConfigDefaults(t *testing.T) {
	config := DefaultConfig()
	apiConfig := config.APIConfig

	assert.Equal(t, 30*time.Second, apiConfig.Timeout)
	assert.Equal(t, 3, apiConfig.RetryCount)
	assert.Equal(t, 1*time.Second, apiConfig.RetryDelay)
	assert.Equal(t, 10, apiConfig.MaxConnections)
	assert.True(t, apiConfig.KeepAlive)
	assert.True(t, apiConfig.FollowRedirects)
}

func TestDatabaseConfigDefaults(t *testing.T) {
	config := DefaultConfig()
	dbConfig := config.DatabaseConfig

	assert.Equal(t, 10, dbConfig.MaxOpenConns)
	assert.Equal(t, 5, dbConfig.MaxIdleConns)
	assert.Equal(t, time.Hour, dbConfig.ConnMaxLife)
	assert.Equal(t, 30*time.Second, dbConfig.QueryTimeout)
	assert.Equal(t, 60*time.Second, dbConfig.TransactionTimeout)
	assert.NotNil(t, dbConfig.Connections)
}

func TestReportConfigDefaults(t *testing.T) {
	config := DefaultConfig()
	reportConfig := config.ReportConfig

	assert.True(t, reportConfig.Enabled)
	assert.Equal(t, "./reports", reportConfig.OutputDir)
	assert.Contains(t, reportConfig.Formats, "html")
	assert.Contains(t, reportConfig.Formats, "json")
	assert.True(t, reportConfig.IncludeLogs)
	assert.False(t, reportConfig.Compress)
}

func TestMobileConfigDefaults(t *testing.T) {
	config := DefaultConfig()
	mobileConfig := config.MobileConfig

	assert.Equal(t, "localhost", mobileConfig.AppiumServer.Host)
	assert.Equal(t, 4723, mobileConfig.AppiumServer.Port)
	assert.Equal(t, 30*time.Second, mobileConfig.Timeout)
	assert.Equal(t, 10*time.Second, mobileConfig.ImplicitWait)
	assert.Equal(t, "./screenshots", mobileConfig.ScreenshotPath)
}

func TestProxyConfig(t *testing.T) {
	proxy := &ProxyConfig{
		Host:     "proxy.example.com",
		Port:     8080,
		Username: "user",
		Password: "pass",
	}

	assert.Equal(t, "proxy.example.com", proxy.Host)
	assert.Equal(t, 8080, proxy.Port)
	assert.Equal(t, "user", proxy.Username)
	assert.Equal(t, "pass", proxy.Password)
}

func TestAuthConfig(t *testing.T) {
	// Test Bearer auth
	bearerAuth := &AuthConfig{
		Type:  "bearer",
		Token: "test-token",
	}
	assert.Equal(t, "bearer", bearerAuth.Type)
	assert.Equal(t, "test-token", bearerAuth.Token)

	// Test Basic auth
	basicAuth := &AuthConfig{
		Type:     "basic",
		Username: "user",
		Password: "pass",
	}
	assert.Equal(t, "basic", basicAuth.Type)
	assert.Equal(t, "user", basicAuth.Username)
	assert.Equal(t, "pass", basicAuth.Password)

	// Test API Key auth
	apiKeyAuth := &AuthConfig{
		Type:   "api_key",
		APIKey: "api-key-123",
	}
	assert.Equal(t, "api_key", apiKeyAuth.Type)
	assert.Equal(t, "api-key-123", apiKeyAuth.APIKey)
}

func TestTLSConfig(t *testing.T) {
	tlsConfig := &TLSConfig{
		InsecureSkipVerify: true,
		CertFile:           "/path/to/cert.pem",
		KeyFile:            "/path/to/key.pem",
		CAFile:             "/path/to/ca.pem",
	}

	assert.True(t, tlsConfig.InsecureSkipVerify)
	assert.Equal(t, "/path/to/cert.pem", tlsConfig.CertFile)
	assert.Equal(t, "/path/to/key.pem", tlsConfig.KeyFile)
	assert.Equal(t, "/path/to/ca.pem", tlsConfig.CAFile)
}

func TestDatabaseConnection(t *testing.T) {
	dbConn := &DatabaseConnection{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "require",
		Options: map[string]string{
			"charset": "utf8",
		},
	}

	assert.Equal(t, "postgres", dbConn.Driver)
	assert.Equal(t, "localhost", dbConn.Host)
	assert.Equal(t, 5432, dbConn.Port)
	assert.Equal(t, "testdb", dbConn.Database)
	assert.Equal(t, "user", dbConn.Username)
	assert.Equal(t, "pass", dbConn.Password)
	assert.Equal(t, "require", dbConn.SSLMode)
	assert.Equal(t, "utf8", dbConn.Options["charset"])
}

func TestDeviceConfig(t *testing.T) {
	device := &DeviceConfig{
		Name:         "iPhone 12",
		PlatformName: "iOS",
		DeviceName:   "iPhone 12 Simulator",
		UDID:         "12345-67890",
		Version:      "14.5",
		Capabilities: map[string]interface{}{
			"autoAcceptAlerts": true,
		},
	}

	assert.Equal(t, "iPhone 12", device.Name)
	assert.Equal(t, "iOS", device.PlatformName)
	assert.Equal(t, "iPhone 12 Simulator", device.DeviceName)
	assert.Equal(t, "12345-67890", device.UDID)
	assert.Equal(t, "14.5", device.Version)
	assert.True(t, device.Capabilities["autoAcceptAlerts"].(bool))
}

func TestConfigModification(t *testing.T) {
	config := DefaultConfig()

	// Modify browser config
	config.BrowserConfig.Headless = false
	config.BrowserConfig.Browser = "firefox"
	config.BrowserConfig.WindowSize = "1366x768"

	assert.False(t, config.BrowserConfig.Headless)
	assert.Equal(t, "firefox", config.BrowserConfig.Browser)
	assert.Equal(t, "1366x768", config.BrowserConfig.WindowSize)

	// Modify API config
	config.APIConfig.BaseURL = "https://api.example.com"
	config.APIConfig.Timeout = 60 * time.Second

	assert.Equal(t, "https://api.example.com", config.APIConfig.BaseURL)
	assert.Equal(t, 60*time.Second, config.APIConfig.Timeout)

	// Add database connection
	config.DatabaseConfig.Connections["test"] = &DatabaseConnection{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}

	testConn := config.DatabaseConfig.Connections["test"]
	assert.NotNil(t, testConn)
	assert.Equal(t, "mysql", testConn.Driver)
	assert.Equal(t, 3306, testConn.Port)
}

func TestConfigCopy(t *testing.T) {
	original := DefaultConfig()
	original.BrowserConfig.Headless = false
	original.APIConfig.BaseURL = "https://api.example.com"

	// Create a copy by creating a new default config
	copy := DefaultConfig()

	// Original modifications should not affect the copy
	assert.True(t, copy.BrowserConfig.Headless) // Should be default value
	assert.Empty(t, copy.APIConfig.BaseURL)     // Should be default value

	// Modify copy
	copy.BrowserConfig.Browser = "safari"
	copy.APIConfig.Timeout = 45 * time.Second

	// Original should not be affected
	assert.Equal(t, "chrome", original.BrowserConfig.Browser)
	assert.Equal(t, 30*time.Second, original.APIConfig.Timeout)
}
