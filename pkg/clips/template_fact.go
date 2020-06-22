package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
import "C"
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
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

// TemplateFact is an unordered fact
type TemplateFact struct {
	env     *Environment
	factptr unsafe.Pointer
}

func createTemplateFact(env *Environment, factptr unsafe.Pointer) *TemplateFact {
	ret := &TemplateFact{
		env:     env,
		factptr: factptr,
	}
	C.EnvIncrementFactCount(env.env, factptr)
	runtime.SetFinalizer(ret, func(*TemplateFact) {
		ret.Drop()
	})
	return ret
}

// Drop drops the reference to the fact in CLIPS. should be called when done with the fact
func (f *TemplateFact) Drop() {
	if f.factptr != nil {
		C.EnvDecrementFactCount(f.env.env, f.factptr)
		f.factptr = nil
	}
}

// Index returns the index number of this fact within CLIPS
func (f *TemplateFact) Index() int {
	return int(C.EnvFactIndex(f.env.env, f.factptr))
}

// Asserted returns true if the fact has been asserted.
func (f *TemplateFact) Asserted() bool {
	if f.Index() == 0 {
		return false
	}
	if C.EnvFactExistp(f.env.env, f.factptr) != 1 {
		return false
	}
	return true
}

// Assert asserts the fact
func (f *TemplateFact) Assert() error {
	if f.Asserted() {
		return fmt.Errorf("Fact already asserted")
	}

	ret := C.EnvAssignFactSlotDefaults(f.env.env, f.factptr)
	if ret != 1 {
		return EnvError(f.env, "Unable to set defaults for fact")
	}

	factptr := C.EnvAssert(f.env.env, f.factptr)
	if factptr == nil {
		return EnvError(f.env, "Unable to assert fact")
	}
	return nil
}

// Retract retracts the fact from CLIPS
func (f *TemplateFact) Retract() error {
	ret := C.EnvRetract(f.env.env, f.factptr)
	if ret != 1 {
		return EnvError(f.env, "Unable to retract fact")
	}
	return nil
}

// Template returns the template defining this fact
func (f *TemplateFact) Template() *Template {
	tplptr := C.EnvFactDeftemplate(f.env.env, f.factptr)
	return createTemplate(f.env, tplptr)
}

// String returns a string representation of the fact
func (f *TemplateFact) String() string {
	ret := factPPString(f.env, f.factptr)
	split := strings.SplitN(ret, "     ", 2)
	return strings.TrimRight(split[len(split)-1], "\n")
}

// Equal returns true if this fact equal the given fact
func (f *TemplateFact) Equal(otherfact Fact) bool {
	other, ok := otherfact.(*TemplateFact)
	if !ok {
		return false
	}
	return f.factptr == other.factptr
}

// Slots returns a function that can be called to get the next slot for this fact. Will return nil when no more slots remain
func (f *TemplateFact) Slots() (map[string]interface{}, error) {
	data := createDataObject(f.env)
	defer data.Delete()

	tplptr := C.EnvFactDeftemplate(f.env.env, f.factptr)
	C.EnvDeftemplateSlotNames(f.env.env, tplptr, data.byRef())
	namesblob := data.Value()
	names, ok := namesblob.([]interface{})
	if !ok {
		panic("Unexpected data returned from CLIPS for slot names")
	}

	ret := make(map[string]interface{}, len(names))
	var err error
	for _, name := range names {
		namestr, ok := name.(Symbol)
		if !ok {
			panic("Unexpected data returned from CLIPS for slot names")
		}
		data, err = slotValue(f.env, f.factptr, namestr)
		if err != nil {
			return nil, err
		}
		defer data.Delete()
		ret[string(namestr)] = data.Value()
	}
	return ret, nil
}

// Slot returns the value stored in the given slot
func (f *TemplateFact) Slot(name string) (interface{}, error) {
	data, err := slotValue(f.env, f.factptr, Symbol(name))
	if err != nil {
		return nil, err
	}
	defer data.Delete()
	return data.Value(), nil
}

// ExtractSlot unmarshals the given slot value into the object provided by the user
func (f *TemplateFact) ExtractSlot(retval interface{}, name string) error {
	data, err := slotValue(f.env, f.factptr, Symbol(name))
	if err != nil {
		return err
	}
	defer data.Delete()
	return data.ExtractValue(retval, false)
}

// Set alters the item at a specific in the multifield
func (f *TemplateFact) Set(slot string, value interface{}) error {
	if f.Asserted() {
		return fmt.Errorf("Unable to change asserted fact")
	}
	data := createDataObject(f.env)
	defer data.Delete()
	cslot := C.CString(slot)
	defer C.free(unsafe.Pointer(cslot))

	data.SetValue(value)

	ret := C.EnvPutFactSlot(f.env.env, f.factptr, cslot, data.byRef())
	if ret != 1 {
		slots := f.Template().Slots()
		_, ok := slots[slot]
		if !ok {
			return fmt.Errorf(`Fact %d does not have slot "%s"`, f.Index(), slot)
		}
		return EnvError(f.env, "Unable to set slot value")
	}
	return nil
}

// Extract unmarshals this fact into the user provided object
func (f *TemplateFact) Extract(retval interface{}) error {
	slots, err := f.Slots()
	if err != nil {
		return err
	}
	knownInstances := make(map[InstanceName]interface{})
	return f.env.structuredExtract(retval, slots, false, knownInstances)
}
