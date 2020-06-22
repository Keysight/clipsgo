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

func TestGenericEnv(t *testing.T) {
	t.Run("List Generics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)
		err = env.Build(`(defgeneric bar "lame generic")`)
		assert.NilError(t, err)

		generics := env.Generics()
		assert.Equal(t, len(generics), 2)
		assert.Equal(t, generics[0].Name(), "foo")
		assert.Equal(t, generics[1].Name(), "bar")
	})

	t.Run("Find Generic", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)
		err = env.Build(`(defgeneric bar "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)
		assert.Equal(t, generic.Name(), "foo")

		_, err = env.FindGeneric("baz")
		assert.ErrorContains(t, err, "not found")
	})
}

func TestGeneric(t *testing.T) {
	t.Run("Generics basic values", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)
		assert.Equal(t, generic.Name(), "foo")
		assert.Equal(t, generic.String(), `(defgeneric MAIN::foo "lame generic")`)
	})

	t.Run("Generics equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)
		err = env.Build(`(defgeneric bar "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		generic2, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		assert.Assert(t, generic.Equal(generic2))

		generic, err = env.FindGeneric("bar")
		assert.NilError(t, err)
		assert.Assert(t, !generic.Equal(generic2))
	})

	t.Run("Generics call", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)
		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		ret, err := generic.Call("1 2")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(3))

		ret, err = generic.Call("1")
		assert.ErrorContains(t, err, "No applicable methods")
	})

	t.Run("Module", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		mod := generic.Module()
		assert.Equal(t, mod.Name(), "MAIN")
	})

	t.Run("Deletable", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)
		assert.Assert(t, generic.Deletable())

		err = env.Build(`(defrule fooref => (foo 1 2))`)
		assert.NilError(t, err)

		assert.Assert(t, !generic.Deletable())
	})

	t.Run("Undefine", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)
		err = env.Build(`(defrule fooref => (foo 1 2))`)
		assert.NilError(t, err)

		err = generic.Undefine()
		assert.ErrorContains(t, err, "Unable")

		_, err = env.Eval(`(undefrule fooref)`)
		assert.NilError(t, err)

		err = generic.Undefine()
		assert.NilError(t, err)
	})

	t.Run("Watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		assert.Assert(t, !generic.Watched())
		generic.Watch(true)
		assert.Assert(t, generic.Watched())
	})

	t.Run("Methods", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 0)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)
		err = env.Build(`(defmethod foo ((?a FLOAT) (?b FLOAT)) (+ ?a ?b))`)
		assert.NilError(t, err)

		methods = generic.Methods()
		assert.Equal(t, len(methods), 2)
	})
}

func TestMethod(t *testing.T) {
	t.Run("Method basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 1)

		method := methods[0]
		assert.Equal(t, method.String(), `(defmethod MAIN::foo
   ((?a INTEGER)
    (?b INTEGER))
   (+ ?a ?b))`)
	})

	t.Run("Method equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)
		err = env.Build(`(defmethod foo ((?a FLOAT) (?b FLOAT)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 2)
		methods2 := generic.Methods()
		assert.Equal(t, len(methods2), 2)

		assert.Assert(t, methods[0].Equal(methods2[0]))
		assert.Assert(t, !methods[0].Equal(methods2[1]))
	})

	t.Run("Method watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 1)

		method := methods[0]
		assert.Assert(t, !method.Watched())
		method.Watch(true)
		assert.Assert(t, method.Watched())
	})

	t.Run("Deletable", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 1)

		method := methods[0]
		assert.Assert(t, method.Deletable())
	})

	t.Run("Undefine", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)
		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)
		methods := generic.Methods()
		assert.Equal(t, len(methods), 1)

		method := methods[0]

		err = method.Undefine()
		assert.NilError(t, err)
	})

	t.Run("Method description", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 1)

		method := methods[0]
		assert.Equal(t, method.Description(), "1  (INTEGER) (INTEGER)")
	})

	t.Run("Method restrictions", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defgeneric foo "lame generic")`)
		assert.NilError(t, err)

		err = env.Build(`(defmethod foo ((?a INTEGER) (?b INTEGER)) (+ ?a ?b))`)
		assert.NilError(t, err)

		generic, err := env.FindGeneric("foo")
		assert.NilError(t, err)

		methods := generic.Methods()
		assert.Equal(t, len(methods), 1)

		method := methods[0]
		assert.DeepEqual(t, method.Restrictions(), []interface{}{
			// min-max args
			int64(2), int64(2),
			// number of restrictions
			int64(2),
			int64(6),
			int64(9),
			false,
			int64(1),
			Symbol("INTEGER"),
			false,
			int64(1),
			Symbol("INTEGER"),
		})
	})
}
