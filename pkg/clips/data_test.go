package clips
/*
   Copyright 2020 Keysight Technologies

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*/

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"gotest.tools/assert"
)

func TestDataFromClips(t *testing.T) {
	t.Run("nil Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "nil"
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret), nil)
		assert.Equal(t, ret, nil)

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, nil)

		// interface
		ifval := fmt.Errorf("error interface")
		err = env.ExtractEval(&ifval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, ifval, nil)

		intval := 5
		err = env.ExtractEval(&intval, inputval)
		assert.ErrorContains(t, err, "Unable to convert")
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

		inputval := "TRUE"
		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, true)

		boolval := false
		err = env.ExtractEval(&boolval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, boolval, true)

		intval := 5
		err = env.ExtractEval(&intval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("Float Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "12.0"
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, ret, float64(12.0))

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, 12.0)

		floatval := 4.5
		err = env.ExtractEval(&floatval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, floatval, 12.0)

		float32val := float32(4.5)
		err = env.ExtractEval(&float32val, inputval)
		assert.NilError(t, err)
		assert.Equal(t, float32val, float32(12.0))

		intval := 5
		err = env.ExtractEval(&intval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("Integer Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "12"
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Int64)
		assert.Equal(t, ret, int64(12))

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, int64(12))

		intval := 4
		err = env.ExtractEval(&intval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, intval, 12)

		int64val := int64(4)
		err = env.ExtractEval(&int64val, inputval)
		assert.NilError(t, err)
		assert.Equal(t, int64val, int64(12))

		int32val := int32(4)
		err = env.ExtractEval(&int32val, inputval)
		assert.NilError(t, err)
		assert.Equal(t, int32val, int32(12))

		int16val := int16(4)
		err = env.ExtractEval(&int16val, inputval)
		assert.NilError(t, err)
		assert.Equal(t, int16val, int16(12))

		int8val := int8(4)
		err = env.ExtractEval(&int8val, inputval)
		assert.NilError(t, err)
		assert.Equal(t, int8val, int8(12))

		type Special int8
		var sval Special
		err = env.ExtractEval(&sval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, sval, Special(12))

		floatval := 5.0
		err = env.ExtractEval(&floatval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("String Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := `"Hello World!"`
		outputval := "Hello World!"
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.String)
		assert.Equal(t, ret, outputval)

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, outputval)

		stringval := "foo"
		err = env.ExtractEval(&stringval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, stringval, outputval)

		type Specialized string
		var sval Specialized
		err = env.ExtractEval(&sval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, sval, Specialized(outputval))

		intval := 5
		err = env.ExtractEval(&intval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("External Address Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var successful bool
		callback1 := func(arg unsafe.Pointer) {
			successful = arg == unsafe.Pointer(nil)
		}

		err := env.DefineFunction("test-callback", callback1)
		assert.NilError(t, err)

		callback2 := func() unsafe.Pointer {
			return unsafe.Pointer(nil)
		}

		err = env.DefineFunction("generate-external", callback2)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback (generate-external))")
		assert.NilError(t, err)
		assert.Assert(t, successful)
	})

	t.Run("Symbol Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "UnadornedSymbol"
		outputval := Symbol("UnadornedSymbol")
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		//assert.Equal(t, reflect.TypeOf(ret).String(), "clips.Symbol")
		assert.Equal(t, ret, outputval)

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, outputval)

		stringval := "foo"
		err = env.ExtractEval(&stringval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, stringval, string(outputval))

		symbolval := Symbol("foo")
		err = env.ExtractEval(&symbolval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, symbolval, outputval)

		intval := 5
		err = env.ExtractEval(&intval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("InstanceName Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "[gen1]"
		outputval := InstanceName("gen1")
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		//assert.Equal(t, reflect.TypeOf(ret).String(), "clips.Symbol")
		assert.Equal(t, ret, outputval)

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, *retval, outputval)

		stringval := "foo"
		err = env.ExtractEval(&stringval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, stringval, string(outputval))

		inameval := InstanceName("foo")
		err = env.ExtractEval(&inameval, inputval)
		assert.NilError(t, err)
		assert.Equal(t, inameval, outputval)

		intval := 5
		err = env.ExtractEval(&intval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("MULTIFIELD Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "(create$ a b \"c\" 1.0 2 3)"
		outputval := []interface{}{
			Symbol("a"),
			Symbol("b"),
			"c",
			1.0,
			int64(2),
			int64(3),
		}
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).Kind(), reflect.Slice)

		val, ok := ret.([]interface{})
		assert.Assert(t, ok)
		assert.DeepEqual(t, val, outputval)

		retval := new(interface{})
		err = env.ExtractEval(retval, inputval)
		assert.NilError(t, err)
		assert.DeepEqual(t, *retval, outputval)

		sliceval := make([]string, 0)
		err = env.ExtractEval(&sliceval, "(create$ a b c d e f)")
		assert.NilError(t, err)
		assert.DeepEqual(t, sliceval, []string{
			"a",
			"b",
			"c",
			"d",
			"e",
			"f",
		})

		err = env.ExtractEval(&sliceval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("Implied Fact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		inputval := "(bind ?ret (assert (foo a b c)))"
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).String(), "*clips.ImpliedFact")

		f, ok := ret.(*ImpliedFact)
		assert.Assert(t, ok)
		assert.Equal(t, f.String(), "(foo a b c)")

		// reset to try again
		err = f.Retract()
		assert.NilError(t, err)

		var factval Fact
		err = env.ExtractEval(&factval, inputval)
		assert.NilError(t, err)
		assert.Assert(t, factval != nil)
		assert.Equal(t, factval.String(), "(foo a b c)")

		// reset to try again
		err = factval.Retract()
		assert.NilError(t, err)

		var ptrval *ImpliedFact
		err = env.ExtractEval(&ptrval, inputval)
		assert.NilError(t, err)
		assert.Assert(t, ptrval != nil)
		assert.Equal(t, ptrval.String(), "(foo a b c)")

		// reset to try again
		err = ptrval.Retract()
		assert.NilError(t, err)

		var wptrval *TemplateFact
		err = env.ExtractEval(&wptrval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("Template Fact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		inputval := `(bind ?ret (assert (foo)))`
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).String(), "*clips.TemplateFact")

		f, ok := ret.(*TemplateFact)
		assert.Assert(t, ok)

		// reset to try again
		err = f.Retract()
		assert.NilError(t, err)

		var factval Fact
		err = env.ExtractEval(&factval, inputval)
		assert.NilError(t, err)
		assert.Assert(t, factval != nil)
		assert.Equal(t, factval.String(), "(foo (bar nil) (baz))")

		// reset to try again
		err = factval.Retract()
		assert.NilError(t, err)

		var ptrval *TemplateFact
		err = env.ExtractEval(&ptrval, inputval)
		assert.NilError(t, err)
		assert.Assert(t, ptrval != nil)
		assert.Equal(t, ptrval.String(), "(foo (bar nil) (baz))")

		// reset to try again
		err = ptrval.Retract()
		assert.NilError(t, err)

		var wptrval *ImpliedFact
		err = env.ExtractEval(&wptrval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("Instance Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(foo of Foo (bar 12))`)
		defer inst.Drop()
		assert.NilError(t, err)

		inputval := `(bind ?ret (instance-address [foo]))`
		outputval := "[foo] of Foo (bar 12) (baz)"
		ret, err := env.Eval(inputval)
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(ret).String(), "*clips.Instance")

		inst, ok := ret.(*Instance)
		assert.Assert(t, ok)
		assert.Equal(t, inst.String(), outputval)

		var ptrval *Instance
		err = env.ExtractEval(&ptrval, inputval)
		assert.NilError(t, err)
		assert.Assert(t, ptrval != nil)
		assert.Equal(t, ptrval.String(), outputval)

		var wptrval *ImpliedFact
		err = env.ExtractEval(&wptrval, inputval)
		assert.ErrorContains(t, err, "Invalid type")
	})
}

func TestDataIntoClips(t *testing.T) {
	t.Run("nil Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func() interface{} {
			return nil
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
		callback := func() bool {
			return ret
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

		callback := func() interface{} {
			return 1.7E12
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

		callback := func() interface{} {
			return 112
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

		callback := func() interface{} {
			return "Test String"
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.Equal(t, ret, "Test String")
	})

	t.Run("Specialized String Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type Specialized string
		callback := func() Specialized {
			return Specialized("Test String")
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

		callback := func() interface{} {
			return unsafe.Pointer(nil)
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

		callback := func() interface{} {
			return Symbol("TestSymbol")
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

		callback := func() interface{} {
			return InstanceName("testname")
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
		callback := func() interface{} {
			return in
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

	t.Run("MULTIFIELD Conversion noninterface", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		in := []string{
			"a",
			"b",
		}
		callback := func() interface{} {
			return in
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		ret, err := env.Eval("(test-callback)")
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, []interface{}{
			"a",
			"b",
		})
	})

	t.Run("ImpliedFact Conversion", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		callback := func(arg interface{}) interface{} {
			return arg
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

		callback := func(arg interface{}) interface{} {
			return arg
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

		callback := func(arg *Instance) *Instance {
			return arg
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
