# Dependency Installation Removal Plan

This document lists all the files that currently contain logic for automatically installing external dependencies. As requested, this logic will be removed. The `setup` command will be refactored to check for these dependencies and report their status to the user.

## Files with Installation Logic

The following files contain calls to `utils.IsPackageInstalled`, `utils.DetectPackageManager`, or `utils.InstallPackage` and will be modified:

- **`cmd/format/format.go`**

  - _Dependency:_ `qmlformat`
  - _Action:_ Remove automatic installation prompt.

- **`cmd/i18n/compile.go`**

  - _Dependency:_ `gettext`
  - _Action:_ Remove automatic installation prompt.

- **`cmd/i18n/extract.go`**

  - _Dependency:_ `gettext`
  - _Action:_ Remove automatic installation prompt.

- **`cmd/init/init.go`**

  - _Dependency:_ `git`
  - _Action:_ Remove the check for `git`. The command will just attempt to run `git init` and fail if it's not present, which is standard behavior.

- **`cmd/preview/preview.go`**

  - _Dependency:_ `plasmoidviewer`
  - _Action:_ Remove automatic installation prompt.

- **`cmd/upgrade/upgrade.go`**

  - _Dependency:_ `curl`
  - _Action:_ Remove automatic installation prompt.

- **`cmd/setup/setup.go`**
  - _Dependency:_ All development dependencies (e.g., `plasmoidviewer`).
  - _Action:_ This command will be completely refactored. Instead of installing dependencies, it will check for their presence and print a helpful status report.

## External Dependencies

- `qmlformat`

  - Used by: prasmoid format
  - Purpose: To format and prettify QML source files.

- `gettext` (provides xgettext, msgfmt, msginit, msgmerge)

  - Used by: prasmoid i18n
  - Purpose: To extract translatable strings from source code and compile translation files.

- `plasmoidviewer`

  - Used by: prasmoid preview
  - Purpose: To launch and display the plasmoid in a preview window for development.

- `curl`

  - Used by: prasmoid upgrade
  - Purpose: To download the update script for the CLI.

- `nano`
  - Used by: prasmoid changeset add
  - Purpose: To allow the user to write a changelog summary. It respects the $EDITOR environment variable.
