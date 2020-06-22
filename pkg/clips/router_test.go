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
	"bytes"
	"log"
	"testing"

	"gotest.tools/assert"
)

func TestLoggingRouter(t *testing.T) {
	t.Run("Create LoggingRouter", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		var buf bytes.Buffer
		logger := log.New(&buf, "logger: ", 0)
		lr := CreateLoggingRouter(env, logger)

		_, err := env.Eval(`(printout t "Testing" crlf)`)
		assert.NilError(t, err)
		_, err = env.Eval(`(printout t "1 2 3")`)
		assert.NilError(t, err)
		// unfinished line should be buffered
		assert.Equal(t, buf.String(), "logger: Testing\n")

		// finishing the line should make it appear
		buf.Reset()
		_, err = env.Eval(`(printout t crlf)`)
		assert.NilError(t, err)
		assert.Equal(t, buf.String(), "logger: 1 2 3\n")

		assert.Equal(t, lr.Name(), "go-logging-router")
		err = lr.Delete()
		assert.NilError(t, err)

		_, err = env.Eval(`(printout t "Testing" crlf)`)
		assert.NilError(t, err)
	})
}
