package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
)

func TestI18nCompileCommand(t *testing.T) {
	// Set up a temporary project
	projectDir, cleanup := tests.SetupTestProject(t)
	defer cleanup()

	// Create a dummy config
	config := types.Config{
		I18n: types.ConfigI18n{
			Dir:     "translations",
			Locales: []string{"en"},
		},
	}

	// Create a dummy .po file
	poDir := filepath.Join(projectDir, config.I18n.Dir)
	_ = os.MkdirAll(poDir, 0755)
	poContent := `
msgid "Hello World"
msgstr "Hello World"
`
	if err := os.WriteFile(filepath.Join(poDir, "en.po"), []byte(poContent), 0644); err != nil {
		t.Fatalf("Failed to write en.po: %v", err)
	}

	// 2. Execute the CompileMessages function
	if err := CompileI18n(config, true); err != nil {
		t.Fatalf("CompileMessages failed: %v", err)
	}

	// 3. Verify the output
	plasmoidId, _ := utils.GetDataFromMetadata("Id")
	moFile := filepath.Join(projectDir, "contents", "locale", "en", "LC_MESSAGES", "plasma_applet_"+plasmoidId.(string)+".mo")

	if _, err := os.Stat(moFile); os.IsNotExist(err) {
		t.Fatalf("Expected .mo file to be created at %s, but it was not", moFile)
	}

	// Check that the .mo file is not empty
	info, _ := os.Stat(moFile)
	if info.Size() == 0 {
		t.Errorf("Expected .mo file to not be empty, but it was")
	}
}
