# Contributing to Webull OpenAPI CLI

Thank you for your interest in contributing! This document provides guidelines for contributing to this project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/webull-openapi-cli.git`
3. Create a feature branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run tests: `make test`
6. Run linter: `make lint`
7. Commit your changes: `git commit -m "feat: add your feature"`
8. Push to your fork: `git push origin feature/your-feature`
9. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.22 or later
- Make

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- All exported types and functions must have doc comments
- Keep functions focused and small
- Error messages should be lowercase and not end with punctuation
- All code, comments, and documentation must be in English

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation only
- `refactor:` — Code change that neither fixes a bug nor adds a feature
- `test:` — Adding or updating tests
- `chore:` — Maintenance tasks

## Pull Request Guidelines

- One PR per feature or fix
- Include tests for new functionality
- Update documentation if needed
- Ensure CI passes before requesting review
- Keep PRs focused and reasonable in size

## Reporting Issues

- Use GitHub Issues
- Include steps to reproduce
- Include expected vs actual behavior
- Include CLI version (`webull version`)
- Include OS and Go version

## Code of Conduct

Be respectful, constructive, and professional in all interactions.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
