# Contributing to enva

Thank you for considering contributing to this project! Here are the guidelines to help you get started.

## How to Contribute

### Reporting Bugs
- Use the GitHub Issues tab
- Include: OS, Go version, steps to reproduce, expected vs actual behavior
- If possible, include error logs or screenshots

### Suggesting Features
- Open an issue with the "enhancement" label
- Describe the feature and why it would be useful
- Consider if it aligns with the project's scope

### Submitting Changes
1. **Fork** the repository
2. **Create a branch**: `git checkout -b feature/your-feature-name`
3. **Make your changes**
4. **Test your changes**: Run `go test ./...` and ensure all tests pass
5. **Commit with Conventional Commits**:
   ```bash
   # Format: type(scope): description
   git commit -m "feat: add new validation rule"
   git commit -m "fix: resolve null pointer in parser"
   git commit -m "docs: update installation instructions"
   ```
6. **Push to your fork**: git push origin feature/your-feature-name
7. **Open a Pull Request**
