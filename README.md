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

> [!IMPORTANT]
> The installer script requires jq to be installed for parsing GitHub API responses.
> Install it via `sudo apt install jq`, `sudo dnf install jq`, `sudo pacman -S jq` depending on your distro.

#### Recommended: Standard Build (Native)

- **Best for**: `Arch`, `Fedora`, `Ubuntu`, `Debian`, and most general-purpose Linux distros.
- This version is compiled with CGO enabled and links to your system's native libraries for full integration and performance.
- Might not work on minimal or stripped-down systems without standard dev libraries.

```bash
sudo curl -sSL https://raw.githubusercontent.com/PRASSamin/prasmoid/main/install | bash -s 1
```

#### Portable Build (Static)

- **Best for**: `Alpine`, `NixOS`, minimal Docker images, CI/CD environments, or any system without full libc/glibc.
- Fully statically linked (CGO disabled), built to just run anywhere, even on weird-ass environments where shared libs are missing.
- Slightly larger binary, but way more portable.

```bash
sudo curl -sSL https://raw.githubusercontent.com/PRASSamin/prasmoid/main/install | bash -s 2
```

#### Packages

#### Arch Linux (AUR)

- **Best for**: `Arch`, `Manjaro`, and other Arch-based distros.
- **_Tested on_**: 'Arch Linux' (fully up-to-date as of 2025-08-05)
  ```bash
  yay -S prasmoid
  ```

#### Debian/Ubuntu (.deb)

