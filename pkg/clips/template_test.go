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

func TestTemplate(t *testing.T) {
	t.Run("Get Template", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)
		defer fact.Drop()

		assert.Equal(t, fact.Index(), 1)
		assert.Assert(t, fact.Asserted())

		tmpl := fact.Template()
		assert.Assert(t, tmpl.Implied())
		assert.Equal(t, tmpl.Name(), "foo")
	})

	t.Run("Template Equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)
		defer fact.Drop()

		tmpl := fact.Template()

		tmpl2, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		assert.Assert(t, tmpl.Equal(tmpl2))
	})

	t.Run("Template Not Equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)
		defer fact.Drop()

		tmpl := fact.Template()

		fact, err = env.AssertString(`(bar a b c)`)
		assert.NilError(t, err)
		defer fact.Drop()

		tmpl2 := fact.Template()

		assert.Assert(t, !tmpl.Equal(tmpl2))
	})

	t.Run("Template String", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		assert.Equal(t, tmpl.String(), `(deftemplate MAIN::foo
   (slot bar)
   (multislot baz))`)
	})

	t.Run("Template Watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		assert.Assert(t, !tmpl.Watched())
		tmpl.Watch(true)
		assert.NilError(t, err)
		assert.Assert(t, tmpl.Watched())
	})

	t.Run("Template Deletable", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		assert.Assert(t, tmpl.Deletable())

		_, err = env.AssertString("(foo (bar a))")
		assert.NilError(t, err)
		assert.Assert(t, !tmpl.Deletable())
	})

	t.Run("Unsuccessful Undefine Template", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		_, err = env.AssertString("(foo (bar a))")
		assert.NilError(t, err)
		assert.Assert(t, !tmpl.Deletable())

		err = tmpl.Undefine()
		assert.ErrorContains(t, err, "Unable to undefine")
	})

	t.Run("Successful Undefine Template", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		assert.Assert(t, tmpl.Deletable())

		err = tmpl.Undefine()
		assert.NilError(t, err)
	})

	t.Run("Module", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		mod := tmpl.Module()
		assert.Equal(t, mod.Name(), "MAIN")
	})
}

func TestTemplateSlot(t *testing.T) {
	t.Run("TemplateSlot basic values", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		bar, ok := slots["bar"]
		assert.Assert(t, ok)
		_, ok = slots["baz"]
		assert.Assert(t, ok)

		assert.Equal(t, bar.String(), "bar")
		assert.Equal(t, bar.Name(), "bar")
	})

	t.Run("TemplateSlot equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		bar, ok := slots["bar"]
		assert.Assert(t, ok)
		baz, ok := slots["baz"]
		assert.Assert(t, ok)

		assert.Assert(t, !bar.Equal(baz))
		slots = tmpl.Slots()
		bar2, ok := slots["bar"]
		assert.Assert(t, ok)
		assert.Assert(t, bar.Equal(bar2))
	})

	t.Run("TemplateSlot multifield", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		bar, ok := slots["bar"]
		assert.Assert(t, ok)
		baz, ok := slots["baz"]
		assert.Assert(t, ok)

		assert.Assert(t, !bar.Multifield())
		assert.Assert(t, baz.Multifield())
	})

	t.Run("TemplateSlot types", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(deftemplate foo
			(slot bar
				(type SYMBOL))
			(multislot baz
				(type STRING)))`)
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		bar, ok := slots["bar"]
		assert.Assert(t, ok)
		baz, ok := slots["baz"]
		assert.Assert(t, ok)

		barTypes := bar.Types()
		bazTypes := baz.Types()
		assert.Equal(t, len(barTypes), 1)
		assert.Equal(t, barTypes[0], Symbol("SYMBOL"))
		assert.Equal(t, len(bazTypes), 1)
		assert.Equal(t, bazTypes[0], Symbol("STRING"))
	})

	t.Run("TemplateSlot range", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(deftemplate foo
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

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		norange, ok := slots["norange"]
		assert.Assert(t, ok)
		lowrange, ok := slots["lowrange"]
		assert.Assert(t, ok)
		highrange, ok := slots["highrange"]
		assert.Assert(t, ok)
		fullrange, ok := slots["fullrange"]
		assert.Assert(t, ok)

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

	t.Run("TemplateSlot cardinality", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(deftemplate foo
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

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		notmulti, ok := slots["notmulti"]
		assert.Assert(t, ok)
		nocard, ok := slots["nocard"]
		assert.Assert(t, ok)
		lowcard, ok := slots["lowcard"]
		assert.Assert(t, ok)
		highcard, ok := slots["highcard"]
		assert.Assert(t, ok)
		fullcard, ok := slots["fullcard"]
		assert.Assert(t, ok)

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

	t.Run("TemplateSlot default", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(deftemplate foo
			(slot nodefault)
			(slot explicitno
				(default ?NONE))
			(slot static
				(default 4))
			(multislot dynamic
				(default-dynamic (gensym)))
		)`)
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		nodefault, ok := slots["nodefault"]
		assert.Assert(t, ok)
		explicitno, ok := slots["explicitno"]
		assert.Assert(t, ok)
		static, ok := slots["static"]
		assert.Assert(t, ok)
		dynamic, ok := slots["dynamic"]
		assert.Assert(t, ok)

		assert.Equal(t, nodefault.DefaultType(), STATIC_DEFAULT)
		assert.Equal(t, explicitno.DefaultType(), NO_DEFAULT)
		assert.Equal(t, static.DefaultType(), STATIC_DEFAULT)
		assert.Equal(t, dynamic.DefaultType(), DYNAMIC_DEFAULT)

		assert.Equal(t, nodefault.DefaultValue(), nil)
		assert.Equal(t, explicitno.DefaultValue(), Symbol("?NONE"))
		assert.Equal(t, static.DefaultValue(), int64(4))
		assert.DeepEqual(t, dynamic.DefaultValue(), []interface{}{Symbol("gen1")})
	})

	t.Run("TemplateSlot allowed values", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(deftemplate foo
			(slot bar)
			(slot baz
				(allowed-values a b c))
		)`)
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)

		slots := tmpl.Slots()
		bar, ok := slots["bar"]
		assert.Assert(t, ok)
		baz, ok := slots["baz"]
		assert.Assert(t, ok)

		_, ok = bar.AllowedValues()
		assert.Assert(t, !ok)
		av, ok := baz.AllowedValues()
		assert.Assert(t, ok)
		assert.DeepEqual(t, av, []interface{}{
			Symbol("a"),
			Symbol("b"),
			Symbol("c"),
		})
	})

}
