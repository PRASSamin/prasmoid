package runtime

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dop251/goja"
)

// CommandConfig represents the configuration for a command
type CommandConfig struct {
	Run   goja.Callable `json:"-"` // Goja function, skip JSON export
	Short string        `json:"short"`
	Long  string        `json:"long"`
	Alias []string      `json:"alias"`
	Flags []CommandFlag `json:"flags"`
}

type CommandFlag struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Value       interface{} `json:"usage"`
	Shorthand   string      `json:"shorthand"`
	Description string      `json:"description"`
}

// CommandStorage stores the command configuration globally
var CommandStorage CommandConfig

func Prasmoid(vm *goja.Runtime, module *goja.Object) {
	exports := module.Get("exports").(*goja.Object)

    exports.Set("getMetadata", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("prasmoid.getMetadata: missing key")
		}
		if len(call.Arguments) > 1 {
			return vm.ToValue("prasmoid.getMetadata: too many arguments")
		}
		data, err := GetDataFromMetadata(call.Arguments[0].String())
		if err != nil {
			return vm.ToValue("prasmoid.getMetadata: " + err.Error())
		}
		return vm.ToValue(data)
	})

	exports.Set("Command", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic("prasmoid.Command: exactly 1 argument required")
		}
	
		// Check if it's an object
		arg := call.Argument(0)
		if arg == nil || goja.IsUndefined(arg) || goja.IsNull(arg) {
			panic("prasmoid.Command: argument must be a JS object")
		}
	
		_, ok := goja.AssertFunction(arg)
		if ok {
			panic("prasmoid.Command: expected object, got function")
		}
	
		cmdObj := arg.ToObject(vm)
		if cmdObj == nil {
			panic("prasmoid.Command: failed to convert argument to object")
		}
	
		runVal := cmdObj.Get("run")
		if goja.IsUndefined(runVal) || runVal == nil {
			panic("prasmoid.Command: missing 'run' function")
		}
	
		runFunc, ok := goja.AssertFunction(runVal)
		if !ok {
			panic("prasmoid.Command: 'run' must be a function")
		}
	
		config := CommandConfig{Run: runFunc}
		
		// Optional fields
		if shortVal := cmdObj.Get("short"); !goja.IsUndefined(shortVal) && shortVal != nil {
			config.Short = shortVal.String()
		}

		if longVal := cmdObj.Get("long"); !goja.IsUndefined(longVal) && longVal != nil {
			config.Long = longVal.String()
		}

		if aliasVal := cmdObj.Get("alias"); !goja.IsUndefined(aliasVal) && aliasVal != nil {
			if aliasArr, ok := aliasVal.Export().([]interface{}); ok {
				for _, a := range aliasArr {
					if str, ok := a.(string); ok {
						config.Alias = append(config.Alias, str)
					}
				}
			}
		}


		if flagsVal := cmdObj.Get("flags"); !goja.IsUndefined(flagsVal) && flagsVal != nil {
			if flagArr, ok := flagsVal.Export().([]interface{}); ok {
				for _, f := range flagArr {
					if flagMap, ok := f.(map[string]interface{}); ok {
						typ := asString(flagMap["type"])
						  if typ == "bool" {
							switch flagMap["value"].(type){
							case bool:
							default:
								panic("non-bool value not allowed in boolean flag")
							}
						  }
						  if flagMap["value"] == nil {
							switch typ {
							case "bool":
								flagMap["value"] = false
							case "string":
								flagMap["value"] = ""
							}
						}

						config.Flags = append(config.Flags, CommandFlag{
							Name:        asString(flagMap["name"]),
							Type:        typ,
							Value:       flagMap["default"],
							Shorthand:   asString(flagMap["shorthand"]),
							Description: asString(flagMap["description"]),
						})
					}
				}
			}
		}
	
		CommandStorage = config
		return nil
	})
	
}

func asString(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
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