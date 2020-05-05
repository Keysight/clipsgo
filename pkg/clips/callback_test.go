package clips

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestCallback(t *testing.T) {
	t.Run("DefineFunction", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		argcount := 0
		callback := func(args []interface{}) (interface{}, error) {
			argcount = len(args)
			return 7, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback a b c)")
		assert.NilError(t, err)
		assert.Equal(t, argcount, 3)
		assert.Equal(t, ret, int64(7))
	})

	t.Run("Function error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return nil, fmt.Errorf("expected")
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, "*errors.errorString: expected")
	})

	t.Run("Invalid args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return nil, fmt.Errorf("expected")
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, "*errors.errorString: expected")
	})
}
