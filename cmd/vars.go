package cmd

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/PRASSamin/prasmoid/cmd/extendcli"
	"github.com/PRASSamin/prasmoid/internal"
	"github.com/PRASSamin/prasmoid/utils"
	"golang.org/x/term"
)

// mockable functions for testing
var (
	// utils
	utilsLoadConfigRC = utils.LoadConfigRC

	// extendcli
	extendcliDiscoverAndRegisterCustomCommands = extendcli.DiscoverAndRegisterCustomCommands

	// cobra
	rootCmdExecute = RootCmd.Execute

	// os
	osExit = os.Exit
	osUserCacheDir = os.UserCacheDir
	osTempDir = os.TempDir
	osReadFile = os.ReadFile
	osWriteFile = os.WriteFile

	// time
	timeParse = time.Parse
	timeSince = time.Since
	timeNow = time.Now

	// net/tls
	tlsDial = tls.Dial
	connWrite = func(conn *tls.Conn, b []byte) (n int, err error) { return conn.Write(b) }
	connClose = func(conn *tls.Conn) error { return conn.Close() }

	// io
	ioReadAll = io.ReadAll

	// encoding/json
	jsonUnmarshal = json.Unmarshal
	jsonMarshal = json.Marshal

	// golang.org/x/term
	termGetSize = term.GetSize

	// internal
	internalAppMetaDataVersion = internal.AppMetaData.Version

	// for testing purposes
	logPrintf = log.Printf
)
