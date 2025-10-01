package fix

import (
	"os/exec"

	"github.com/PRASSamin/prasmoid/utils"
)

var (
	execCommand          = exec.Command
	
	utilsIsPackageInstalled = utils.IsPackageInstalled
	utilsCheckRoot = utils.CheckRoot

	scriptURL = "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/scripts/fix"
)
