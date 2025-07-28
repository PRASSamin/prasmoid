package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PRASSamin/prasmoid/deps"
	"github.com/PRASSamin/prasmoid/internal/runtime"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/fatih/color"
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
	return filepath.Join(os.Getenv("HOME"), ".local/share/plasma/plasmoids", id), nil
}

// Get metadata from metadata.json
func GetDataFromMetadata(key string) (string, error) {
	data, err := os.ReadFile("metadata.json")
	if err != nil {
		return "", fmt.Errorf("metadata.json not found")
	}
	var meta map[string]interface{}
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return "", err
	}
	if plugin, ok := meta["KPlugin"].(map[string]interface{}); ok {
		if id, ok := plugin[key].(string); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("%s not found in metadata.json", key)
}

// Check if plasmoid is installed
func IsInstalled() (bool, string, error) {
	id, err := GetDataFromMetadata("Id")
	if err != nil {
		return false, "", err
	}

	// Check user directory
	userPath := filepath.Join(os.Getenv("HOME"), ".local/share/plasma/plasmoids", id)
	if _, err := os.Stat(userPath); err == nil {
		return true, userPath, nil
	}

	// Check system directory
	systemPath := filepath.Join("/usr/share/plasma/plasmoids", id)
	if _, err := os.Stat(systemPath); err == nil {
		return true, systemPath, nil
	}

	return false, userPath, nil // userPath is default
}

// Check if package is installed
func IsPackageInstalled(name string) bool {
	_, err := exec.LookPath(name)
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
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to update metadata.json")
	}

	return nil
}

var supportedPackageManagers = map[string]string{
	"apt":    "apt",
	"dnf":    "dnf",
	"pacman": "pacman",
	"nix-env": "nix",
}

// Detect package manager
func DetectPackageManager() (string, error) {
	for binary, pm := range supportedPackageManagers {
		_, err := exec.LookPath(binary)
		if err == nil {
			return pm, nil
		}
	}
	return "", fmt.Errorf("no supported package manager found: %+v", supportedPackageManagers)
}

func GetBinPath() (string, error) {
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

func InstallPackage(pm, binName string, pkgNames map[string]string) error {
	binPath, err := GetBinPath()
	if err != nil {
		return fmt.Errorf("failed to get bin path: %v", err)
	}

	pkgName, ok := pkgNames[pm]
	if !ok {
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	var cmd *exec.Cmd
	switch pm {
	case "nix":
		cmd = exec.Command("nix-env", "-iA", pkgName)
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", pkgName)
	default:
		cmd = exec.Command("sudo", pm, "install", "-y", pkgName)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		color.Yellow("Warning: install command exited with error: %v", err)
	}

	if err := ensureBinaryLinked(binName, binPath); err != nil {
		return err
	}

	color.Green("%s installed!", binName)
	return nil
}

// ensureBinaryLinked looks for a binary and symlinks it into our binPath.
func ensureBinaryLinked(binName, binPath string) error {
	if _, err := exec.LookPath(binName); err == nil {
		return nil // already found in PATH
	}

	color.Yellow("Binary %s not in PATH, searching manually...", binName)
	findCmd := exec.Command("sudo", "find", "/", "-type", "f", "-name", binName)
	out, err := findCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to locate %s binary: %v", binName, err)
	}

	path := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	if path == "" {
		return fmt.Errorf("%s not found on system", binName)
	}

	link := filepath.Join(binPath, binName)
	if _, err := os.Lstat(link); err == nil {
		color.Yellow("Warning: symlink already exists at %s, skipping...", link)
		return nil
	}

	if err := os.Symlink(path, link); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

func InstallQmlformat(pm string) error {
	return InstallPackage(pm, deps.QmlFormatPackageName["binary"], deps.QmlFormatPackageName)
}

func InstallPlasmoidPreview(pm string) error {
	return InstallPackage(pm, deps.PlasmoidPreviewPackageName["binary"], deps.PlasmoidPreviewPackageName)
}


func IsValidPlasmoid() bool {
	if _, err := os.Stat("metadata.json"); err != nil {
		return false
	}
	if _, err := os.Stat("contents"); err != nil {
		return false
	}
	return true
}


func LoadConfigRC() (types.Config, error) {
    var configFileName string = "prasmoid.config.js"
	
	defaultConfig := types.Config{
		Commands: types.ConfigCommands{
			Dir: ".prasmoid/commands",
			DefaultRT: "js",
		},
	}

    data, err := os.ReadFile(configFileName)
    if err != nil {
        return defaultConfig, err
    }
    vm := runtime.NewRuntime()
    _, err = vm.RunString(string(data))
    if err != nil {
        return defaultConfig, err
    }
    config := vm.Get("config")
    if config == nil {
        return defaultConfig, fmt.Errorf("config not found in %s", configFileName)
    }
    
    // Convert to JSON bytes
    configBytes, err := json.Marshal(config.Export())
    if err != nil {
        return defaultConfig, fmt.Errorf("failed to marshal config: %v", err)
    }
    
    // Unmarshal into Config struct
    var result types.Config
    if err := json.Unmarshal(configBytes, &result); err != nil {
        return defaultConfig, fmt.Errorf("failed to unmarshal config: %v", err)
    }
    
    return result, nil
}