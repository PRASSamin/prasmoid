package link

import (
	"os"

	"github.com/PRASSamin/prasmoid/utils"
)

var (
	osRemoveAll          = os.RemoveAll
	osGetwd              = os.Getwd
	osSymlink            = os.Symlink
	utilsIsValidPlasmoid = utils.IsValidPlasmoid
	utilsGetDevDest      = utils.GetDevDest
)
