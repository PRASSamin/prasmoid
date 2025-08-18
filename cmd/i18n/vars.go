package i18n

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/PRASSamin/prasmoid/utils"
	"github.com/bmatcuk/doublestar/v4"
)

// mockable functions for testing
var (
	// os functions
	osMkdirAll  = os.MkdirAll
	osStat      = os.Stat
	osRemove    = os.Remove
	osRename    = os.Rename
	osReadFile  = os.ReadFile
	osWriteFile = os.WriteFile

	// filepath functions
	filepathGlob = filepath.Glob
	filepathWalk = filepath.Walk

	// exec functions
	execCommand = exec.Command
	execLookPath = exec.LookPath

	// other libraries
	doublestarGlob = func(fsys fs.FS, pattern string) ([]string, error) {
		return doublestar.Glob(fsys, pattern)
	}

	// utils functions
	GetDataFromMetadata = utils.GetDataFromMetadata

	// command runner
	runCommand = func(cmd *exec.Cmd) error {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("command \"%s\" failed: %v\n%s", cmd.String(), err, stderr.String())
		}
		return nil
	}
)
