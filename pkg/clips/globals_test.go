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

func TestGlobalsEnv(t *testing.T) {
	t.Run("List globals", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)
		err = env.Build(`(defglobal ?*bar* = (create$ a b c))`)
		assert.NilError(t, err)

		globs := env.Globals()
		assert.Equal(t, len(globs), 2)

		val, err := globs[0].Value()
		assert.NilError(t, err)
		assert.Equal(t, globs[0].Name(), "foo")
		assert.Equal(t, val, int64(17))

		val, err = globs[1].Value()
		assert.NilError(t, err)
		assert.Equal(t, globs[1].Name(), "bar")
		assert.DeepEqual(t, val, []interface{}{
			Symbol("a"),
			Symbol("b"),
			Symbol("c"),
		})
	})

	t.Run("Successful find global", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)
		err = env.Build(`(defglobal ?*bar* = (create$ a b c))`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)
		val, err := glob.Value()
		assert.NilError(t, err)
		assert.Equal(t, glob.Name(), "foo")
		assert.Equal(t, val, int64(17))
	})

	t.Run("Unsuccessful find global", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)
		err = env.Build(`(defglobal ?*bar* = (create$ a b c))`)
		assert.NilError(t, err)

		_, err = env.FindGlobal("baz")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("Check global change", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		assert.Assert(t, env.GlobalsChanged())
		assert.Assert(t, !env.GlobalsChanged())

		err = env.Build(`(defglobal ?*foo* = 20)`)
		assert.NilError(t, err)
		assert.Assert(t, env.GlobalsChanged())
		assert.Assert(t, !env.GlobalsChanged())
	})
}

func TestGlobal(t *testing.T) {
	t.Run("Global basic values", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)
		assert.Equal(t, glob.Name(), "foo")
		assert.Equal(t, glob.String(), "(defglobal MAIN ?*foo* = 17)")

		glob2, err := env.FindGlobal("foo")
		assert.NilError(t, err)
		assert.Assert(t, glob.Equal(glob2))

		err = env.Build(`(defglobal ?*bar* = "another")`)
		assert.NilError(t, err)
		glob2, err = env.FindGlobal("bar")
		assert.NilError(t, err)
		assert.Assert(t, !glob.Equal(glob2))
	})

	t.Run("Set Global", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)

		glob.SetValue("newval")
		val, err := glob.Value()
		assert.NilError(t, err)
		assert.Equal(t, val, "newval")
	})

	t.Run("Deletable", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)

		assert.Equal(t, glob.Deletable(), true)

		err = env.Build(`(defrule globref "cause a reference to the global"
			(foo ?var)
		=>
			(printout t ?*foo*))`)
		assert.NilError(t, err)

		assert.Equal(t, glob.Deletable(), false)
	})

	t.Run("Watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)

		assert.Equal(t, glob.Watched(), false)
		glob.Watch(true)
		assert.Equal(t, glob.Watched(), true)
	})

	t.Run("Undefine", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)

		mod := glob.Module()
		assert.Equal(t, mod.Name(), "MAIN")
	})

	t.Run("Undefine", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defglobal ?*foo* = 17)`)
		assert.NilError(t, err)

		glob, err := env.FindGlobal("foo")
		assert.NilError(t, err)

		err = env.Build(`(defrule globref "cause a reference to the global"
			(foo ?var)
		=>
			(printout t ?*foo*))`)
		assert.NilError(t, err)

		assert.Equal(t, glob.Deletable(), false)
		err = glob.Undefine()
		assert.ErrorContains(t, err, "Unable")

		_, err = env.Eval(`(undefrule globref)`)
		assert.NilError(t, err)
	})
}
