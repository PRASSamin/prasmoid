# CHANGELOG

---

## [v0.0.4] - 2025-09-12

#### Changed

- Changed `prasmoid update me` to `prasmoid upgrade` for better clarity and consistency
- Renamed subcommand `commands` to `command`. Custom commands are now managed via `prasmoid command [add|remove]`
- Improved installation script with clearer and more helpful user feedback
- Reorganized project layout to support `go install` by moving command files to `cmd/` and build scripts to `dev/`
- Fixed handling of default values for command flags
- Enhanced string formatting and improved handling of complex data types in logs
- Streamlined environment setup by removing unnecessary error handling
- Refactored i18n commands to use mockable functions for better testability
- Enhanced initialization functionality
- Added new `dev` command group for development-related utilities

## [v0.0.3] - 2025-07-29

#### Changed

- **Regeneration Command**: Added `regen` command to regenerate `prasmoid.d.ts` and `prasmoid.config.js` if lost or modified
- **Internationalization (i18n) Support**:
  _[Suggested by this awesome community member](https://www.reddit.com/r/kde/comments/1mb9paz/comment/n5mt6tg/?utm_source=share&utm_medium=web3x&utm_name=web3xcss&utm_term=1&utm_content=share_button)_

  - **i18n Extraction**: Extracts translatable strings from metadata and QML files
  - **i18n Compilation**: Compiles `.po` files into binary `.mo` format for use in plasmoids
  - **i18n Locale Management**: Adds or removes locales from your plasmoid

- **Extended Configuration System**:

  - The `prasmoid.config.js` now supports:

    - **i18n Locales**: Define supported locale codes (e.g., `en`, `fr`, `pt_BR`, etc.)
    - **i18n Directory**: Specify your translation directory

- **Build Command**: Now auto-compiles translations before packaging
- **Init Command**: Now allows interactive locale selection, auto-generating localized `metadata.json` entries
- **Version Checking**: Refined CLI version checking with smarter, more accurate version comparisons
- **Performance and Usability**: Various commands have been tightened up for performance and usability
- **Constant Management**: Renamed `deps` to `consts` for cleaner architecture and more intuitive constant management

---

## [v0.0.2] - 2025-07-28

#### Changed

- **Setup Command**: Added `setup` command to install all necessary dependencies for development environment
- **Package Management**: Refactored package management to support multiple package managers (apt, dnf, pacman, nix)
- **Build Process**: Updated the build process to be more flexible
- **Binary Format**: Removed support for compressed version of the binary

---

## [v0.0.1] - 2025-07-28

- **Initial Release**: Initial release of the project

