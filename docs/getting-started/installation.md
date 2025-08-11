# Installation

This guide covers installing Gowright and setting up your development environment for testing.

## Prerequisites

### Go Version
Gowright requires **Go 1.22 or later**. Check your Go version:

```bash
go version
```

If you need to update Go, visit the [official Go installation guide](https://golang.org/doc/install).

### System Dependencies

#### For UI Testing (Optional)
If you plan to use UI testing features, you'll need Chrome or Chromium:

=== "Ubuntu/Debian"
    ```bash
    # Install Chrome
    wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
    echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
    sudo apt update
    sudo apt install google-chrome-stable
    
    # Or install Chromium
    sudo apt install chromium-browser
    ```

=== "CentOS/RHEL/Fedora"
    ```bash
    # Install Chrome
    sudo dnf install -y google-chrome-stable
    
    # Or install Chromium
    sudo dnf install -y chromium
    ```

=== "macOS"
    ```bash
    # Using Homebrew
    brew install --cask google-chrome
    
    # Or install Chromium
    brew install --cask chromium
    ```

=== "Windows"
    Download and install Chrome from [https://www.google.com/chrome/](https://www.google.com/chrome/)

#### For Mobile Testing (Optional)
If you plan to use mobile testing features, you'll need Appium server and mobile development tools:

=== "Appium Server"
    ```bash
    # Install Node.js (required for Appium)
    # Ubuntu/Debian
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
    
    # macOS
    brew install node
    
    # Windows - Download from https://nodejs.org/
    
    # Install Appium globally
    npm install -g appium
    
    # Install Appium drivers
    appium driver install uiautomator2  # For Android
    appium driver install xcuitest      # For iOS (macOS only)
    
    # Start Appium server
    appium
    ```

=== "Android Setup"
    ```bash
    # Install Android SDK
    # Ubuntu/Debian
    sudo apt install android-sdk
    
    # macOS
    brew install --cask android-studio
    
    # Set environment variables
    export ANDROID_HOME=$HOME/Android/Sdk
    export PATH=$PATH:$ANDROID_HOME/emulator
    export PATH=$PATH:$ANDROID_HOME/tools
    export PATH=$PATH:$ANDROID_HOME/tools/bin
    export PATH=$PATH:$ANDROID_HOME/platform-tools
    
    # Create Android Virtual Device (AVD)
    avdmanager create avd -n test_device -k "system-images;android-30;google_apis;x86_64"
    
    # Start emulator
    emulator -avd test_device
    ```

=== "iOS Setup (macOS only)"
    ```bash
    # Install Xcode from App Store
    
    # Install Xcode command line tools
    xcode-select --install
    
    # Install iOS Simulator
    # This is included with Xcode
    
    # List available simulators
    xcrun simctl list devices
    
    # Start iOS Simulator
    open -a Simulator
    ```

#### For Database Testing (Optional)
Database drivers are included, but you may need database servers for testing:

=== "PostgreSQL"
    ```bash
    # Ubuntu/Debian
    sudo apt install postgresql postgresql-contrib
    
    # macOS
    brew install postgresql
    
    # Start service
    sudo systemctl start postgresql  # Linux
    brew services start postgresql   # macOS
    ```

=== "MySQL"
    ```bash
    # Ubuntu/Debian
    sudo apt install mysql-server
    
    # macOS
    brew install mysql
    
    # Start service
    sudo systemctl start mysql       # Linux
    brew services start mysql        # macOS
    ```

=== "SQLite"
    SQLite support is built-in, no additional installation required.

## Installing Gowright

### Using Go Modules (Recommended)

Add Gowright to your Go project:

```bash
# Initialize Go module (if not already done)
go mod init your-project-name

# Add Gowright dependency
go get github.com/gowright/framework
```

### Verify Installation

Create a simple test file to verify the installation:

```go title="test_installation.go"
package main

import (
    "fmt"
    "log"
    
    "github.com/gowright/framework/pkg/gowright"
)

func main() {
    // Create framework with default configuration
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Initialize the framework
    if err := framework.Initialize(); err != nil {
        log.Fatal("Failed to initialize Gowright:", err)
    }
    
    fmt.Println("✅ Gowright installed and initialized successfully!")
    
    // Print version information
    fmt.Printf("Framework version: %s\n", framework.Version())
}
```

Run the test:

```bash
go run test_installation.go
```

Expected output:
```
✅ Gowright installed and initialized successfully!
Framework version: v1.0.0
```

## Project Structure

Here's a recommended project structure for Gowright tests:

```
your-project/
├── go.mod
├── go.sum
├── main.go
├── tests/
│   ├── api/
│   │   ├── user_api_test.go
│   │   └── product_api_test.go
│   ├── ui/
│   │   ├── login_ui_test.go
│   │   └── checkout_ui_test.go
│   ├── database/
│   │   ├── migration_test.go
│   │   └── integrity_test.go
│   ├── integration/
│   │   └── e2e_workflow_test.go
│   └── config/
│       ├── gowright-config.json
│       └── test-data.json
├── reports/
│   ├── json/
│   └── html/
└── scripts/
    ├── setup-test-db.sh
    └── run-tests.sh
```

## Configuration Setup

### Basic Configuration File

Create a configuration file for your tests:

```json title="gowright-config.json"
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
      "User-Agent": "Gowright-Test-Client/1.0"
    }
  },
  "database_config": {
    "connections": {
      "test": {
        "driver": "sqlite3",
        "dsn": ":memory:",
        "max_open_conns": 10,
        "max_idle_conns": 5
      }
    }
  },
  "report_config": {
    "local_reports": {
      "json": true,
      "html": true,
      "output_dir": "./reports"
    }
  }
}
```

### Environment-Specific Configuration

Create different configuration files for different environments:

=== "Development"
    ```json title="gowright-config.dev.json"
    {
      "log_level": "debug",
      "browser_config": {
        "headless": false,
        "timeout": "60s"
      },
      "api_config": {
        "base_url": "http://localhost:8080"
      }
    }
    ```

=== "CI/CD"
    ```json title="gowright-config.ci.json"
    {
      "log_level": "warn",
      "browser_config": {
        "headless": true,
        "timeout": "30s"
      },
      "parallel": true,
      "max_retries": 1
    }
    ```

=== "Production"
    ```json title="gowright-config.prod.json"
    {
      "log_level": "error",
      "browser_config": {
        "headless": true,
        "timeout": "45s"
      },
      "api_config": {
        "base_url": "https://api.production.com"
      }
    }
    ```

## IDE Setup

### VS Code

Install recommended extensions for better development experience:

```json title=".vscode/extensions.json"
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-vscode.test-adapter-converter"
  ]
}
```

Configure VS Code settings:

```json title=".vscode/settings.json"
{
  "go.testFlags": ["-v"],
  "go.testTimeout": "300s",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter"
  }
}
```

### GoLand/IntelliJ

1. Install the Go plugin
2. Configure test runner settings:
   - Go to **Settings** → **Go** → **Build Tags & Vendoring**
   - Add build tags if needed: `integration,ui,database`

## Docker Setup (Optional)

For containerized testing environments:

```dockerfile title="Dockerfile.test"
FROM golang:1.22-alpine AS builder

# Install Chrome for UI testing
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont

# Set Chrome path
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o test-runner ./tests/

FROM alpine:latest
RUN apk --no-cache add ca-certificates chromium
WORKDIR /root/

COPY --from=builder /app/test-runner .
COPY --from=builder /app/gowright-config.json .

CMD ["./test-runner"]
```

Build and run:

```bash
# Build test image
docker build -f Dockerfile.test -t gowright-tests .

# Run tests
docker run --rm gowright-tests
```

## Troubleshooting Installation

### Common Issues

#### Go Version Compatibility
```bash
# Error: "go: module requires Go 1.22"
# Solution: Update Go version
go version  # Check current version
# Update Go from https://golang.org/dl/
```

#### Chrome/Chromium Not Found
```bash
# Error: "chrome executable not found"
# Solution: Install Chrome or set CHROME_BIN environment variable
export CHROME_BIN=/usr/bin/google-chrome
# Or
export CHROME_BIN=/usr/bin/chromium-browser
```

#### Database Driver Issues
```bash
# Error: "sql: unknown driver"
# Solution: Import the database driver
import _ "github.com/lib/pq"           // PostgreSQL
import _ "github.com/go-sql-driver/mysql" // MySQL
import _ "github.com/mattn/go-sqlite3"    // SQLite
```

#### Permission Issues
```bash
# Error: "permission denied" when creating reports
# Solution: Ensure write permissions for report directory
chmod 755 ./reports
mkdir -p ./reports/json ./reports/html
```

### Verification Script

Create a comprehensive verification script:

```go title="verify_setup.go"
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
)

func main() {
    fmt.Println("🔍 Verifying Gowright installation...")
    
    // Test 1: Framework initialization
    fmt.Print("✓ Testing framework initialization... ")
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    if err := framework.Initialize(); err != nil {
        log.Fatal("❌ Failed:", err)
    }
    fmt.Println("✅ OK")
    
    // Test 2: Configuration loading
    fmt.Print("✓ Testing configuration loading... ")
    if _, err := os.Stat("gowright-config.json"); err == nil {
        config, err := gowright.LoadConfigFromFile("gowright-config.json")
        if err != nil {
            log.Fatal("❌ Failed:", err)
        }
        _ = config
    }
    fmt.Println("✅ OK")
    
    // Test 3: UI testing capability
    fmt.Print("✓ Testing UI testing capability... ")
    uiTester := gowright.NewRodUITester()
    browserConfig := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  10 * time.Second,
    }
    if err := uiTester.Initialize(browserConfig); err != nil {
        fmt.Printf("⚠️  Warning: UI testing not available: %v\n", err)
    } else {
        uiTester.Cleanup()
        fmt.Println("✅ OK")
    }
    
    // Test 4: API testing capability
    fmt.Print("✓ Testing API testing capability... ")
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 5 * time.Second,
    })
    if err := apiTester.Initialize(nil); err != nil {
        log.Fatal("❌ Failed:", err)
    }
    apiTester.Cleanup()
    fmt.Println("✅ OK")
    
    // Test 5: Database testing capability
    fmt.Print("✓ Testing database testing capability... ")
    dbTester := gowright.NewDatabaseTester()
    dbConfig := &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "test": {
                Driver: "sqlite3",
                DSN:    ":memory:",
            },
        },
    }
    if err := dbTester.Initialize(dbConfig); err != nil {
        log.Fatal("❌ Failed:", err)
    }
    dbTester.Cleanup()
    fmt.Println("✅ OK")
    
    fmt.Println("\n🎉 All verification tests passed!")
    fmt.Println("Gowright is ready to use!")
}
```

Run the verification:

```bash
go run verify_setup.go
```

## Next Steps

Now that Gowright is installed and configured:

1. **[Quick Start](quick-start.md)** - Create your first test
2. **[Configuration](configuration.md)** - Learn about advanced configuration options
3. **[API Testing](../testing-modules/api-testing.md)** - Start with API testing
4. **[Examples](../examples/basic-usage.md)** - Explore practical examples

## Getting Help

If you encounter issues during installation:

1. Check the [Troubleshooting Guide](../reference/troubleshooting.md)
2. Search [GitHub Issues](https://github.com/gowright/framework/issues)
3. Ask questions in [GitHub Discussions](https://github.com/gowright/framework/discussions)
4. Contact support at [support@gowright.dev](mailto:support@gowright.dev)