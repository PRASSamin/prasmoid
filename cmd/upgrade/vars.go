package upgrade

import (
	"os"
	"os/exec"
	"os/user"

	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
)

var (
	osExecutable         = os.Executable
	execCommand          = exec.Command
	osRemove             = os.Remove
	userCurrent          = user.Current
	rootGetCacheFilePath = root.GetCacheFilePath
	
	utilsIsPackageInstalled = utils.IsPackageInstalled

	scriptURL = "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update"
)
