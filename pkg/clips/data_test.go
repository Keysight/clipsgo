package clips

import (
	"reflect"
	"testing"

	"gotest.tools/assert"
)

func TestDataFromClips(t *testing.T) {
	t.Run("Float Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("12.0")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Float64)
		assert.Equal(t, ret, 12.0)
	})

	t.Run("Integer Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("12")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Int64)
		assert.Equal(t, ret, int64(12))
	})

	t.Run("String Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("\"Hello World!\"")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.String)
		assert.Equal(t, ret, "Hello World!")
	})

	/*jjj
	t.Run("External Address Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("\"Hello World!\"")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.String)
		assert.Equal(t, ret, "Hello World!")
	})
	*/

	t.Run("Symbol Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("UnadornedSymbol")
		assert.NilError(t, err)
		//assert.Equal(t, reflect.TypeOf(ret).String(), "clips.Symbol")
		assert.Equal(t, ret, Symbol("UnadornedSymbol"))
	})

	t.Run("InstanceName Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("[gen1]")
		assert.NilError(t, err)
		//assert.Equal(t, reflect.TypeOf(ret).String(), "clips.Symbol")
		assert.Equal(t, ret, InstanceName("gen1"))
	})

	t.Run("MULTIFIELD Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		ret, err := env.Eval("(create$ a b \"c\" 1.0 2 3)")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Slice)

		val, ok := ret.([]interface{})
		assert.Assert(t, ok)
		assert.Equal(t, len(val), 6)
		assert.Equal(t, val[0], Symbol("a"))
		assert.Equal(t, val[1], Symbol("b"))
		assert.Equal(t, val[2], "c")
		assert.Equal(t, val[3], float64(1.0))
		assert.Equal(t, val[4], int64(2))
		assert.Equal(t, val[5], int64(3))
	})

}

func TestDataIntoClips(t *testing.T) {
	/*
		t.Run("Float Conversion", func(t *testing.T) {
			env := CreateEnvironment()
			defer env.Close()

			ret, err := env.Eval("12.0")
			assert.NilError(t, err)
			assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Float64)
			assert.Equal(t, ret, 12.0)
		})
	*/
}
