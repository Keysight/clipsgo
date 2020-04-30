package clips

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
		defer fact.Delete()

		assert.Equal(t, fact.Index(), 1)
		assert.Assert(t, fact.Asserted())

		tmpl := fact.Template()
		defer tmpl.Delete()
		assert.Assert(t, tmpl.Implied())
		assert.Equal(t, tmpl.Name(), "foo")
	})

	t.Run("Template Equals", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)
		defer fact.Delete()

		tmpl := fact.Template()
		defer tmpl.Delete()

		tmpl2, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		defer tmpl2.Delete()

		assert.Assert(t, tmpl.Equals(tmpl2))
	})

	t.Run("Template Not Equals", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		fact, err := env.AssertString(`(foo a b c)`)
		assert.NilError(t, err)
		defer fact.Delete()

		tmpl := fact.Template()
		defer tmpl.Delete()

		fact, err = env.AssertString(`(bar a b c)`)
		assert.NilError(t, err)
		defer fact.Delete()

		tmpl2 := fact.Template()
		defer tmpl2.Delete()

		assert.Assert(t, !tmpl.Equals(tmpl2))
	})

	t.Run("Template String", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		defer tmpl.Delete()

		assert.Equal(t, tmpl.String(), `(deftemplate MAIN::foo
   (slot bar)
   (multislot baz))
`)
	})

	// TODO Module

	t.Run("Template Watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build("(deftemplate foo (slot bar) (multislot baz))")
		assert.NilError(t, err)

		tmpl, err := env.FindTemplate("foo")
		assert.NilError(t, err)
		defer tmpl.Delete()

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
		defer tmpl.Delete()

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
		defer tmpl.Delete()

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
		defer tmpl.Delete()

		assert.Assert(t, tmpl.Deletable())

		err = tmpl.Undefine()
		assert.NilError(t, err)
	})

	// TODO TemplateSlots
}
