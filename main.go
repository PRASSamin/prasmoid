/*
Copyright © 2025 PRAS
*/
package main

import (
	"github.com/PRASSamin/prasmoid/cmd"
	_ "github.com/PRASSamin/prasmoid/cmd/build"
	_ "github.com/PRASSamin/prasmoid/cmd/changeset"
	_ "github.com/PRASSamin/prasmoid/cmd/command"
	_ "github.com/PRASSamin/prasmoid/cmd/format"
	_ "github.com/PRASSamin/prasmoid/cmd/i18n"
	_ "github.com/PRASSamin/prasmoid/cmd/i18n/locales"
	_ "github.com/PRASSamin/prasmoid/cmd/init"
	_ "github.com/PRASSamin/prasmoid/cmd/install"
	_ "github.com/PRASSamin/prasmoid/cmd/link"
	_ "github.com/PRASSamin/prasmoid/cmd/preview"
	_ "github.com/PRASSamin/prasmoid/cmd/regen"
	_ "github.com/PRASSamin/prasmoid/cmd/uninstall"
	_ "github.com/PRASSamin/prasmoid/cmd/unlink"
	_ "github.com/PRASSamin/prasmoid/cmd/upgrade"
)

func main() {
	cmd.Execute()
}
