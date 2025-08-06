# Configuration

Gowright provides flexible configuration options to customize the framework for your specific testing needs. Configuration can be provided through code, JSON files, or environment variables.

## Configuration Methods

### 1. Default Configuration

The simplest way to get started:

```go
framework := gowright.NewWithDefaults()
```

This creates a framework with sensible defaults for all modules.

### 2. Programmatic Configuration

Create configuration in code for maximum flexibility:

```go
config := &gowright.Config{
    LogLevel:   "info",
    Parallel:   true,
    MaxRetries: 3,
    
    BrowserConfig: &gowright.BrowserConfig{
        Headless:   true,
        Timeout:    30 * time.Second,
        WindowSize: &gowright.WindowSize{Width: 1920, Height: 1080},
    },
    
    APIConfig: &gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
        Headers: map[string]string{
            "User-Agent": "Gowright-Test-Client",
        },
    },
    
    DatabaseConfig: &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "main": {
                Driver: "postgres",
                DSN:    "postgres://user:pass@localhost/testdb?sslmode=disable",
            },
        },
    },
    
    AppiumConfig: &gowright.AppiumConfig{
        ServerURL: "http://localhost:4723",
        Timeout:   30 * time.Second,
    },
    
    ReportConfig: &gowright.ReportConfig{
        LocalReports: gowright.LocalReportConfig{
            JSON:      true,
            HTML:      true,
            OutputDir: "./test-reports",
        },
    },
}

framework := gowright.New(config)
```

### 3. Configuration from File

Load configuration from JSON files:

```go
config, err := gowright.LoadConfigFromFile("gowright-config.json")
if err != nil {
    panic(err)
}

framework := gowright.New(config)
```

### 4. Environment Variables

Override configuration with environment variables:

```bash
export GOWRIGHT_LOG_LEVEL=debug
export GOWRIGHT_BROWSER_HEADLESS=false
export GOWRIGHT_API_BASE_URL=https://staging.api.com
export GOWRIGHT_DB_DSN=postgres://user:pass@localhost/testdb
```

## Configuration Structure

### Core Configuration

```go
type Config struct {
    LogLevel            string                   `json:"log_level"`
    Parallel            bool                     `json:"parallel"`
    MaxRetries          int                      `json:"max_retries"`
    BrowserConfig       *BrowserConfig           `json:"browser_config"`
    APIConfig           *APIConfig               `json:"api_config"`
    DatabaseConfig      *DatabaseConfig          `json:"database_config"`
    AppiumConfig        *AppiumConfig            `json:"appium_config"`
    ReportConfig        *ReportConfig            `json:"report_config"`
    ParallelRunnerConfig *ParallelRunnerConfig   `json:"parallel_runner_config"`
}
```

### Browser Configuration

```go
type BrowserConfig struct {
    Headless    bool          `json:"headless"`
    Timeout     time.Duration `json:"timeout"`
    UserAgent   string        `json:"user_agent"`
    WindowSize  *WindowSize   `json:"window_size"`
    ChromePath  string        `json:"chrome_path"`
    ChromeArgs  []string      `json:"chrome_args"`
    Extensions  []string      `json:"extensions"`
}

type WindowSize struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}
```

**Example:**
```json
{
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "user_agent": "Gowright-Tester/1.0",
    "window_size": {
      "width": 1920,
      "height": 1080
    },
    "chrome_path": "/usr/bin/google-chrome",
    "chrome_args": [
      "--no-sandbox",
      "--disable-dev-shm-usage"
    ]
  }
}
```

### API Configuration

```go
type APIConfig struct {
    BaseURL     string            `json:"base_url"`
    Timeout     time.Duration     `json:"timeout"`
    Headers     map[string]string `json:"headers"`
    AuthConfig  *AuthConfig       `json:"auth_config"`
    RetryConfig *RetryConfig      `json:"retry_config"`
}

type AuthConfig struct {
    Type     string `json:"type"`     // "basic", "bearer", "api_key"
    Username string `json:"username"`
    Password string `json:"password"`
    Token    string `json:"token"`
    APIKey   string `json:"api_key"`
    Header   string `json:"header"`   // Custom auth header name
}
```

