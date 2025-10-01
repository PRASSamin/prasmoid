package init

import (
	"encoding/json"
	"os"
	"os/exec"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/utils"
)

// mockable functions for testing
var (
	// survey
	surveyAskOne = survey.AskOne
	surveyAsk    = survey.Ask

	// os
	osReadDir   = os.ReadDir
	osStat      = os.Stat
	osGetwd     = os.Getwd
	osMkdirAll  = os.MkdirAll
	osWriteFile = os.WriteFile
	osSymlink   = os.Symlink
	osRemoveAll = os.RemoveAll

	// exec
	execCommand = exec.Command

	// utils
	utilsIsPackageInstalled  = utils.IsPackageInstalled
	utilsAskForLocales       = utils.AskForLocales
	
	// json
	jsonMarshalIndent = json.MarshalIndent

	// text/template
	templateNew = template.New
)
