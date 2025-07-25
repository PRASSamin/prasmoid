# Prasmoid CLI

<p align="center">
  <img src="https://i.imgur.com/aJ4p3fR.png" alt="Prasmoid Logo" width="150">
</p>

<p align="center">
  <strong>The all-in-one CLI for KDE Plasmoid development.</strong>
  <br />
  Build, test, and manage your plasmoids with ease.
</p>

<p align="center">
    <a href="#">
        <img src="https://img.shields.io/badge/PR-Welcome-brightgreen" alt="PRs Welcome">
    </a>
    <a href="#">
        <img src="https://img.shields.io/badge/Go-1.22-blue.svg" alt="Go Version">
    </a>
    <a href="#">
        <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License">
    </a>
</p>

---

Prasmoid is a command-line tool built with Go that dramatically simplifies the development of KDE Plasma plasmoids. It provides a complete suite of commands to handle everything from project creation and live previews to versioning and packaging, allowing you to focus on building great plasmoids.

One of its most powerful features is a built-in JavaScript runtime, which allows you to extend the CLI with your own custom commands tailored to your project's needs.

## Features

- **Project Scaffolding**: Create a new, fully-structured plasmoid project in seconds with `prasmoid init`.
- **Live Preview**: Test your plasmoid in real-time with `plasmoidviewer` and enjoy automatic reloading on file changes.
- **Automated Building**: Package your project into a distributable `.plasmoid` file with a single command.
- **Code Formatting**: Keep your QML code clean and consistent with the built-in `qmlformat` integration.
- **Smart Versioning**: Manage your project's version and changelog with a powerful `changeset` system.
- **Extensible**: Write your own custom commands in JavaScript to automate any workflow you can imagine.
- **User-Friendly**: Interactive prompts and automatic dependency checks make for a smooth development experience.

## Installation

You can install Prasmoid directly from the source. Make sure you have **Go 1.22+** installed on your system.

```bash
go install github.com/PRASSamin/prasmoid@latest
```

This will compile and place the `prasmoid` binary in your Go bin directory (`$HOME/go/bin`). Ensure this directory is in your system's `PATH`.

## Getting Started: Your First Plasmoid

Creating a new plasmoid is as simple as running `prasmoid init`.

1.  **Run the init command:**

    ```bash
    prasmoid init
    ```

2.  **Answer the interactive prompts:** The CLI will guide you through configuring your project name, description, author details, and license.

    ```
    ╭────────────────────────────────────────────╮
    │    💠 Plasmoid Applet Project Generator    │
    ╰────────────────────────────────────────────╯

    ? Project name: MyAwesomePlasmoid
    ? Description: A new plasmoid project.
    ? Choose a license: GPL-3.0+
    ? Author: Alex Doe
    ? Author email: alex@example.com
    ? Plasmoid ID: org.kde.myawesomeplasmoid
    ? Initialize a git repository? Yes
    ```

3.  **Navigate to your new project and start the preview:**
    ```bash
    cd MyAwesomePlasmoid
    prasmoid preview --watch
    ```
    This will open your plasmoid in `plasmoidviewer`. The `--watch` flag tells Prasmoid to monitor your files for changes and automatically restart the preview.

## Command Reference

### `prasmoid init`

Initializes a new plasmoid project.

```bash
prasmoid init [flags]
```

- **-n, --name**: The name of the project. If not provided, you will be prompted.

### `prasmoid build`

Builds and packages the project into a `.plasmoid` archive in the `./build` directory.

```bash
prasmoid build [flags]
```

- **-o, --output**: Specify a different output directory.

### `prasmoid preview`

Launches the plasmoid in a live preview window.

```bash
prasmoid preview [flags]
```

- **-w, --watch**: Watches for file changes and automatically restarts the previewer.

### `prasmoid format`

Formats all `.qml` files in the `contents` directory using `qmlformat`.

```bash
prasmoid format [flags]
```

- **-d, --dir**: The directory to format (defaults to `./contents`).
- **-w, --watch**: Watches for file changes and automatically formats them.

### `prasmoid link` / `unlink`

- `link`: Creates a symbolic link from your project directory to the KDE plasmoids directory (`~/.local/share/plasma/plasmoids/`), which is necessary for development.
- `unlink`: Removes the symbolic link.

### `prasmoid install` / `uninstall`

- `install`: Copies the project files to the system-wide plasmoids directory for use.
- `uninstall`: Removes the plasmoid from the system.

### `prasmoid changeset`

Manages versioning and changelogs.

- **`prasmoid changeset add`**: Creates a new changeset file. You'll be prompted to select a version bump (patch, minor, major) and to write a summary of the changes.
- **`prasmoid changeset apply`**: Applies all pending changesets, updating the version in `metadata.json` and prepending the changes to `CHANGELOG.md`.

### `prasmoid commands`

Manages custom, project-specific commands.

- **`prasmoid commands add`**: Creates a new JavaScript file in your commands directory (`.prasmoid/commands/` by default) from a template.
- **`prasmoid commands remove`**: Deletes a custom command file.

## Extending Prasmoid with Custom Commands

Prasmoid's most powerful feature is its extensibility. You can write your own commands in JavaScript to automate project-specific tasks.

**Creating a Custom Command**

1.  Run `prasmoid commands add` and give your command a name (e.g., `deploy`).
2.  This creates a file like `.prasmoid/commands/deploy.js`.
3.  Edit the file to define your command's logic.

**Example: A simple "hello" command**

```javascript
// .prasmoid/commands/hello.js

/// <reference path="../../prasmoid.d.ts" />
const prasmoid = require("prasmoid");

prasmoid.Command({
  // The main function to run
  run: (ctx) => {
    const name = ctx.Flags().get("name") || "World";
    console.green(`Hello, ${name}!`);

    const args = ctx.Args();
    if (args.length > 0) {
      console.yellow("Arguments received:", args.join(", "));
    }
  },

  // A short description shown in the help list
  short: "Prints a greeting.",

  // A longer, more detailed description
  long: "A simple command that prints a greeting. You can use the --name flag to customize the greeting.",

  // Command-line flags
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

Now you can run your custom command:

```bash
$ prasmoid hello
# Output: Hello, World!

$ prasmoid hello --name "Prasmoid" an-argument
# Output: Hello, Prasmoid!
#         Arguments received: an-argument
```

### Custom Command API

Inside your JavaScript commands, you have access to a special runtime environment with built-in modules:

- **`prasmoid`**: Interact with project metadata.
- **`console`**: Print colored output to the terminal (`console.green`, `console.red`, etc.).
- **`fs`**: A synchronous file system API, similar to Node.js's `fs` module.
- **`os`**: Provides operating system-level information.
- **`child_process`**: Execute shell commands.

Type definitions are provided in `prasmoid.d.ts` for full autocompletion and type-checking in supported editors like VS Code.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
