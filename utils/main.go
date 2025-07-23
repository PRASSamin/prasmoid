package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// Detect package manager
func DetectPackageManager() (string, error) {
	pmCommands := []string{"dnf", "apt"}
	for _, cmd := range pmCommands {
		_, err := exec.LookPath(cmd)
		if err == nil {
			return cmd, nil
		}
	}
	fmt.Println((pmCommands))
	return "", fmt.Errorf("no supported package manager found (dnf, apt)")
}


// Get package name for a binary
func GetPackageForBinary(name string) (string, error) {
    path, err := exec.LookPath(name)
    if err != nil {
        return "", fmt.Errorf("binary %s not found", name)
    }

	pm, err := DetectPackageManager()
	if err != nil {
		return "", err
	}

	if pm == "apt" {
		out, err := exec.Command("dpkg", "-S", path).Output()
		if err != nil {
			return "", fmt.Errorf("dpkg -S failed: %v", err)
		}
		parts := strings.SplitN(string(out), ":", 2)
		return strings.TrimSpace(parts[0]), nil
	} 
	
	if pm == "dnf"{
		 out, err := exec.Command("rpm", "-qf", path).Output()
    if err != nil {
        return "", fmt.Errorf("rpm -qf failed: %v", err)
    }
	return strings.TrimSpace(string(out)), nil
	}
   

    return "", fmt.Errorf("")
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


func TryInstallPackage(pkg string) (bool, error) {
	pm, err := DetectPackageManager()
	if err != nil {
		return false, err
	}

	switch pm {
	case "dnf":
		cmd := exec.Command("sudo", "dnf", "install", "-y", pkg)
		cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
		if cmd.Run() == nil {
			return true, nil
		} else {
			return false, fmt.Errorf("failed to install %s", pkg)
		}

	case "apt":
		cmd := exec.Command("sudo", "apt", "install", "-y", pkg)
		cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
		if cmd.Run() == nil {
			return true, nil
		} else {
			return false, fmt.Errorf("failed to install %s", pkg)
		}
	default:
		return false, fmt.Errorf("unsupported package manager: %s", pm)
	}
}


func InstallPackages(packages []string) error {
	pm, err := DetectPackageManager()
	highlight := color.New(color.FgGreen, color.Bold).SprintFunc()
	if err != nil {
		return err
	}

	switch pm {
	case "dnf":
        fmt.Printf("Trying to install %s with dnf...\n", highlight(strings.Join(packages, " ")))

		cmd := exec.Command("sudo", "dnf", "install", "-y", "--skip-unavailable")
		cmd.Args = append(cmd.Args, packages...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case "apt":
		for _, pkg := range packages {
			fmt.Printf("Trying to install '%s' with apt...\n", highlight(pkg))
			cmd := exec.Command("sudo", "apt", "install", "-y", pkg)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				color.Yellow("Skipping '%s': not found or failed to install.\n", pkg)
			}
		}
		return nil

	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}
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