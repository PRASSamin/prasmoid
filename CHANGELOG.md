# CHANGELOG

---

## [Unreleased]

#### Changes

- The command to update the tool itself has been changed, `prasmoid update me` ➜ `prasmoid upgrade`.
- The build script now automatically updates package files (`PKGBUILD`) with the correct version and file checksums to help packager.
- The installation script now shows clearer and more helpful messages.
- The project layout was changed to make sure `go install` works correctly. This involved moving command files to a root `cmd/` folder and moving the build script into a `dev/` folder.

## [v0.0.3] - 2025-07-29

#### Changes

- Added `regen` command to regenerate `prasmoid.d.ts` and `prasmoid.config.js` if lost or modified.
- **Internationalization (i18n) Support**:
  _[Suggested by this awesome community member](https://www.reddit.com/r/kde/comments/1mb9paz/comment/n5mt6tg/?utm_source=share&utm_medium=web3x&utm_name=web3xcss&utm_term=1&utm_content=share_button)_

  - `i18n extract`: Extracts translatable strings from metadata and QML files.
  - `i18n compile`: Compiles `.po` files into binary `.mo` format for use in plasmoids.
  - `i18n locales edit`: Adds or removes locales from your plasmoid.

- **Extended Configuration System**:

  - The `prasmoid.config.js` now supports:

    - `i18n.locales`: Define supported locale codes (e.g., `en`, `fr`, `pt_BR`, etc.)
    - `i18n.dir`: Specify your translation directory

- `build` command now **auto-compiles translations** before packaging.
- `init` command now allows **interactive locale selection**, auto-generating localized `metadata.json` entries.
- CLI version checking has been **refined** with smarter, more accurate version comparisons.
- Various commands have been tightened up for performance and usability.
- Renamed `deps` ➜ `consts` for cleaner architecture and more intuitive constant management.

---

## [v0.0.2] - 2025-07-28

#### Changes

- Add `setup` command to install all necessary dependencies for development environment.
- Refactor package management to support multiple package managers (apt, dnf, pacman, nix).
- Update the build process to be more flexible.
- Remove support for compressed version of the binary.

---

## [v0.0.1] - 2025-07-28

- Initial release.
