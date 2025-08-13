package runtime

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dop251/goja"
)

func ChildProcess(vm *goja.Runtime, module *goja.Object) {
	_cp := module.Get("exports").(*goja.Object)

	// === execSync remains unchanged ===
	execSync := func(call goja.FunctionCall) goja.Value {
		cmdStr := call.Argument(0).String()
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			return vm.ToValue("No command given")
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return vm.ToValue(err.Error())
		}
		return vm.ToValue(string(out))
	}

	execAsync := func(call goja.FunctionCall) goja.Value {
		return vm.ToValue("child_process.exec is not implemented in this runtime, use execSync instead")
	}

	if err := _cp.Set("execSync", execSync); err != nil {
		fmt.Printf("Error setting execSync: %v\n", err)
	}
	if err := _cp.Set("exec", execAsync); err != nil {
		fmt.Printf("Error setting exec: %v\n", err)
	}

	// === NOT IMPLEMENTED FUNCTIONS ===
	notImplemented := func(name string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(fmt.Sprintf("child_process.%s is not implemented in this runtime", name))
		}
	}

	for _, name := range []string{
		"execFileSync", "execFile", "spawn", "spawnSync",
	} {
		if err := _cp.Set(name, notImplemented(name)); err != nil {
			fmt.Printf("Error setting %s: %v\n", name, err)
		}
	}
}
