# Contributing to Gowright

Thank you for your interest in contributing to Gowright! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- Chrome/Chromium browser (for UI testing)
- SQLite, PostgreSQL, or MySQL (for database testing)

### Development Setup

1. **Fork the repository**
   ```bash
   # Fork the repo on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/gowright.git
   cd gowright
   ```

2. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/original-org/gowright.git
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Verify setup**
   ```bash
   # Run tests to ensure everything works
   go test ./...
   
   # Run integration tests
   go run integration_test_runner.go
   
   # Run benchmarks
   go test -bench=. ./...
   ```

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

- **Bug fixes**: Fix issues in existing functionality
- **New features**: Add new testing capabilities or improvements
- **Documentation**: Improve or add documentation
- **Performance improvements**: Optimize existing code
- **Test coverage**: Add or improve tests
- **Examples**: Add usage examples or tutorials

### Before You Start

1. **Check existing issues**: Look for existing issues or discussions about your idea
2. **Create an issue**: For significant changes, create an issue to discuss the approach
3. **Get feedback**: Engage with maintainers and community members

### Coding Standards

#### Go Style Guide

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Write clear, self-documenting code

#### Code Organization

```
pkg/gowright/
├── core/           # Core framework functionality
├── api/            # API testing module
├── ui/             # UI testing module
├── database/       # Database testing module
├── integration/    # Integration testing module
├── reporting/      # Reporting system
└── utils/          # Utility functions
```

#### Naming Conventions

- **Packages**: Use lowercase, single words when possible
- **Types**: Use PascalCase (e.g., `APITester`, `DatabaseConfig`)
- **Functions**: Use PascalCase for exported, camelCase for unexported
- **Variables**: Use camelCase
- **Constants**: Use PascalCase or UPPER_CASE for package-level constants

#### Documentation

- Add godoc comments for all exported types, functions, and methods
- Use complete sentences in comments
- Provide examples for complex functionality

```go
// APITester provides HTTP client capabilities for testing REST APIs.
// It supports various HTTP methods, authentication, and response validation.
//
// Example:
//   config := &APIConfig{BaseURL: "https://api.example.com"}
//   tester := NewAPITester(config)
//   response, err := tester.Get("/users", nil)
type APITester interface {
    // Get performs a GET request to the specified endpoint.
    Get(endpoint string, headers map[string]string) (*APIResponse, error)
}
```

#### Error Handling

- Use the `GowrightError` type for framework-specific errors
- Provide context information in errors
- Handle errors gracefully with appropriate fallbacks

```go
if err != nil {
    return NewGowrightError(APIError, "failed to execute request", err).
        WithContext("endpoint", endpoint).
        WithContext("method", "GET")
}
```

#### Testing

- Write tests for all new functionality
- Maintain or improve test coverage
- Use table-driven tests when appropriate
- Mock external dependencies

```go
func TestAPITester_Get(t *testing.T) {
    tests := []struct {
        name           string
        endpoint       string
        expectedStatus int
        expectError    bool
    }{
        {
            name:           "successful request",
            endpoint:       "/users",
            expectedStatus: 200,
            expectError:    false,
        },
        {
            name:        "invalid endpoint",
            endpoint:    "/invalid",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Pull Request Process

### 1. Create a Branch

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Or a bugfix branch
git checkout -b bugfix/issue-number-description
```

### 2. Make Changes

- Write clean, well-documented code
- Add or update tests as needed
- Update documentation if necessary
- Ensure all tests pass

### 3. Commit Changes

Use clear, descriptive commit messages:

```bash
# Good commit messages
git commit -m "Add support for custom HTTP headers in API tests"
git commit -m "Fix memory leak in browser pool management"
git commit -m "Update README with new configuration options"

# Follow conventional commits format
git commit -m "feat: add support for mobile device emulation"
git commit -m "fix: resolve database connection timeout issue"
git commit -m "docs: update API documentation for new features"
```

### 4. Push and Create PR

```bash
# Push your branch
git push origin feature/your-feature-name

# Create a pull request on GitHub
```

### 5. PR Requirements

Your pull request should:

- **Have a clear title and description**
- **Reference related issues** (e.g., "Fixes #123")
- **Include tests** for new functionality
- **Update documentation** if needed
- **Pass all CI checks**
- **Be reviewed** by at least one maintainer

### PR Template

```markdown
## Description
Brief description of changes made.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Other (please describe)

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] No breaking changes (or clearly documented)

## Related Issues
Fixes #(issue number)
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./pkg/gowright/

# Run integration tests
go run integration_test_runner.go

# Run performance benchmarks
go test -bench=. ./...

# Run tests with race detection
go test -race ./...
```

### Test Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test module interactions
3. **End-to-End Tests**: Test complete workflows
4. **Performance Tests**: Benchmark critical paths
5. **Regression Tests**: Prevent known issues from reoccurring

### Writing Tests

#### Unit Tests

```go
func TestNewAPITester(t *testing.T) {
    config := &APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
    }
    
    tester := NewAPITester(config)
    
    assert.NotNil(t, tester)
    assert.Equal(t, config.BaseURL, tester.GetConfig().BaseURL)
}
```

#### Integration Tests

```go
func TestAPITesterIntegration(t *testing.T) {
    // Setup test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()
    
    // Test with real HTTP calls
    config := &APIConfig{BaseURL: server.URL}
    tester := NewAPITester(config)
    
    response, err := tester.Get("/test", nil)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
}
```

#### Benchmark Tests

```go
func BenchmarkAPITester_Get(b *testing.B) {
    tester := setupAPITester()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := tester.Get("/benchmark", nil)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Documentation

### Types of Documentation

1. **Code Documentation**: Godoc comments
2. **API Documentation**: Comprehensive API reference
3. **User Guide**: Usage examples and tutorials
4. **Contributing Guide**: This document
5. **Changelog**: Record of changes between versions

### Writing Documentation

- Use clear, concise language
- Provide practical examples
- Keep documentation up-to-date with code changes
- Use proper markdown formatting

### Generating Documentation

```bash
# Generate godoc documentation
godoc -http=:6060

# View documentation at http://localhost:6060/pkg/github.com/your-org/gowright/
```

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

1. **Update version numbers**
2. **Update CHANGELOG.md**
3. **Run full test suite**
4. **Update documentation**
5. **Create release tag**
6. **Publish release notes**

### Creating a Release

```bash
# Create and push tag
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3

# Create release on GitHub with release notes
```

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Email**: [maintainers@gowright.dev](mailto:maintainers@gowright.dev)

### Asking Questions

When asking for help:

1. **Search existing issues** first
2. **Provide context** about what you're trying to do
3. **Include relevant code** and error messages
4. **Specify your environment** (Go version, OS, etc.)

### Reporting Bugs

Use the bug report template:

```markdown
## Bug Description
Clear description of the bug.

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen.

## Actual Behavior
What actually happens.

## Environment
- Go version:
- Gowright version:
- OS:
- Browser (if UI testing):

## Additional Context
Any other relevant information.
```

## Recognition

Contributors are recognized in:

- **CONTRIBUTORS.md**: List of all contributors
- **Release notes**: Major contributions highlighted
- **GitHub**: Contributor statistics and graphs

Thank you for contributing to Gowright! Your efforts help make testing in Go more comprehensive and enjoyable for everyone. 🚀