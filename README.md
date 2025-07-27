# Prasmoid CLI

<p align="center">
  <img src="logo.svg" alt="Prasmoid Logo" width="300">
</p>

<p align="center">
  <strong>The All in One Development Toolkit for KDE Plasmoids</strong>
  <br />
  Build, test, and manage your plasmoids with unparalleled ease and efficiency.
</p>

<p align="center">
    <a href="https://github.com/PRASSamin/prasmoid/pulls">
        <img src="https://img.shields.io/badge/PR-Welcome-brightgreen" alt="PRs Welcome">
    </a>
    <a href="https://go.dev/">
        <img src="https://img.shields.io/badge/Go-1.24-blue.svg" alt="Go Version">
    </a>
    <a href="./LICENSE">
        <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License">
    </a>
</p>

---

## Why Prasmoid?

While the core structure of KDE Plasma plasmoids is straightforward, the surrounding development workflow, including setup, building, testing, and deployment often involves repetitive manual steps. **Prasmoid CLI** is designed to abstract away these boring tasks. It's a powerful command-line tool, crafted with Go, that provides a seamless, integrated experience, allowing you to **focus solely on writing your plasmoid's code**.

**Focus on your code, not the boilerplate.** Prasmoid handles the heavy lifting, from project scaffolding and live previews to smart versioning and packaging, allowing you to concentrate on creating amazing plasmoids.

One of its most revolutionary features is a **built-in, zero-dependency JavaScript runtime**. This allows you to extend the CLI with your own custom commands, automating any workflow imaginable, directly within your project – no Node.js installation required!

## Getting Started

### Installation

Prasmoid is designed for quick and easy installation. Choose your preferred method:

> [!NOTE]
> The installer script requires `jq` to be installed for parsing GitHub API responses.

#### Recommended: Standard CLI

```bash
curl -sSL https://raw.githubusercontent.com/PRASSamin/prasmoid/main/install | bash -s 1
```

#### Compact: Compressed CLI

```bash
curl -sSL https://raw.githubusercontent.com/PRASSamin/prasmoid/main/install | bash -s 2
```

### Updating Prasmoid

Keep Prasmoid up to date with the latest features and improvements using one of these methods:

#### 1. CLI Method (Recommended)

The simplest way to update Prasmoid is by using the built-in update command:

```bash
prasmoid update me
```

> [!TIP]
> This command is a convenient wrapper around the manual update method. It's designed to be lightweight and efficient, avoiding the need for additional internal update logic that would increase the binary size.

#### 2. Manual Update via Curl

If you prefer a manual update, you can use curl:

```bash
curl -sSL https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update | bash -s $(which prasmoid)
```

## Your First Plasmoid Project

Creating a new plasmoid is incredibly simple with `prasmoid init`.

1.  **Run the initialization command:**

    ```bash
    prasmoid init
    ```

2.  **Follow the interactive prompts:** Prasmoid will guide you through setting up your project details:

    ```
    ? Project name: MyAwesomePlasmoid
    ? Description: A new plasmoid project.
    ? Choose a license: GPL-3.0+
    ? Author: Alex Doe
    ? Author email: alex@example.com
    ? Plasmoid ID: org.kde.myawesomeplasmoid
    ? Initialize a git repository? Yes
    ```

    Once completed, a new project directory will be created with your chosen configuration, ready for development!

## Command Reference

Prasmoid provides a comprehensive set of commands to manage your plasmoid projects.

| Command           | Description                                                                                                                                                       | Usage & Flags                                                                                                                                                                             |
| :---------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `init`            | Initializes a new plasmoid project.                                                                                                                               | `prasmoid init [-n <name>]` <br> `-n, --name`: Specify the project name directly.                                                                                                         |
| `build`           | Builds and packages the project into a distributable `.plasmoid` archive.                                                                                         | `prasmoid build [-o <output_dir>]` <br> `-o, --output`: Specify a custom output directory (defaults to `./build`).                                                                        |
| `preview`         | Launches the plasmoid in a live preview window.                                                                                                                   | `prasmoid preview [-w]` <br> `-w, --watch`: Enable automatic window restart on file changes.                                                                                              |
| `format`          | Formats all `.qml` files in the `contents` directory using `qmlformat`.                                                                                           | `prasmoid format [-d <dir>] [-w]` <br> `-d, --dir`: Specify the directory to format (defaults to `./contents`). <br> `-w, --watch`: Watch for file changes and automatically format them. |
| `link`            | Creates a symbolic link from your project to the KDE plasmoids development directory (`~/.local/share/plasma/plasmoids/`). Essential for development and preview. | `prasmoid link [-w]` <br> `-w, --where`: Show the linked path without performing the link operation.                                                                                      |
| `unlink`          | Removes the symbolic link created by `prasmoid link`.                                                                                                             | `prasmoid unlink`                                                                                                                                                                         |
| `install`         | Installs the current plasmoid project to the system-wide plasmoids directory for production use.                                                                  | `prasmoid install`                                                                                                                                                                        |
| `uninstall`       | Removes the plasmoid from the system.                                                                                                                             | `prasmoid uninstall`                                                                                                                                                                      |
| `changeset`       | Manages versioning and changelogs for your project.                                                                                                               | See subcommands below.                                                                                                                                                                    |
| `changeset add`   | Creates a new changeset file, prompting for version bump and summary.                                                                                             | `prasmoid changeset add [-b <type>] [-s <summary>]` <br> `-b, --bump`: Specify version bump type (`patch`, `minor`, `major`). <br> `-s, --summary`: Provide a changelog summary directly. |
| `changeset apply` | Applies all pending changesets, updating `metadata.json` and `CHANGELOG.md`.                                                                                      | `prasmoid changeset apply`                                                                                                                                                                |
| `commands`        | Manages custom, project-specific JavaScript commands.                                                                                                             | See subcommands below.                                                                                                                                                                    |
| `commands add`    | Creates a new JavaScript command file from a template in `.prasmoid/commands/`.                                                                                   | `prasmoid commands add [-n <name>]` <br> `-n, --name`: Specify the command name.                                                                                                          |
| `commands remove` | Deletes a custom command.                                                                                                                                         | `prasmoid commands remove [-n <name>]` <br> `-n, --name`: Specify the command name to remove.                                                                                             |
| `update`          | Manage update operations.                                                                                                                                         | See subcommands below.                                                                                                                                                                    |
| `me`              | Updates Prasmoid to the latest version.                                                                                                                           | `prasmoid update me`                                                                                                                                                                      |
| `version`         | Displays the current version of Prasmoid.                                                                                                                         | `prasmoid version`                                                                                                                                                                        |

