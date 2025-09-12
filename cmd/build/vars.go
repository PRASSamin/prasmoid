package build

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/PRASSamin/prasmoid/cmd/i18n"
	"github.com/PRASSamin/prasmoid/utils"
)

var (
	utilsIsValidPlasmoid     = utils.IsValidPlasmoid
	i18nCompileI18n          = i18n.CompileI18n
	utilsGetDataFromMetadata = utils.GetDataFromMetadata
	osRemoveAll              = os.RemoveAll
	osMkdirAll               = os.MkdirAll
	osCreate                 = os.Create
	zipNewWriter             = zip.NewWriter
	osOpen                   = os.Open
	ioCopy                   = io.Copy
	filepathWalk             = filepath.Walk
)
