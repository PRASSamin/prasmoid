package extendcli

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	osStat              = os.Stat
	osReadDir           = os.ReadDir
	osReadFile          = os.ReadFile
	osMkdirAll          = os.MkdirAll
	osWriteFile         = os.WriteFile
	osRemoveAll         = os.RemoveAll
	osCreateTemp        = os.CreateTemp
	filepathJoin        = filepath.Join
	doublestarPathMatch = doublestar.PathMatch
)
