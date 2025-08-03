---
title: Installation
description: Detailed installation instructions for Gowright
---

This page provides comprehensive installation instructions for the Gowright testing framework.

## System Requirements

### Go Version
- **Go 1.22 or later** is required
- Check your Go version: `go version`

### Operating Systems
- **Linux** (Ubuntu 18.04+, CentOS 7+, etc.)
- **macOS** (10.14+)
- **Windows** (10+)

### Browser Requirements (for UI Testing)
- **Chrome** or **Chromium** browser
- Automatically downloaded by go-rod if not present

### Database Requirements (for Database Testing)
- **PostgreSQL** 12+
- **MySQL** 8.0+
- **SQLite** 3.x
- **Redis** 6.0+ (for caching tests)

## Installation Methods

### Method 1: Go Modules (Recommended)

Add Gowright to your Go project:

```bash
# Initialize Go module if not already done
go mod init your-project-name

# Add Gowright dependency
go get github/gowright/framework

# Install specific version (optional)
go get github/gowright/framework@v1.0.0
```

### Method 2: Direct Download

Download and install directly:

```bash
# Clone the repository
git clone https://github.com/your-org/gowright.git
cd gowright

# Install dependencies
go mod download

# Run tests to verify installation
go test ./...
```

## Database Drivers

Install database drivers for the databases you plan to test:

### PostgreSQL
```bash
go get github.com/lib/pq
```

### MySQL
```bash
go get github.com/go-sql-driver/mysql
```

### SQLite
```bash
go get github.com/mattn/go-sqlite3
```

### Redis
```bash
go get github.com/go-redis/redis/v8
```

## Browser Setup

### Automatic Setup (Recommended)
Gowright automatically downloads Chrome/Chromium when needed:

```go
config := &gowright.BrowserConfig{
    Headless: true, // Will auto-download browser if needed
}
```

### Manual Browser Installation

#### Ubuntu/Debian
```bash
# Chrome
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
sudo apt update
sudo apt install google-chrome-stable

# Or Chromium
sudo apt install chromium-browser
```

#### CentOS/RHEL/Fedora
```bash
# Chrome
sudo dnf install google-chrome-stable

# Or Chromium
sudo dnf install chromium
```

#### macOS
```bash
# Using Homebrew
brew install --cask google-chrome

# Or Chromium
brew install --cask chromium
```

#### Windows
Download and install from:
- [Google Chrome](https://www.google.com/chrome/)
- [Chromium](https://www.chromium.org/getting-involved/download-chromium)

## Verification

Create a test file to verify your installation:

```go
// main_test.go
package main

import (
    "testing"
    "time"
    
    "github/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestInstallation(t *testing.T) {
    // Test framework initialization
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    err := framework.Initialize()
    require.NoError(t, err, "Framework should initialize successfully")
    
    // Test API tester
    apiConfig := &gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
    }
    
    apiTester := gowright.NewAPITester(apiConfig)
    err = apiTester.Initialize(apiConfig)
    require.NoError(t, err, "API tester should initialize")
    defer apiTester.Cleanup()
    
    response, err := apiTester.Get("/get", nil)
    require.NoError(t, err, "API request should succeed")
    assert.Equal(t, 200, response.StatusCode)
    
    // Test UI tester (headless)
    browserConfig := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err = uiTester.Initialize(browserConfig)
    require.NoError(t, err, "UI tester should initialize")
    defer uiTester.Cleanup()
    
    err = uiTester.Navigate("https://example.com")
    require.NoError(t, err, "Navigation should succeed")
    
    t.Log("✅ Gowright installation verified successfully!")
}
```

Run the verification:

```bash
go test -v
```

## Troubleshooting

### Common Issues

#### "Chrome not found" Error
```bash
# Solution 1: Let Gowright auto-download
config.BrowserConfig.AutoDownload = true

# Solution 2: Install Chrome manually (see Browser Setup above)

# Solution 3: Specify Chrome path
config.BrowserConfig.ExecutablePath = "/path/to/chrome"
```

#### Database Connection Issues
```bash
# Check if database is running
sudo systemctl status postgresql  # PostgreSQL
sudo systemctl status mysql       # MySQL

# Test connection manually
psql -h localhost -U username -d database  # PostgreSQL
mysql -h localhost -u username -p database # MySQL
```

#### Go Module Issues
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Verify dependencies
go mod verify
```

#### Permission Issues (Linux/macOS)
```bash
# Fix Chrome sandbox issues
sudo sysctl kernel.unprivileged_userns_clone=1

# Or run with --no-sandbox flag
config.BrowserConfig.Arguments = []string{"--no-sandbox"}
```

### Getting Help

If you encounter issues:

1. **Check the logs**: Enable debug logging with `LogLevel: "debug"`
2. **Search existing issues**: [GitHub Issues](https://github.com/your-org/gowright/issues)
3. **Ask for help**: [GitHub Discussions](https://github.com/your-org/gowright/discussions)
4. **Contact support**: [support@gowright.dev](mailto:support@gowright.dev)

## Next Steps

After successful installation:

1. [Quick Start Guide](/getting-started/quick-start/) - Your first Gowright test
2. [Configuration](/configuration/) - Learn about configuration options
3. [Examples](/examples/) - See practical examples
4. [API Reference](/api/) - Detailed API documentation