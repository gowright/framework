---
title: Configuration Overview
description: Learn how to configure Gowright for your testing needs
---

This guide covers all configuration options available in the Gowright testing framework.

## Basic Configuration

### Default Configuration

```go
config := gowright.DefaultConfig()
framework := gowright.New(config)
```

### Custom Configuration

```go
config := &gowright.Config{
    LogLevel:   "info",
    Parallel:   true,
    MaxRetries: 3,
    BrowserConfig: &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    },
    APIConfig: &gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
    },
}

framework := gowright.New(config)
```

## Configuration Sources

### 1. JSON Configuration File

Create a `gowright-config.json` file:

```json
{
  "log_level": "info",
  "parallel": true,
  "max_retries": 3,
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "window_size": {
      "width": 1920,
      "height": 1080
    }
  },
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "10s",
    "headers": {
      "User-Agent": "Gowright-Test-Client"
    }
  }
}
```

Load the configuration:

```go
config, err := gowright.LoadConfigFromFile("gowright-config.json")
if err != nil {
    log.Fatal(err)
}

framework := gowright.New(config)
```

### 2. Environment Variables

Set configuration via environment variables:

```bash
export GOWRIGHT_LOG_LEVEL=debug
export GOWRIGHT_PARALLEL=true
export GOWRIGHT_API_BASE_URL=https://api.example.com
export GOWRIGHT_BROWSER_HEADLESS=false
```

Load environment configuration:

```go
config := gowright.LoadConfigFromEnv()
framework := gowright.New(config)
```

### 3. Programmatic Configuration

```go
config := &gowright.Config{
    LogLevel: "debug",
    Parallel: true,
    BrowserConfig: &gowright.BrowserConfig{
        Headless: false,
        Timeout:  60 * time.Second,
        WindowSize: &gowright.WindowSize{
            Width:  1920,
            Height: 1080,
        },
    },
}

framework := gowright.New(config)
```

## Configuration Sections

The main configuration is divided into several sections:

- **[Browser Configuration](/configuration/browser/)** - UI testing settings
- **[API Configuration](/configuration/api/)** - HTTP client settings  
- **[Database Configuration](/configuration/database/)** - Database connections
- **[Reporting Configuration](/configuration/reporting/)** - Test reporting options

## Environment-Specific Configuration

### Multi-Environment Setup

```go
func loadEnvironmentConfig() *gowright.Config {
    env := os.Getenv("TEST_ENV")
    if env == "" {
        env = "development"
    }
    
    configFile := fmt.Sprintf("config/%s.json", env)
    
    config, err := gowright.LoadConfigFromFile(configFile)
    if err != nil {
        log.Printf("Failed to load config file %s: %v", configFile, err)
        config = gowright.DefaultConfig()
    }
    
    // Override with environment variables
    gowright.MergeConfigFromEnv(config)
    
    return config
}
```

## Configuration Validation

### Built-in Validation

```go
config := &gowright.Config{
    // ... your configuration
}

if err := gowright.ValidateConfig(config); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

### Custom Validation

```go
func validateCustomConfig(config *gowright.Config) error {
    if config.APIConfig != nil {
        if config.APIConfig.BaseURL == "" {
            return fmt.Errorf("API base URL is required")
        }
        
        if !strings.HasPrefix(config.APIConfig.BaseURL, "https://") {
            return fmt.Errorf("API base URL must use HTTPS")
        }
    }
    
    return nil
}
```

## Next Steps

- Learn about specific configuration sections:
  - [Browser Configuration](/configuration/browser/)
  - [API Configuration](/configuration/api/)
  - [Database Configuration](/configuration/database/)
  - [Reporting Configuration](/configuration/reporting/)
- See [configuration examples](/examples/) for real-world usage