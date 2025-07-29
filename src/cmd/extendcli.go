package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PRASSamin/prasmoid/internal/runtime"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/dop251/goja"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// DiscoverAndRegisterCustomCommands scans for Js, Go files and registers them as cobra commands.
func DiscoverAndRegisterCustomCommands(rootCmd *cobra.Command) {	
	commandsDir := ConfigRC.Commands.Dir

	// Check if command directory exists
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		return
	}

	files, err := os.ReadDir(commandsDir)
	if err != nil {
		color.Red("Error reading commands directory: %v", err)
		return
	}

	// Filter out ignored files
	var filteredFiles []os.DirEntry
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if isIgnored(file.Name(), ConfigRC.Commands.Ignore, commandsDir) {
			continue
		}
		filteredFiles = append(filteredFiles, file)
	}

	// Register commands
	for _, file := range filteredFiles {
		registerJSCommand(rootCmd, filepath.Join(commandsDir, file.Name()))
	}
}

// isIgnored checks whether a filename matches any ignore rule (direct match or glob)
func isIgnored(filename string, ignoreList []string, root string) bool {
	for _, rule := range ignoreList {
		if strings.ContainsAny(rule, "*?[") {
			// Glob pattern
			pattern := filepath.Join(root, rule)
			match, _ := doublestar.PathMatch(pattern, filepath.Join(root, filename))
			if match {
				return true
			}
		} else if rule == filename {
			// Exact filename match
			return true
		}
	}
	return false
}

type flagValues struct {
	Strings map[string]*string
	Bools   map[string]*bool
}

func registerJSCommand(rootCmd *cobra.Command, path string) {
	// Read the JS file
    src, err := os.ReadFile(path)
    if err != nil {
        color.Red("Failed to read JS script %s: %v", path, err)
        return
    }

	// Create new runtime instance
	vm := runtime.NewRuntime()

	_, err = vm.RunString(string(src))
    if err != nil {
        color.Red("Error running script:", err)
    }

	cmd := &cobra.Command{}
	flagVals := &flagValues{
		Strings: make(map[string]*string),
		Bools:   make(map[string]*bool),
	}

	command := runtime.CommandStorage
	filename := strings.TrimSpace(filepath.Base(path))
	cmd.Use = strings.ReplaceAll(strings.TrimSuffix(filename, filepath.Ext(filename)), " ", "")

	cmd.Short = command.Short
	cmd.Long = command.Long
	if len(command.Alias) > 0 {
			cmd.Aliases = make([]string, len(command.Alias))
			copy(cmd.Aliases, command.Alias)
		}

		for _, flag := range command.Flags {
			if flag.Name == "" {
				color.Yellow("Flag name is required")
				return
			}
			if flag.Type == "" {
				color.Yellow("Flag type is required")
				return
			}



			switch flag.Type {
			case "string":
				var val string
				flagVals.Strings[flag.Name] = &val
				if flag.Shorthand == "" {
					cmd.Flags().StringVar(flagVals.Strings[flag.Name], flag.Name, fmt.Sprintf("%v", flag.Value), flag.Description)
				} else {
					cmd.Flags().StringVarP(flagVals.Strings[flag.Name], flag.Name, flag.Shorthand, fmt.Sprintf("%v", flag.Value), flag.Description)
				}
			case "bool":
				var val bool
				flagVals.Bools[flag.Name] = &val
				if flag.Shorthand == "" {
					cmd.Flags().BoolVar(flagVals.Bools[flag.Name], flag.Name, flag.Value.(bool), flag.Description)
				} else {
					cmd.Flags().BoolVarP(flagVals.Bools[flag.Name], flag.Name, flag.Shorthand, flag.Value.(bool), flag.Description)
				}
			default:
				color.Yellow("Unsupported flag type: %s", flag.Type)
				return
			}
		}
		
		cmd.Run = func(cmd *cobra.Command, args []string) {	
			// Create JavaScript object for context
			ctxObj := vm.NewObject()
			
			// Add Args method
			ctxObj.Set("Args", func(call goja.FunctionCall) goja.Value {
				return vm.ToValue(args)
			})
			
			// Create flags object
			flagsObj := vm.NewObject()
			
			// Add raw flag values as properties
			for k, v := range flagVals.Strings {
				flagsObj.Set(k, vm.ToValue(*v))
			}
			for k, v := range flagVals.Bools {
				flagsObj.Set(k, vm.ToValue(*v))
			}
			
			// Add getFlag method
			flagsObj.Set("get", func(call goja.FunctionCall) goja.Value {
				name := call.Argument(0).String()
				if val, ok := flagVals.Strings[name]; ok {
					return vm.ToValue(*val)
				} else if val, ok := flagVals.Bools[name]; ok {
					return vm.ToValue(*val)
				}
				return goja.Undefined()
			})
			
			// Add Flags method
			ctxObj.Set("Flags", func(call goja.FunctionCall) goja.Value {
				return flagsObj
			})
			
			// Pass the context to the JS function
			_, err := command.Run(goja.Undefined(), ctxObj)
			if err != nil {
				color.Red("JS command error (%s): %v", path, err)
			}
			runtime.EventLoop.Wait()
		}

		cmd.GroupID = "custom"
		rootCmd.AddCommand(cmd)
}