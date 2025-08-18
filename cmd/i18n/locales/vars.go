package locales

import (
	"os"

	"github.com/PRASSamin/prasmoid/utils"
)

// mockable functions for testing
var (
	utilsIsValidPlasmoid = utils.IsValidPlasmoid
	utilsAskForLocales   = utils.AskForLocales
	osWriteFile          = os.WriteFile
)
