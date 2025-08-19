package preview

import (
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/cmd/link"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fsnotify/fsnotify"
)

// To enable mocking
var (
	// os/exec
	execCommand        = exec.Command

	// utils
	utilsIsValidPlasmoid      = utils.IsValidPlasmoid
	utilsIsLinked             = utils.IsLinked
	utilsGetDevDest           = utils.GetDevDest
	utilsIsPackageInstalled   = utils.IsPackageInstalled
	utilsDetectPackageManager = utils.DetectPackageManager
	utilsInstallPackage       = utils.InstallPackage
	utilsGetDataFromMetadata  = utils.GetDataFromMetadata
	utilsIsQmlFile            = utils.IsQmlFile

	// link
	linkLinkPlasmoid = link.LinkPlasmoid

	// survey
	surveyAskOne = survey.AskOne

	// fsnotify
	fsnotifyNewWatcher = fsnotify.NewWatcher
	currentViewer *exec.Cmd

	// filepath
	filepathWalk = filepath.Walk

	// os/signal
	signalNotify = signal.Notify

	// time
	timeAfterFunc = time.AfterFunc

	// confirmation
	confirmInstallation bool
	confirmLink bool
)
