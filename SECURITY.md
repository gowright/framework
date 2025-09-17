# Security Policy

## Supported Versions

We actively maintain security for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.1.x   | :white_check_mark: |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it by emailing [security@gowright.dev](mailto:security@gowright.dev). Please do not report security vulnerabilities through public GitHub issues.

## Security Scanning

This project uses [gosec](https://github.com/secureco/gosec) for static security analysis. The following security findings have been reviewed and are considered acceptable:

### Validated Security Findings

#### G204 - Command Injection (pkg/openapi/openapi_tester.go)
- **Finding**: `exec.Command` with potentially tainted input
- **Mitigation**: Input is validated using `isValidCommitHash()` and `isValidSpecPath()` functions
- **Risk**: Low - inputs are sanitized before use

#### G304 - Path Traversal (pkg/ui/memory_efficient_capture.go)
- **Finding**: File operations with variable paths
- **Mitigation**: All file paths are validated using `isValidFilePath()` function
- **Risk**: Low - paths are sanitized to prevent directory traversal

#### G301/G306 - File Permissions (pkg/ui/ui_tester.go)
- **Finding**: Directory (0750) and file (0600) permissions
- **Mitigation**: Permissions have been set to secure values
- **Risk**: None - permissions are appropriately restrictive

## Security Best Practices

### Input Validation
- All external inputs are validated before use
- Path traversal protection is implemented for file operations
- Command injection protection is implemented for subprocess execution

### File System Security
- Screenshot directories are created with 0750 permissions
- Screenshot files are created with 0600 permissions
- File paths are validated to prevent directory traversal

### Browser Security
- Chrome is launched with security-focused arguments:
  - `--no-sandbox` (required for CI environments)
  - `--disable-dev-shm-usage` (prevents container issues)
  - `--disable-gpu` (reduces attack surface in headless mode)

### CI/CD Security
- Security scanning is performed on every build
- Dependency scanning is enabled
- SARIF reports are uploaded to GitHub Security tab

## Security Contact

For security-related questions or concerns, please contact:
- Email: [security@gowright.dev](mailto:security@gowright.dev)
- GitHub: Create a private security advisory

## Acknowledgments

We appreciate the security research community and welcome responsible disclosure of security vulnerabilities.