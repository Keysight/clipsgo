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

func TestModulesEnv(t *testing.T) {
	t.Run("Current module", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		current := env.CurrentModule()
		assert.Equal(t, current.Name(), "MAIN")

		err := env.Build(`(defmodule Foo "lame module" (export ?ALL))`)
		assert.NilError(t, err)

		current = env.CurrentModule()
		assert.Equal(t, current.Name(), "Foo")
	})

	t.Run("List modules", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defmodule Foo "lame module" (export ?ALL))`)
		assert.NilError(t, err)
		err = env.Build(`(defmodule Bar "lame module" (export ?NONE))`)
		assert.NilError(t, err)

		modules := env.Modules()
		assert.Equal(t, len(modules), 3)
		assert.Equal(t, modules[0].Name(), "MAIN")
		assert.Equal(t, modules[1].Name(), "Foo")
		assert.Equal(t, modules[2].Name(), "Bar")
	})

	t.Run("Find module", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defmodule Foo "lame module" (export ?ALL))`)
		assert.NilError(t, err)
		err = env.Build(`(defmodule Bar "lame module" (export ?NONE))`)
		assert.NilError(t, err)

		module, err := env.FindModule("Foo")
		assert.NilError(t, err)
		assert.Equal(t, module.Name(), "Foo")

		_, err = env.FindModule("Bif")
		assert.ErrorContains(t, err, "not found")
	})
}

func TestModules(t *testing.T) {
	t.Run("Module basic values", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defmodule Foo "lame module" (export ?ALL))`)
		assert.NilError(t, err)

		module, err := env.FindModule("Foo")
		assert.NilError(t, err)
		assert.Equal(t, module.Name(), "Foo")
		assert.Equal(t, module.String(), `(defmodule Foo "lame module"
   (export ?ALL))`)
	})

	t.Run("Module equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defmodule Foo "lame module" (export ?ALL))`)
		assert.NilError(t, err)
		err = env.Build(`(defmodule Bar "lame module" (export ?NONE))`)
		assert.NilError(t, err)

		module, err := env.FindModule("Foo")
		assert.NilError(t, err)

		module2, err := env.FindModule("Foo")
		assert.NilError(t, err)
		assert.Assert(t, module.Equal(module2))

		module2, err = env.FindModule("Bar")
		assert.NilError(t, err)
		assert.Assert(t, !module.Equal(module2))
	})
}
