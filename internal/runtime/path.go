package runtime

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)


func Path(vm *goja.Runtime, module *goja.Object) {
	_path := module.Get("exports").(*goja.Object)

	// path.resolve(...paths)
	_path.Set("resolve", func(call goja.FunctionCall) goja.Value {
		parts := []string{}
		for _, arg := range call.Arguments {
			parts = append(parts, arg.String())
		}
		return vm.ToValue(filepath.Join(parts...))
	})

	// path.normalize(path)
	_path.Set("normalize", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		return vm.ToValue(filepath.Clean(call.Arguments[0].String()))
	})

	// path.isAbsolute(path)
	_path.Set("isAbsolute", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(false)
		}
		return vm.ToValue(filepath.IsAbs(call.Arguments[0].String()))
	})

	// path.join(...paths)
	_path.Set("join", func(call goja.FunctionCall) goja.Value {
		parts := []string{}
		for _, arg := range call.Arguments {
			parts = append(parts, arg.String())
		}
		return vm.ToValue(filepath.Join(parts...))
	})

	// path.relative(from, to)
	_path.Set("relative", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("path.relative: needs from and to")
		} 
		from := call.Arguments[0].String()
		to := call.Arguments[1].String()
		
		// Convert both paths to absolute paths
		absFrom, err1 := filepath.Abs(from)
		absTo, err2 := filepath.Abs(to)
		if err1 != nil || err2 != nil {
			return vm.ToValue(fmt.Sprintf("path.relative error: %v", err1))
		}
		
		rel, err := filepath.Rel(absFrom, absTo)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("path.relative error: %v", err))
		}
		return vm.ToValue(rel)
	})

	// path.toNamespacedPath(path)
	_path.Set("toNamespacedPath", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		p := call.Arguments[0].String()
		return vm.ToValue(filepath.Clean(p))
	})

	// path.dirname(path)
	_path.Set("dirname", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		return vm.ToValue(filepath.Dir(call.Arguments[0].String()))
	})

	// path.basename(path)
	_path.Set("basename", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		return vm.ToValue(filepath.Base(call.Arguments[0].String()))
	})

	// path.extname(path)
	_path.Set("extname", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		return vm.ToValue(filepath.Ext(call.Arguments[0].String()))
	})

	// path.parse(path)
	_path.Set("parse", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(nil)
		}
		p := call.Arguments[0].String()
		root := "/" 
		p = filepath.Clean(p) 
		if !strings.HasPrefix(p, "/") {
			root = "" 
		}
		dir := filepath.Dir(p)
		base := filepath.Base(p)
		ext := filepath.Ext(p)
		name := strings.TrimSuffix(base, ext)
		
		result := vm.NewObject()
		result.Set("root", root)
		result.Set("dir", dir)
		result.Set("base", base)
		result.Set("ext", ext)
		result.Set("name", name)
		return result
	})

	// path.matchesGlob(path, pattern)
	_path.Set("matchesGlob", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("path.matchesGlob: missing path or pattern")
		}
		p := call.Arguments[0].String()
		pattern := call.Arguments[1].String()
		match, err := filepath.Match(pattern, filepath.Base(p))
		if err != nil {
			return vm.ToValue(fmt.Sprintf("glob error: %v", err))
		}
		return vm.ToValue(match)
	})


	_path.Set("format", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue("Error: path.format not implemented")
	})
}