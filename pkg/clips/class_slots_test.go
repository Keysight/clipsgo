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

func TestClassSlots(t *testing.T) {
	t.Run("ClassSlot basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slot, err := class.Slot("bar")
		assert.NilError(t, err)

		assert.Equal(t, slot.Name(), "bar")
		assert.Equal(t, slot.String(), "bar")
	})

	t.Run("ClassSlot basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)
		err = env.Build(`(defclass Bar (is-a USER) (slot bar))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)
		Bar, err := env.FindClass("Bar")
		assert.NilError(t, err)

		slot, err := Foo.Slot("bar")
		assert.NilError(t, err)

		slot2, err := Foo.Slot("bar")
		assert.NilError(t, err)

		assert.Assert(t, slot.Equal(slot2))

		slot2, err = Foo.Slot("baz")
		assert.NilError(t, err)
		assert.Assert(t, !slot.Equal(slot2))

		slot2, err = Bar.Slot("bar")
		assert.NilError(t, err)
		assert.Assert(t, !slot.Equal(slot2))
	})

	t.Run("ClassSlot queries", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slot, err := Foo.Slot("bar")
		assert.NilError(t, err)

		assert.Assert(t, !slot.Public())
		assert.Assert(t, slot.Initable())
		assert.Assert(t, slot.Writable())
		assert.Assert(t, slot.Accessible())
	})

	t.Run("ClassSlot types", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slot, err := Foo.Slot("bar")
		assert.NilError(t, err)

		types := slot.Types()
		assert.DeepEqual(t, types, []Symbol{
			Symbol("INTEGER"),
		})
	})

	t.Run("ClassSlot sources", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slot, err := Foo.Slot("bar")
		assert.NilError(t, err)

		sources := slot.Sources()
		assert.DeepEqual(t, sources, []Symbol{
			Symbol("Foo"),
		})
	})

	t.Run("ClassSlot range", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot norange
				(type INTEGER))
			(slot lowrange
				(type FLOAT)
				(range -10.0 ?VARIABLE))
			(slot highrange
				(type INTEGER)
				(range ?VARIABLE 10))
			(slot fullrange
				(type NUMBER)
				(range -15.0 10))
		)`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		norange, err := Foo.Slot("norange")
		assert.NilError(t, err)
		lowrange, err := Foo.Slot("lowrange")
		assert.NilError(t, err)
		highrange, err := Foo.Slot("highrange")
		assert.NilError(t, err)
		fullrange, err := Foo.Slot("fullrange")
		assert.NilError(t, err)

		_, hasLow, _, hasHigh := norange.IntRange()
		assert.Assert(t, !hasLow)
		assert.Assert(t, !hasHigh)
		_, hasLow, _, hasHigh = norange.FloatRange()
		assert.Assert(t, !hasLow)
		assert.Assert(t, !hasHigh)

		_, hasLow, _, hasHigh = lowrange.IntRange()
		assert.Assert(t, !hasLow)
		assert.Assert(t, !hasHigh)
		floatLow, hasLow, _, hasHigh := lowrange.FloatRange()
		assert.Assert(t, hasLow)
		assert.Assert(t, !hasHigh)
		assert.Equal(t, floatLow, float64(-10))

		_, hasLow, intHigh, hasHigh := highrange.IntRange()
		assert.Assert(t, !hasLow)
		assert.Assert(t, hasHigh)
		assert.Equal(t, intHigh, int64(10))
		_, hasLow, _, hasHigh = highrange.FloatRange()
		assert.Assert(t, !hasLow)
		assert.Assert(t, !hasHigh)

		_, hasLow, intHigh, hasHigh = fullrange.IntRange()
		assert.Assert(t, !hasLow)
		assert.Assert(t, hasHigh)
		assert.Equal(t, intHigh, int64(10))
		floatLow, hasLow, _, hasHigh = fullrange.FloatRange()
		assert.Assert(t, hasLow)
		assert.Assert(t, !hasHigh)
		assert.Equal(t, floatLow, float64(-15))
	})

	t.Run("ClassSlot facets", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slot, err := Foo.Slot("bar")
		assert.NilError(t, err)

		sources := slot.Facets()
		assert.DeepEqual(t, sources, []Symbol{
			"SGL",
			"STC",
			"INH",
			"RW",
			"LCL",
			"RCT",
			"EXC",
			"PRV",
			"RW",
			"put-bar",
		})
	})

	t.Run("ClassSlot cardinality", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
            (slot notmulti)
            (multislot nocard)
            (multislot lowcard
                (cardinality 2 ?VARIABLE))
            (multislot highcard
                (cardinality ?VARIABLE 7))
            (multislot fullcard
                (cardinality 2 7))
        )`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		notmulti, err := Foo.Slot("notmulti")
		assert.NilError(t, err)
		nocard, err := Foo.Slot("nocard")
		assert.NilError(t, err)
		lowcard, err := Foo.Slot("lowcard")
		assert.NilError(t, err)
		highcard, err := Foo.Slot("highcard")
		assert.NilError(t, err)
		fullcard, err := Foo.Slot("fullcard")
		assert.NilError(t, err)

		low, _, hasHigh := notmulti.Cardinality()
		assert.Assert(t, !hasHigh)
		assert.Equal(t, low, int64(0))

		_, _, hasHigh = nocard.Cardinality()
		assert.Assert(t, !hasHigh)
		assert.Equal(t, low, int64(0))

		low, _, hasHigh = lowcard.Cardinality()
		assert.Assert(t, !hasHigh)
		assert.Equal(t, low, int64(2))

		low, high, hasHigh := highcard.Cardinality()
		assert.Assert(t, hasHigh)
		assert.Equal(t, low, int64(0))
		assert.Equal(t, high, int64(7))

		low, high, hasHigh = fullcard.Cardinality()
		assert.Assert(t, hasHigh)
		assert.Equal(t, low, int64(2))
		assert.Equal(t, high, int64(7))
	})

	t.Run("ClassSlot default", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
            (slot nodefault)
            (slot explicitno
                (default ?NONE))
            (slot static
                (default 4))
            (multislot dynamic
                (default-dynamic (gensym)))
        )`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		nodefault, err := Foo.Slot("nodefault")
		assert.NilError(t, err)
		explicitno, err := Foo.Slot("explicitno")
		assert.NilError(t, err)
		static, err := Foo.Slot("static")
		assert.NilError(t, err)
		dynamic, err := Foo.Slot("dynamic")
		assert.NilError(t, err)

		assert.Equal(t, nodefault.DefaultValue(), nil)
		assert.Equal(t, explicitno.DefaultValue(), Symbol("?NONE"))
		assert.Equal(t, static.DefaultValue(), int64(4))
		assert.DeepEqual(t, dynamic.DefaultValue(), []interface{}{Symbol("gen1")})
	})

	t.Run("ClassSlot allowed values", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
            (slot bar)
            (slot baz
                (allowed-values a b c))
        )`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		bar, err := Foo.Slot("bar")
		assert.NilError(t, err)
		baz, err := Foo.Slot("baz")
		assert.NilError(t, err)

		_, ok := bar.AllowedValues()
		assert.Assert(t, !ok)
		av, ok := baz.AllowedValues()
		assert.Assert(t, ok)
		assert.DeepEqual(t, av, []interface{}{
			Symbol("a"),
			Symbol("b"),
			Symbol("c"),
		})
	})

	t.Run("ClassSlot allowed classes", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER)
			(slot bar)
			(slot baz
				(type INSTANCE-NAME)
                (allowed-classes Foo))
        )`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		bar, err := Foo.Slot("bar")
		assert.NilError(t, err)
		baz, err := Foo.Slot("baz")
		assert.NilError(t, err)

		_, ok := bar.AllowedClasses()
		assert.Assert(t, !ok)
		av, ok := baz.AllowedClasses()
		assert.Assert(t, ok)
		assert.DeepEqual(t, av, []Symbol{
			"Foo",
		})
	})
}
