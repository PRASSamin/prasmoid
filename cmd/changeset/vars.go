package changeset

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/utils"
)

var (
	utilsIsValidPlasmoid     = utils.IsValidPlasmoid
	utilsGetDataFromMetadata = utils.GetDataFromMetadata
	utilsUpdateMetadata      = utils.UpdateMetadata
	surveyAskOne             = survey.AskOne
	osMkdirAll               = os.MkdirAll
	osWriteFile              = os.WriteFile
	osReadFile               = os.ReadFile
	osRemove                 = os.Remove
	osStat                   = os.Stat
	osGetenv                 = os.Getenv
	osCreateTemp             = os.CreateTemp
	execCommand              = exec.Command
	filepathWalk             = filepath.Walk
)
