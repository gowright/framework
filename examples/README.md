# Gowright Testing Framework Examples

This directory contains comprehensive examples demonstrating the full capabilities of the Gowright testing framework. Each example showcases different aspects of the framework and can be run independently or used as templates for your own tests.

## Overview

The Gowright framework provides unified testing capabilities across:
- **UI Testing**: Browser automation using go-rod/rod
- **API Testing**: HTTP/REST endpoint testing using go-resty/resty  
- **Database Testing**: Database operations and validations
- **Integration Testing**: Multi-system workflow orchestration
- **Reporting**: Comprehensive test reporting in multiple formats

## Example Files

### 1. Basic Usage (`basic_usage.go`)

**Purpose**: Demonstrates framework initialization and basic configuration.

**Key Features**:
- Framework initialization with default configuration
- Configuration loading from environment variables
- Configuration file management
- Basic test suite setup

**Run**: `go run examples/basic_usage.go`

### 2. UI Testing (`ui_testing_example.go`)

**Purpose**: Comprehensive browser automation testing examples.

**Key Features**:
- Basic navigation and form interaction
- Element visibility and interaction testing
- Screenshot capture and page validation
- Dynamic content and AJAX testing
- Multi-step user workflows (login/logout)
- Error handling and negative testing

**Test Scenarios**:
- Form submission with validation
- Dropdown selection and interaction
- Screenshot capture for visual validation
- AJAX loading and dynamic content
- Complete user authentication workflow
- Error scenario testing with invalid credentials

**Run**: `go run examples/ui_testing_example.go`

### 3. API Testing (`api_testing_example.go`)

**Purpose**: REST API endpoint testing with various HTTP methods and validations.

**Key Features**:
- GET, POST, PUT, DELETE request testing
- JSON response validation with JSONPath
- Header validation and custom headers
- Error scenario testing (404, validation errors)
- Request/response logging and debugging

**Test Scenarios**:
- Simple GET request with response validation
- POST request with JSON body and creation validation
- Bulk data retrieval and validation
- Error handling for non-existent resources
- PUT request for data updates
- DELETE request testing

**Run**: `go run examples/api_testing_example.go`

### 4. Database Testing (`database_testing_example.go`)

**Purpose**: Comprehensive database testing including transactions, constraints, and performance.

**Key Features**:
- Table creation and data insertion
- Transaction management with rollback testing
- Complex JOIN queries and data validation
- Data integrity and constraint testing
- Performance testing with large datasets
- Multi-database operations
- Schema migration testing

**Test Scenarios**:
- Database schema setup and data insertion
- Transaction rollback verification
- Complex JOIN queries with custom assertions
- Constraint violation testing (UNIQUE, FOREIGN KEY)
- Performance testing with 1000+ records
- Multi-database audit logging
- Schema migration with column additions

**Run**: `go run examples/database_testing_example.go`

### 5. Integration Testing (`integration_testing_example.go`)

**Purpose**: Multi-system workflow testing combining UI, API, and database operations.

**Key Features**:
- E-commerce workflow orchestration
- User registration and profile management
- Data synchronization across systems
- Error handling and rollback mechanisms
- Performance testing across multiple systems

**Test Scenarios**:
- Complete e-commerce purchase workflow (DB → API → UI → DB)
- User registration flow with API creation and UI verification
- Data synchronization between source and destination systems
- Error handling with automatic rollback on failures
- Performance testing with bulk operations across systems

**Run**: `go run examples/integration_testing_example.go`

### 6. Test Suite with Assertions (`test_suite_with_assertions.go`)

**Purpose**: Demonstrates the assertion system and comprehensive test reporting.

**Key Features**:
- Comprehensive assertion methods (Equal, NotNil, Contains, etc.)
- Test suite execution with setup/teardown
- Detailed logging and step tracking
- Professional HTML and JSON report generation
- Error handling and failure reporting

**Test Scenarios**:
- User authentication validation
- Data structure validation
- Error response handling
- Performance metrics validation
- Integration testing with multiple systems

**Run**: `go run examples/test_suite_with_assertions.go`

### 7. Assertion Reporting (`assertion_reporting_example.go`)

**Purpose**: Focused example of the assertion system with detailed reporting.

**Key Features**:
- Individual test execution with assertions
- Step-by-step assertion tracking
- Detailed error reporting and logging
- HTML and JSON report generation

**Run**: `go run examples/assertion_reporting_example.go`

### 8. Reporting Example (`reporting_example.go`)

**Purpose**: Demonstrates the reporting system capabilities.

**Key Features**:
- JSON and HTML report generation
- Test result aggregation and summary
- Screenshot and log attachment
- Multiple report format support

**Run**: `go run examples/reporting_example.go`

## Running the Examples

### Prerequisites

1. **Go 1.19+**: Ensure you have Go installed
2. **Dependencies**: Run `go mod tidy` to install required dependencies
3. **Browser**: For UI testing examples, Chrome/Chromium should be installed
4. **Database**: SQLite is used for database examples (no additional setup required)

### Individual Examples

Run any example individually:

