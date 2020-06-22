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

func TestTemplateFact(t *testing.T) {
	t.Run("Fact Index", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo)`)
		defer fact.Drop()
		assert.NilError(t, err)

		assert.Equal(t, fact.Index(), 1)
		assert.Assert(t, fact.Asserted())
	})

	t.Run("Fact String", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo)`)
		defer fact.Drop()
		assert.NilError(t, err)

		assert.Equal(t, fact.String(), "(foo (bar nil) (baz))")
	})

	t.Run("Fact Slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo (bar 4) (baz a b c))`)
		defer fact.Drop()
		assert.NilError(t, err)

		slots, err := fact.Slots()
		assert.NilError(t, err)
		assert.Equal(t, len(slots), 2)

		val, ok := slots["bar"]
		assert.Assert(t, ok)
		assert.Equal(t, val, int64(4))

		val, ok = slots["baz"]
		assert.Assert(t, ok)
		arr, ok := val.([]interface{})
		assert.Assert(t, ok)

		assert.DeepEqual(t, arr, []interface{}{
			Symbol("a"),
			Symbol("b"),
			Symbol("c"),
		})

		tfact, ok := fact.(*TemplateFact)
		assert.Assert(t, ok)
		val, err = tfact.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(4))

		_, err = tfact.Slot("bif")
		assert.ErrorContains(t, err, "Unable to get slot")
	})

	t.Run("Fact Extract Slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo (bar 4) (baz a b c))`)
		defer fact.Drop()
		assert.NilError(t, err)

		tf, ok := fact.(*TemplateFact)
		assert.Assert(t, ok)

		var intvar int
		err = tf.ExtractSlot(&intvar, "bar")
		assert.NilError(t, err)
		assert.Equal(t, intvar, 4)

		var msvar []Symbol
		err = tf.ExtractSlot(&msvar, "baz")
		assert.NilError(t, err)
		assert.DeepEqual(t, msvar, []Symbol{
			Symbol("a"),
			Symbol("b"),
			Symbol("c"),
		})
	})

	t.Run("Fact Extract", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo (bar 4) (baz a b c))`)
		defer fact.Drop()
		assert.NilError(t, err)

		type Foo struct {
			Bar int      `clips:"bar"`
			Baz []string `clips:"baz"`
		}
		var foovar Foo
		err = fact.Extract(&foovar)
		assert.NilError(t, err)
		assert.DeepEqual(t, foovar, Foo{
			Bar: 4,
			Baz: []string{
				"a", "b", "c",
			},
		})
	})

	t.Run("Fact Retract", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo)`)
		defer fact.Drop()
		assert.NilError(t, err)

		err = fact.Retract()
		assert.NilError(t, err)

		assert.Assert(t, !fact.Asserted())
	})

	t.Run("Assert already asserted", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo)`)
		defer fact.Drop()
		assert.NilError(t, err)
		err = fact.Assert()
		assert.ErrorContains(t, err, "Fact already asserted")
	})

	t.Run("Set and Assert", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		assert.Assert(t, !tmpl.Implied())
		fact, err := tmpl.NewFact()
		assert.NilError(t, err)

		ifact, ok := fact.(*TemplateFact)
		assert.Assert(t, ok)

		err = ifact.Set("bar", 4)
		assert.NilError(t, err)
		err = ifact.Set("baz", []interface{}{
			Symbol("b"),
			3,
		})
		assert.NilError(t, err)
		err = ifact.Set("bif", 3)
		assert.ErrorContains(t, err, "does not have slot")

		err = ifact.Assert()
		assert.NilError(t, err)

		err = ifact.Set("bar", 5)
		assert.ErrorContains(t, err, "Unable to change")
	})

	t.Run("Assert Defaults", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		assert.Assert(t, !tmpl.Implied())
		fact, err := tmpl.NewFact()
		assert.NilError(t, err)

		ifact, ok := fact.(*TemplateFact)
		assert.Assert(t, ok)

		err = ifact.Assert()
		assert.NilError(t, err)
	})

	t.Run("Bad Set", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar (type INTEGER)) (multislot baz (type STRING)))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		assert.Assert(t, !tmpl.Implied())
		fact, err := tmpl.NewFact()
		assert.NilError(t, err)

		ifact, ok := fact.(*TemplateFact)
		assert.Assert(t, ok)

		// multislot value in single
		err = ifact.Set("bar", []interface{}{
			Symbol("b"),
			3,
		})
		assert.ErrorContains(t, err, "Unable")
		err = ifact.Set("bar", Symbol("foo"))
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("Equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar (type INTEGER)) (multislot baz (type STRING)))")
		assert.NilError(t, err)

		fact, err := env.AssertString(`(foo)`)
		defer fact.Drop()
		assert.NilError(t, err)

		factlist := env.Facts()
		assert.Equal(t, len(factlist), 2)
		assert.Assert(t, fact.Equal(factlist[1]))
	})
}
