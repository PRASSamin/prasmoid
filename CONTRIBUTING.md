# Contributing to Prasmoid

First of all, thank you for considering contributing to Prasmoid! We welcome any and all contributions, from bug reports to new features. This document will guide you through the process.

## How to Contribute

We use the standard GitHub flow for contributions. Here's a quick overview:

1.  **Find an issue to work on.** You can find a list of open issues [here](https://github.com/PRASSamin/prasmoid/issues). Look for issues tagged with `good first issue` or `help wanted` if you're new to the project.
2.  **Fork the repository** to your own GitHub account.
3.  **Create a new branch** for your changes.
4.  **Make your changes** and commit them with a clear and descriptive commit message.
5.  **Run the local CI checks** to ensure your changes meet our standards.
6.  **Push your changes** to your fork.
7.  **Submit a pull request** to the `main` branch of the Prasmoid repository.

## Development Setup

To get started with development, you'll need to set up your local environment. Here's what you'll need:

- **Go:** Version 1.24.x or later.
- **gettext:** A tool for internationalization.
- **plasma-sdk:** The SDK for developing Plasma widgets.

```bash
# For Debian/Ubuntu
sudo apt install gettext plasma-sdk

# For Fedora
sudo dnf install gettext plasma-sdk
```

Once you have these dependencies installed, you can set up the project with the following commands:

```bash
# Clone your fork of the repository
git clone https://github.com/YOUR_USERNAME/prasmoid.git
cd prasmoid

# Install Go dependencies
go mod download
```

## Coding Standards

We use `golangci-lint` to enforce code quality. You can run all the necessary checks with our local CI command:

```bash
go run ./dev/main.go ci
```

This command will run all the necessary checks.

## Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for our commit messages. This helps us maintain a clear and descriptive commit history.

Each commit message should consist of a **type**, a **scope** (optional), and a **description**.

```
<type>(<scope>): <description>
```

- **type:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- **scope:** The part of the codebase you're changing (e.g., `preview`, `ci`, `docs`).
- **description:** A short summary of the changes.

## Pull Request Process

Before submitting a pull request, please ensure the following:

- Your code builds and runs without errors.
- All tests pass. You can run them with `go test ./...`.
- The local CI checks pass. Run `go run ./dev/main.go ci`.

Your pull request should have:

- A clear and descriptive title.
- A detailed description of the changes you've made.
- A reference to the issue it resolves (e.g., `Closes #123`).

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior.