**Example:**
```json
{
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "30s",
    "headers": {
      "User-Agent": "Gowright-API-Tester/1.0",
      "Content-Type": "application/json"
    },
    "auth_config": {
      "type": "bearer",
      "token": "${API_TOKEN}"
    }
  }
}
```

### Database Configuration

```go
type DatabaseConfig struct {
    Connections map[string]*DBConnection `json:"connections"`
}

type DBConnection struct {
    Driver       string `json:"driver"`
    DSN          string `json:"dsn"`
    MaxOpenConns int    `json:"max_open_conns"`
    MaxIdleConns int    `json:"max_idle_conns"`
    MaxLifetime  string `json:"max_lifetime"`
}
```

**Example:**
```json
{
  "database_config": {
    "connections": {
      "primary": {
        "driver": "postgres",
        "dsn": "postgres://user:pass@localhost/testdb?sslmode=disable",
        "max_open_conns": 10,
        "max_idle_conns": 5,
        "max_lifetime": "1h"
      },
      "secondary": {
        "driver": "mysql",
        "dsn": "user:pass@tcp(localhost:3306)/testdb",
        "max_open_conns": 5,
        "max_idle_conns": 2
      }
    }
  }
}
```

### Appium Configuration

```go
type AppiumConfig struct {
    ServerURL string        `json:"server_url"`
    Timeout   time.Duration `json:"timeout"`
}
```

**Example:**
```json
{
  "appium_config": {
    "server_url": "http://localhost:4723",
    "timeout": "30s"
  }
}
```

**Environment Variables:**
- `GOWRIGHT_APPIUM_SERVER_URL` - Appium server URL
- `GOWRIGHT_APPIUM_TIMEOUT` - Request timeout

### Report Configuration

```go
type ReportConfig struct {
    LocalReports  LocalReportConfig  `json:"local_reports"`
    RemoteReports RemoteReportConfig `json:"remote_reports"`
}

type LocalReportConfig struct {
    JSON      bool   `json:"json"`
    HTML      bool   `json:"html"`
    OutputDir string `json:"output_dir"`
}

type RemoteReportConfig struct {
    JiraXray     *JiraXrayConfig     `json:"jira_xray"`
    AIOTest      *AIOTestConfig      `json:"aio_test"`
    ReportPortal *ReportPortalConfig `json:"report_portal"`
}
```

**Example:**
```json
{
  "report_config": {
    "local_reports": {
      "json": true,
      "html": true,
      "output_dir": "./test-reports"
    },
    "remote_reports": {
      "jira_xray": {
        "url": "https://your-jira.atlassian.net",
        "username": "${JIRA_USERNAME}",
        "password": "${JIRA_API_TOKEN}",
        "project_key": "TEST"
      }
    }
  }
}
```

### Parallel Execution Configuration

```go
type ParallelRunnerConfig struct {
    MaxConcurrency int            `json:"max_concurrency"`
    ResourceLimits ResourceLimits `json:"resource_limits"`
}

type ResourceLimits struct {
    MaxMemoryMB     int `json:"max_memory_mb"`
    MaxCPUPercent   int `json:"max_cpu_percent"`
    MaxOpenFiles    int `json:"max_open_files"`
    MaxNetworkConns int `json:"max_network_conns"`
}
```

**Example:**
```json
{
  "parallel_runner_config": {
    "max_concurrency": 4,
    "resource_limits": {
      "max_memory_mb": 1024,
      "max_cpu_percent": 80,
      "max_open_files": 100,
      "max_network_conns": 50
    }
  }
}
```

## Environment-Specific Configuration

### Development Configuration

```json title="gowright-config.dev.json"
{
  "log_level": "debug",
  "parallel": false,
  "browser_config": {
    "headless": false,
    "timeout": "60s"
  },
  "api_config": {
    "base_url": "http://localhost:8080",
    "timeout": "30s"
  },
  "database_config": {
    "connections": {
      "main": {
        "driver": "sqlite3",
        "dsn": "./test.db"
      }
    }
  }
}
```

### CI/CD Configuration

