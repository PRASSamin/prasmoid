package runtime

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/fatih/color"
)

// console provides console.log, console.warn, console.error, console.debug, console.info,
// console.red, console.green, console.yellow, console.color.
//
// console.log, console.warn, console.error, console.debug, console.info are plain
// loggers.
//
// console.red, console.green, console.yellow are colored loggers.
//
// console.color is a flexible color logger. It takes multiple arguments where
// the last argument is the color name and the rest of the arguments are the text
// parts.
//
// Example:
// console.color("This is a", "red", " warning!")
func Console(vm *goja.Runtime, module *goja.Object) {
	_console := module.Get("exports").(*goja.Object)

	createPlainLogger := func(prefix string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			var output []string
			for _, arg := range call.Arguments {
				output = append(output, stringifyJS(arg.Export()))
			}
			if prefix != "" {
				fmt.Println(prefix, strings.Join(output, " "))
			} else {
				fmt.Println(strings.Join(output, " "))
			}
			return goja.Undefined()
		}
	}
	
// Helper: For red, green, yellow
createColorLogger := func(colorFunc func(format string, a ...interface{})) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var parts []string
		for _, arg := range call.Arguments {
			parts = append(parts, stringifyJS(arg.Export()))
		}
		colorFunc(strings.Join(parts, " "))
		return goja.Undefined()
	}
}

// Standard logs
_console.Set("log", createPlainLogger(""))
_console.Set("warn", createPlainLogger(""))
_console.Set("error", createPlainLogger(""))
_console.Set("debug", createPlainLogger(""))
_console.Set("info", createPlainLogger(""))

// Colored logs
_console.Set("red", createColorLogger(color.Red))
_console.Set("green", createColorLogger(color.Green))
_console.Set("yellow", createColorLogger(color.Yellow))

// Flexible color log: console.color("msg", "colorName")
_console.Set("color", func(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		fmt.Println("Usage: console.color(...textParts, colorName)")
		return goja.Undefined()
	}

	colorName := fmt.Sprintf("%v", call.Arguments[len(call.Arguments)-1].Export())
	var parts []string
	for _, arg := range call.Arguments[:len(call.Arguments)-1] {
		parts = append(parts, fmt.Sprintf("%v", arg.Export()))
	}
	if len(call.Arguments) == 1 {
		colorName = "white"
		for _, arg := range call.Arguments {
			parts = append(parts, fmt.Sprintf("%v", arg.Export()))
		}
	}

	text := strings.Join(parts, " ")
	c := color.New(color.FgHiWhite)

	switch strings.ToLower(colorName) {
	case "red":
		c = color.New(color.FgHiRed)
	case "green":
		c = color.New(color.FgHiGreen)
	case "yellow":
		c = color.New(color.FgHiYellow)
	case "blue":
		c = color.New(color.FgHiBlue)
	case "magenta":
		c = color.New(color.FgHiMagenta)
	case "cyan":
		c = color.New(color.FgHiCyan)
	case "black":
		c = color.New(color.FgHiBlack)
	default:
		c = color.New(color.FgHiWhite)
	}

	c.Println(text)
	return goja.Undefined()
})
}

func stringifyJS(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("\"%s\"", val)
	case []interface{}:
		var parts []string
		for _, item := range val {
			parts = append(parts, stringifyJS(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		var parts []string
		for k, v := range val {
			parts = append(parts, fmt.Sprintf("%s: %s", k, stringifyJS(v)))
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case map[interface{}]interface{}:
		var parts []string
		for k, v := range val {
			parts = append(parts, fmt.Sprintf("%v: %s", k, stringifyJS(v)))
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case func(goja.FunctionCall) goja.Value:
		return "[Function]"
	default:
		if fn, ok := val.(fmt.Stringer); ok {
			return fn.String()
		}
		return fmt.Sprintf("%v", val)
	}
}
