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

func TestAgendaEnv(t *testing.T) {
	t.Run("AgendaChanged", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		assert.Assert(t, !env.AgendaChanged())
		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		assert.Assert(t, env.AgendaChanged())
	})

	t.Run("Focus", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()
		err := env.Build(`(defmodule Foo "lame module")`)
		assert.NilError(t, err)

		f := env.Focus()
		assert.Equal(t, f.Name(), "MAIN")

		f, err = env.FindModule("Foo")
		assert.NilError(t, err)

		env.SetFocus(f)

		f = env.Focus()
		assert.Equal(t, f.Name(), "Foo")
	})

	t.Run("Strategy", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()
		assert.Equal(t, env.Strategy(), DEPTH)

		env.SetStrategy(BREADTH)

		assert.Equal(t, env.Strategy(), BREADTH)
	})

	t.Run("SalienceEvaluation", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()
		assert.Equal(t, env.SalienceEvaluation(), WHEN_DEFINED)

		env.SetSalienceEvaluation(WHEN_ACTIVATED)

		assert.Equal(t, env.SalienceEvaluation(), WHEN_ACTIVATED)
	})

	t.Run("List Rules", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		rules := env.Rules()
		assert.Equal(t, len(rules), 2)
		assert.Equal(t, rules[0].Name(), "foo")
		assert.Equal(t, rules[1].Name(), "bar")
	})

	t.Run("find Rule", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)
		assert.Equal(t, rule.Name(), "foo")

		_, err = env.FindRule("baz")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("Reorder", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		acts := env.Activations()
		assert.Equal(t, len(acts), 2)
		assert.Equal(t, acts[0].Name(), "bar")
		assert.Equal(t, acts[1].Name(), "foo")

		env.SetStrategy(BREADTH)
		env.Reorder(nil)

		acts = env.Activations()
		assert.Equal(t, len(acts), 2)
		assert.Equal(t, acts[0].Name(), "foo")
		assert.Equal(t, acts[1].Name(), "bar")

		env.SetStrategy(MEA)
		env.Reorder(env.Focus())
	})

	t.Run("Refresh", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		acts := env.Activations()
		assert.Equal(t, len(acts), 2)
		assert.Equal(t, acts[0].Name(), "bar")
		assert.Equal(t, acts[1].Name(), "foo")

		env.SetStrategy(BREADTH)
		env.Refresh(nil)

		acts = env.Activations()
		assert.Equal(t, len(acts), 2)
		assert.Equal(t, acts[0].Name(), "foo")
		assert.Equal(t, acts[1].Name(), "bar")

		env.SetStrategy(MEA)
		env.Refresh(env.Focus())
	})

	t.Run("Activations", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		acts := env.Activations()
		assert.Equal(t, len(acts), 2)
		assert.Equal(t, acts[0].Name(), "bar")
		assert.Equal(t, acts[1].Name(), "foo")
	})

	t.Run("Clear agenda", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		acts := env.Activations()
		assert.Equal(t, len(acts), 2)

		err = env.ClearAgenda()
		assert.NilError(t, err)

		acts = env.Activations()
		assert.Equal(t, len(acts), 0)
	})

	t.Run("Clear Focus", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()
		err := env.Build(`(defmodule Foo "lame module")`)
		assert.NilError(t, err)

		f := env.Focus()
		assert.Equal(t, f.Name(), "MAIN")

		f, err = env.FindModule("Foo")
		assert.NilError(t, err)

		env.SetFocus(f)

		f = env.Focus()
		assert.Equal(t, f.Name(), "Foo")

		env.ClearFocus()
		env.SetFocus(f)
		// focus at this point is nil and has to be set
		f = env.Focus()
		assert.Equal(t, f.Name(), "Foo")
	})

	t.Run("Run", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule baz => (printout t "fired"))`)
		assert.NilError(t, err)

		acts := env.Activations()
		assert.Equal(t, len(acts), 3)

		ran := env.Run(1)
		assert.Equal(t, ran, int64(1))
		acts = env.Activations()
		assert.Equal(t, len(acts), 2)

		// I don't know why this possibility is supported, but it is
		ran = env.Run(0)
		assert.Equal(t, ran, int64(0))
		acts = env.Activations()
		assert.Equal(t, len(acts), 2)

		ran = env.Run(-1)
		assert.Equal(t, ran, int64(2))
		acts = env.Activations()
		assert.Equal(t, len(acts), 0)
	})
}

func TestRule(t *testing.T) {
	t.Run("Rule basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)
		assert.Equal(t, rule.Name(), "foo")
		assert.Equal(t, rule.String(), `(defrule MAIN::foo
   =>
   (printout t "fired"))`)
	})

	t.Run("Rule equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)
		rule2, err := env.FindRule("foo")
		assert.NilError(t, err)
		assert.Assert(t, rule.Equal(rule2))

		rule2, err = env.FindRule("bar")
		assert.NilError(t, err)
		assert.Assert(t, !rule.Equal(rule2))
	})

	t.Run("Module", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)

		mod := rule.Module()
		assert.Equal(t, mod.Name(), "MAIN")
	})

	t.Run("Deletable", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)

		assert.Assert(t, rule.Deletable())
	})

	t.Run("Watch firings", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)

		assert.Assert(t, !rule.WatchedFirings())
		rule.WatchFirings(true)
		assert.Assert(t, rule.WatchedFirings())
	})

	t.Run("Watch activations", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)

		assert.Assert(t, !rule.WatchedActivations())
		rule.WatchActivations(true)
		assert.Assert(t, rule.WatchedActivations())
	})

	t.Run("Matches", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)

		ms, err := rule.Matches(TERSE)
		assert.NilError(t, err)
		assert.DeepEqual(t, ms, []interface{}{
			int64(1),
			int64(0),
			int64(1),
		})
	})

	t.Run("Refresh", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)
		rule.Refresh()
		// TODO find some way to validate something happened
	})

	t.Run("Breakpoint", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)
		rule.AddBreakpoint()
		err = rule.RemoveBreakpoint()
		assert.NilError(t, err)
	})

	t.Run("Breakpoint", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		rule, err := env.FindRule("foo")
		assert.NilError(t, err)
		rule.Undefine()
		_, err = env.FindRule("foo")
		assert.ErrorContains(t, err, "not found")
	})
}

func TestActivations(t *testing.T) {
	t.Run("Activation basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)

		activations := env.Activations()
		assert.Equal(t, len(activations), 1)

		activation := activations[0]

		assert.Equal(t, activation.Name(), "foo")
		assert.Equal(t, activation.String(), `0      foo: *`)
	})

	t.Run("Activation equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		activations := env.Activations()
		assert.Equal(t, len(activations), 2)

		activation := activations[0]

		activations = env.Activations()
		assert.Equal(t, len(activations), 2)
		activation2 := activations[0]

		assert.Assert(t, activation.Equal(activation2))
		activation2 = activations[1]
		assert.Assert(t, !activation.Equal(activation2))
	})

	t.Run("Activation salience", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo (declare (salience 100)) => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		activations := env.Activations()
		assert.Equal(t, len(activations), 2)

		activation := activations[0]
		assert.Equal(t, activation.Salience(), 100)
		activation.SetSalience(-100)
		assert.Equal(t, activation.Salience(), -100)
	})

	t.Run("Activation remove", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defrule foo (declare (salience 100)) => (printout t "fired"))`)
		assert.NilError(t, err)
		err = env.Build(`(defrule bar => (printout t "fired"))`)
		assert.NilError(t, err)

		activations := env.Activations()
		assert.Equal(t, len(activations), 2)

		activation := activations[0]
		err = activation.Remove()
		assert.NilError(t, err)

		activations = env.Activations()
		assert.Equal(t, len(activations), 1)
	})
}
