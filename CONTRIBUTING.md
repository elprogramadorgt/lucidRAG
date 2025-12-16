# Contributing to lucidRAG

Thank you for your interest in contributing to lucidRAG! This document provides guidelines and best practices for contributing to the project.

## Development Setup

Please refer to the [README.md](README.md) for instructions on setting up your development environment.

## Code Style

### Go

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `go vet` before committing
- Write tests for new functionality
- Keep functions small and focused
- Use meaningful variable and function names

### Angular/TypeScript

- Follow the [Angular Style Guide](https://angular.io/guide/styleguide)
- Use TypeScript strict mode
- Write unit tests for components and services
- Use meaningful component and service names
- Follow component-based architecture

## Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(rag): add document chunking algorithm
fix(whatsapp): handle webhook timeout errors
docs(readme): update installation instructions
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Write or update tests
5. Ensure all tests pass
6. Update documentation if needed
7. Submit a pull request

## Testing

### Go Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/rag -v
```

### Angular Tests

```bash
cd admin-ui
npm test
```

## Code Review

All pull requests require at least one approval before merging. Reviewers will check:

- Code quality and readability
- Test coverage
- Documentation updates
- Adherence to project conventions
- Performance implications
- Security considerations

## Questions?

If you have questions, please open an issue on GitHub.
