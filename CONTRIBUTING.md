# Contributing to I2C Display Controller

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Adding New Display Drivers](#adding-new-display-drivers)

## Code of Conduct

This project follows the standard open source code of conduct. Be respectful, constructive, and professional in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone git@github.com:YOUR_USERNAME/i2c-display.git`
3. Add upstream remote: `git remote add upstream git@github.com:ausil/i2c-display.git`
4. Create a feature branch: `git checkout -b feature/my-feature`

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint (for linting)

### Install Dependencies

```bash
go mod download
```

### Install Development Tools

```bash
# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Linter

```bash
make lint
```

## Making Changes

### Branch Naming

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test improvements

### Commit Messages

Use clear, descriptive commit messages:

```
Short summary (50 chars or less)

Detailed explanation if needed. Wrap at 72 characters.

- Bullet points are okay
- Use present tense: "Add feature" not "Added feature"
- Reference issues: "Fixes #123" or "Related to #456"
```

## Testing

### Unit Tests

All new code must include unit tests. Aim for >80% coverage.

```bash
make test
```

### Mock Display Testing

Test without hardware using the mock display:

```bash
make run-mock
```

### Hardware Tests

If you have access to hardware, run integration tests:

```bash
make test-hardware
```

## Code Style

### Go Code Standards

- Follow standard Go conventions
- Use `gofmt` and `goimports`
- Run `make fmt` before committing
- Follow the existing code structure

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"

    // External packages
    "github.com/rs/zerolog"

    // Internal packages
    "github.com/ausil/i2c-display/internal/config"
    "github.com/ausil/i2c-display/internal/logger"
)
```

### Error Handling

- Always check and handle errors
- Use structured logging with context
- Return errors with `fmt.Errorf` and `%w` for wrapping

```go
if err != nil {
    return fmt.Errorf("failed to initialize display: %w", err)
}
```

### Logging

Use the structured logger, not `fmt.Print` or `log`:

```go
log.With().
    Str("display_type", displayType).
    Int("width", width).
    Logger().Info("Initializing display")
```

## Pull Request Process

1. **Update your fork**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**
   ```bash
   make fmt
   make lint
   make test
   ```

3. **Push to your fork**
   ```bash
   git push origin feature/my-feature
   ```

4. **Create Pull Request**
   - Provide a clear description
   - Reference related issues
   - Include screenshots for UI changes
   - Ensure CI passes

5. **Code Review**
   - Address review comments
   - Keep commits clean (squash if needed)
   - Be responsive to feedback

## Adding New Display Drivers

See [DISPLAY_TYPES.md](DISPLAY_TYPES.md) for detailed instructions on adding support for new display types.

### Quick Steps

1. **Create driver file** - `internal/display/YOUR_DISPLAY.go`
2. **Implement Display interface** - All required methods
3. **Update factory** - Add to `internal/display/factory.go`
4. **Add display specs** - Update `internal/config/display_specs.go`
5. **Create example config** - `configs/config.YOUR_DISPLAY.json`
6. **Add tests** - Unit tests for the new driver
7. **Update documentation** - README.md and DISPLAY_TYPES.md

### Example Template

Use `internal/display/TEMPLATE.go.example` as a starting point for new drivers.

## Project Structure

```
i2c-display/
├── cmd/
│   └── i2c-displayd/           # Main application
├── internal/
│   ├── config/                 # Configuration management
│   ├── display/                # Display drivers
│   ├── health/                 # Health checking
│   ├── logger/                 # Structured logging
│   ├── metrics/                # Prometheus metrics
│   ├── renderer/               # Page rendering
│   ├── retry/                  # Retry logic
│   ├── rotation/               # Page rotation
│   └── stats/                  # System statistics
├── configs/                    # Example configurations
├── testdata/                   # Test fixtures
└── .github/workflows/          # CI/CD pipelines
```

## Questions?

- Open an issue for bugs or feature requests
- Tag issues appropriately
- Be patient and respectful

Thank you for contributing!
