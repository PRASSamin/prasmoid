package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/PRASSamin/prasmoid/tests"
)

func TestI18nExtractCommand(t *testing.T) {
	// Set up a temporary project
	projectDir, cleanup := tests.SetupTestProject(t)
	defer cleanup()

	// Create a dummy config file
	configContent := `
const config = {
  commands: {
    dir: ".prasmoid/commands",
    ignore: []
  },
  i18n: {
    dir: "translations",
    locales: ["en", "fr"]
  }
};
`
	if err := os.WriteFile(filepath.Join(projectDir, "prasmoid.config.js"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cmd.ConfigRC = utils.LoadConfigRC()

	// Create a dummy QML file with a translatable string
	qmlDir := filepath.Join(projectDir, "contents", "ui")
	_ = os.MkdirAll(qmlDir, 0755)
	qmlContent := `import QtQuick 2.0; Text { text: i18n("Hello World") }`
	if err := os.WriteFile(filepath.Join(qmlDir, "main.qml"), []byte(qmlContent), 0644); err != nil {
		t.Fatalf("Failed to write QML file: %v", err)
	}

	flagStates := []string{"true", "false"}
	for _, state := range flagStates {
		// Initialize the command
		if err := I18nExtractCmd.Flags().Set("no-po", state); err != nil {
			t.Errorf("Error setting flag 'no-po': %v", err)
		}

		//  Execute the i18n extract command
		I18nExtractCmd.Run(I18nExtractCmd, []string{})

		// Verify the output
		translationsDir := filepath.Join(projectDir, "translations")
		potFile := filepath.Join(translationsDir, "template.pot")
		enPoFile := filepath.Join(translationsDir, "en.po")
		frPoFile := filepath.Join(translationsDir, "fr.po")

		// Check if template.pot was created
		if _, err := os.Stat(potFile); os.IsNotExist(err) {
			t.Fatalf("Expected template.pot to be created, but it was not")
		}

		// Check if the pot file contains the extracted string
		potContent, err := os.ReadFile(potFile)
		if err != nil {
			t.Fatalf("Failed to read template.pot: %v", err)
		}

		if !strings.Contains(string(potContent), `msgid "Hello World"`) {
			t.Errorf("Expected template.pot to contain 'msgid \"Hello World\"', but it did not")
		}

		if state == "false" {
			// Check if en.po and fr.po were created
			if _, err := os.Stat(enPoFile); os.IsNotExist(err) {
				t.Errorf("Expected en.po to be created, but it was not")
			}
			if _, err := os.Stat(frPoFile); os.IsNotExist(err) {
				t.Errorf("Expected fr.po to be created, but it was not")
			}
		} else {
			if _, err := os.Stat(enPoFile); os.IsExist(err) {
				t.Errorf("Expected en.po to not be created, but it was")
			}
			if _, err := os.Stat(frPoFile); os.IsExist(err) {
				t.Errorf("Expected fr.po to not be created, but it was")
			}
		}

		// Check for backup files
		backupFiles, _ := filepath.Glob(filepath.Join(translationsDir, "*.po~"))
		if len(backupFiles) > 0 {
			t.Errorf("Found unexpected backup files: %v", backupFiles)
		}
	}
}
