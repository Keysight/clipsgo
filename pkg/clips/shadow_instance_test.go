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
	"testing"

	"gotest.tools/assert"
)

func TestInsert(t *testing.T) {
	t.Run("Basic Insert", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type TestClass struct {
			Intval   int
			Floatval float64
			IntSlice []int
			SymSlice []Symbol
			GenSlice []interface{}
		}
		template := TestClass{
			Intval:   7,
			Floatval: 15.0,
			SymSlice: []Symbol{
				"a", "b", "c",
			},
			GenSlice: []interface{}{
				"a", Symbol("b"), int64(2),
			},
		}

		inst, err := env.Insert("", template)
		assert.NilError(t, err)
		assert.Equal(t, inst.Class().String(), `(defclass MAIN::TestClass
   (is-a USER)
   (slot Intval
      (type INTEGER))
   (slot Floatval
      (type FLOAT))
   (multislot IntSlice
      (type INTEGER))
   (multislot SymSlice
      (type SYMBOL))
   (multislot GenSlice
      (type ?VARIABLE)))`)

		var out TestClass
		err = inst.Extract(&out)
		assert.NilError(t, err)
		assert.DeepEqual(t, out, template)
	})

	t.Run("Nested insert - direct", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type ChildClass struct {
			Intval   *int
			Floatval *float64
		}
		type ParentClass struct {
			Str   string
			Child ChildClass
		}
		intval := 99
		floatval := 107.0
		template := ParentClass{
			Str: "with actual value",
			Child: ChildClass{
				Intval:   &intval,
				Floatval: &floatval,
			},
		}

		inst, err := env.Insert("", template)
		assert.NilError(t, err)
		assert.Equal(t, inst.Class().String(), `(defclass MAIN::ParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (slot Child
      (type INSTANCE-NAME)
      (allowed-classes ChildClass)))`)

		assert.Equal(t, inst.String(), `[gen1] of ParentClass (Str "with actual value") (Child [gen2])`)

		subinst, err := env.FindInstance("gen2", "")
		assert.NilError(t, err)
		assert.Equal(t, subinst.String(), `[gen2] of ChildClass (Intval 99) (Floatval 107.0)`)

		var ret *ParentClass
		err = inst.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &template)
	})

	t.Run("Nested insert - pointer", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type ChildClass struct {
			Intval   int
			Floatval float64
		}
		type ParentClass struct {
			Str   string
			Child *ChildClass
		}
		template1 := ParentClass{
			Str: "with nil value",
		}
		template2 := ParentClass{
			Str: "with actual value",
			Child: &ChildClass{
				Intval:   99,
				Floatval: 107.0,
			},
		}

		inst, err := env.Insert("", &template1)
		assert.NilError(t, err)
		assert.Equal(t, inst.Class().String(), `(defclass MAIN::ParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (slot Child
      (type INSTANCE-NAME)
      (allowed-classes ChildClass)))`)

		var ret *ParentClass
		err = inst.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &template1)

		inst, err = env.Insert("", &template2)
		assert.NilError(t, err)
		ret = nil
		err = inst.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &template2)
	})

	t.Run("Nested insert - recursion loop", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		template := ComposeParentClass{
			Str: "with actual value",
			Child: &ComposeChildClass{
				Intval:   99,
				Floatval: 107.0,
			},
		}
		template.Child.Recurse = &template

		var ret *ComposeParentClass
		inst, err := env.Insert("", &template)
		assert.NilError(t, err)
		ret = nil
		err = inst.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &template)
	})
}
