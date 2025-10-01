package upgrade

import (
	"os"
	"os/exec"

	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
)

var (
	osExecutable         = os.Executable
	execCommand          = exec.Command
	osRemove             = os.Remove
	rootGetCacheFilePath = root.GetCacheFilePath
	
	utilsCheckRoot       = utils.CheckRoot
	utilsIsPackageInstalled = utils.IsPackageInstalled

	scriptURL = "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update"
)
