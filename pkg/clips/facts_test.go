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
	"io/ioutil"
	"os"
	"testing"

	"github.com/udhos/equalfile"
	"gotest.tools/assert"
)

func TestFacts(t *testing.T) {
	t.Run("Create Fact String", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)

		assert.Equal(t, ret.String(), "(foo a b c)")
	})

	t.Run("Iterate over facts", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		_, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)

		_, err = env.AssertString(`(bar)`)
		assert.NilError(t, err)

		_, err = env.AssertString(`(foo 1 2 3)`)
		assert.NilError(t, err)

		// There is an "initial fact" so count is one more than we created ourselves
		facts := env.Facts()
		assert.Equal(t, len(facts), 4)
	})

	t.Run("Load facts", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.LoadFactsFromString(`
		(initial-fact)
		(foo a b c)
		(bar)
		(foo 1 2 3)
		`)
		assert.NilError(t, err)

		// There is an initialfact to start with, so expect one extra
		assert.Equal(t, len(env.Facts()), 4)
	})

	t.Run("Load facts from file", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.LoadFacts("testdata/factfile.clp")
		assert.NilError(t, err)

		// There is an initialfact to start with, so expect one extra
		assert.Equal(t, len(env.Facts()), 4)
	})

	t.Run("Save facts", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		_, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)

		_, err = env.AssertString(`(bar)`)
		assert.NilError(t, err)

		_, err = env.AssertString(`(foo 1 2 3)`)
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.SaveFacts(tmpfile.Name(), LOCAL_SAVE)
		assert.NilError(t, err)

		cmp := equalfile.New(nil, equalfile.Options{})
		equal, err := cmp.CompareFile("testdata/factfile.clp", tmpfile.Name())
		assert.NilError(t, err)
		assert.Equal(t, equal, true)
	})

	t.Run("Get templates", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar))")
		assert.NilError(t, err)

		err = env.Build("(deftemplate bar (slot foo))")
		assert.NilError(t, err)

		templates := env.Templates()
		assert.Assert(t, templates != nil)
		assert.Equal(t, len(templates), 3)
		assert.Equal(t, templates[0].Name(), "initial-fact")
		assert.Equal(t, templates[1].Name(), "foo")
		assert.Equal(t, templates[2].Name(), "bar")
	})

	t.Run("Successful template lookup", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar))")
		assert.NilError(t, err)

		err = env.Build("(deftemplate bar (slot foo))")
		assert.NilError(t, err)

		template, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		assert.Equal(t, template.Name(), "foo")
	})

	t.Run("Failed template lookup", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar))")
		assert.NilError(t, err)

		err = env.Build("(deftemplate bar (slot foo))")
		assert.NilError(t, err)

		_, err = env.FindTemplate("bif")
		assert.ErrorContains(t, err, "not found")
	})
}
