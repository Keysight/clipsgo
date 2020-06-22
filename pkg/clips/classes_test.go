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

func TestClassEnv(t *testing.T) {
	t.Run("defaults mode", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		assert.Equal(t, env.ClassDefaultsMode(), CONVENIENCE_MODE)
		env.SetClassDefaultsMode(CONSERVATION_MODE)
		assert.Equal(t, env.ClassDefaultsMode(), CONSERVATION_MODE)
	})

	t.Run("Classes", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)
		err = env.Build(`(defclass Bar (is-a USER))`)
		assert.NilError(t, err)

		classes := env.Classes()

		assert.Equal(t, len(classes), 19)
		assert.Equal(t, classes[17].Name(), "Foo")
		assert.Equal(t, classes[18].Name(), "Bar")
	})

	t.Run("FindClass", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		assert.Equal(t, class.Name(), "Foo")
	})
}

func TestClass(t *testing.T) {
	t.Run("Class basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		assert.Equal(t, class.Name(), "Foo")
		assert.Equal(t, class.String(), `(defclass MAIN::Foo
   (is-a USER)
   (slot bar)
   (multislot baz))`)
	})

	t.Run("Class equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)
		err = env.Build(`(defclass Bar (is-a USER))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)
		class2, err := env.FindClass("Foo")
		assert.NilError(t, err)

		assert.Assert(t, class.Equal(class2))

		class2, err = env.FindClass("Bar")
		assert.NilError(t, err)
		assert.Assert(t, !class.Equal(class2))
	})

	t.Run("Class queries", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		assert.Assert(t, !class.Abstract())
		assert.Assert(t, class.Reactive())
		assert.Assert(t, class.Deletable())
	})

	t.Run("Class module", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mod := class.Module()
		assert.Equal(t, mod.Name(), "MAIN")
	})

	t.Run("Class watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		assert.Assert(t, !class.WatchedInstances())
		class.WatchInstances(true)
		assert.Assert(t, class.WatchedInstances())

		assert.Assert(t, !class.WatchedSlots())
		class.WatchSlots(true)
		assert.Assert(t, class.WatchedSlots())
	})

	t.Run("New instance", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar (type INTEGER)) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		inst, err := class.NewInstance("named", false)
		assert.NilError(t, err)
		assert.Equal(t, inst.Name(), InstanceName("named"))
		initval, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, initval, int64(0))

		inst, err = class.NewInstance("named", true)
		assert.NilError(t, err)
		assert.Equal(t, inst.Name(), InstanceName("named"))

		// try and retrieve an uninitialized slot
		ret, err := inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, nil)

		// now initialize it
		inst.SetSlot("bar", 7)
		ret, err = inst.Slot("bar")
		assert.NilError(t, err)
		assert.Equal(t, ret, int64(7))
	})

	t.Run("MessageHandlers", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mhs := class.MessageHandlers()
		assert.Equal(t, len(mhs), 4)
		assert.Equal(t, mhs[0].Name(), "get-bar")
		assert.Equal(t, mhs[1].Name(), "put-bar")
		assert.Equal(t, mhs[2].Name(), "get-baz")
		assert.Equal(t, mhs[3].Name(), "put-baz")
	})

	t.Run("Find MessageHandler", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mh, err := class.FindMessageHandler("get-bar", PRIMARY)
		assert.NilError(t, err)
		assert.Equal(t, mh.Name(), "get-bar")

		_, err = class.FindMessageHandler("get-bar", BEFORE)
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("sub / super class", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)
		err = env.Build(`(defclass Bar (is-a Foo))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)
		Bar, err := env.FindClass("Bar")
		assert.NilError(t, err)

		assert.Assert(t, !Foo.Subclass(Bar))
		assert.Assert(t, Bar.Subclass(Foo))

		assert.Assert(t, Foo.Superclass(Bar))
		assert.Assert(t, !Bar.Superclass(Foo))
	})

	t.Run("Slots", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slots := Foo.Slots(true)
		assert.Equal(t, len(slots), 2)
		assert.Equal(t, slots[0].Name(), "bar")
		assert.Equal(t, slots[1].Name(), "baz")
	})

	t.Run("Slot", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		slot, err := Foo.Slot("bar")
		assert.Equal(t, slot.Name(), "bar")

		_, err = Foo.Slot("bif")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("Instances", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)
		_, err = env.MakeInstance(`(of Foo)`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		insts := Foo.Instances()
		assert.Equal(t, len(insts), 2)
		assert.Equal(t, insts[0].Name(), InstanceName("gen1"))
		assert.Equal(t, insts[1].Name(), InstanceName("gen2"))
	})

	t.Run("sub / super classes", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)
		err = env.Build(`(defclass Bar (is-a Foo))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)
		Bar, err := env.FindClass("Bar")
		assert.NilError(t, err)

		subc, err := Foo.Subclasses(true)
		assert.NilError(t, err)
		assert.Equal(t, len(subc), 1)
		assert.Equal(t, subc[0].Name(), "Bar")

		subc, err = Bar.Subclasses(true)
		assert.NilError(t, err)
		assert.Equal(t, len(subc), 0)

		supc, err := Bar.Superclasses(false)
		assert.NilError(t, err)
		assert.Equal(t, len(supc), 1)
		assert.Equal(t, supc[0].Name(), "Foo")
	})

	t.Run("Undefine", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		Foo, err := env.FindClass("Foo")
		assert.NilError(t, err)

		err = Foo.Undefine()
		assert.NilError(t, err)

		_, err = env.FindClass("Foo")
		assert.ErrorContains(t, err, "not found")
	})
}

func TestMessageHandler(t *testing.T) {
	t.Run("MessageHandler basics", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mh, err := class.FindMessageHandler("get-bar", PRIMARY)
		assert.NilError(t, err)
		assert.Equal(t, mh.Name(), "get-bar")
		assert.Equal(t, mh.String(), "")
		assert.Equal(t, mh.Type(), PRIMARY)
		assert.Equal(t, !mh.Deletable(), true)
	})

	t.Run("MessageHandler equal", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mh, err := class.FindMessageHandler("get-bar", PRIMARY)
		assert.NilError(t, err)
		mh2, err := class.FindMessageHandler("get-bar", PRIMARY)
		assert.NilError(t, err)
		assert.Assert(t, mh.Equal(mh2))

		mh2, err = class.FindMessageHandler("put-bar", PRIMARY)
		assert.NilError(t, err)
		assert.Assert(t, !mh.Equal(mh2))
	})

	t.Run("MessageHandler watch", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mh, err := class.FindMessageHandler("get-bar", PRIMARY)
		assert.NilError(t, err)

		assert.Assert(t, !mh.Watched())
		mh.Watch(true)
		assert.Assert(t, mh.Watched())
	})

	t.Run("MessageHandler undefine", func(t *testing.T) {
		env := CreateEnvironment()
		defer env.Delete()

		err := env.Build(`(defclass Foo (is-a USER) (slot bar) (multislot baz))`)
		assert.NilError(t, err)

		class, err := env.FindClass("Foo")
		assert.NilError(t, err)

		mh, err := class.FindMessageHandler("get-bar", PRIMARY)
		assert.NilError(t, err)

		err = mh.Undefine()
		assert.NilError(t, err)

		_, err = class.FindMessageHandler("get-bar", PRIMARY)
		assert.ErrorContains(t, err, "not found")
	})
}
