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

func TestCreateEnvironment(t *testing.T) {
	t.Run("Explicit delete", func(t *testing.T) {
		env := CreateEnvironment()
		env.Delete()
	})

	t.Run("Load text", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Load("testdata/dopey.save")
		assert.NilError(t, err)
	})

	t.Run("Load binary", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Load("testdata/dopey.bsave")
		assert.NilError(t, err)
	})

	t.Run("Load failure", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Load("testdata/file_not_found")
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("Save text", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Load("testdata/dopey.save")
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.Save(tmpfile.Name(), false)
		assert.NilError(t, err)

		cmp := equalfile.New(nil, equalfile.Options{})
		equal, err := cmp.CompareFile("testdata/dopey.save", tmpfile.Name())
		assert.NilError(t, err)
		assert.Equal(t, equal, true)
	})

	t.Run("Save binary", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Load("testdata/dopey.save")
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.Save(tmpfile.Name(), true)
		assert.NilError(t, err)

		// Binary output is not consistent; not sure how to verify
		/*
			cmp := equalfile.New(nil, equalfile.Options{})
			equal, err := cmp.CompareFile("testdata/dopey.bsave", tmpfile.Name())
			assert.NilError(t, err)
			assert.Equal(t, equal, true)
		*/
	})

	t.Run("Save failure", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Load("testdata/dopey.save")
		assert.NilError(t, err)

		err = env.Save("/not_writable", true)
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("BatchStar", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.BatchStar("testdata/dopey.clp")
		assert.NilError(t, err)
	})

	t.Run("BatchStar failure", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.BatchStar("testdata/file_not_found")
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("Build", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar))")
		assert.NilError(t, err)
	})

	t.Run("Build failure", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar")
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("Eval", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		ret, err := env.Eval("(rules)")
		assert.NilError(t, err)
		assert.Equal(t, ret, nil)
	})

	t.Run("Eval Failure", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		_, err := env.Eval("(create$ 1 2 3")
		assert.ErrorContains(t, err, "Unable to parse")
	})

	t.Run("Clear", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		env.Clear()
	})

	t.Run("Reset", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		env.Reset()
	})

	t.Run("DefineFunction", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		argcount := 0
		callback := func(args []interface{}) {
			argcount = len(args)
		}

		err := env.DefineFunction("test-callback", callback)
		assert.NilError(t, err)

		_, err = env.Eval("(test-callback (create$ a b c))")
		assert.NilError(t, err)
		assert.Equal(t, argcount, 3)
	})

	t.Run("CompleteCommand", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		complete, err := env.CompleteCommand("(deftemplate bar (slot bar))")
		assert.NilError(t, err)
		assert.Assert(t, complete)
		complete, err = env.CompleteCommand("(deftemplate bar (slot bar)")
		assert.NilError(t, err)
		assert.Assert(t, !complete)
	})

	t.Run("SendCommand", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		// try some stuff that Eval chokes on
		err := env.SendCommand("(assert (foo a b c))")
		assert.NilError(t, err)

		err = env.SendCommand("(deftemplate bar (slot bar))")
		assert.NilError(t, err)

		err = env.SendCommand("(deftemplate baz (slot bar")
		assert.ErrorContains(t, err, "Syntax Error")

		err = env.SendCommand("(deftemplate foo (slot bar))")
		assert.ErrorContains(t, err, "in use")
	})
}
