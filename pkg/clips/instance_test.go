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
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/udhos/equalfile"
	"gotest.tools/assert"
)

func TestInstanceEnv(t *testing.T) {
	t.Run("Instances Changed", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER))`)
		assert.NilError(t, err)

		assert.Assert(t, env.InstancesChanged())
		assert.Assert(t, !env.InstancesChanged())
		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		assert.Assert(t, env.InstancesChanged())
	})

	t.Run("Instances", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), InstanceName("initial-object"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[2].Name(), InstanceName("named"))
	})

	t.Run("Find instance", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		inst, err := env.FindInstance("named", "")
		assert.Equal(t, inst.Name(), InstanceName("named"))

		_, err = env.FindInstance("foo", "")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("Save instances text", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.SaveInstances(tmpfile.Name(), false, LOCAL_SAVE)
		assert.NilError(t, err)

		cmp := equalfile.New(nil, equalfile.Options{})
		equal, err := cmp.CompareFile("testdata/instancesfile.clp", tmpfile.Name())
		assert.NilError(t, err)
		assert.Equal(t, equal, true)
	})

	t.Run("Save instances binary", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.SaveInstances(tmpfile.Name(), true, LOCAL_SAVE)
		assert.NilError(t, err)

		/*
			cmp := equalfile.New(nil, equalfile.Options{})
			equal, err := cmp.CompareFile("testdata/instances.bsave", tmpfile.Name())
			assert.NilError(t, err)
			assert.Assert(t, equal)
		*/
	})
	t.Run("Load instances text", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.LoadInstances("testdata/instancesfile.clp")
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), InstanceName("initial-object"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[2].Name(), InstanceName("named"))
	})

	t.Run("Load instances binary", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.LoadInstances("testdata/instances.bsave")
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), InstanceName("initial-object"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[2].Name(), InstanceName("named"))
	})

	t.Run("Load instances string", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.LoadInstancesFromString(`
		([initial-object] of INITIAL-OBJECT)

		([gen1] of Foo
		   (bar nil)
		   (baz))
		
		([named] of Foo
		   (bar nil)
		   (baz))
		`)
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), InstanceName("initial-object"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[2].Name(), InstanceName("named"))
	})

	t.Run("Restore instances", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.RestoreInstances("testdata/instancesfile.clp")
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), InstanceName("initial-object"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[2].Name(), InstanceName("named"))
	})

	t.Run("Restore instances string", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.RestoreInstancesFromString(`
		([initial-object] of INITIAL-OBJECT)

		([gen1] of Foo
		   (bar nil)
		   (baz))
		
		([named] of Foo
		   (bar nil)
		   (baz))
		`)
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), InstanceName("initial-object"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[2].Name(), InstanceName("named"))
	})

	t.Run("Make instance", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		defer inst.Drop()
		assert.NilError(t, err)
		assert.Equal(t, inst.Name(), InstanceName("gen1"))

		/* CLIPS actually doesn't error on this
		_, err = env.MakeInstance(`(of Foo (bar "testing"))`)
		assert.ErrorContains(t, err, "invalid")
		*/
	})
}

func TestInstance(t *testing.T) {
	t.Run("Instance basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)
		assert.Equal(t, inst.Name(), InstanceName("gen1"))
		assert.Equal(t, inst.String(), "[gen1] of Foo (bar 12)")
	})

	t.Run("Instance equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		inst2, err := env.FindInstance("gen1", "")
		assert.NilError(t, err)

		assert.Assert(t, inst.Equal(inst2))

		inst2, err = env.MakeInstance(`(of Foo (bar 77))`)
		assert.NilError(t, err)
		assert.Assert(t, !inst.Equal(inst2))
	})

	t.Run("Class", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		class := inst.Class()
		assert.Equal(t, class.Name(), "Foo")
	})

	t.Run("Slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		slots := inst.Slots(true)
		assert.NilError(t, err)
		assert.Equal(t, len(slots), 2)
		_, ok := slots["bar"]
		assert.Assert(t, ok)
		_, ok = slots["baz"]
		assert.Assert(t, ok)
	})

	t.Run("Slot", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(12))

		ret, err = inst.Slot("bif")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("SetSlot", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(12))

		err = inst.SetSlot("bar", 77)
		assert.NilError(t, err)

		ret, err = inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(77))

		err = inst.SetSlot("bif", 77)
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("SetSlot type violation", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(12))

		err = inst.SetSlot("bar", "forty two")
		// I would think this should be rejected by CLIPS, but it isn't
		assert.NilError(t, err)

		err = inst.SetSlot("baz", []string{"forty two"})
		assert.NilError(t, err)
	})

	t.Run("Send", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret := inst.Send("get-bar", "")
		assert.Equal(t, ret, int64(12))

		ret = inst.Send("put-bar", "77")
		assert.Equal(t, ret, int64(77))

		ret, err = inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(77))

		ret = inst.Send("garbage", "")
		assert.Equal(t, ret, false)
	})

	t.Run("Delete", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		err = inst.Delete()
		assert.NilError(t, err)
	})

	t.Run("Unmake", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		err = inst.Unmake()
		assert.NilError(t, err)
	})
}

func TestExtract(t *testing.T) {
	t.Run("cant extract", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot int (type INTEGER))
			(slot float (type FLOAT))
			(slot sym (type SYMBOL))
			(multislot ms))
		`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (int 12) (float 28.0) (sym bar) (ms a b c))`)
		assert.NilError(t, err)

		var retval int
		err = inst.Extract(retval)
		assert.ErrorContains(t, err, "non-pointer")
		err = inst.Extract(nil)
		assert.ErrorContains(t, err, "non-pointer")
		var mapval map[interface{}]interface{}
		err = inst.Extract(&mapval)
		assert.ErrorContains(t, err, "must be type string")
	})

	t.Run("Flat to map", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot int (type INTEGER))
			(slot float (type FLOAT))
			(slot sym (type SYMBOL))
			(multislot ms))
		`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (int 12) (float 28.0) (sym bar) (ms a b c))`)
		assert.NilError(t, err)

		output := map[string]interface{}{
			"int":   int64(12),
			"float": 28.0,
			"sym":   Symbol("bar"),
			"ms": []interface{}{
				Symbol("a"),
				Symbol("b"),
				Symbol("c"),
			},
		}

		// starting from nil
		var retval map[string]interface{}
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output)

		// populating an existing map
		retval = make(map[string]interface{})
		retval2 := retval
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval2, output)

		// individual slots
		var intval int
		err = inst.ExtractSlot(&intval, "int")
		assert.NilError(t, err)
		assert.Equal(t, intval, 12)

		var floatval float64
		err = inst.ExtractSlot(&floatval, "float")
		assert.NilError(t, err)
		assert.Equal(t, floatval, 28.0)

		var symvar Symbol
		err = inst.ExtractSlot(&symvar, "sym")
		assert.NilError(t, err)
		assert.Equal(t, symvar, Symbol("bar"))

		var msvar []string
		err = inst.ExtractSlot(&msvar, "ms")
		assert.NilError(t, err)
		assert.DeepEqual(t, msvar, []string{
			"a", "b", "c",
		})
	})

	t.Run("Restrictive type map", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot str (type STRING))
			(slot sym (type SYMBOL))
		)`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (str "foo") (sym bar))`)
		assert.NilError(t, err)

		output := map[string]string{
			"str": "foo",
			"sym": "bar",
		}

		// starting from nil
		var retval map[string]string
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output)

		// populating an existing map
		retval = make(map[string]string)
		retval2 := retval
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval2, output)
	})

	t.Run("Restrictive type map - cant", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot str (type STRING))
			(slot sym (type SYMBOL))
			(multislot ms (type SYMBOL))
		)`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (str "foo") (sym bar))`)
		assert.NilError(t, err)

		// starting from nil
		var retval map[string]string
		err = inst.Extract(&retval)
		assert.ErrorContains(t, err, "Invalid type")
	})

	t.Run("Nested to map", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot sym (type SYMBOL))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		err = env.Build(`(defclass Bar (is-a USER)
			(slot foo (type INSTANCE-ADDRESS) (allowed-classes Foo))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(fooinst of Foo (sym bar) (ms a b c))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(barinst of Bar)`)
		assert.NilError(t, err)

		_, err = env.Eval(`(send [barinst] put-foo (instance-address [fooinst]))`)
		assert.NilError(t, err)

		output := map[string]interface{}{
			"foo": map[string]interface{}{
				"sym": Symbol("bar"),
				"ms": []interface{}{
					Symbol("a"),
					Symbol("b"),
					Symbol("c"),
				},
			},
			"ms": []interface{}{},
		}

		// starting from nil
		var retval map[string]interface{}
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output)

		// populating an existing map
		retval = make(map[string]interface{})
		retval2 := retval
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval2, output)
	})

	t.Run("Flat to struct", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot Int (type INTEGER))
			(slot Float (type FLOAT))
			(slot Sym (type SYMBOL))
			(multislot MS))
		`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (Int 12) (Float 28.0) (Sym bar) (MS a b c))`)
		assert.NilError(t, err)

		type Foo struct {
			private   int
			IntVal    int     `json:"Int"`
			FloatVal  float64 `clips:"Float"`
			Sym       Symbol
			MultiSlot []interface{} `json:"MS,omitempty"`
		}

		output := Foo{
			IntVal:   12,
			FloatVal: 28.0,
			Sym:      Symbol("bar"),
			MultiSlot: []interface{}{
				Symbol("a"),
				Symbol("b"),
				Symbol("c"),
			},
		}

		// in-place
		var retval Foo
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output, cmpopts.IgnoreUnexported(output))
	})

	t.Run("Flat to struct with anonymous field", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot Int (type INTEGER))
			(slot Float (type FLOAT))
			(slot Sym (type SYMBOL))
			(multislot MS))
		`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (Int 12) (Float 28.0) (Sym bar) (MS a b c))`)
		assert.NilError(t, err)

		type Bar struct {
			IntVal   int     `json:"Int"`
			FloatVal float64 `clips:"Float"`
			Sym      Symbol
			private  int
		}
		type Foo struct {
			Bar
			MultiSlot []interface{} `json:"MS,omitempty"`
		}

		output := Foo{
			Bar: Bar{
				IntVal:   12,
				FloatVal: 28.0,
				Sym:      Symbol("bar"),
			},
			MultiSlot: []interface{}{
				Symbol("a"),
				Symbol("b"),
				Symbol("c"),
			},
		}

		// in-place
		var retval Foo
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output, cmpopts.IgnoreUnexported(output.Bar))
	})

	t.Run("Pointer to slice", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot Int (type INTEGER))
			(slot Float (type FLOAT))
			(slot Sym (type SYMBOL))
			(multislot MS))
		`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (Int 12) (Float 28.0) (Sym bar) (MS a b c))`)
		assert.NilError(t, err)

		type Foo struct {
			private   int
			IntVal    int     `json:"Int"`
			FloatVal  float64 `clips:"Float"`
			Sym       Symbol
			MultiSlot *[]interface{} `json:"MS,omitempty"`
		}

		output := Foo{
			IntVal:   12,
			FloatVal: 28.0,
			Sym:      Symbol("bar"),
			MultiSlot: &[]interface{}{
				Symbol("a"),
				Symbol("b"),
				Symbol("c"),
			},
		}

		// in-place
		var retval Foo
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output, cmpopts.IgnoreUnexported(output))
	})

	t.Run("Nested to struct", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot Int (type INTEGER))
			(slot Float (type FLOAT))
			(slot Sym (type SYMBOL))
			(multislot MS))
		`)
		assert.NilError(t, err)

		err = env.Build(`(defclass Bar (is-a USER)
			(slot foo (type INSTANCE-ADDRESS) (allowed-classes Foo))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(fooinst of Foo (Int 12) (Float 28.0) (Sym bar) (MS a b c))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(barinst of Bar)`)
		assert.NilError(t, err)

		_, err = env.Eval(`(send [barinst] put-foo (instance-address [fooinst]))`)
		assert.NilError(t, err)

		type Foo struct {
			private   int
			IntVal    int     `json:"Int"`
			FloatVal  float64 `clips:"Float"`
			Sym       Symbol
			MultiSlot []interface{} `json:"MS,omitempty"`
		}
		type Bar struct {
			FooVal Foo `clips:"foo"`
		}

		output := Bar{
			FooVal: Foo{
				IntVal:   12,
				FloatVal: 28.0,
				Sym:      Symbol("bar"),
				MultiSlot: []interface{}{
					Symbol("a"),
					Symbol("b"),
					Symbol("c"),
				},
			},
		}

		// nil ptr
		var retval *Bar
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, &output, cmpopts.IgnoreUnexported(output.FooVal))

		// in-place ptr
		retval = &Bar{}
		retval2 := retval
		err = inst.Extract(retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval.FooVal, retval2.FooVal, cmpopts.IgnoreUnexported(output.FooVal))
		assert.DeepEqual(t, retval2, &output, cmpopts.IgnoreUnexported(output.FooVal))
	})

	t.Run("Nested to struct by ref", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot Int (type INTEGER))
			(slot Float (type FLOAT))
			(slot Sym (type SYMBOL))
			(multislot MS))
		`)
		assert.NilError(t, err)

		err = env.Build(`(defclass Bar (is-a USER)
			(slot foo (type INSTANCE-ADDRESS) (allowed-classes Foo))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(fooinst of Foo (Int 12) (Float 28.0) (Sym bar) (MS a b c))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(barinst of Bar)`)
		assert.NilError(t, err)

		_, err = env.Eval(`(send [barinst] put-foo (instance-address [fooinst]))`)
		assert.NilError(t, err)

		type Foo struct {
			private   int
			IntVal    int     `json:"Int"`
			FloatVal  float64 `clips:"Float"`
			Sym       Symbol
			MultiSlot []interface{} `json:"MS,omitempty"`
		}
		type Bar struct {
			FooVal *Foo `clips:"foo"`
		}

		output := Bar{
			FooVal: &Foo{
				IntVal:   12,
				FloatVal: 28.0,
				Sym:      Symbol("bar"),
				MultiSlot: []interface{}{
					Symbol("a"),
					Symbol("b"),
					Symbol("c"),
				},
			},
		}

		// in-place, nil reference
		var retval Bar
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output, cmpopts.IgnoreUnexported(*output.FooVal))

		// in-place, existing reference
		var fooval Foo
		retval2 := Bar{
			FooVal: &fooval,
		}
		err = inst.Extract(&retval2)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output, cmpopts.IgnoreUnexported(*output.FooVal))
		// existing ref should have been updated, not replaced
		assert.DeepEqual(t, &fooval, output.FooVal, cmpopts.IgnoreUnexported(*output.FooVal))
	})

	t.Run("Nested to map by name", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot sym (type SYMBOL))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		err = env.Build(`(defclass Bar (is-a USER)
			(slot foo (type INSTANCE-NAME) (allowed-classes Foo))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(fooinst of Foo (sym bar) (ms a b c))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(barinst of Bar)`)
		assert.NilError(t, err)

		_, err = env.Eval(`(send [barinst] put-foo [fooinst])`)
		assert.NilError(t, err)

		output := map[string]interface{}{
			"foo": map[string]interface{}{
				"sym": Symbol("bar"),
				"ms": []interface{}{
					Symbol("a"),
					Symbol("b"),
					Symbol("c"),
				},
			},
			"ms": []interface{}{},
		}

		// starting from nil
		var retval map[string]interface{}
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output)

		// populating an existing map
		retval = make(map[string]interface{})
		retval2 := retval
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval2, output)
	})

	t.Run("Nested to map nil ref", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot sym (type SYMBOL))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		err = env.Build(`(defclass Bar (is-a USER)
			(slot foo (type INSTANCE-NAME) (allowed-classes Foo))
			(multislot ms)
		)`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(barinst of Bar)`)
		assert.NilError(t, err)

		output := map[string]interface{}{
			"foo": nil,
			"ms":  []interface{}{},
		}

		// starting from nil
		var retval map[string]interface{}
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, output)

		// populating an existing map
		retval = make(map[string]interface{})
		retval2 := retval
		err = inst.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval2, output)
	})

	t.Run("implicit extract in function call", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot Int (type INTEGER))
			(slot Float (type FLOAT))
			(slot Sym (type SYMBOL))
			(multislot MS))
		`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(fooinst of Foo (Int 12) (Float 28.0) (Sym bar) (MS a b c))`)
		assert.NilError(t, err)

		type Foo struct {
			private   int
			IntVal    int     `json:"Int"`
			FloatVal  float64 `clips:"Float"`
			Sym       Symbol
			MultiSlot []interface{} `json:"MS,omitempty"`
		}

		callback := func(fooval *Foo) bool {
			return fooval != nil && fooval.IntVal == 12 && fooval.FloatVal == 28.0 && fooval.Sym == Symbol("bar")
		}

		err = env.DefineFunction("check", callback)
		assert.NilError(t, err)

		// in-place
		ret, err := env.Eval(`(check [fooinst])`)
		assert.NilError(t, err)
		assert.Equal(t, ret, true)
	})
}
