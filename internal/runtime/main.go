package runtime

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/url"
)

func NewRuntime() *goja.Runtime {
	vm := goja.New()
	
	registry := new(require.Registry)
	registry.Enable(vm)
	url.Enable(vm)
	Register(vm, "process", Process)
	Register(vm, "os", OS)
	Register(vm, "fs", FS)
	Register(vm, "path", Path)
	Register(vm, "child_process", ChildProcess)
	Register(vm, "prasmoid", Prasmoid)
	Register(vm, "console", Console)

	return vm
}

func Register(vm *goja.Runtime, name string, module func(vm *goja.Runtime, module *goja.Object)) {
	require.RegisterCoreModule(name, module)
	vm.Set(name, require.Require(vm, name))
}