```json title="gowright-config.ci.json"
{
  "log_level": "warn",
  "parallel": true,
  "max_retries": 1,
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "chrome_args": [
      "--no-sandbox",
      "--disable-dev-shm-usage",
      "--disable-gpu"
    ]
  },
  "parallel_runner_config": {
    "max_concurrency": 2,
    "resource_limits": {
      "max_memory_mb": 512,
      "max_cpu_percent": 70
    }
  }
}
```

### Production Configuration

```json title="gowright-config.prod.json"
{
  "log_level": "error",
  "parallel": true,
  "max_retries": 3,
  "browser_config": {
    "headless": true,
    "timeout": "45s"
  },
  "api_config": {
    "base_url": "https://api.production.com",
    "timeout": "20s"
  },
  "report_config": {
    "remote_reports": {
      "jira_xray": {
        "url": "${JIRA_URL}",
        "username": "${JIRA_USERNAME}",
        "password": "${JIRA_API_TOKEN}",
        "project_key": "PROD"
      }
    }
  }
}
```

## Environment Variables

### Core Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOWRIGHT_LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `GOWRIGHT_PARALLEL` | Enable parallel execution | `false` |
| `GOWRIGHT_MAX_RETRIES` | Maximum retry attempts | `3` |

### Browser Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOWRIGHT_BROWSER_HEADLESS` | Run browser in headless mode | `true` |
| `GOWRIGHT_BROWSER_TIMEOUT` | Browser operation timeout | `30s` |
| `GOWRIGHT_CHROME_PATH` | Path to Chrome executable | Auto-detected |

### API Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOWRIGHT_API_BASE_URL` | Base URL for API requests | - |
| `GOWRIGHT_API_TIMEOUT` | API request timeout | `10s` |
| `GOWRIGHT_API_TOKEN` | Bearer token for authentication | - |

### Database Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOWRIGHT_DB_DRIVER` | Database driver | `sqlite3` |
| `GOWRIGHT_DB_DSN` | Database connection string | `:memory:` |
| `GOWRIGHT_DB_MAX_CONNECTIONS` | Maximum database connections | `10` |

## Configuration Loading Priority

Gowright loads configuration in the following order (later sources override earlier ones):

1. **Default values** - Built-in defaults
2. **Configuration file** - JSON configuration file
3. **Environment variables** - OS environment variables
4. **Programmatic overrides** - Values set in code

### Example Loading Strategy

```go
func loadConfiguration() *gowright.Config {
    // Start with defaults
    config := gowright.DefaultConfig()
    
    // Load from file if exists
    if configFile := os.Getenv("GOWRIGHT_CONFIG_FILE"); configFile != "" {
        if fileConfig, err := gowright.LoadConfigFromFile(configFile); err == nil {
            config = gowright.MergeConfigs(config, fileConfig)
        }
    }
    
    // Apply environment variable overrides
    config = gowright.ApplyEnvironmentOverrides(config)
    
    // Apply any programmatic overrides
    if os.Getenv("CI") == "true" {
        config.Parallel = true
        config.BrowserConfig.Headless = true
    }
    
    return config
}
```

## Configuration Validation

Gowright validates configuration at startup:

```go
func validateConfig(config *gowright.Config) error {
    if config.BrowserConfig != nil {
        if config.BrowserConfig.Timeout <= 0 {
            return errors.New("browser timeout must be positive")
        }
    }
    
    if config.APIConfig != nil {
        if config.APIConfig.BaseURL == "" {
            return errors.New("API base URL is required")
        }
    }
    
    // Additional validation...
    return nil
}
```

## Configuration Templates

### Minimal Configuration

```json title="minimal-config.json"
{
  "log_level": "info"
}
```

### API-Only Configuration

```json title="api-only-config.json"
{
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "30s"
  },
  "report_config": {
    "local_reports": {
      "json": true,
      "output_dir": "./api-reports"
    }
  }
}
```

### UI-Only Configuration

```json title="ui-only-config.json"
{
  "browser_config": {
    "headless": false,
    "timeout": "60s",
    "window_size": {
      "width": 1280,
      "height": 720
    }
  },
  "report_config": {
    "local_reports": {
      "html": true,
      "output_dir": "./ui-reports"
    }
  }
}
```

