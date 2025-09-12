package regen

import (
	initCmd "github.com/PRASSamin/prasmoid/cmd/init"
	"github.com/PRASSamin/prasmoid/utils"
)

var (
	utilsAskForLocales            = utils.AskForLocales
	initCmdCreateConfigFile       = initCmd.CreateConfigFile
	initCmdCreateFileFromTemplate = initCmd.CreateFileFromTemplate
)
