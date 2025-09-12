package command

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
)

// mockable functions for testing
var (
	// survey
	surveyAskOne = survey.AskOne

	// os
	osStat      = os.Stat
	osMkdirAll  = os.MkdirAll
	osGetwd     = os.Getwd
	osRemove    = os.Remove
	osWriteFile = os.WriteFile

	// filepath
	filepathAbs  = filepath.Abs
	filepathRel  = filepath.Rel
	filepathWalk = filepath.Walk

	// regexp
	regexpMustCompile = regexp.MustCompile

	invalidChars = regexpMustCompile(`[\\/:*?"<>|\s@]`)
)
