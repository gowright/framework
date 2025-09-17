# Changelog

All notable changes to the Gowright Testing Framework will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.1.0] - 2024-12-XX

### Added
- **UI Testing with Rod**: Complete browser automation implementation using go-rod/rod
  - Full Chrome/Chromium browser automation via Chrome DevTools Protocol
  - Element interactions: click, type, scroll, attribute access
  - Advanced assertions: text validation, element existence, visibility checks
  - Screenshot capture with automatic file management
  - JavaScript execution in browser context
  - Wait strategies for elements and custom conditions
  - Configurable browser options: headless/headed mode, window size, timeouts
  - Support for custom user agents and browser arguments
  - Automatic browser lifecycle management (launch, connect, cleanup)

### Enhanced
- **UIAssertion struct**: Added `Attribute` field for attribute-based assertions
- **Error handling**: Improved error messages, error type consistency, and proper cleanup error handling
- **Documentation**: Comprehensive UI testing documentation and examples
- **Default Chrome Arguments**: Automatically applies `--no-default-browser-check`, `--no-first-run`, and `--disable-fre` for better automation experience
- **Cookie Notice Handling**: Added `DismissCookieNotices()` method with comprehensive JavaScript-based dismissal

### Dependencies
- Added `github.com/go-rod/rod v0.116.2` for browser automation
- Updated go.mod with rod dependencies and transitive packages

### Technical Details
- Implemented 15+ UI testing methods with full rod integration
- Added support for 7 different assertion types
- Added support for 6 different action types
- Created example applications demonstrating UI testing capabilities
- Updated unit tests to work with real browser automation (using data URLs for test HTML)
- Full compatibility with existing Gowright framework architecture

## [v1.0.0] - Previous Release

### Added
- Initial framework release with API, Database, Mobile, and Integration testing
- OpenAPI specification validation and testing
- Comprehensive reporting system
- Parallel execution capabilities
- Modular architecture design