package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestPrasmoidModule(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "prasmoid-test-")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Change to the temp directory for the duration of the test
	originalWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	// Create a dummy metadata.json
	metadataContent := `
	{
	  "KPlugin": {
	    "Id": "org.kde.testplasmoid",
	    "Version": "1.0.0",
	    "Name": "Test Plasmoid"
	  }
	}
	`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))

	vm := NewRuntime()
	_, err = vm.RunString(`const prasmoid = require('prasmoid');`)
	require.NoError(t, err)

	t.Run("getMetadata", func(t *testing.T) {
		t.Run("existing key", func(t *testing.T) {
			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "org.kde.testplasmoid", val.String())
		})

		t.Run("non-existing key", func(t *testing.T) {
			script := `prasmoid.getMetadata('NonExistent');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: NonExistent not found in metadata.json")
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `prasmoid.getMetadata();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "prasmoid.getMetadata: missing key", val.String())
		})

		t.Run("too many arguments", func(t *testing.T) {
			script := `prasmoid.getMetadata('Id', 'extra');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "prasmoid.getMetadata: too many arguments", val.String())
		})

		t.Run("metadata.json not found", func(t *testing.T) {
			// Temporarily remove metadata.json
			require.NoError(t, os.Remove(filepath.Join(tmpDir, "metadata.json")))

			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: metadata.json not found")

			// Recreate metadata.json for subsequent tests
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))
		})

		t.Run("invalid metadata.json format", func(t *testing.T) {
			// Corrupt metadata.json
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(`{invalid json`), 0644))

			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: invalid character")

			// Recreate metadata.json
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))
		})

		t.Run("KPlugin not found in metadata.json", func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(`{"OtherKey": {}}`), 0644))

			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: Id not found in metadata.json")

			// Recreate metadata.json
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))
		})

		t.Run("KPlugin is not a map", func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(`{"KPlugin": "not a map"}`), 0644))

			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: Id not found in metadata.json")

			// Recreate metadata.json
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))
		})

		t.Run("key not found within KPlugin", func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(`{"KPlugin": {"OtherId": "value"}}`), 0644))

			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: Id not found in metadata.json")

			// Recreate metadata.json
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))
		})

		t.Run("key is not a string within KPlugin", func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(`{"KPlugin": {"Id": 123}}`), 0644))

			script := `prasmoid.getMetadata('Id');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "prasmoid.getMetadata: Id not found in metadata.json")

			// Recreate metadata.json
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644))
		})
	})

	t.Run("Command", func(t *testing.T) {
		// Test panic conditions
		t.Run("no arguments panics", func(t *testing.T) {
			require.PanicsWithValue(t, "prasmoid.Command: exactly 1 argument required", func() {
				_, _ = vm.RunString(`prasmoid.Command();`)
			})
		})

		t.Run("undefined argument panics", func(t *testing.T) {
			require.PanicsWithValue(t, "prasmoid.Command: argument must be a JS object", func() {
				_, _ = vm.RunString(`prasmoid.Command(undefined);`)
			})
		})

		t.Run("null argument panics", func(t *testing.T) {
			require.PanicsWithValue(t, "prasmoid.Command: argument must be a JS object", func() {
				_, _ = vm.RunString(`prasmoid.Command(null);`)
			})
		})

		t.Run("function argument panics", func(t *testing.T) {
			require.PanicsWithValue(t, "prasmoid.Command: expected object, got function", func() {
				_, _ = vm.RunString(`prasmoid.Command(function(){});`)
			})
		})

		t.Run("missing run function panics", func(t *testing.T) {
			require.PanicsWithValue(t, "prasmoid.Command: missing 'run' function", func() {
				_, _ = vm.RunString(`prasmoid.Command({});`)
			})
		})

		t.Run("run is not a function panics", func(t *testing.T) {
			require.PanicsWithValue(t, "prasmoid.Command: 'run' must be a function", func() {
				_, _ = vm.RunString(`prasmoid.Command({run: "not a function"});`)
			})
		})

		t.Run("non-bool value in boolean flag panics", func(t *testing.T) {
			require.PanicsWithValue(t, "non-bool value not allowed in boolean flag", func() {
				script := `
					prasmoid.Command({
						run: function(){},
						flags: [{name: "myflag", type: "bool", value: "true"}]
					});
				`
				_, _ = vm.RunString(script)
			})
		})

		// Test successful command registration
		t.Run("basic command registration", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){ return "command executed"; }
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.NotNil(t, CommandStorage.Run)
			// Verify the run function can be called
			val, err := CommandStorage.Run(goja.Undefined())
			require.NoError(t, err)
			require.Equal(t, "command executed", val.String())
		})

		t.Run("command with short and long", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					short: "Short description",
					long: "Long description"
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "Short description", CommandStorage.Short)
			require.Equal(t, "Long description", CommandStorage.Long)
		})

		t.Run("command with alias", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					alias: ["cmd1", "c1"]
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, CommandStorage.Alias, "cmd1")
			require.Contains(t, CommandStorage.Alias, "c1")
			require.Len(t, CommandStorage.Alias, 2)
		})

		t.Run("command with empty alias", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					alias: []
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Empty(t, CommandStorage.Alias)
		})

		t.Run("command with flags", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					flags: [
						{name: "myString", type: "string", value: "default"},
						{name: "myBool", type: "bool", value: true},
						{name: "myStringNoValue", type: "string"},
						{name: "myBoolNoValue", type: "bool"}
					]
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Len(t, CommandStorage.Flags, 4)

			flag1 := CommandStorage.Flags[0]
			require.Equal(t, "myString", flag1.Name)
			require.Equal(t, "string", flag1.Type)
			require.Equal(t, "default", flag1.Value)

			flag2 := CommandStorage.Flags[1]
			require.Equal(t, "myBool", flag2.Name)
			require.Equal(t, "bool", flag2.Type)
			require.Equal(t, true, flag2.Value)

			flag3 := CommandStorage.Flags[2]
			require.Equal(t, "myStringNoValue", flag3.Name)
			require.Equal(t, "string", flag3.Type)
			require.Equal(t, "", flag3.Value)

			flag4 := CommandStorage.Flags[3]
			require.Equal(t, "myBoolNoValue", flag4.Name)
			require.Equal(t, "bool", flag4.Type)
			require.Equal(t, false, flag4.Value)
		})

		t.Run("command with flags and shorthand/description", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					flags: [
						{name: "verbose", type: "bool", shorthand: "v", description: "Enable verbose output"}
					]
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Len(t, CommandStorage.Flags, 1)
			flag := CommandStorage.Flags[0]
			require.Equal(t, "verbose", flag.Name)
			require.Equal(t, "bool", flag.Type)
			require.Equal(t, false, flag.Value)
			require.Equal(t, "v", flag.Shorthand)
			require.Equal(t, "Enable verbose output", flag.Description)
		})

		t.Run("command with flags containing non-map elements", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					flags: [
						{name: "validFlag", type: "string", value: "test"},
						"not a flag object",
						123
					]
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Len(t, CommandStorage.Flags, 1)
			require.Equal(t, "validFlag", CommandStorage.Flags[0].Name)
		})

		t.Run("command with alias containing non-string elements", func(t *testing.T) {
			script := `
				prasmoid.Command({
					run: function(){},
					alias: ["alias1", 123, true, "alias2"]
				});
			`
			_, err := vm.RunString(script)
			require.NoError(t, err)
			require.Len(t, CommandStorage.Alias, 2)
			require.Contains(t, CommandStorage.Alias, "alias1")
			require.Contains(t, CommandStorage.Alias, "alias2")
		})
	})
}