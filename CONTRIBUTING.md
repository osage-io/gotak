# Contributing to GoTAK

First off, thank you for considering contributing to GoTAK! It's people like you that make GoTAK such a great tool for the tactical communications community.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Security](#security)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

**In short: Be respectful, be inclusive, and help make this a welcoming community for everyone.**

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: [Download here](https://golang.org/dl/)
- **Docker & Docker Compose**: [Get Docker](https://docs.docker.com/get-docker/)
- **Make**: Usually pre-installed on Unix systems
- **Git**: [Install Git](https://git-scm.com/)

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gotak.git
   cd gotak
   ```

3. **Set up the upstream remote**:
   ```bash
   git remote add upstream https://github.com/dfedick/gotak.git
   ```

4. **Install development tools**:
   ```bash
   make install-tools
   make precommit-install
   ```

5. **Start the development environment**:
   ```bash
   make dev-up
   ```

6. **Verify everything works**:
   ```bash
   make test
   make build
   ```

## Development Workflow

### Before You Start

1. **Check existing issues** to see if your idea/bug is already being worked on
2. **Create an issue** if one doesn't exist to discuss your proposed changes
3. **Get assignment** or confirmation from maintainers before starting major work

### Making Changes

1. **Create a feature branch** from `main`:
   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards
3. **Add tests** for new functionality
4. **Run the full test suite**:
   ```bash
   make test
   make test-integration
   make lint
   make security
   ```

5. **Commit your changes** using conventional commits:
   ```bash
   git commit -m "feat: add new CoT message type support"
   ```

### Keeping Your Branch Updated

```bash
# Fetch latest changes from upstream
git fetch upstream

# Rebase your branch on main
git rebase upstream/main
```

## Coding Standards

### Go Code Style

- **Follow Go conventions**: Use `gofmt`, `goimports`, and `go vet`
- **Use golangci-lint**: Our CI runs comprehensive linting
- **Effective Go**: Follow the guidelines in [Effective Go](https://golang.org/doc/effective_go.html)
- **Package documentation**: Every public package should have meaningful documentation
- **Function comments**: Public functions should have descriptive comments

### Code Organization

```go
// ✅ Good: Clear, descriptive function with proper error handling
func (s *Server) handleCoTMessage(client *Client, message []byte) error {
    event, err := cot.ParseCoT(message)
    if err != nil {
        return fmt.Errorf("failed to parse CoT message: %w", err)
    }
    
    return s.processEvent(client, event)
}

// ❌ Bad: Unclear function with poor error handling
func (s *Server) handle(c *Client, m []byte) {
    e, _ := cot.Parse(m)  // Ignoring errors
    s.process(c, e)
}
```

### Project Structure Conventions

- **cmd/**: Application entry points only
- **internal/**: Private application code
- **pkg/**: Reusable library code
- **tests/**: Test files and test utilities
- **config/**: Configuration files and examples
- **docs/**: Documentation

### Configuration

- Use YAML for configuration files
- Provide reasonable defaults
- Include validation for all configuration values
- Document all configuration options

## Testing Requirements

### Unit Tests

- **Coverage requirement**: Maintain >80% test coverage
- **Test file naming**: `*_test.go` alongside the code being tested
- **Test function naming**: `TestFunctionName` or `TestType_Method`
- **Use testify**: Prefer `assert` and `require` from testify

```go
func TestCoTParser_ParseEvent(t *testing.T) {
    parser := cot.NewParser()
    
    testCases := []struct {
        name     string
        input    []byte
        expected *cot.Event
        wantErr  bool
    }{
        {
            name:  "valid position report",
            input: validPositionXML,
            expected: &cot.Event{
                Type: "a-f-G",
                UID:  "test-uid",
            },
            wantErr: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := parser.ParseEvent(tc.input)
            
            if tc.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tc.expected.Type, result.Type)
            assert.Equal(t, tc.expected.UID, result.UID)
        })
    }
}
```

### Integration Tests

- **Test realistic scenarios**: Full client-server communication
- **Use Docker Compose**: Integration tests should use our dev environment
- **Build tag**: Mark integration tests with `//go:build integration`
- **Clean state**: Each test should start with a clean environment

### Benchmarks

- **Performance critical code**: Add benchmarks for performance-sensitive functions
- **Benchmark naming**: `BenchmarkFunctionName`
- **Include in CI**: Performance regressions should be caught

## Documentation

### Code Documentation

- **Package documentation**: Every package should have a doc.go file
- **Public API documentation**: All exported types, functions, and variables
- **Examples**: Include examples in documentation when helpful

### User Documentation

- **README updates**: Keep README.md current with new features
- **Configuration docs**: Document all configuration options
- **API documentation**: Update API docs for any endpoint changes

### Architecture Decisions

- **ADRs**: Document significant architectural decisions
- **Location**: Store in `docs/architecture/` directory
- **Template**: Use the ADR template provided

## Pull Request Process

### Before Submitting

1. **Rebase on main**: Ensure your branch is up to date
2. **Run all checks**: Tests, linting, and security scans must pass
3. **Update documentation**: Include relevant documentation updates
4. **Clean commit history**: Squash commits if necessary

### PR Description Template

```markdown
## Summary
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] No new security vulnerabilities introduced
```

### Review Process

1. **Automated checks**: CI must pass before review
2. **Peer review**: At least one maintainer review required
3. **Address feedback**: Respond to all review comments
4. **Final approval**: Maintainer approval required for merge

## Issue Reporting

### Bug Reports

Use our bug report template and include:

- **Environment details**: OS, Go version, deployment method
- **Steps to reproduce**: Clear, step-by-step instructions
- **Expected vs actual behavior**: What should happen vs what actually happens
- **Logs**: Relevant log output (sanitized of sensitive data)
- **Configuration**: Relevant configuration (sanitized)

### Feature Requests

- **Use case**: Describe the problem you're trying to solve
- **Proposed solution**: Your idea for implementing the feature
- **Alternatives**: Other solutions you've considered
- **Additional context**: Any other relevant information

### Performance Issues

- **Benchmark data**: Include performance measurements
- **Environment**: Detailed system specifications
- **Profiling data**: CPU/memory profiles when relevant
- **Workload description**: What workload causes the issue

## Security

### Security Vulnerabilities

**Do not open GitHub issues for security vulnerabilities.**

Instead, please email security details to: [security@gotak.dev](mailto:security@gotak.dev)

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Security Guidelines

- **No secrets in code**: Never commit API keys, passwords, or certificates
- **Input validation**: Always validate and sanitize user input
- **Secure defaults**: Choose secure configuration defaults
- **Dependency updates**: Keep dependencies current
- **Security testing**: Include security considerations in testing

## Development Best Practices

### Git Workflow

- **Commit messages**: Use conventional commits format
  ```
  feat: add new CoT message type support
  fix: resolve memory leak in client handler
  docs: update API documentation
  test: add integration tests for authentication
  ```

- **Branch naming**: Use descriptive branch names
  ```
  feature/jwt-authentication
  bugfix/memory-leak-client-handler
  docs/api-documentation-update
  ```

### Performance Considerations

- **Memory efficiency**: Be mindful of memory allocations in hot paths
- **Goroutine management**: Properly manage goroutine lifecycles
- **Connection pooling**: Reuse connections where appropriate
- **Profiling**: Use Go's profiling tools for performance optimization

### Error Handling

- **Wrap errors**: Use `fmt.Errorf` with `%w` for error wrapping
- **Context**: Provide meaningful error context
- **Logging**: Log errors at appropriate levels
- **Recovery**: Handle panics gracefully in server code

```go
// ✅ Good error handling
func (s *Server) processMessage(msg []byte) error {
    event, err := cot.ParseCoT(msg)
    if err != nil {
        return fmt.Errorf("failed to parse CoT message: %w", err)
    }
    
    if err := s.validateEvent(event); err != nil {
        return fmt.Errorf("event validation failed: %w", err)
    }
    
    return nil
}
```

## Getting Help

### Communication Channels

- **GitHub Discussions**: For general questions and discussions
- **GitHub Issues**: For bug reports and feature requests
- **Email**: For security-related concerns

### Useful Resources

- [Go Documentation](https://golang.org/doc/)
- [TAK Protocol Documentation](https://tak.gov/)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://postgresql.org/docs/)

## Recognition

Contributors will be recognized in:
- **README.md**: Contributors section
- **CHANGELOG.md**: Release notes
- **Git history**: Proper attribution in commits

Thank you for contributing to GoTAK! 🚀

---

*This contributing guide is inspired by open source best practices and tailored for the GoTAK project. It will evolve as our project grows.*
