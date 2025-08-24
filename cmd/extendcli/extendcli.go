/*
Copyright 2025 PRAS
*/
package extendcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/internal/runtime"
	"github.com/PRASSamin/prasmoid/types"
)

// DiscoverAndRegisterCustomCommands scans for Js, Go files and registers them as cobra commands.
func DiscoverAndRegisterCustomCommands(rootCmd *cobra.Command, ConfigRC types.Config) {
	commandsDir := ConfigRC.Commands.Dir

	// Check if command directory exists
	if _, err := osStat(commandsDir); os.IsNotExist(err) {
		return
	}

	files, err := osReadDir(commandsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error reading commands directory: %v", err))
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
		_ = registerJSCommand(rootCmd, filepath.Join(commandsDir, file.Name()))
	}
}

// isIgnored checks whether a filename matches any ignore rule (direct match or glob)
func isIgnored(filename string, ignoreList []string, root string) bool {
	for _, rule := range ignoreList {
		if strings.ContainsAny(rule, "*?[") {
			// Glob pattern
			pattern := filepath.Join(root, rule)
			match, _ := doublestarPathMatch(pattern, filepath.Join(root, filename))
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

var registerJSCommand = func(rootCmd *cobra.Command, path string) error {
	// Read the JS file
	src, err := osReadFile(path)
	if err != nil {
		fmt.Println(color.RedString("Failed to read JS script %s: %v", path, err))
		return fmt.Errorf("failed to read JS script %s: %v", path, err)
	}

	// Create new runtime instance
	vm := runtime.NewRuntime()

	_, err = vm.RunString(string(src))
	if err != nil {
		fmt.Println(color.RedString("Error running script: %v", err))
		return fmt.Errorf("error running script: %v", err)
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
			fmt.Println(color.YellowString("Flag name is required"))
			return fmt.Errorf("flag name is required")
		}
		if flag.Type == "" {
			fmt.Println(color.YellowString("Flag type is required"))
			return fmt.Errorf("flag type is required")
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
			fmt.Fprintln(os.Stderr, color.RedString("Unsupported flag type: %s", flag.Type))
			return fmt.Errorf("unsupported flag type: %s", flag.Type)
		}
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Create JavaScript object for context
		ctxObj := vm.NewObject()

		// Add Args method
		_ = ctxObj.Set("Args", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(args)
		})

		// Create flags object
		flagsObj := vm.NewObject()

		// Add raw flag values as properties
		for k, v := range flagVals.Strings {
			_ = flagsObj.Set(k, vm.ToValue(*v))
		}
		for k, v := range flagVals.Bools {
			_ = flagsObj.Set(k, vm.ToValue(*v))
		}

		// Add getFlag method
		_ = flagsObj.Set("get", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			if val, ok := flagVals.Strings[name]; ok {
				return vm.ToValue(*val)
			} else if val, ok := flagVals.Bools[name]; ok {
				return vm.ToValue(*val)
			}
			return goja.Undefined()
		})

		// Add Flags method
		_ = ctxObj.Set("Flags", func(call goja.FunctionCall) goja.Value {
			return flagsObj
		})

		// Pass the context to the JS function
		_, err := command.Run(goja.Undefined(), ctxObj)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("JS command error (%s): %v", path, err))
		}
		runtime.EventLoop.Wait()
	}

	cmd.GroupID = "custom"
	rootCmd.AddCommand(cmd)
	return nil
}
