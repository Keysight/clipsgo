package clips

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
