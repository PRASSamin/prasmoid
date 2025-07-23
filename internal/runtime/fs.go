package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)


func FS(vm *goja.Runtime, module *goja.Object) {
	_fs := module.Get("exports").(*goja.Object)

	// readFileSync(path: string) => string
	_fs.Set("readFileSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.readFileSync: missing path")
		}
		path := call.Arguments[0].String()
		data, err := os.ReadFile(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(string(data))
	})

	// writeFileSync(path: string, content: string)
	_fs.Set("writeFileSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("fs.writeFileSync: missing path or content")
		}
		path := call.Arguments[0].String()
		content := call.Arguments[1].String()

		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	})

	// appendFileSync(path: string, content: string)
	_fs.Set("appendFileSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("fs.appendFileSync: missing path or content")
		}
		path := call.Arguments[0].String()
		content := call.Arguments[1].String()

		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		defer f.Close()

		if _, err := f.WriteString(content); err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	})

	// existsSync(path: string) => boolean
	_fs.Set("existsSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.existsSync: missing path")
		}
		path := call.Arguments[0].String()
		_, err := os.Stat(path)
		exists := !os.IsNotExist(err)
		return vm.ToValue(exists)
	})

	// readdirSync(path: string) => string[]
	_fs.Set("readdirSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.readdirSync: missing path")
		}
		path := call.Arguments[0].String()
		files, err := os.ReadDir(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		
		result := vm.NewArray()
		index := 0
		for _, file := range files {
			result.Set(fmt.Sprintf("%d", index), vm.ToValue(file.Name()))
			index++
		}
		return result
	})
	
	// mkdirSync(path: string)
	_fs.Set("mkdirSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.mkdirSync: missing path")
		}
		path := call.Arguments[0].String()
		err := os.Mkdir(path, 0755)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	})
	
	rm := func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.rmSync: missing path")
		}
		path := call.Arguments[0].String()
		recursive := false
		if len(call.Arguments) > 1 {
			options := call.Arguments[1].ToObject(vm)
			
			recursiveProp := options.Get("recursive")
			if recursiveProp != nil {
				recursive = recursiveProp.ToBoolean()
			}
		}
		
		var err error
		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	}

	// rmdirSync(path: string, options?: { recursive?: boolean })
	_fs.Set("rmdirSync", rm)
	_fs.Set("rmSync", rm)

	// copyFileSync(src: string, dest: string)
	_fs.Set("copyFileSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("fs.copyFileSync: missing src or dest")
		}
		src := call.Arguments[0].String()
		dest := call.Arguments[1].String()

		// Open source file
		srcFile, err := os.Open(src)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		defer srcFile.Close()

		// Create destination file
		destFile, err := os.Create(dest)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		defer destFile.Close()

		// Copy contents
		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}

		return goja.Undefined()
	})

	// renameSync(oldPath: string, newPath: string)
	_fs.Set("renameSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("fs.renameSync: missing oldPath or newPath")
		}
		oldPath := call.Arguments[0].String()
		newPath := call.Arguments[1].String()
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	})

	// unlinkSync(path: string)
	_fs.Set("unlinkSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.unlinkSync: missing path")
		}
		path := call.Arguments[0].String()
		err := os.Remove(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	})

	_fs.Set("realpathSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.realpathSync: missing path")
		}
		path := call.Arguments[0].String()
		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(realPath)
	})


	_fs.Set("readlinkSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.readlinkSync: missing path")
		}
		path := call.Arguments[0].String()
		link, err := os.Readlink(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(link)
	})
}
