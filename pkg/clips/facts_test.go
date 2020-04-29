package clips

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
		defer env.Close()

		_, err := env.InsertString(`(foo a b c)`)
		assert.NilError(t, err)

		// TODO check out ret
	})

	t.Run("Iterate over facts", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		_, err := env.InsertString(`(foo a b c)`)
		assert.NilError(t, err)

		_, err = env.InsertString(`(bar)`)
		assert.NilError(t, err)

		_, err = env.InsertString(`(foo 1 2 3)`)
		assert.NilError(t, err)

		count := 0
		nextFact := env.FactIterator()
		fact := nextFact()
		for ; fact != nil; fact = nextFact() {
			count++
			// TODO assert something about fact
		}
		// There is an initialfact to start with, so expect one extra
		assert.Equal(t, count, 4)
	})

	t.Run("Load facts", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		err := env.LoadFacts(`
		(initial-fact)
		(foo a b c)
		(bar)
		(foo 1 2 3)
		`)
		assert.NilError(t, err)

		count := 0
		nextFact := env.FactIterator()
		fact := nextFact()
		for ; fact != nil; fact = nextFact() {
			count++
			// TODO assert something about fact
		}
		// There is an initialfact to start with, so expect one extra
		assert.Equal(t, count, 4)
	})

	t.Run("Load facts from file", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		err := env.LoadFactsFromFile("testdata/factfile.clp")
		assert.NilError(t, err)

		count := 0
		nextFact := env.FactIterator()
		fact := nextFact()
		for ; fact != nil; fact = nextFact() {
			count++
			// TODO assert something about fact
		}
		// There is an initialfact to start with, so expect one extra
		assert.Equal(t, count, 4)
	})

	t.Run("Save facts", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Close()

		_, err := env.InsertString(`(foo a b c)`)
		assert.NilError(t, err)

		_, err = env.InsertString(`(bar)`)
		assert.NilError(t, err)

		_, err = env.InsertString(`(foo 1 2 3)`)
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
}
