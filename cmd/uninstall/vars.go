package uninstall

import (
	"os"

	"github.com/PRASSamin/prasmoid/utils"
)

var (
	utilsIsValidPlasmoid = utils.IsValidPlasmoid
	utilsIsInstalled     = utils.IsInstalled
	osRemoveAll          = os.RemoveAll
)
