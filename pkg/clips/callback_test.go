package clips

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestCallback(t *testing.T) {
	t.Run("NoArgs", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func() {
			called = true
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.Equal(t, ret, false)
	})

	t.Run("Args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a, b, c interface{}) {
			called = true
			return
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback a b c)")
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.Equal(t, ret, false)
	})

	t.Run("Not enough args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a, b, c interface{}) {
			called = true
			return
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback a b)")
		assert.ErrorContains(t, err, "expected exactly 3")
		assert.Assert(t, !called)
	})

	t.Run("Too many args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a, b, c interface{}) {
			called = true
			return
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback a b c d)")
		assert.ErrorContains(t, err, "expected exactly 3")
		assert.Assert(t, !called)
	})

	t.Run("Variadic args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var argcount int
		callback := func(a, b interface{}, c ...interface{}) {
			argcount = len(c)
			return
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback a b)")
		assert.NilError(t, err)
		assert.Equal(t, argcount, 0)
		assert.Equal(t, ret, false)
	})

	t.Run("Variadic args with more", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var argcount int
		callback := func(a, b interface{}, c ...interface{}) {
			argcount = len(c)
			return
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback a b c d e f)")
		assert.NilError(t, err)
		assert.Equal(t, argcount, 4)
		assert.Equal(t, ret, false)
	})

	t.Run("Variadic - not enough args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a, b interface{}, c ...interface{}) {
			called = true
			return
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback a)")
		assert.ErrorContains(t, err, "expected at least 2")
		assert.Assert(t, !called)
	})

	t.Run("Typed args", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int64, b float64, c string, d Symbol) bool {
			called = true
			return a == 7 && b == 15.0 && c == "testing" && d == Symbol("foo")
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval(`(test-callback 7 15.0 "testing" foo)`)
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.Equal(t, ret, true)
	})

	t.Run("multifield arg", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int64, b []interface{}, c []interface{}) bool {
			called = true
			return a == 7 && len(b) == 2
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval(`(test-callback 7 (create$ a b) (create$ c d))`)
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.Equal(t, ret, true)
	})

	t.Run("Scale reduction", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int, b float32) bool {
			called = true
			return a == 7 && b == 15.0
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval(`(test-callback 7 15.0)`)
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.Equal(t, ret, true)
	})

	t.Run("Scale loss - int", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int32, b float32) bool {
			called = true
			return a == 7 && b == 15.0
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval(`(test-callback 9223372036854775807 15.0)`)
		assert.ErrorContains(t, err, "too large")
		assert.Assert(t, !called)
	})

	t.Run("Scale loss - float", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int32, b float32) bool {
			called = true
			return a == 7 && b == 15.0
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval(`(test-callback 7 15.0123456789012345678901234567890)`)
		assert.ErrorContains(t, err, "too precise")
		assert.Assert(t, !called)
	})

	t.Run("Bad type", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int, b float32) bool {
			called = true
			return a == 7 && b == 15.0
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval(`(test-callback 7.0 15.0)`)
		assert.ErrorContains(t, err, "Invalid type")
		assert.Assert(t, !called)
	})

	t.Run("Bad variadic type", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func(a int, b ...float32) bool {
			called = true
			return a == 7 && len(b) == 2 && b[0] == 15.0 && b[1] == 3
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval(`(test-callback 7 15.0 3)`)
		assert.ErrorContains(t, err, "Invalid type")
		assert.Assert(t, !called)
	})

	t.Run("Single return", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func() int {
			called = true
			return 7
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.Equal(t, ret, int64(7))
	})

	t.Run("error return - error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() error {
			return fmt.Errorf("expected")
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback)")
		assert.ErrorContains(t, err, "expected")
	})

	t.Run("error return - no error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() error {
			return nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, false)
	})

	t.Run("Double return", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func() (int, float32) {
			called = true
			return 7, 15.0
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.DeepEqual(t, ret, []interface{}{
			int64(7),
			float64(15.0),
		})
	})

	t.Run("Double with error return - error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() (int, error) {
			return 7, fmt.Errorf("expected")
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback)")
		assert.ErrorContains(t, err, "expected")
	})

	t.Run("Double with error return - no error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() (int, error) {
			return 7, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(7))
	})

	t.Run("More return", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var called bool
		callback := func() (int, float32, bool) {
			called = true
			return 7, 15.0, true
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Assert(t, called)
		assert.DeepEqual(t, ret, []interface{}{
			int64(7),
			float64(15.0),
			true,
		})
	})

	t.Run("More with error return - error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() (int, bool, error) {
			return 7, true, fmt.Errorf("expected")
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback)")
		assert.ErrorContains(t, err, "expected")
	})

	t.Run("Double with error return - no error", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() (int, bool, error) {
			return 7, true, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, []interface{}{
			int64(7),
			true,
		})
	})
}