### Complete Configuration

```json title="complete-config.json"
{
  "log_level": "info",
  "parallel": true,
  "max_retries": 3,
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "user_agent": "Gowright-Tester/1.0",
    "window_size": {
      "width": 1920,
      "height": 1080
    },
    "chrome_args": [
      "--no-sandbox",
      "--disable-dev-shm-usage"
    ]
  },
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "30s",
    "headers": {
      "User-Agent": "Gowright-API-Tester/1.0",
      "Content-Type": "application/json"
    },
    "auth_config": {
      "type": "bearer",
      "token": "${API_TOKEN}"
    }
  },
  "database_config": {
    "connections": {
      "primary": {
        "driver": "postgres",
        "dsn": "${DATABASE_URL}",
        "max_open_conns": 10,
        "max_idle_conns": 5
      }
    }
  },
  "report_config": {
    "local_reports": {
      "json": true,
      "html": true,
      "output_dir": "./test-reports"
    },
    "remote_reports": {
      "jira_xray": {
        "url": "${JIRA_URL}",
        "username": "${JIRA_USERNAME}",
        "password": "${JIRA_API_TOKEN}",
        "project_key": "TEST"
      }
    }
  },
  "parallel_runner_config": {
    "max_concurrency": 4,
    "resource_limits": {
      "max_memory_mb": 1024,
      "max_cpu_percent": 80,
      "max_open_files": 100,
      "max_network_conns": 50
    }
  }
}
```

## Best Practices

### 1. Use Environment-Specific Configs

Create separate configuration files for different environments:

```bash
configs/
├── gowright-config.dev.json
├── gowright-config.staging.json
├── gowright-config.prod.json
└── gowright-config.ci.json
```

### 2. Secure Sensitive Data

Never commit sensitive data to version control:

```json
{
  "api_config": {
    "auth_config": {
      "token": "${API_TOKEN}"  // Use environment variables
    }
  },
  "database_config": {
    "connections": {
      "main": {
        "dsn": "${DATABASE_URL}"  // Use environment variables
      }
    }
  }
}
```

### 3. Validate Configuration

Always validate configuration before use:

```go
config, err := gowright.LoadConfigFromFile("config.json")
if err != nil {
    log.Fatal("Failed to load config:", err)
}

if err := gowright.ValidateConfig(config); err != nil {
    log.Fatal("Invalid config:", err)
}
```

### 4. Use Reasonable Defaults

Set sensible defaults for your testing environment:

```go
config := gowright.DefaultConfig()
config.BrowserConfig.Timeout = 45 * time.Second  // Longer timeout for slow tests
config.APIConfig.Timeout = 20 * time.Second      // Reasonable API timeout
config.MaxRetries = 2                            // Retry failed tests once
```

## Troubleshooting Configuration

### Common Issues

**Configuration file not found:**
```go
// Check if file exists before loading
if _, err := os.Stat("config.json"); os.IsNotExist(err) {
    log.Println("Config file not found, using defaults")
    config = gowright.DefaultConfig()
} else {
    config, err = gowright.LoadConfigFromFile("config.json")
}
```

**Invalid JSON format:**
```bash
# Validate JSON syntax
python -m json.tool gowright-config.json
```

**Environment variable not set:**
```go
// Provide fallback values
baseURL := os.Getenv("API_BASE_URL")
if baseURL == "" {
    baseURL = "http://localhost:8080"  // Fallback
}
```

### Debug Configuration

Enable debug logging to see configuration loading:

```go
config := &gowright.Config{
    LogLevel: "debug",
}

// This will log configuration loading details
framework := gowright.New(config)
```

## Next Steps

- [Quick Start](quick-start.md) - Create your first test with configuration
- [API Testing](../testing-modules/api-testing.md) - Configure API testing
- [UI Testing](../testing-modules/ui-testing.md) - Configure browser automation
- [Database Testing](../testing-modules/database-testing.md) - Configure database connections
- [Best Practices](../reference/best-practices.md) - Configuration best practices