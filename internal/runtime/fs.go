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
    _fs.Set("Stats", statsCtor)


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

	// cpSync(src: string, dest: string) :::::::::::::::::::: copy entire directory
	_fs.Set("cpSync", func(call goja.FunctionCall) goja.Value {
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
	})

	// globSync(pattern: string) => string[]
	_fs.Set("globSync", func(call goja.FunctionCall) goja.Value {
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
	})

	// mkdtempSync(prefix: string) => string
	_fs.Set("mkdtempSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("fs.mkdtempSync: missing prefix")
		}
		prefix := call.Arguments[0].String()
		dir, err := os.MkdirTemp("", prefix)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%v", err))
		}
		return vm.ToValue(dir)
	})

	// symlinkSync(target: string, link: string)
	_fs.Set("symlinkSync", func(call goja.FunctionCall) goja.Value {
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
	})

	// statSync(path: string) => Stats
	_fs.Set("statSync", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("fs.statSync: missing path"))
		}
		path := call.Arguments[0].String()

		info, err := os.Stat(path)
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("fs.statSync error: %v", err)))
		}

		return toJsStats(info, vm, statsCtor)
	})

	// Stats Class
	_fs.Set("Stats", statsCtor)


	// watch
	_fs.Set("watch", func(call goja.FunctionCall) goja.Value {
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
	
		err = watcher.Add(path)
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("watcher add error: %v", err)))
		}
	
		fileWatchers[path] = watcher
		if listener != nil {
			watchCallbacks[path] = append(watchCallbacks[path], listener)
		}

		info, _ := os.Stat(path)
		if info.IsDir() && options != nil && options["recursive"] != nil && options["recursive"] == true {
			filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					watcher.Add(path)
				}
				return nil
			})
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
							cb(goja.Undefined(), vm.ToValue(eventType), vm.ToValue(ev.Name))
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
		watcherObj.Set("close", func(goja.FunctionCall) goja.Value {
			watcher.Close()
			delete(fileWatchers, path)
			delete(watchCallbacks, path)
			EventLoop.Done()
			return goja.Undefined()
		})
	
		return watcherObj
	})

	_fs.Set("watchFile", func(call goja.FunctionCall) goja.Value {
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
		listenerObj := call.Argument(len(call.Arguments)-1).ToObject(vm)
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
								cb(goja.Undefined(), jsCurr, jsPrev)
							}
						}
						prev = curr
					}
				}
			}()
		}
	
		return goja.Undefined()
	})

	_fs.Set("unwatchFile", func(call goja.FunctionCall) goja.Value {
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
	})
	
	
	
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
	"readdir", "mkdir", "exists", "readFile", "symlink", "readlink", "unlink", "realpath", "rename","rm", "rmdir","stat",}

	notImplList := []string{"access","accessSync","chown","chownSync","chmod","chmodSync","close","closeSync","createReadStream","createWriteStream","fchown","fchownSync","fchmod","fchmodSync","fdatasync","fdatasyncSync","fstat","fstatSync","fsync","fsyncSync","ftruncate","ftruncateSync","futimes","futimesSync","lchown","lchownSync","lstat","lstatSync","lutimes","lutimesSync","open","openSync","openAsBlob","read","readSync","readv","readvSync","statfs","statfsSync","truncate","truncateSync","utimes","utimesSync","write","writeSync","writev","writevSync","Dirent","ReadStream","WriteStream","FileReadStream","FileWriteStream","Dir","opendir","opendirSync"}

	for _, name := range asyncNotImplList {
		_fs.Set(name, asyncNotImplemented(name))
	}
	for _, name := range notImplList {
		_fs.Set(name, notImplemented(name))
	}
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dest, srcInfo.Mode())
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
	obj.Set("dev", stat.Dev)
	obj.Set("ino", stat.Ino)
	obj.Set("mode", stat.Mode)
	obj.Set("nlink", stat.Nlink)
	obj.Set("uid", stat.Uid)
	obj.Set("gid", stat.Gid)
	obj.Set("rdev", stat.Rdev)
	obj.Set("size", stat.Size)
	obj.Set("blksize", stat.Blksize)
	obj.Set("blocks", stat.Blocks)

	// Timestamps in milliseconds
	obj.Set("atimeMs", getTimeMs(time.Unix(stat.Atim.Sec, stat.Atim.Nsec)))
	obj.Set("mtimeMs", getTimeMs(time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)))
	obj.Set("ctimeMs", getTimeMs(time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)))
	
	// Add ISO 8601 formatted times in UTC
	obj.Set("atime", time.Unix(stat.Atim.Sec, stat.Atim.Nsec).UTC().Format(time.RFC3339Nano))
	obj.Set("mtime", time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec).UTC().Format(time.RFC3339Nano))
	obj.Set("ctime", time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC().Format(time.RFC3339Nano))

	obj.Set("isFile", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode().IsRegular())
	})
	obj.Set("isDirectory", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.IsDir())
	})
	obj.Set("isSymbolicLink", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeSymlink != 0)
	})
	obj.Set("isSocket", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeSocket != 0)
	})
	obj.Set("isFIFO", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeNamedPipe != 0)
	})
	obj.Set("isCharacterDevice", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Mode()&os.ModeCharDevice != 0)
	})
	obj.Set("isBlockDevice", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(false) 
	})
	
	instance, err := vm.New(statsCtor, obj)
	if err != nil {
		panic(err)
	}
	return vm.ToValue(instance)
}