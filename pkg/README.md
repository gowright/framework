# Gowright Framework - Modular Package Structure

The Gowright framework has been refactored into a modular package structure to improve maintainability, testability, and separation of concerns. Each package focuses on a specific aspect of testing while the integration package orchestrates them all.

## Package Structure

```
pkg/
├── gowright.go              # Main entry point with backward compatibility
├── core/                    # Core framework and interfaces
│   ├── gowright.go         # Main framework orchestrator
│   ├── interfaces.go       # Core interfaces for all testers
│   ├── types.go           # Core types and error handling
│   └── test_types.go      # Test type definitions
├── config/                  # Configuration management
│   └── config.go          # All configuration types and defaults
├── ui/                     # UI/Browser testing
│   └── ui_tester.go       # UI testing implementation
├── api/                    # API/HTTP testing
│   └── api_tester.go      # API testing implementation
├── database/               # Database testing
│   └── database_tester.go # Database testing implementation
├── mobile/                 # Mobile/Appium testing
│   └── mobile_tester.go   # Mobile testing implementation
├── integration/            # Integration testing orchestrator
│   └── integration.go     # Orchestrates all other packages
├── reporting/              # Test result reporting
│   └── reporter.go        # Multiple report format support
└── assertions/             # Common assertion utilities
    └── assertions.go      # Shared assertion functions
```

## Key Benefits

### 1. **Separation of Concerns**
- Each package has a single, well-defined responsibility
- UI testing logic is separate from API testing logic
- Configuration is centralized but modular
- Reporting is independent of test execution

### 2. **Integration Package as Orchestrator**
- The `integration` package brings together all other packages
- Enables complex workflows that span multiple testing domains
- Provides rollback capabilities for failed integration tests
- Maintains state across different test types

### 3. **Modular Usage**
- Use individual packages independently
- Mix and match only the testing capabilities you need
- Easier to mock and test individual components
- Reduced dependencies for specific use cases

### 4. **Backward Compatibility**
- Main `pkg/gowright.go` re-exports all types and functions
- Existing code continues to work without changes
- Gradual migration path for users who want to adopt modular approach

## Usage Examples

### Complete Framework Setup
```go
import "github.com/gowright/framework/pkg/gowright"

// Create with all testers
gw := gowright.NewGowrightWithAllTesters(gowright.DefaultConfig())
gw.Initialize()
defer gw.Close()
```

### Individual Package Usage
```go
import (
    "github.com/gowright/framework/pkg/ui"
    "github.com/gowright/framework/pkg/api"
    "github.com/gowright/framework/pkg/config"
)

// Use only UI testing
uiTester := ui.NewUITester()
uiTester.Initialize(config.DefaultConfig().BrowserConfig)
```

### Integration Testing
```go
// Integration tests orchestrate multiple test types
integrationTest := &gowright.IntegrationTest{
    Name: "User Registration Flow",
    Steps: []gowright.IntegrationStep{
        {
            Name: "UI Step",
            Type: gowright.StepTypeUI,
            Action: &gowright.UIStepAction{...},
        },
        {
            Name: "API Step", 
            Type: gowright.StepTypeAPI,
            Action: &gowright.APIStepAction{...},
        },
        {
            Name: "Database Step",
            Type: gowright.StepTypeDatabase,
            Action: &gowright.DatabaseStepAction{...},
        },
    },
}

result := gw.ExecuteIntegrationTest(integrationTest)
```

## Package Details

### Core Package (`pkg/core/`)
- **Purpose**: Framework orchestration and core interfaces
- **Key Components**: Main Gowright struct, all tester interfaces, core types
- **Dependencies**: Only config package

### Config Package (`pkg/config/`)
- **Purpose**: Centralized configuration management
- **Key Components**: All configuration structs and defaults
- **Dependencies**: None (pure configuration)

### UI Package (`pkg/ui/`)
- **Purpose**: Browser automation and UI testing
- **Key Components**: UITester implementation, browser interactions
- **Dependencies**: core, config, assertions

### API Package (`pkg/api/`)
- **Purpose**: HTTP API testing and validation
- **Key Components**: APITester implementation, HTTP client management
- **Dependencies**: core, config, assertions

### Database Package (`pkg/database/`)
- **Purpose**: Database testing and validation
- **Key Components**: DatabaseTester implementation, connection management
- **Dependencies**: core, config, assertions

### Mobile Package (`pkg/mobile/`)
- **Purpose**: Mobile app testing via Appium
- **Key Components**: MobileTester implementation, Appium integration
- **Dependencies**: core, config, assertions

### Integration Package (`pkg/integration/`)
- **Purpose**: Orchestrates all other testing packages
- **Key Components**: IntegrationTester, workflow execution, rollback
- **Dependencies**: All other packages

### Reporting Package (`pkg/reporting/`)
- **Purpose**: Test result reporting in multiple formats
- **Key Components**: Multiple reporter implementations (JSON, HTML, XML, JUnit)
- **Dependencies**: core, config

### Assertions Package (`pkg/assertions/`)
- **Purpose**: Common assertion utilities for all test types
- **Key Components**: Asserter with various assertion methods
- **Dependencies**: core

## Migration Guide

### For Existing Users
No changes required! The main package re-exports everything:
```go
// This continues to work exactly as before
import "github.com/gowright/framework/pkg/gowright"
gw := gowright.New(gowright.DefaultConfig())
```

### For New Modular Approach
```go
// Import specific packages as needed
import (
    "github.com/gowright/framework/pkg/core"
    "github.com/gowright/framework/pkg/ui"
    "github.com/gowright/framework/pkg/api"
)

// Use packages independently
uiTester := ui.NewUITester()
apiTester := api.NewAPITester()
```

## Future Extensibility

The modular structure makes it easy to:
- Add new testing domains (e.g., performance, security)
- Implement different backends for existing domains
- Create custom reporters or assertion libraries
- Build domain-specific testing tools on top of core interfaces

Each package can evolve independently while maintaining the overall framework contract through the core interfaces.