```bash
# Basic framework usage
go run examples/basic_usage.go

# UI testing with browser automation
go run examples/ui_testing_example.go

# API testing with REST endpoints
go run examples/api_testing_example.go

# Database testing with transactions
go run examples/database_testing_example.go

# Integration testing across systems
go run examples/integration_testing_example.go

# Assertion system demonstration
go run examples/test_suite_with_assertions.go
```

### All Examples

Run all examples in sequence:

```bash
# Create a simple script to run all examples
for example in basic_usage ui_testing_example api_testing_example database_testing_example integration_testing_example test_suite_with_assertions; do
    echo "Running $example..."
    go run examples/${example}.go
    echo "Completed $example"
    echo "---"
done
```

## Generated Reports

Each example generates reports in its respective directory:

- `./ui-test-reports/` - UI testing reports
- `./api-test-reports/` - API testing reports  
- `./database-test-reports/` - Database testing reports
- `./integration-test-reports/` - Integration testing reports
- `./comprehensive-reports/` - Comprehensive test suite reports
- `./assertion-reports/` - Assertion-focused reports
- `./example-reports/` - Basic reporting examples

### Report Formats

**JSON Reports**: Machine-readable test results with detailed metadata
- Test execution times
- Pass/fail status for each test
- Error messages and stack traces
- Log entries and debug information

**HTML Reports**: Human-readable reports with professional styling
- Visual test result dashboard
- Detailed test step breakdown
- Screenshot attachments (for UI tests)
- Interactive filtering and sorting

## Configuration Examples

### Browser Configuration

```go
config := &gowright.BrowserConfig{
    Headless:   false,                    // Set to true for CI/CD
    Timeout:    30 * time.Second,         // Page load timeout
    UserAgent:  "Gowright-Tester/1.0",    // Custom user agent
    WindowSize: &gowright.WindowSize{     // Browser window size
        Width:  1920,
        Height: 1080,
    },
}
```

### API Configuration

```go
config := &gowright.APIConfig{
    BaseURL: "https://api.example.com",   // Base URL for all requests
    Timeout: 30 * time.Second,            // Request timeout
    Headers: map[string]string{           // Default headers
        "User-Agent":    "Gowright-API-Tester/1.0",
        "Content-Type":  "application/json",
    },
    AuthConfig: &gowright.AuthConfig{     // Authentication config
        Type:  "bearer",
        Token: "your-api-token",
    },
}
```

### Database Configuration

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "primary": {
            Driver:       "postgres",              // Database driver
            DSN:          "postgres://user:pass@localhost/db",
            MaxOpenConns: 10,                      // Connection pool size
            MaxIdleConns: 5,                       // Idle connections
        },
        "secondary": {
            Driver:       "mysql",
            DSN:          "user:pass@tcp(localhost:3306)/db",
            MaxOpenConns: 5,
            MaxIdleConns: 2,
        },
    },
}
```

## Best Practices

### 1. Test Organization

- Group related tests into suites
- Use descriptive test names
- Include setup and teardown functions
- Implement proper error handling

### 2. Assertions

- Use specific assertion methods for better error messages
- Include descriptive assertion messages
- Validate both positive and negative scenarios
- Log important test steps for debugging

### 3. Configuration Management

- Use environment variables for sensitive data
- Separate configuration for different environments
- Use reasonable timeouts for network operations
- Configure appropriate connection pools for databases

### 4. Reporting

- Enable both JSON and HTML reports for different audiences
- Include screenshots for UI test failures
- Log detailed information for debugging
- Use structured logging for better report parsing

### 5. Integration Testing

- Design tests to be independent and repeatable
- Implement proper cleanup and rollback mechanisms
- Use realistic test data and scenarios
- Test error conditions and edge cases

## Troubleshooting

### Common Issues

1. **Browser not found**: Ensure Chrome/Chromium is installed for UI tests
2. **Database connection errors**: Check database configuration and permissions
3. **API timeout errors**: Verify network connectivity and increase timeouts if needed
4. **Report generation failures**: Ensure output directories have write permissions

### Debug Mode

Enable debug logging by setting the log level:

```go
config := &gowright.Config{
    LogLevel: "DEBUG",  // Enable detailed logging
    // ... other config
}
```

### Environment Variables

Set environment variables for configuration:

```bash
export GOWRIGHT_LOG_LEVEL=DEBUG
export GOWRIGHT_BROWSER_HEADLESS=true
export GOWRIGHT_API_TIMEOUT=60s
export GOWRIGHT_DB_MAX_CONNECTIONS=20
```

## Contributing

When adding new examples:

1. Follow the existing naming convention
2. Include comprehensive comments and documentation
3. Demonstrate both success and failure scenarios
4. Generate appropriate reports
5. Update this README with the new example

## Support

For questions, issues, or contributions:

1. Check the main framework documentation
2. Review existing examples for similar use cases
3. Create issues for bugs or feature requests
4. Submit pull requests for improvements

---

These examples provide a comprehensive foundation for using the Gowright testing framework in real-world scenarios. Each example can be customized and extended based on your specific testing requirements.