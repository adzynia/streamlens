# Contributing to StreamLens

Thank you for your interest in contributing to StreamLens! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/streamlens.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make test`
6. Commit your changes: `git commit -am 'Add some feature'`
7. Push to the branch: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Setup

### Prerequisites
- Go 1.22 or higher
- Docker & Docker Compose
- Make (optional but recommended)

### Local Development
```bash
# Install dependencies
make deps

# Start infrastructure (Redpanda + Postgres)
docker compose -f deploy/docker-compose.yml up -d redpanda postgres

# Run services locally
make run-ingestion    # Terminal 1
make run-processor    # Terminal 2
make run-metrics-api  # Terminal 3
```

## Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and concise

## Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR: `make test`
- Aim for reasonable test coverage of critical paths
- Include both happy path and error cases

Example test structure:
```go
func TestFunctionName(t *testing.T) {
    // Setup
    // Execute
    // Assert
}
```

## Commit Messages

Follow the conventional commits format:

- `feat: add new feature`
- `fix: resolve bug in processor`
- `docs: update README with examples`
- `test: add unit tests for handlers`
- `refactor: simplify aggregation logic`
- `chore: update dependencies`

## Pull Request Process

1. **Update documentation** if you're changing functionality
2. **Add tests** for new features or bug fixes
3. **Ensure CI passes** - all builds and tests must succeed
4. **Keep PRs focused** - one feature/fix per PR
5. **Respond to feedback** - address review comments promptly

### PR Description Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe how you tested your changes

## Checklist
- [ ] Code follows project style guidelines
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] All tests passing
```

## Areas for Contribution

We welcome contributions in these areas:

### Features
- Distributed tracing (OpenTelemetry) integration
- Prometheus metrics endpoint
- Additional aggregation windows (5min, 15min, 1hr)
- Support for `llm.evaluations` topic
- Rate limiting on ingestion API
- Schema registry integration

### Infrastructure
- Kubernetes deployment manifests
- Terraform configurations
- Grafana dashboard templates
- Alerting rules

### Documentation
- Additional examples and use cases
- Architecture diagrams
- Performance tuning guides
- Deployment guides for different platforms

### Testing
- Integration tests
- Load testing scenarios
- Chaos engineering tests
- End-to-end test suite

## Reporting Bugs

Use GitHub Issues to report bugs. Include:

- **Description**: Clear description of the bug
- **Steps to Reproduce**: Minimal steps to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: OS, Go version, Docker version
- **Logs**: Relevant log output (use code blocks)

## Suggesting Features

We love feature suggestions! Open an issue with:

- **Use Case**: Why is this feature needed?
- **Proposed Solution**: How should it work?
- **Alternatives**: Other approaches you considered
- **Additional Context**: Any other relevant information

## Code Review Process

- Maintainers will review PRs within a few days
- Address feedback and push updates to your branch
- Once approved, maintainers will merge your PR
- PRs may be closed if inactive for 30+ days

## Community

- Be respectful and inclusive
- Follow our [Code of Conduct](CODE_OF_CONDUCT.md)
- Help others and share knowledge
- Have fun building great software!

## Questions?

Open a GitHub issue with the `question` label, and we'll help you out.

Thank you for contributing to StreamLens! 🚀
