package runtime

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dop251/goja"
	"github.com/fsnotify/fsnotify"
)

func FS(vm *goja.Runtime, module *goja.Object) {
	_fs := module.Get("exports").(*goja.Object)
	var (
		fileWatchers   = make(map[string]*fsnotify.Watcher)
		watchCallbacks = make(map[string][]goja.Callable)

		filePollers   = make(map[string]chan struct{})
		pollCallbacks = make(map[string]map[*goja.Object]goja.Callable)
	)

	_, err := vm.RunString(`
        class Stats {
            constructor(data) {
                Object.assign(this, data)
            }
            isFile() { return this.mode && (this.mode & 0o170000) === 0o100000 }
            isDirectory() { return this.mode && (this.mode & 0o170000) === 0o040000 }
            isSymbolicLink() { return this.mode && (this.mode & 0o170000) === 0o120000 }
            isSocket() { return this.mode && (this.mode & 0o170000) === 0o140000 }
            isFIFO() { return this.mode && (this.mode & 0o170000) === 0o010000 }
            isCharacterDevice() { return this.mode && (this.mode & 0o170000) === 0o020000 }
            isBlockDevice() { return this.mode && (this.mode & 0o170000) === 0o060000 }
        }
    `)
	if err != nil {
		panic(err)
	}

	// Then set it in the module
	statsCtor := vm.Get("Stats").ToObject(vm)
	if err := _fs.Set("Stats", statsCtor); err != nil {
		fmt.Printf("Error setting Stats on _fs: %v\n", err)
	}

	// readFileSync(path: string) => string
	if err := _fs.Set("readFileSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.readFileSync: missing path")
		}
		path := call.Arguments[0].String()
		data, err := os.ReadFile(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(string(data))
	}); err != nil {
		fmt.Printf("Error setting fs.readFileSync: %v\n", err)
	}

	// writeFileSync(path: string, content: string)
	if err := _fs.Set("writeFileSync", func(call goja.FunctionCall) goja.Value {
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
	}); err != nil {
		fmt.Printf("Error setting fs.writeFileSync: %v\n", err)
	}

	// appendFileSync(path: string, content: string)
	if err := _fs.Set("appendFileSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("fs.appendFileSync: missing path or content")
		}
		path := call.Arguments[0].String()
		content := call.Arguments[1].String()

		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Printf("Error closing file in appendFileSync: %v\n", err)
			}
		}()

		if _, err := f.WriteString(content); err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.appendFileSync: %v\n", err)
	}

	// existsSync(path: string) => boolean
	if err := _fs.Set("existsSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.existsSync: missing path")
		}
		path := call.Arguments[0].String()
		_, err := os.Stat(path)
		exists := !os.IsNotExist(err)
		return vm.ToValue(exists)
	}); err != nil {
		fmt.Printf("Error setting fs.existsSync: %v\n", err)
	}

	// readdirSync(path: string) => string[]
	if err := _fs.Set("readdirSync", func(call goja.FunctionCall) goja.Value {
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
			if err := result.Set(fmt.Sprintf("%d", index), vm.ToValue(file.Name())); err != nil {
				fmt.Printf("Error setting result in readdirSync: %v\n", err)
			}
			index++
		}
		return result
	}); err != nil {
		fmt.Printf("Error setting fs.readdirSync: %v\n", err)
	}

	// mkdirSync(path: string)
	if err := _fs.Set("mkdirSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.mkdirSync: missing path")
		}
		path := call.Arguments[0].String()
		err := os.Mkdir(path, 0755)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.mkdirSync: %v\n", err)
	}

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
	if err := _fs.Set("rmdirSync", rm); err != nil {
		fmt.Printf("Error setting fs.rmdirSync: %v\n", err)
	}
	if err := _fs.Set("rmSync", rm); err != nil {
		fmt.Printf("Error setting fs.rmSync: %v\n", err)
	}

	// copyFileSync(src: string, dest: string)
	if err := _fs.Set("copyFileSync", func(call goja.FunctionCall) goja.Value {
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
		defer func() {
			if err := srcFile.Close(); err != nil {
				fmt.Printf("Error closing source file in copyFileSync: %v\n", err)
			}
		}()

		// Create destination file
		destFile, err := os.Create(dest)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		defer func() {
			if err := destFile.Close(); err != nil {
				fmt.Printf("Error closing destination file in copyFileSync: %v\n", err)
			}
		}()

		// Copy contents
		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}

		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.copyFileSync: %v\n", err)
	}

	// renameSync(oldPath: string, newPath: string)
	if err := _fs.Set("renameSync", func(call goja.FunctionCall) goja.Value {
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
	}); err != nil {
		fmt.Printf("Error setting fs.renameSync: %v\n", err)
	}

	// unlinkSync(path: string)
	if err := _fs.Set("unlinkSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.unlinkSync: missing path")
		}
		path := call.Arguments[0].String()
		err := os.Remove(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.unlinkSync: %v\n", err)
	}

	if err := _fs.Set("realpathSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.realpathSync: missing path")
		}
		path := call.Arguments[0].String()
		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(realPath)
	}); err != nil {
		fmt.Printf("Error setting fs.realpathSync: %v\n", err)
	}

	if err := _fs.Set("readlinkSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.readlinkSync: missing path")
		}
		path := call.Arguments[0].String()
		link, err := os.Readlink(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(link)
	}); err != nil {
		fmt.Printf("Error setting fs.readlinkSync: %v\n", err)
	}

	// cpSync(src: string, dest: string) :::::::::::::::::::: copy entire directory
	if err := _fs.Set("cpSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("fs.cpSync: missing src or dest")
		}
		src := call.Arguments[0].String()
		dest := call.Arguments[1].String()

		// Check if source exists and is a directory
		srcInfo, err := os.Stat(src)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}

		if !srcInfo.IsDir() {
			if err := copyFile(src, dest); err != nil {
				return vm.ToValue(fmt.Sprintf("%v", err))
			}
			return goja.Undefined()
		}

		// Create destination directory if it doesn't exist
		err = os.MkdirAll(dest, srcInfo.Mode())
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}

		// Walk through the source directory and copy files
		err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip the root directory
			if path == src {
				return nil
			}

			// Calculate relative path and destination path
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			destPath := filepath.Join(dest, relPath)

			// If it's a directory, create it
			if d.IsDir() {
				return os.MkdirAll(destPath, srcInfo.Mode())
			}

			// If it's a file, copy it
			return copyFile(path, destPath)
		})

		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.cpSync: %v\n", err)
	}

	// globSync(pattern: string) => string[]
	if err := _fs.Set("globSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.globSync: missing pattern")
		}
		pattern := call.Arguments[0].String()
		fsys := os.DirFS(".")
		matches, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("glob error: %v", err))
		}
		return vm.ToValue(matches)
	}); err != nil {
		fmt.Printf("Error setting fs.globSync: %v\n", err)
	}

	// mkdtempSync(prefix: string) => string
	if err := _fs.Set("mkdtempSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.mkdtempSync: missing prefix")
		}
		prefix := call.Arguments[0].String()
		dir, err := os.MkdirTemp("", prefix)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(dir)
	}); err != nil {
		fmt.Printf("Error setting fs.mkdtempSync: %v\n", err)
	}

	// symlinkSync(target: string, link: string)
	if err := _fs.Set("symlinkSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("fs.symlinkSync requires 2 arguments: target and path"))
		}

		targetVal := call.Arguments[0]
		linkVal := call.Arguments[1]

		if goja.IsUndefined(targetVal) || goja.IsNull(targetVal) || goja.IsUndefined(linkVal) {
			panic(vm.ToValue("fs.symlinkSync: target or link path is undefined"))
		}

		target := targetVal.String()
		link := linkVal.String()

		if target == "" || link == "" {
			panic(vm.ToValue("fs.symlinkSync: target or link path cannot be empty"))
		}

		err := os.Symlink(target, link)
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("fs.symlinkSync: %v", err)))
		}

		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.symlinkSync: %v\n", err)
	}

	// statSync(path: string) => Stats
	if err := _fs.Set("statSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("fs.statSync: missing path"))
		}
		path := call.Arguments[0].String()

		info, err := os.Stat(path)
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("fs.statSync error: %v", err)))
		}

		return toJsStats(info, vm, statsCtor)
	}); err != nil {
		fmt.Printf("Error setting fs.statSync: %v\n", err)
	}

	// Stats Class
	if err := _fs.Set("Stats", statsCtor); err != nil {
		fmt.Printf("Error setting fs.Stats: %v\n", err)
	}

	// watch
	if err := _fs.Set("watch", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		var listener goja.Callable
		var options map[string]interface{}

		if len(call.Arguments) >= 2 {
			if cb, ok := goja.AssertFunction(call.Arguments[len(call.Arguments)-1]); ok {
				listener = cb
			} else {
				panic(vm.ToValue("watch: last argument must be a function"))
			}
		}

		if len(call.Arguments) >= 2 {
			if optObj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				options = optObj
			}
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("watcher error: %v", err)))
		}

		if err = watcher.Add(path); err != nil {
			panic(vm.ToValue(fmt.Sprintf("watcher add error: %v", err)))
		}

		fileWatchers[path] = watcher
		if listener != nil {
			watchCallbacks[path] = append(watchCallbacks[path], listener)
		}

		info, _ := os.Stat(path)
		if info.IsDir() && options != nil && options["recursive"] != nil && options["recursive"] == true {
			if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if err := watcher.Add(path); err != nil {
						fmt.Printf("watcher add error: %v\n", err)
					}
				}
				return nil
			}); err != nil {
				fmt.Printf("watcher walk error: %v\n", err)
			}
		}

		// Background event loop
		if options == nil || options["persistent"] == nil || options["persistent"] == true {
			EventLoop.Add(1)
		}
		go func() {
			defer func() {
				if options == nil || options["persistent"] == nil || options["persistent"] == true {
					EventLoop.Done()
				}
			}()
			debounceTimers := make(map[string]*time.Timer)
			debounceDuration := 100 * time.Millisecond

			for {
				select {
				case ev, ok := <-watcher.Events:
					if !ok {
						return
					}

					if timer, exists := debounceTimers[ev.Name]; exists {
						timer.Stop()
					}

					debounceTimers[ev.Name] = time.AfterFunc(debounceDuration, func() {
						eventType := "change"
						if ev.Op&fsnotify.Rename == fsnotify.Rename {
							eventType = "rename"
						}
						for _, cb := range watchCallbacks[path] {
							if _, err := cb(goja.Undefined(), vm.ToValue(eventType), vm.ToValue(ev.Name)); err != nil {
								fmt.Printf("Error in watch callback: %v\n", err)
							}
						}
					})

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					fmt.Println("watch error:", err)
				}
			}
		}()

		// Return a "close" function
		watcherObj := vm.NewObject()
		if err := watcherObj.Set("close", func(goja.FunctionCall) goja.Value {
			if err := watcher.Close(); err != nil {
				fmt.Printf("Error closing watcher in JS close function: %v\n", err)
			}
			delete(fileWatchers, path)
			delete(watchCallbacks, path)
			EventLoop.Done()
			return goja.Undefined()
		}); err != nil {
			fmt.Printf("Error setting close on watcherObj: %v\n", err)
		}

		return watcherObj
	}); err != nil {
		fmt.Printf("Error setting fs.watch: %v\n", err)
	}

	if err := _fs.Set("watchFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		var interval = time.Second
		var listener goja.Callable

		if len(call.Arguments) >= 2 {
			if cb, ok := goja.AssertFunction(call.Arguments[len(call.Arguments)-1]); ok {
				listener = cb
			} else {
				panic(vm.ToValue("watchFile: last argument must be a function"))
			}
		}

		if len(call.Arguments) >= 2 {
			if optObj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if ival, exists := optObj["interval"]; exists {
					if intval, ok := ival.(float64); ok {
						interval = time.Millisecond * time.Duration(intval)
					}
				}
			}
		}

		// function object as a key
		listenerObj := call.Argument(len(call.Arguments) - 1).ToObject(vm)
		if pollCallbacks[path] == nil {
			pollCallbacks[path] = make(map[*goja.Object]goja.Callable)
		}
		pollCallbacks[path][listenerObj] = listener

		// If poller not started, start it
		if _, running := filePollers[path]; !running {
			stop := make(chan struct{})
			filePollers[path] = stop

			EventLoop.Add(1)
			go func() {
				defer EventLoop.Done()
				var prev os.FileInfo
				ticker := time.NewTicker(interval)
				defer ticker.Stop()

				for {
					select {
					case <-stop:
						return
					case <-ticker.C:
						curr, err := os.Stat(path)
						if err != nil {
							continue
						}
						if prev != nil && curr.ModTime() != prev.ModTime() {
							jsPrev := toJsStats(prev, vm, statsCtor)
							jsCurr := toJsStats(curr, vm, statsCtor)
							for _, cb := range pollCallbacks[path] {
								if _, err := cb(goja.Undefined(), jsCurr, jsPrev); err != nil {
									fmt.Printf("Error in watchFile callback: %v\n", err)
								}
							}
						}
						prev = curr
					}
				}
			}()
		}

		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.watchFile: %v\n", err)
	}

	if err := _fs.Set("unwatchFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		var listener goja.Callable

		if len(call.Arguments) > 1 {
			if cb, ok := goja.AssertFunction(call.Argument(1)); ok {
				listener = cb
			}
		}

		if listener != nil {
			// Remove specific listener
			listenerObj := call.Argument(1).ToObject(vm)
			delete(pollCallbacks[path], listenerObj)
		} else {
			// Remove all listeners
			delete(pollCallbacks, path)
		}

		// If no listeners left, stop the poller
		if len(pollCallbacks[path]) == 0 {
			if stop, ok := filePollers[path]; ok {
				close(stop)
				delete(filePollers, path)
			}
		}

		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting fs.unwatchFile: %v\n", err)
	}

	// ============== Not implemented ================
	asyncNotImplemented := func(name string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(fmt.Sprintf("fs.%s is not implemented in this runtime, use sync version instead %sSync", name, name))
		}
	}

	notImplemented := func(name string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(fmt.Sprintf("fs.%s is not implemented in this runtime", name))
		}
	}

	asyncNotImplList := []string{"glob", "link", "cp", "copyFile", "appendFile", "writeFile", "mkdtemp",
		"readdir", "mkdir", "exists", "readFile", "symlink", "readlink", "unlink", "realpath", "rename", "rm", "rmdir", "stat"}

	notImplList := []string{"access", "accessSync", "chown", "chownSync", "chmod", "chmodSync", "close", "closeSync", "createReadStream", "createWriteStream", "fchown", "fchownSync", "fchmod", "fchmodSync", "fdatasync", "fdatasyncSync", "fstat", "fstatSync", "fsync", "fsyncSync", "ftruncate", "ftruncateSync", "futimes", "futimesSync", "lchown", "lchownSync", "lstat", "lstatSync", "lutimes", "lutimesSync", "open", "openSync", "openAsBlob", "read", "readSync", "readv", "readvSync", "statfs", "statfsSync", "truncate", "truncateSync", "utimes", "utimesSync", "write", "writeSync", "writev", "writevSync", "Dirent", "ReadStream", "WriteStream", "FileReadStream", "FileWriteStream", "Dir", "opendir", "opendirSync"}

	for _, name := range asyncNotImplList {
		if err := _fs.Set(name, asyncNotImplemented(name)); err != nil {
			fmt.Printf("Error setting fs.%s: %v\n", name, err)
		}
	}
	for _, name := range notImplList {
		if err := _fs.Set(name, notImplemented(name)); err != nil {
			fmt.Printf("Error setting fs.%s: %v\n", name, err)
		}
	}
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			fmt.Printf("Error closing source file in copyFile: %v\n", err)
		}
	}()

	dstFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			fmt.Printf("Error closing destination file in copyFile: %v\n", err)
		}
	}()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dest, srcInfo.Mode())
}

