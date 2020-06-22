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

func TestError(t *testing.T) {
	t.Run("Error code", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		_, err := env.Eval("(create$ 1 2 3")
		assert.ErrorContains(t, err, "Unable to parse")
		shellError, ok := err.(*Error)
		assert.Assert(t, ok)
		assert.Equal(t, shellError.Code, "EXPRNPSR2")
	})

	t.Run("Error message", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		_, err := env.Eval("(create$ 1 2 3")
		assert.ErrorContains(t, err, "Unable to parse")
		assert.Equal(t, err.Error(), "Unable to parse construct \"(create$ 1 2 3\": [EXPRNPSR2] Expected a constant, variable, or expression.")
	})
}
