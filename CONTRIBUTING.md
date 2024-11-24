# Contributing to Go-Again

First off, thank you for considering contributing to Go-Again! It's people like you that make Go-Again such a great tool.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct (please be kind and respectful to others).

## How Can I Contribute?

### Testing Framework Compatibility

One of the most valuable ways you can contribute is by testing and adding support for different Go web frameworks. Here's how:

1. Pick an untested framework from the README
2. Create a minimal working example using Go-Again
3. Document any special configuration needed
4. Update the compatibility table in the README
5. Submit a PR with your example and documentation

### Reporting Bugs

When filing a bug report, please include:

- A clear, descriptive title
- Go version (`go version`)
- Go-Again version
- Framework and version being used
- Minimal example to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant logs or error messages

### Suggesting Enhancements

We love new ideas! When suggesting enhancements:

- Use a clear, descriptive title
- Provide a step-by-step description of the suggested enhancement
- Explain why this enhancement would be useful
- List some example use cases

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure all tests pass
5. Make sure your code follows the existing style

#### Commit Message Guidelines

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or fewer
- Reference issues and pull requests in the description

Example:
```
Add support for Echo framework

- Add Echo framework example
- Update compatibility table
- Add Echo-specific configuration docs
- Add tests for Echo integration

Closes #123
```

### Development Process

1. Install dependencies:
```bash
go mod download
```

2. Run tests:
```bash
go test ./...
```

3. Format code:
```bash
go fmt ./...
```

### Project Structure

```
go-again/
├── examples/          # Framework-specific examples
│   ├── gin/
│   ├── echo/
│   └── ...
├── internal/         # Internal packages
├── tests/           # Test files
└── reloader.go      # Main package code
```

### Adding Framework Support

When adding support for a new framework:

1. Create a new example in the `examples/` directory
2. Include a minimal working example
3. Document any framework-specific configuration
4. Add tests for the framework integration
5. Update the compatibility table in README.md

Example structure for framework support:
```go
examples/
└── framework-name/
    ├── main.go
    ├── templates/
    │   └── index.html
    └── README.md    # Framework-specific instructions
```

### Testing

- Write tests for any new functionality
- Run the full test suite before submitting PRs
- Include both unit and integration tests where appropriate

### Documentation

- Update README.md for any user-facing changes
- Add godoc comments for exported functions
- Include examples in doc comments where helpful
- Update framework compatibility table when adding support

### Review Process

1. Create a Pull Request
2. CI will run automated tests
3. Maintainers will review your code
4. Address any requested changes
5. Once approved, your PR will be merged

### Release Process

1. Version numbers follow [Semantic Versioning](https://semver.org/)
2. Document all changes in CHANGELOG.md
3. Tag releases with appropriate version number

## Need Help?

Feel free to:
- Open an issue with questions
- Ask for clarification on existing issues
- Join discussions on open issues and PRs

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes for their contributions
- Credited in framework-specific documentation they create

Thank you for contributing to Go-Again! 🎉