## Extending Prasmoid with Custom Commands

Prasmoid's most powerful and unique feature is its extensibility through custom JavaScript commands. This allows you to automate any project-specific workflow directly within your CLI, without needing Node.js installed on your system.

### How it Works: The Embedded JavaScript Runtime

Prasmoid includes a lightweight, high-performance JavaScript runtime embedded directly within its Go binary. This runtime provides a Node.js-like environment, offering synchronous APIs for common modules such as `fs`, `os`, `path`, `child_process`, and a custom `prasmoid` module for CLI-specific interactions.

This means you can write powerful automation scripts in JavaScript, and Prasmoid will execute them natively, making your custom commands fast, portable, and truly zero-dependency for end-users.

### Creating a Custom Command

1.  **Generate the command file:**

    ```bash
    prasmoid commands add deploy
    ```

    This will create a file like `.prasmoid/commands/deploy.js`.

2.  **Edit the file** to define your command's logic. Prasmoid automatically adds type definitions (`prasmoid.d.ts`) for autocompletion and type-checking in editors like VS Code.

    ```javascript
    // .prasmoid/commands/hello.js

    /// <reference path="../../prasmoid.d.ts" />
    const prasmoid = require("prasmoid");
    const fs = require("fs"); // Example: You can use fs module

    prasmoid.Command({
      // The main function to run when the command is executed
      run: (ctx) => {
        const name = ctx.Flags().get("name") || "World";
        console.green(`Hello, ${name}!`); // Use Prasmoid's colored console

        const args = ctx.Args();
        if (args.length > 0) {
          console.yellow("Arguments received:", args.join(", "));
        }

        // Example: Read a file using the embedded fs module
        try {
          const content = fs.readFileSync("somefile.txt", "utf8");
          console.log("File content:", content);
        } catch (e) {
          console.red("Error reading file:", e.message);
        }
      },

      // A short description shown in the help list
      short: "Prints a greeting.",

      // A longer, more detailed description for 'prasmoid help hello'
      long: "A simple command that prints a greeting. You can use the --name flag to customize the greeting.\n\nExample:\n  prasmoid hello --name 'Alice'",

      // Define command-line flags
      flags: [
        {
          name: "name",
          shorthand: "n",
          type: "string",
          default: "World",
          description: "The name to greet.",
        },
      ],
    });
    ```

3.  **Run your custom command:**

    ```bash
    $ prasmoid hello
    # Output: Hello, World!

    $ prasmoid hello --name "Prasmoid" an-argument
    # Output: Hello, Prasmoid!
    #         Arguments received: an-argument
    ```

### Available JavaScript Modules & APIs

The embedded runtime provides a subset of Node.js-like APIs, focusing on synchronous operations suitable for CLI scripting:

- **`prasmoid`**: Custom module for CLI interactions.
  - `prasmoid.Command(config)`: Registers a new command.
  - `prasmoid.getMetadata(key)`: Reads values from `metadata.json`.
  - `ctx.Args()`: Get command-line arguments.
  - `ctx.Flags().get(name)`: Get flag values.
- **`console`**: Enhanced logging with color support (`console.log`, `console.red`, `console.green`, `console.color`, etc.).
- **`fs`**: Synchronous file system operations (`fs.readFileSync`, `fs.writeFileSync`, `fs.existsSync`, `fs.readdirSync`, etc.).
- **`os`**: Operating system information (`os.arch`, `os.platform`, `os.homedir`, `os.tmpdir`, etc.).
- **`child_process`**: Execute shell commands synchronously (`child_process.execSync`).
- **`process`**: Process information and control (`process.exit`, `process.cwd`, `process.env`, `process.uptime`, `process.memoryUsage`, `process.nextTick`).
- **`path`**: Utilities for working with file paths (`path.join`, `path.resolve`, `path.basename`, `path.extname`, etc.).

> [!NOTE]
> The embedded runtime currently supports **synchronous** file system operations only. Asynchronous functions (e.g., `fs.readFile`) are not implemented.

## Contributing

We welcome contributions from the community! Whether it's bug reports, feature requests, or code contributions, your help is invaluable.

- **Report Bugs**: If you find an issue, please open a [GitHub Issue](https://github.com/PRASSamin/prasmoid/issues).
- **Suggest Features**: Have an idea for a new feature? Open an issue to discuss it.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
