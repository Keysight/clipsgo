package clips

import (
	"reflect"
	"testing"
	"unsafe"

	"gotest.tools/assert"
)

func TestDataFromClips(t *testing.T) {
	t.Run("nil Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("nil")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret), nil)
		assert.Equal(t, ret, nil)
	})

	t.Run("Boolean Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("TRUE")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Bool)
		assert.Equal(t, ret, true)

		ret, err = env.Eval("FALSE")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Bool)
		assert.Equal(t, ret, false)
	})

	t.Run("Float Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("12.0")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Float64)
		assert.Equal(t, ret, 12.0)
	})

	t.Run("Integer Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("12")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Int64)
		assert.Equal(t, ret, int64(12))
	})

	t.Run("String Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("\"Hello World!\"")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.String)
		assert.Equal(t, ret, "Hello World!")
	})

	t.Run("External Address Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback1 := func(args []interface{}) (interface{}, error) {
			assert.Equal(t, args[0], unsafe.Pointer(nil))
			return nil, nil
		}

		err := env.DefineFunction("test-callback", callback1)
		assert.NilError(t, err)

		callback2 := func(args []interface{}) (interface{}, error) {
			return unsafe.Pointer(nil), nil
		}

		err = env.DefineFunction("generate-external", callback2)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback (generate-external))")
		assert.NilError(t, err)
	})

	t.Run("Symbol Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("UnadornedSymbol")
		assert.NilError(t, err)
		//assert.Equal(t, reflect.TypeOf(ret).String(), "clips.Symbol")
		assert.Equal(t, ret, Symbol("UnadornedSymbol"))
	})

	t.Run("InstanceName Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("[gen1]")
		assert.NilError(t, err)
		//assert.Equal(t, reflect.TypeOf(ret).String(), "clips.Symbol")
		assert.Equal(t, ret, InstanceName("gen1"))
	})

	t.Run("MULTIFIELD Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

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

	t.Run("Implied Fact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("(bind ?ret (assert (foo a b c)))")
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).String(), "*clips.ImpliedFact")

		f, ok := ret.(*ImpliedFact)
		assert.Assert(t, ok)
		assert.Equal(t, f.String(), "(foo a b c)")
	})

	t.Run("Template Fact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		ret, err := env.Eval(`(bind ?ret (assert (foo)))`)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).String(), "*clips.TemplateFact")

		f, ok := ret.(*TemplateFact)
		assert.Assert(t, ok)
		assert.Equal(t, f.String(), "(foo (bar nil) (baz))")
	})

	t.Run("Instance Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(foo of Foo (bar 12))`)
		defer inst.Drop()
		assert.NilError(t, err)

		ret, err := env.Eval(`(bind ?ret (instance-address [foo]))`)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).String(), "*clips.Instance")

		inst, ok := ret.(*Instance)
		assert.Assert(t, ok)
		assert.Equal(t, inst.String(), "[foo] of Foo (bar 12) (baz)")
	})
}

func TestDataIntoClips(t *testing.T) {
	t.Run("nil Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return nil, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, nil)
	})

	t.Run("Bool Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret := true
		callback := func(args []interface{}) (interface{}, error) {
			return ret, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, true)

		ret = false
		_, err = env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, false)
	})

	t.Run("Float Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return 1.7E12, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, 1.7E12)
	})

	t.Run("Integer Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return 112, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(112))
	})

	t.Run("String Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return "Test String", nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, "Test String")
	})

	t.Run("External Address Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return unsafe.Pointer(nil), nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, unsafe.Pointer(nil))
	})

	t.Run("Symbol Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return Symbol("TestSymbol"), nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, Symbol("TestSymbol"))
	})

	t.Run("InstanceName Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return InstanceName("testname"), nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, InstanceName("testname"))
	})

	t.Run("MULTIFIELD Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		in := []interface{}{
			"a",
			Symbol("b"),
			int64(1),
			int32(2),
			int(3),
			float32(2.0),
			float64(1.7E12),
			InstanceName("gen7"),
		}
		callback := func(args []interface{}) (interface{}, error) {
			return in, nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, []interface{}{
			"a",
			Symbol("b"),
			int64(1),
			int64(2),
			int64(3),
			float64(2.0),
			float64(1.7E12),
			InstanceName("gen7"),
		})
	})

	t.Run("ImpliedFact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return args[0], nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback (assert (foo a b c)))")
		assert.NilError(t, err)

		fact, ok := ret.(*ImpliedFact)
		assert.Assert(t, ok)
		assert.Equal(t, fact.Index(), 1)
	})

	t.Run("Template Fact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return args[0], nil
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)
		err = env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback (assert (foo)))")
		assert.NilError(t, err)

		fact, ok := ret.(*TemplateFact)
		assert.Assert(t, ok)
		assert.Equal(t, fact.Index(), 1)
	})

	t.Run("Instance Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(args []interface{}) (interface{}, error) {
			return args[0], nil
		}
		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		err = env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(foo of Foo (bar 12))`)
		defer inst.Drop()
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback (instance-address [foo]))")
		assert.NilError(t, err)

		inst, ok := ret.(*Instance)
		assert.Assert(t, ok)
		assert.Equal(t, inst.String(), "[foo] of Foo (bar 12) (baz)")
	})
}
