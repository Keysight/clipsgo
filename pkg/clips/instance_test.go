package clips

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/udhos/equalfile"
	"gotest.tools/assert"
)

func TestInstanceEnv(t *testing.T) {
	t.Run("Instances Changed", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER))`)
		assert.NilError(t, err)

		assert.Assert(t, env.InstancesChanged())
		assert.Assert(t, !env.InstancesChanged())
		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		assert.Assert(t, env.InstancesChanged())
	})

	t.Run("Instances", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), "initial-object")
		assert.Equal(t, insts[1].Name(), "gen1")
		assert.Equal(t, insts[2].Name(), "named")
	})

	t.Run("Find instance", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		inst, err := env.FindInstance("named", "")
		assert.Equal(t, inst.Name(), "named")

		_, err = env.FindInstance("foo", "")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("Save instances text", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.SaveInstances(tmpfile.Name(), false, LOCAL_SAVE)
		assert.NilError(t, err)

		cmp := equalfile.New(nil, equalfile.Options{})
		equal, err := cmp.CompareFile("testdata/instancesfile.clp", tmpfile.Name())
		assert.NilError(t, err)
		assert.Equal(t, equal, true)
	})

	t.Run("Save instances binary", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(named of Foo)`)
		assert.NilError(t, err)

		tmpfile, err := ioutil.TempFile("", "test.*.save")
		assert.NilError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		err = env.SaveInstances(tmpfile.Name(), true, LOCAL_SAVE)
		assert.NilError(t, err)

		/*
			cmp := equalfile.New(nil, equalfile.Options{})
			equal, err := cmp.CompareFile("testdata/instances.bsave", tmpfile.Name())
			assert.NilError(t, err)
			assert.Assert(t, equal)
		*/
	})
	t.Run("Load instances text", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.LoadInstances("testdata/instancesfile.clp")
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), "initial-object")
		assert.Equal(t, insts[1].Name(), "gen1")
		assert.Equal(t, insts[2].Name(), "named")
	})

	t.Run("Load instances binary", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.LoadInstances("testdata/instances.bsave")
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), "initial-object")
		assert.Equal(t, insts[1].Name(), "gen1")
		assert.Equal(t, insts[2].Name(), "named")
	})

	t.Run("Load instances string", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.LoadInstancesFromString(`
		([initial-object] of INITIAL-OBJECT)

		([gen1] of Foo
		   (bar nil)
		   (baz))
		
		([named] of Foo
		   (bar nil)
		   (baz))
		`)
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), "initial-object")
		assert.Equal(t, insts[1].Name(), "gen1")
		assert.Equal(t, insts[2].Name(), "named")
	})

	t.Run("Restore instances", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.RestoreInstances("testdata/instancesfile.clp")
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), "initial-object")
		assert.Equal(t, insts[1].Name(), "gen1")
		assert.Equal(t, insts[2].Name(), "named")
	})

	t.Run("Restore instances string", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		err = env.RestoreInstancesFromString(`
		([initial-object] of INITIAL-OBJECT)

		([gen1] of Foo
		   (bar nil)
		   (baz))
		
		([named] of Foo
		   (bar nil)
		   (baz))
		`)
		assert.NilError(t, err)

		insts := env.Instances()
		assert.Assert(t, insts != nil)
		assert.Equal(t, len(insts), 3)
		assert.Equal(t, insts[0].Name(), "initial-object")
		assert.Equal(t, insts[1].Name(), "gen1")
		assert.Equal(t, insts[2].Name(), "named")
	})

	t.Run("Make instance", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		defer inst.Drop()
		assert.NilError(t, err)
		assert.Equal(t, inst.Name(), "gen1")

		/* CLIPS actually doesn't error on this
		_, err = env.MakeInstance(`(of Foo (bar "testing"))`)
		assert.ErrorContains(t, err, "invalid")
		*/
	})
}

func TestInstance(t *testing.T) {
	t.Run("Instance basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)
		assert.Equal(t, inst.Name(), "gen1")
		assert.Equal(t, inst.String(), "[gen1] of Foo (bar 12)")
	})

	t.Run("Instance equals", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		inst2, err := env.FindInstance("gen1", "")
		assert.NilError(t, err)

		assert.Assert(t, inst.Equals(inst2))

		inst2, err = env.MakeInstance(`(of Foo (bar 77))`)
		assert.NilError(t, err)
		assert.Assert(t, !inst.Equals(inst2))
	})

	t.Run("Class", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		class := inst.Class()
		assert.Equal(t, class.Name(), "Foo")
	})

	t.Run("Slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		slots := inst.Slots(true)
		assert.NilError(t, err)
		assert.Equal(t, len(slots), 2)
		_, ok := slots["bar"]
		assert.Assert(t, ok)
		_, ok = slots["baz"]
		assert.Assert(t, ok)
	})

	t.Run("Slot", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(12))

		ret, err = inst.Slot("bif")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("SetSlot", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(12))

		err = inst.SetSlot("bar", 77)
		assert.NilError(t, err)

		ret, err = inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(77))

		err = inst.SetSlot("bif", 77)
		assert.ErrorContains(t, err, "Unable")
	})

	t.Run("SetSlot type violation", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz (type INTEGER)))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(12))

		err = inst.SetSlot("bar", "forty two")
		// I would think this should be rejected by CLIPS, but it isn't
		assert.NilError(t, err)

		err = inst.SetSlot("baz", []string{"forty two"})
		assert.NilError(t, err)
	})

	t.Run("Send", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		ret := inst.Send("get-bar", "")
		assert.Equal(t, ret, int64(12))

		ret = inst.Send("put-bar", "77")
		assert.Equal(t, ret, int64(77))

		ret, err = inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(77))

		ret = inst.Send("garbage", "")
		assert.Equal(t, ret, false)
	})

	t.Run("Delete", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		err = inst.Delete()
		assert.NilError(t, err)
	})

	t.Run("Unmake", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		inst, err := env.MakeInstance(`(of Foo (bar 12))`)
		assert.NilError(t, err)

		err = inst.Unmake()
		assert.NilError(t, err)
	})
}
