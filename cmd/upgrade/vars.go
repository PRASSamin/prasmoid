package upgrade

import (
	"os"
	"os/exec"
	"os/user"

	"github.com/AlecAivazis/survey/v2"
	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
)

var (
	utilsIsPackageInstalled = utils.IsPackageInstalled
	utilsDetectPackageManager = utils.DetectPackageManager
	surveyAskOne = survey.AskOne
	utilsInstallPackage = utils.InstallPackage
	osExecutable = os.Executable
	execCommand = exec.Command
	osRemove = os.Remove
	userCurrent = user.Current
	rootGetCacheFilePath = root.GetCacheFilePath

	confirmInstallation bool
	scriptURL = "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update"
)
