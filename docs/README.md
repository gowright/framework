# Gowright Testing Framework

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://github.com/gowright/framework/blob/main/LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()
[![Documentation](https://img.shields.io/badge/docs-Docsify-blue)](https://gowright.github.io/framework/)

Gowright is a comprehensive testing framework for Go that provides unified testing capabilities across UI (browser, mobile), API, database, and integration testing scenarios. Built with a focus on simplicity, performance, and extensibility.

## ✨ Key Features

<div class="grid cards">

<div>

### 🌐 UI Testing
Browser automation using Chrome DevTools Protocol via [go-rod/rod](https://github.com/go-rod/rod)

</div>

<div>

### 🚀 API Testing
HTTP/REST API testing with [go-resty/resty](https://github.com/go-resty/resty/v2)

</div>

<div>

### 💾 Database Testing
Multi-database support with transaction management

</div>

<div>

### 🔗 Integration Testing
Complex workflows spanning multiple systems

</div>

<div>

### 📊 Flexible Reporting
Local (JSON, HTML) and remote reporting (Jira Xray, AIOTest, Report Portal)

</div>

<div>

### 🧪 Testify Integration
Compatible with [stretchr/testify](https://github.com/stretchr/testify)

</div>

<div>

### ⚡ Parallel Execution
Concurrent test execution with resource management

</div>

<div>

### 🛡️ Error Recovery
Graceful error handling and retry mechanisms

</div>

</div>

## 🚀 Quick Start

### Installation

```bash
go get github.com/gowright/framework
```

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
)

func main() {
    // Create framework with default configuration
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Initialize the framework
    if err := framework.Initialize(); err != nil {
        panic(err)
    }
    
    fmt.Println("Gowright framework initialized successfully!")
}
```

## 📖 Documentation Sections

### Getting Started
- [Introduction](getting-started/introduction.md) - Overview and core concepts
- [Installation](getting-started/installation.md) - Setup and installation guide
- [Quick Start](getting-started/quick-start.md) - Get up and running quickly
- [Configuration](getting-started/configuration.md) - Configure Gowright for your needs

### Testing Modules
- [API Testing](testing-modules/api-testing.md) - REST API testing with validation
- [UI Testing](testing-modules/ui-testing.md) - Browser automation and UI testing
- [Database Testing](testing-modules/database-testing.md) - Database operations and validation
- [Integration Testing](testing-modules/integration-testing.md) - Multi-system workflows

### Advanced Features
- [Test Suites](advanced/test-suites.md) - Organizing and running test collections
- [Assertions](advanced/assertions.md) - Custom assertion system
- [Reporting](advanced/reporting.md) - Professional HTML and JSON reports
- [Parallel Execution](advanced/parallel-execution.md) - Concurrent test execution
- [Resource Management](advanced/resource-management.md) - Memory and CPU monitoring

### Examples
- [Basic Usage](examples/basic-usage.md) - Framework initialization examples
- [API Testing Examples](examples/api-testing.md) - Comprehensive API testing scenarios
- [UI Testing Examples](examples/ui-testing.md) - Browser automation examples
- [Database Examples](examples/database-testing.md) - Database testing patterns
- [Integration Examples](examples/integration-testing.md) - End-to-end workflows

### Reference
- [API Reference](reference/api.md) - Complete API documentation
- [Configuration Reference](reference/configuration.md) - All configuration options
- [Best Practices](reference/best-practices.md) - Recommended patterns
- [Troubleshooting](reference/troubleshooting.md) - Common issues and solutions
- [Migration Guide](reference/migration.md) - Migrating from other frameworks

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](contributing/guide.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/gowright/framework/blob/main/LICENSE) file for details.

## 🙏 Acknowledgments

- [go-rod/rod](https://github.com/go-rod/rod) for browser automation
- [go-resty/resty](https://github.com/go-resty/resty) for HTTP client
- [stretchr/testify](https://github.com/stretchr/testify) for testing utilities

## 📞 Support

- 📖 [Documentation](https://gowright.github.io/framework/)
- 🐛 [Issue Tracker](https://github.com/gowright/framework/issues)
- 💬 [Discussions](https://github.com/gowright/framework/discussions)
- 📧 [Email Support](mailto:support@gowright.dev)

---

**Gowright** - Making Go testing comprehensive and enjoyable! 🚀