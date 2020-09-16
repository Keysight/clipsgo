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
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"testing"

	"gotest.tools/assert"
)

type ComposeChildClass struct {
	Intval   int
	Floatval float64
	Recurse  *ComposeParentClass
}
type ComposeParentClass struct {
	Str   string
	Child *ComposeChildClass
}

func TestInsertFields(t *testing.T) {
	t.Run("Basic insert", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type TestClass struct {
			Intval   int `json:"name"`
			Floatval float64
			IntSlice []int
			SymSlice *[]Symbol
			GenSlice []interface{}
		}
		var template *TestClass

		cls, err := env.InsertClass(template)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::TestClass
   (is-a USER)
   (slot _name
      (type INTEGER))
   (slot Floatval
      (type FLOAT))
   (multislot IntSlice
      (type INTEGER))
   (multislot SymSlice
      (type SYMBOL))
   (multislot GenSlice
      (type ?VARIABLE)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 5)

		inst, err := cls.NewInstance("", false)
		assert.NilError(t, err)
		err = inst.SetSlot("_name", 7)
		assert.NilError(t, err)
		err = inst.SetSlot("Floatval", 15.0)
		assert.NilError(t, err)
		err = inst.SetSlot("SymSlice", []Symbol{"a", "b", "c"})
		assert.NilError(t, err)
		err = inst.SetSlot("GenSlice", []interface{}{"a", Symbol("b"), 2})
		assert.NilError(t, err)

		var out TestClass
		err = inst.Extract(&out)
		assert.NilError(t, err)
		assert.DeepEqual(t, out, TestClass{
			Intval:   7,
			Floatval: 15.0,
			SymSlice: &[]Symbol{
				"a", "b", "c",
			},
			GenSlice: []interface{}{
				"a", Symbol("b"), int64(2),
			},
		})
	})

	t.Run("insert with anonymous", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type Bar struct {
			Intval     int
			Floatval   float64
			IntSlice   []int
			Undeclared struct{ Val string }
		}
		type TestClass struct {
			Bar
			SymSlice []Symbol
			GenSlice []interface{}
		}
		var template *TestClass

		cls, err := env.InsertClass(template)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::TestClass
   (is-a USER)
   (slot Intval
      (type INTEGER))
   (slot Floatval
      (type FLOAT))
   (multislot IntSlice
      (type INTEGER))
   (slot Undeclared
      (type INSTANCE-NAME)
      (allowed-classes struct___Val_string__))
   (multislot SymSlice
      (type SYMBOL))
   (multislot GenSlice
      (type ?VARIABLE)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 6)

		inst, err := cls.NewInstance("", false)
		assert.NilError(t, err)
		err = inst.SetSlot("Intval", 7)
		assert.NilError(t, err)
		err = inst.SetSlot("Floatval", 15.0)
		assert.NilError(t, err)
		err = inst.SetSlot("Undeclared", struct{ Val string }{Val: "val"})
		assert.NilError(t, err)
		err = inst.SetSlot("SymSlice", []Symbol{"a", "b", "c"})
		assert.NilError(t, err)
		err = inst.SetSlot("GenSlice", []interface{}{"a", Symbol("b"), 2})
		assert.NilError(t, err)

		var out TestClass
		err = inst.Extract(&out)
		assert.NilError(t, err)
		assert.DeepEqual(t, out, TestClass{
			Bar: Bar{
				Intval:     7,
				Floatval:   15.0,
				Undeclared: struct{ Val string }{Val: "val"},
			},
			SymSlice: []Symbol{
				"a", "b", "c",
			},
			GenSlice: []interface{}{
				"a", Symbol("b"), int64(2),
			},
		})
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
		var template *ParentClass

		cls, err := env.InsertClass(template)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::ParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (slot Child
      (type INSTANCE-NAME)
      (allowed-classes ChildClass)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 2)

		_, err = env.MakeInstance(`(ch of ChildClass (Intval 99) (Floatval 107.0))`)
		assert.NilError(t, err)

		p1, err := env.MakeInstance(`(p1 of ParentClass (Str "with nil value"))`)
		assert.NilError(t, err)

		var ret *ParentClass
		err = p1.Extract(&ret)
		assert.ErrorContains(t, err, "Unable to convert")

		intval := 99
		floatval := 107.0
		p2, err := env.MakeInstance(`(p2 of ParentClass (Str "with actual value") (Child [ch]))`)
		ret = nil
		err = p2.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ParentClass{
			Str: "with actual value",
			Child: ChildClass{
				Intval:   &intval,
				Floatval: &floatval,
			},
		})
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
		var template *ParentClass

		cls, err := env.InsertClass(template)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::ParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (slot Child
      (type INSTANCE-NAME)
      (allowed-classes ChildClass)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 2)

		_, err = env.MakeInstance(`(ch of ChildClass (Intval 99) (Floatval 107.0))`)
		assert.NilError(t, err)

		p1, err := env.MakeInstance(`(p1 of ParentClass (Str "with nil value"))`)
		assert.NilError(t, err)

		var ret *ParentClass
		err = p1.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ParentClass{
			Str:   "with nil value",
			Child: nil,
		})

		p2, err := env.MakeInstance(`(p2 of ParentClass (Str "with actual value") (Child [ch]))`)
		assert.NilError(t, err)
		ret = nil
		err = p2.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ParentClass{
			Str: "with actual value",
			Child: &ChildClass{
				Intval:   99,
				Floatval: 107.0,
			},
		})
	})

	t.Run("Nested insert - pointer w/ nulls allowed", func(t *testing.T) {
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
		var template *ParentClass

		cls, err := env.InsertClass(template, DoNotRestrictAllowedClasses)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::ParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (slot Child
      (type INSTANCE-NAME)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 2)
	})

	t.Run("Nested insert - multi", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		type ChildClass struct {
			Intval   int
			Floatval float64
		}
		type ParentClass struct {
			Str   string
			Child []*ChildClass
		}
		var template *ParentClass

		cls, err := env.InsertClass(template)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::ParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (multislot Child
      (type INSTANCE-NAME)
      (allowed-classes ChildClass)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 2)

		_, err = env.MakeInstance(`(ch1 of ChildClass (Intval 99) (Floatval 107.0))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(ch2 of ChildClass (Intval 99) (Floatval 107.0))`)
		assert.NilError(t, err)

		p1, err := env.MakeInstance(`(p1 of ParentClass (Str "with nil value"))`)
		assert.NilError(t, err)

		var ret *ParentClass
		err = p1.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ParentClass{
			Str:   "with nil value",
			Child: nil,
		})

		p2, err := env.MakeInstance(`(p2 of ParentClass (Str "with actual value") (Child [ch1] [ch2]))`)
		assert.NilError(t, err)
		ret = nil
		err = p2.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ParentClass{
			Str: "with actual value",
			Child: []*ChildClass{
				&ChildClass{
					Intval:   99,
					Floatval: 107.0,
				},
				&ChildClass{
					Intval:   99,
					Floatval: 107.0,
				},
			},
		})
	})

	t.Run("Nested insert - recursive", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var template *ComposeParentClass

		cls, err := env.InsertClass(template)
		assert.NilError(t, err)
		assert.Equal(t, cls.String(), `(defclass MAIN::ComposeParentClass
   (is-a USER)
   (slot Str
      (type STRING))
   (slot Child
      (type INSTANCE-NAME)
      (allowed-classes ComposeChildClass)))`)

		slots := cls.Slots(true)
		assert.Assert(t, slots != nil)
		assert.Equal(t, len(slots), 2)

		_, err = env.MakeInstance(`(ch of ComposeChildClass (Intval 99) (Floatval 107.0))`)
		assert.NilError(t, err)

		p1, err := env.MakeInstance(`(p1 of ComposeParentClass (Str "with nil value"))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(ch2 of ComposeChildClass (Intval 99) (Floatval 107.0) (Recurse [p1]))`)
		assert.NilError(t, err)

		var ret *ComposeParentClass
		err = p1.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ComposeParentClass{
			Str:   "with nil value",
			Child: nil,
		})

		p2, err := env.MakeInstance(`(p2 of ComposeParentClass (Str "with actual value") (Child [ch]))`)
		assert.NilError(t, err)
		ret = nil
		err = p2.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ComposeParentClass{
			Str: "with actual value",
			Child: &ComposeChildClass{
				Intval:   99,
				Floatval: 107.0,
			},
		})

		p3, err := env.MakeInstance(`(p3 of ComposeParentClass (Str "with actual value") (Child [ch2]))`)
		assert.NilError(t, err)
		ret = nil
		err = p3.Extract(&ret)
		assert.NilError(t, err)
		assert.DeepEqual(t, ret, &ComposeParentClass{
			Str: "with actual value",
			Child: &ComposeChildClass{
				Intval:   99,
				Floatval: 107.0,
				Recurse: &ComposeParentClass{
					Str:   "with nil value",
					Child: nil,
				},
			},
		})

		// set up a truly infinite recursion, so we can ensure it's caught
		_, err = env.Eval(`(send [ch2] put-Recurse [p3])`)
		assert.NilError(t, err)
		ret = nil
		err = p3.Extract(&ret)
		assert.NilError(t, err)
		compare := ComposeParentClass{
			Str: "with actual value",
			Child: &ComposeChildClass{
				Intval:   99,
				Floatval: 107.0,
			},
		}
		compare.Child.Recurse = &compare
		assert.DeepEqual(t, ret, &compare)
	})
}
