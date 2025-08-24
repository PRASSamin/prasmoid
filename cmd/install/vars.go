package install

import (
	"os"

	"github.com/PRASSamin/prasmoid/utils"
)

var (
	osRemoveAll          = os.RemoveAll
	osMkdirAll           = os.MkdirAll
	osReadFile           = os.ReadFile
	osWriteFile          = os.WriteFile
	osReadDir            = os.ReadDir
	utilsIsInstalled     = utils.IsInstalled
	utilsIsValidPlasmoid = utils.IsValidPlasmoid
	utilsGetDevDest      = utils.GetDevDest
)