- **Best for**: `Debian`, `(K)Ubuntu`, and other Debian derivatives.
- **_Tested on_**: 'Debian testing' (nightly snapshot 2025-08-05), 'Kubuntu 25.04'
- [Debian package](https://github.com/PRASSamin/prasmoid/releases/download/v0.0.3/prasmoid_0.0.3-1_amd64.deb)
- [PPA repository](https://launchpad.net/~northern-lights/+archive/ubuntu/prasmoid)
  - Pre-requisite - add repo:
  ```bash
  sudo add-apt-repository ppa:northern-lights/prasmoid
  sudo apt update
  ```
  - Install:
  ```bash
  sudo apt install prasmoid
  ```

#### Fedora (.rpm)

- **Best for**: `Fedora`, `RHEL`, `CentOS` and other Fedora derivatives.
- **_Tested on_**: 'Fedora 42'
- [x86_64 Fedora package](https://github.com/PRASSamin/prasmoid/releases/download/v0.0.3/prasmoid-0.0.3-2.fc42.x86_64.rpm) | [source Fedora package](https://github.com/PRASSamin/prasmoid/releases/download/v0.0.3/prasmoid-0.0.3-2.fc42.src.rpm)
- [COPR repository](https://copr.fedorainfracloud.org/coprs/northernlights/prasmoid/)
  - Pre-requisite - add repo:
  ```bash
  sudo dnf copr enable northernlights/prasmoid # for dnf5
  ```
  - Install:
  ```bash
  sudo dnf install prasmoid
  ```

#### Snap

- **Best for**: Anywhere with snaps available.
- **_Tested on_**: 'Kubuntu 25.04'
- [Snap package](https://snapcraft.io/prasmoid)
  ```bash
  snap install prasmoid
  ```

#### Flatpak

- **Best for**: Anywhere with flatpaks available.
- **_Tested on_**: 'Arch Linux' (fully up-to-date as of 2025-08-10)
- [Flatpak package](https://github.com/PRASSamin/prasmoid/releases/download/v0.0.3/prasmoid-v0.0.3.flatpak)

```bash
wget https://github.com/PRASSamin/prasmoid/releases/download/v0.0.3/prasmoid-v0.0.3.flatpak && flatpak install --user prasmoid-v0.0.3.flatpak
```

#### Installation via Go

- **Best for**: `Any system with Go installed`.
- **_Tested on_**: 'Fedora Linux 42(KDE)'

> [!IMPORTANT]
> This method requires Go to be installed on your system.

```bash
go install github.com/PRASSamin/prasmoid
```

### Updating Prasmoid

Keep Prasmoid up to date with the latest features and improvements using one of these methods:

#### 1. CLI Method (Recommended)

The simplest way to update Prasmoid is by using the built-in update command:

```bash
sudo prasmoid upgrade
```

> [!TIP]
> This command is a convenient wrapper around the manual update method. It's designed to be lightweight and efficient, avoiding the need for additional internal update logic that would increase the binary size.

#### 2. Manual Update via Curl

If you prefer a manual update, you can use curl:

```bash
sudo curl -sSL https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update | sudo bash -s $(which prasmoid)
```

#### 3. Go Upgrade

#### Upgrading Go Installations

If you installed Prasmoid using `go install`, use the following command to upgrade:

```bash
sudo env "PATH=$PATH" prasmoid upgrade
```

> [!NOTE]
> You can also use the `Manual Update via Curl` method mentioned above. However, if you want to use the CLI method, you must use the `env "PATH=$PATH"` prefix as shown.

> [!WARNING]
> The `env "PATH=$PATH"` is required because `sudo` uses a restricted PATH by default. This ensures the system can locate the Go-installed `prasmoid` binary in your user's Go bin directory.

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

| Command             | Description                                                             | Usage & Flags                                                                                                                                 |
| :------------------ | :---------------------------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup`             | Bootstraps the development environment (e.g. installs dependencies).    | `prasmoid setup`                                                                                                                              |
| `init`              | Initializes a new plasmoid project.                                     | `prasmoid init [-n <name>]` <br> `-n, --name`: Project name.                                                                                  |
| `build`             | Packages the project into a `.plasmoid` archive.                        | `prasmoid build [-o <output_dir>]` <br> `-o, --output`: Output directory (default: `./build`).                                                |
| `preview`           | Launches the plasmoid in a live preview window.                         | `prasmoid preview [-w]` <br> `-w, --watch`: Auto-restart on file changes.                                                                     |
| `format`            | Formats all `.qml` files in the `contents` directory using `qmlformat`. | `prasmoid format [-d <dir>] [-w]` <br> `-d, --dir`: Directory to format (default: `./contents`). <br> `-w, --watch`: Watch and auto-format.   |
| `link`              | Symlinks the project to KDE’s development plasmoids directory.          | `prasmoid link [-w]` <br> `-w, --where`: Show the linked path only.                                                                           |
| `unlink`            | Removes the symlink created by `link`.                                  | `prasmoid unlink`                                                                                                                             |
| `install`           | Installs the plasmoid system-wide for production use.                   | `prasmoid install`                                                                                                                            |
| `uninstall`         | Uninstalls the plasmoid from the system.                                | `prasmoid uninstall`                                                                                                                          |
| `changeset`         | Manages versioning and changelogs.                                      | See subcommands below.                                                                                                                        |
| `changeset add`     | Creates a new changeset with version bump and summary.                  | `prasmoid changeset add [-b <type>] [-s <summary>]` <br> `-b, --bump`: `patch`, `minor`, or `major`. <br> `-s, --summary`: Changelog summary. |
| `changeset apply`   | Applies pending changesets to `metadata.json` and `CHANGELOG.md`.       | `prasmoid changeset apply`                                                                                                                    |
| `command`           | Manages custom JavaScript CLI commands.                                 | See subcommands below.                                                                                                                        |
| `command add`       | Adds a new custom JS command in `.prasmoid/commands/`.                  | `prasmoid command add [-n <name>]` <br> `-n, --name`: Command name.                                                                          |
| `command remove`    | Removes a custom command.                                               | `prasmoid command remove [-n <name>]` <br> `-n, --name`: Command name.                                                                       |
| `i18n`              | Handles internationalization tasks.                                     | See subcommands below.                                                                                                                        |
| `i18n extract`      | Extracts strings for translation from metadata and QML files.           | `prasmoid i18n extract` <br> `--no-po`: Skip `.po` generation.                                                                                |
| `i18n compile`      | Compiles `.po` files into `.mo` files for use in plasmoids.             | `prasmoid i18n compile` <br> `-s, --silent`: Suppress output.                                                                                 |
| `i18n locales`      | Manages supported locales.                                              | See subcommands below.                                                                                                                        |
| `i18n locales edit` | Launches locale selector to edit supported locales.                     | `prasmoid i18n locales edit`                                                                                                                  |
| `regen`             | Regenerates config or type definition files.                            | See subcommands below.                                                                                                                        |
| `regen types`       | Regenerates `prasmoid.d.ts`.                                            | `prasmoid regen types`                                                                                                                        |
| `regen config`      | Regenerates `prasmoid.config.js`.                                       | `prasmoid regen config`                                                                                                                       |
| `upgrade`           | Updates Prasmoid itself to the latest version.                          | `prasmoid upgrade`                                                                                                                            |
| `version`           | Shows the current version of Prasmoid.                                  | `prasmoid version`                                                                                                                            |

## Extending Prasmoid with Custom Commands

Prasmoid's most powerful and unique feature is its extensibility through custom JavaScript commands. This allows you to automate any project-specific workflow directly within your CLI, without needing Node.js installed on your system.

### How it Works: The Embedded JavaScript Runtime

Prasmoid includes a lightweight, high-performance JavaScript runtime embedded directly within its Go binary. This runtime provides a Node.js-like environment, offering synchronous APIs for common modules such as `fs`, `os`, `path`, `child_process`, and a custom `prasmoid` module for CLI-specific interactions.

This means you can write powerful automation scripts in JavaScript, and Prasmoid will execute them natively, making your custom commands fast, portable, and truly zero-dependency for end-users.

### Creating a Custom Command

1.  **Generate the command file:**

    ```bash
    prasmoid command add deploy
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

---

## Contributing

We welcome contributions from the community! Whether it's bug reports, feature requests, or code contributions, your help is invaluable.

- **Report Bugs**: If you find an issue, please open a [GitHub Issue](https://github.com/PRASSamin/prasmoid/issues).
- **Suggest Features**: Have an idea for a new feature? Open an issue to discuss it.

For more information on how to contribute, see the [CONTRIBUTING.md](CONTRIBUTING.md) file.

---

## 💖 Credits

- **clorteau** – packaging – [GitHub](https://github.com/clorteau)

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
