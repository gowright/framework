package config

import "time"

// Config represents the main configuration for the Gowright framework
type Config struct {
	BrowserConfig  *BrowserConfig  `json:"browser_config"`
	APIConfig      *APIConfig      `json:"api_config"`
	DatabaseConfig *DatabaseConfig `json:"database_config"`
	ReportConfig   *ReportConfig   `json:"report_config"`
	MobileConfig   *MobileConfig   `json:"mobile_config"`
	Parallel       bool            `json:"parallel"`
	MaxWorkers     int             `json:"max_workers"`
	Timeout        time.Duration   `json:"timeout"`
}

// BrowserConfig holds browser-specific configuration
type BrowserConfig struct {
	Browser        string        `json:"browser"`
	Headless       bool          `json:"headless"`
	WindowSize     string        `json:"window_size"`
	Timeout        time.Duration `json:"timeout"`
	ScreenshotPath string        `json:"screenshot_path"`
	DownloadPath   string        `json:"download_path"`
	UserAgent      string        `json:"user_agent"`
	DisableImages  bool          `json:"disable_images"`
	DisableCSS     bool          `json:"disable_css"`
	DisableJS      bool          `json:"disable_js"`
	MaxInstances   int           `json:"max_instances"`
	ReuseInstances bool          `json:"reuse_instances"`
	BrowserArgs    []string      `json:"browser_args"`
	Extensions     []string      `json:"extensions"`
	Proxy          *ProxyConfig  `json:"proxy,omitempty"`
}

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// APIConfig holds API testing configuration
type APIConfig struct {
	BaseURL         string            `json:"base_url"`
	Timeout         time.Duration     `json:"timeout"`
	RetryCount      int               `json:"retry_count"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	DefaultHeaders  map[string]string `json:"default_headers"`
	Auth            *AuthConfig       `json:"auth,omitempty"`
	TLSConfig       *TLSConfig        `json:"tls_config,omitempty"`
	MaxConnections  int               `json:"max_connections"`
	KeepAlive       bool              `json:"keep_alive"`
	FollowRedirects bool              `json:"follow_redirects"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type     string            `json:"type"` // basic, bearer, oauth2, api_key
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Token    string            `json:"token,omitempty"`
	APIKey   string            `json:"api_key,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	CertFile           string `json:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty"`
	CAFile             string `json:"ca_file,omitempty"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Connections        map[string]*DatabaseConnection `json:"connections"`
	DefaultConn        string                         `json:"default_connection"`
	MaxOpenConns       int                            `json:"max_open_connections"`
	MaxIdleConns       int                            `json:"max_idle_connections"`
	ConnMaxLife        time.Duration                  `json:"connection_max_lifetime"`
	QueryTimeout       time.Duration                  `json:"query_timeout"`
	TransactionTimeout time.Duration                  `json:"transaction_timeout"`
}

// DatabaseConnection represents a single database connection configuration
type DatabaseConnection struct {
	Driver   string            `json:"driver"` // mysql, postgres, sqlite3
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Database string            `json:"database"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	SSLMode  string            `json:"ssl_mode,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// ReportConfig holds reporting configuration
type ReportConfig struct {
	Enabled     bool                   `json:"enabled"`
	OutputDir   string                 `json:"output_dir"`
	Formats     []string               `json:"formats"` // html, json, xml, junit
	Template    string                 `json:"template,omitempty"`
	IncludeLogs bool                   `json:"include_logs"`
	Compress    bool                   `json:"compress"`
	Reporters   map[string]interface{} `json:"reporters,omitempty"`
}

// MobileConfig holds mobile testing configuration
type MobileConfig struct {
	AppiumServer   *AppiumServerConfig `json:"appium_server"`
	DefaultDevice  *DeviceConfig       `json:"default_device"`
	Devices        []*DeviceConfig     `json:"devices"`
	AppPath        string              `json:"app_path"`
	Timeout        time.Duration       `json:"timeout"`
	ImplicitWait   time.Duration       `json:"implicit_wait"`
	ScreenshotPath string              `json:"screenshot_path"`
}

// AppiumServerConfig holds Appium server configuration
type AppiumServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// DeviceConfig holds device-specific configuration
type DeviceConfig struct {
	Name         string                 `json:"name"`
	PlatformName string                 `json:"platform_name"` // iOS, Android
	DeviceName   string                 `json:"device_name"`
	UDID         string                 `json:"udid,omitempty"`
	Version      string                 `json:"version"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BrowserConfig: &BrowserConfig{
			Browser:        "chrome",
			Headless:       true,
			WindowSize:     "1920x1080",
			Timeout:        30 * time.Second,
			ScreenshotPath: "./screenshots",
			MaxInstances:   5,
			ReuseInstances: true,
		},
		APIConfig: &APIConfig{
			Timeout:         30 * time.Second,
			RetryCount:      3,
			RetryDelay:      1 * time.Second,
			MaxConnections:  10,
			KeepAlive:       true,
			FollowRedirects: true,
		},
		DatabaseConfig: &DatabaseConfig{
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLife:        time.Hour,
			QueryTimeout:       30 * time.Second,
			TransactionTimeout: 60 * time.Second,
			Connections:        make(map[string]*DatabaseConnection),
		},
		ReportConfig: &ReportConfig{
			Enabled:     true,
			OutputDir:   "./reports",
			Formats:     []string{"html", "json"},
			IncludeLogs: true,
			Compress:    false,
		},
		MobileConfig: &MobileConfig{
			AppiumServer: &AppiumServerConfig{
				Host: "localhost",
				Port: 4723,
			},
			Timeout:        30 * time.Second,
			ImplicitWait:   10 * time.Second,
			ScreenshotPath: "./screenshots",
		},
		Parallel:   false,
		MaxWorkers: 4,
		Timeout:    5 * time.Minute,
	}
}
