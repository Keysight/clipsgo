package clips

import (
	"bytes"
	"log"
	"testing"

	"gotest.tools/assert"
)

func TestLoggingRouter(t *testing.T) {
	t.Run("Create LoggingRouter", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		var buf bytes.Buffer
		logger := log.New(&buf, "logger: ", 0)
		CreateLoggingRouter(env, logger)

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
	})
}
