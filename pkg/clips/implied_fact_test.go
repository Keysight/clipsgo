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

func TestImpliedFact(t *testing.T) {
	t.Run("Fact Index", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		assert.Equal(t, fact.Index(), 1)
		assert.Assert(t, fact.Asserted())
	})

	t.Run("Fact String", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		assert.Equal(t, len(fact.String()), len("(foo a b c)"))
		assert.Equal(t, fact.String(), "(foo a b c)")
	})

	t.Run("Fact Extract", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		var retval []string
		err = fact.Extract(&retval)
		assert.NilError(t, err)
		assert.DeepEqual(t, retval, []string{
			"a", "b", "c",
		})
	})

	t.Run("Fact Slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		slots, err := fact.Slots()
		assert.NilError(t, err)
		assert.Equal(t, len(slots), 1)

		val, ok := slots[""]
		assert.Assert(t, ok)

		arr, ok := val.([]interface{})
		assert.Assert(t, ok)

		assert.Equal(t, len(arr), 3)
		assert.Equal(t, arr[0], Symbol("a"))
		assert.Equal(t, arr[1], Symbol("b"))
		assert.Equal(t, arr[2], Symbol("c"))

		val, err = fact.Slot("")
		assert.NilError(t, err)

		arr, ok = val.([]interface{})
		assert.Assert(t, ok)
		assert.Equal(t, len(arr), 3)
		assert.Equal(t, arr[0], Symbol("a"))
		assert.Equal(t, arr[1], Symbol("b"))
		assert.Equal(t, arr[2], Symbol("c"))

		var strval []string
		err = fact.ExtractSlot(&strval, "")
		assert.NilError(t, err)
		assert.DeepEqual(t, strval, []string{
			"a", "b", "c",
		})
	})

	t.Run("Fact Retract", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		err = fact.Retract()
		assert.NilError(t, err)

		assert.Assert(t, !fact.Asserted())
	})

	t.Run("Assert already asserted", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)
		err = fact.Assert()
		assert.ErrorContains(t, err, "Fact already asserted")
	})

	t.Run("Append and Assert", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		tmpl := fact.Template()
		assert.Assert(t, tmpl.Implied())
		fact, err = tmpl.NewFact()
		assert.NilError(t, err)

		ifact, ok := fact.(*ImpliedFact)
		assert.Assert(t, ok)

		err = ifact.Append("a")
		assert.NilError(t, err)
		ifact.Extend([]interface{}{
			Symbol("b"),
			3,
		})
		assert.NilError(t, err)

		err = ifact.Set(2, "c")
		assert.NilError(t, err)

		err = ifact.Assert()
		assert.NilError(t, err)

		err = ifact.Append("a")
		assert.ErrorContains(t, err, "Unable to change")

		err = ifact.Extend([]interface{}{
			"a",
			Symbol("b"),
		})
		assert.ErrorContains(t, err, "Unable to change")

	})

	t.Run("Read slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		slots, err := fact.Slots()
		assert.NilError(t, err)
		assert.Equal(t, len(slots), 1)

		slot, ok := slots[""]
		assert.Assert(t, ok)
		assert.DeepEqual(t, slot, []interface{}{
			Symbol("a"),
			Symbol("b"),
			Symbol("c"),
		})
	})

	t.Run("Equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		defer fact.Drop()
		assert.NilError(t, err)

		factlist := env.Facts()
		assert.Equal(t, len(factlist), 2)
		assert.Assert(t, fact.Equal(factlist[1]))
	})

}
