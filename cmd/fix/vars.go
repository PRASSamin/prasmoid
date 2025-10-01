package fix

import (
	"os/exec"
	"os/user"

	"github.com/PRASSamin/prasmoid/utils"
)

var (
	execCommand          = exec.Command
	userCurrent          = user.Current
	
	utilsIsPackageInstalled = utils.IsPackageInstalled

	scriptURL = "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/scripts/fix"
)