func SetObjProperty(obj *goja.Object, key string, value interface{}) {
	if err := obj.Set(key, value); err != nil {
		fmt.Printf("Error setting %s on obj: %v\n", key, err)
	}
}

func toJsStats(info os.FileInfo, vm *goja.Runtime, statsCtor *goja.Object) goja.Value {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		panic(vm.ToValue("fs.statSync: failed to read raw stat"))
	}

	getTimeMs := func(t time.Time) int64 {
		return t.UnixNano() / 1e6
	}

	obj := vm.NewObject()
	SetObjProperty(obj, "dev", stat.Dev)
	SetObjProperty(obj, "ino", stat.Ino)
	SetObjProperty(obj, "mode", stat.Mode)
	SetObjProperty(obj, "nlink", stat.Nlink)
	SetObjProperty(obj, "uid", stat.Uid)
	SetObjProperty(obj, "gid", stat.Gid)
	SetObjProperty(obj, "rdev", stat.Rdev)
	SetObjProperty(obj, "size", stat.Size)
	SetObjProperty(obj, "blksize", stat.Blksize)
	SetObjProperty(obj, "blocks", stat.Blocks)

	// Timestamps in milliseconds
	SetObjProperty(obj, "atimeMs", getTimeMs(time.Unix(stat.Atim.Sec, stat.Atim.Nsec)))
	SetObjProperty(obj, "mtimeMs", getTimeMs(time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)))
	SetObjProperty(obj, "ctimeMs", getTimeMs(time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)))

	// Add ISO 8601 formatted times in UTC
	SetObjProperty(obj, "atime", time.Unix(stat.Atim.Sec, stat.Atim.Nsec).UTC().Format(time.RFC3339Nano))
	SetObjProperty(obj, "mtime", time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec).UTC().Format(time.RFC3339Nano))
	SetObjProperty(obj, "ctime", time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC().Format(time.RFC3339Nano))

	SetObjProperty(obj, "isFile", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode().IsRegular())
	})
	SetObjProperty(obj, "isDirectory", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.IsDir())
	})
	SetObjProperty(obj, "isSymbolicLink", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeSymlink != 0)
	})
	SetObjProperty(obj, "isSocket", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeSocket != 0)
	})
	SetObjProperty(obj, "isFIFO", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeNamedPipe != 0)
	})
	SetObjProperty(obj, "isCharacterDevice", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeCharDevice != 0)
	})
	SetObjProperty(obj, "isBlockDevice", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(false)
	})

	instance, err := vm.New(statsCtor, obj)
	if err != nil {
		panic(err)
	}
	return vm.ToValue(instance)
}
