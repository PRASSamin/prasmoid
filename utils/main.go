package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/internal/runtime"
	"github.com/PRASSamin/prasmoid/types"
)

var (
	surveyAskOne = survey.AskOne
	execLookPath = exec.LookPath
	userCurrent  = user.Current
)

// check if plasmoid is linked
func IsLinked() bool {
	dest, err := GetDevDest()
	if err != nil {
		return false
	}
	_, err = os.Stat(dest)
	return err == nil
}

// Get development destination path
func GetDevDest() (string, error) {
	id, err := GetDataFromMetadata("Id")

	if err != nil {
		return "", err
	}
	baseDir := filepath.Join(os.Getenv("HOME"), ".local/share/plasma/plasmoids")
	dest := filepath.Join(baseDir, id.(string))

	// Create only the parent directory (plasmoids) if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directory: %v", err)
	}

	return dest, nil
}

// GetDataFromMetadata reads metadata.json and returns the requested key's value from the KPlugin section
func GetDataFromMetadata(key string) (interface{}, error) {
	const fileName = "metadata.json"

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fileName, err)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", fileName, err)
	}

	kpluginRaw, ok := meta["KPlugin"]
	if !ok {
		return nil, fmt.Errorf("KPlugin section not found in %s", fileName)
	}

	kplugin, ok := kpluginRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("KPlugin section has unexpected structure in %s", fileName)
	}

	value, ok := kplugin[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' not found in KPlugin", key)
	}

	return value, nil
}

// Check if plasmoid is installed
func IsInstalled() (bool, string, error) {
	id, err := GetDataFromMetadata("Id")
	if err != nil {
		return false, "", err
	}

	// Check user directory
	userPath := filepath.Join(os.Getenv("HOME"), ".local/share/plasma/plasmoids", id.(string))
	if _, err := os.Stat(userPath); err == nil {
		return true, userPath, nil
	}

	// Check system directory
	systemPath := filepath.Join("/usr/share/plasma/plasmoids", id.(string))
	if _, err := os.Stat(systemPath); err == nil {
		return true, systemPath, nil
	}

	return false, userPath, nil // userPath is default
}

// Check if package is installed
var IsPackageInstalled = func(name string) bool {
	_, err := execLookPath(name)
	return err == nil
}

// UpdateMetadata updates a key in the metadata.json file.
//
// Parameters:
//   - key: The key inside the section to update or create.
//   - value: The new value to assign to the key.
//   - section: The top-level section in metadata.json where the key is located.
//     By default, this is "KPlugin", which is the standard location for Plasmoid metadata.
//     If you're updating a key at the root level of the JSON, pass "." as the section.
//
// Behavior:
//   - If the section does not exist, it will be created automatically.
//   - If the file cannot be read or parsed, a descriptive error will be returned.
//   - The resulting JSON will be indented for readability.
func UpdateMetadata(key string, value interface{}, sectionOpt ...string) error {
	section := "KPlugin"
	if len(sectionOpt) > 0 {
		section = sectionOpt[0]
	}

	const path = "metadata.json"

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read metadata.json(make sure you're in the correct directory)")
	}

	// Parse the JSON into a generic map
	var meta map[string]interface{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed to parse metadata.json: %w", err)
	}

	// Handle root-level update
	if section == "." {
		meta[key] = value
	} else {
		// Ensure section exists
		sectionMap, ok := meta[section].(map[string]interface{})
		if !ok {
			sectionMap = make(map[string]interface{})
			meta[section] = sectionMap
		}
		sectionMap[key] = value
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to update metadata.json")
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log the error, but don't return it as the function already returns nil
			// or a specific error from encoder.Encode
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to update metadata.json")
	}

	return nil
}

var GetBinPath = func() (string, error) {
	defaultCandidates := []string{
		"/usr/bin",
		"/usr/local/bin",
		"/bin",
		"/nix/var/nix/profiles/default/bin",
	}

	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, ":")

	for _, candidate := range defaultCandidates {
		for _, p := range paths {
			if p == candidate {
				if info, err := os.Stat(p); err == nil && info.IsDir() {
					return p, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no supported bin path found: %+v", defaultCandidates)
}

func IsValidPlasmoid() bool {
	if _, err := os.Stat("metadata.json"); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat("contents"); os.IsNotExist(err) {
		return false
	}
	return true
}

func EnsureStringAndValid(name string, value interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s value is not a string", name)
	}
	return str, nil
}

func LoadConfigRC() types.Config {
	var configFileName = "prasmoid.config.js"
	defaultConfig := types.Config{
		Commands: types.ConfigCommands{
			Dir: ".prasmoid/commands",
		},
		I18n: types.ConfigI18n{
			Dir:     "translations",
			Locales: []string{"en"},
		},
	}

	data, err := os.ReadFile(configFileName)
	if err != nil {
		return defaultConfig
	}
	vm := runtime.NewRuntime()
	_, err = vm.RunString(string(data))
	if err != nil {
		return defaultConfig
	}
	config := vm.Get("config")
	if config == nil {
		return defaultConfig
	}

	// Convert to JSON bytes
	configBytes, err := json.Marshal(config.Export())
	if err != nil {
		return defaultConfig
	}

	// Unmarshal into Config struct
	var result types.Config
	if err := json.Unmarshal(configBytes, &result); err != nil {
		return defaultConfig
	}

	return result
}

func AskForLocales(defaultLocales ...[]string) []string {
	var defaults []string
	if len(defaultLocales) > 0 {
		defaults = defaultLocales[0]
	} else {
		defaults = []string{"en"}
	}

	// Build options: map shortcode ➝ full name
	localeOptions := []string{}
	for code, name := range consts.KDELocales {
		localeOptions = append(localeOptions, fmt.Sprintf("%s (%s)", code, name))
	}
	sort.Strings(localeOptions)

	// Convert passed defaults like "en" ➝ "en (English)"
	defaultDisplay := make([]string, 0, len(defaults))
	for _, code := range defaults {
		if name, ok := consts.KDELocales[code]; ok {
			defaultDisplay = append(defaultDisplay, fmt.Sprintf("%s (%s)", code, name))
		}
	}

	var selectedWithNames []string

	localeQuestion := &survey.MultiSelect{
		Message: "Select locales:",
		Options: localeOptions,
		Default: defaultDisplay,
	}

	if err := surveyAskOne(localeQuestion, &selectedWithNames, survey.WithValidator(func(ans interface{}) error {
		selected, ok := ans.([]core.OptionAnswer)
		if !ok {
			return fmt.Errorf("invalid type for answer")
		}
		if len(selected) == 0 {
			return fmt.Errorf("you must select at least one locale")
		}
		return nil
	})); err != nil {
		return nil
	}

	// Extract shortcodes only, in quoted form
	locales := make([]string, 0, len(selectedWithNames))
	for _, full := range selectedWithNames {
		if parts := strings.SplitN(full, " ", 2); len(parts) > 0 {
			locales = append(locales, parts[0])
		}
	}

	return locales
}

func IsQmlFile(filename string) bool {
	return strings.HasSuffix(filename, ".qml")
}

var CheckRoot = func() error {
	currentUser, err := userCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	if currentUser.Uid != "0" {
		return fmt.Errorf("the requested operation requires superuser privileges. use `sudo %s`", strings.Join(os.Args[0:], " "))
	}
	return nil
}
