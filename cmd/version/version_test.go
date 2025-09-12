package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Act
	VersionCmd.Run(VersionCmd, []string{})
	_ = w.Close()

	// Assert
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	assert.Contains(t, buf.String(), internal.AppMetaData.Version)
}
