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

// ImpliedFact is an ordered fact having an implied definition
type ImpliedFact struct {
	env        *Environment
	factptr    unsafe.Pointer
	multifield []interface{}
}

// ImpliedFactSlot is a hook to the value of a particular slot
type ImpliedFactSlot struct {
	fact  *ImpliedFact
	index int
}

func createImpliedFact(env *Environment, factptr unsafe.Pointer) *ImpliedFact {
	ret := &ImpliedFact{
		env:     env,
		factptr: factptr,
	}
	C.EnvIncrementFactCount(env.env, factptr)
	runtime.SetFinalizer(ret, func(*ImpliedFact) {
		ret.Drop()
	})
	return ret
}

// Drop drops the reference to the fact in CLIPS. should be called when done with the fact
func (f *ImpliedFact) Drop() {
	if f.factptr != nil {
		C.EnvDecrementFactCount(f.env.env, f.factptr)
		f.factptr = nil
	}
}

// Index returns the index number of this fact within CLIPS
func (f *ImpliedFact) Index() int {
	return int(C.EnvFactIndex(f.env.env, f.factptr))
}

// Asserted returns true if the fact has been asserted.
func (f *ImpliedFact) Asserted() bool {
	if f.Index() == 0 {
		return false
	}
	if C.EnvFactExistp(f.env.env, f.factptr) != 1 {
		return false
	}
	return true
}

// Assert asserts the fact
func (f *ImpliedFact) Assert() error {
	if f.Asserted() {
		return fmt.Errorf("Fact already asserted")
	}
	data := createDataObject(f.env)
	defer data.Delete()
	if f.multifield == nil {
		f.multifield = make([]interface{}, 0)
	}
	data.SetValue(f.multifield)
	ret := C.EnvPutFactSlot(f.env.env, f.factptr, nil, data.byRef())
	if ret != 1 {
		return EnvError(f.env, "Unable to set slot for fact")
	}
	ret = C.EnvAssignFactSlotDefaults(f.env.env, f.factptr)
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
func (f *ImpliedFact) Retract() error {
	ret := C.EnvRetract(f.env.env, f.factptr)
	if ret != 1 {
		return EnvError(f.env, "Unable to retract fact")
	}
	return nil
}

// Template returns the template defining this fact
func (f *ImpliedFact) Template() *Template {
	tplptr := C.EnvFactDeftemplate(f.env.env, f.factptr)
	return createTemplate(f.env, tplptr)
}

// String returns a string representation of the fact
func (f *ImpliedFact) String() string {
	ret := factPPString(f.env, f.factptr)
	split := strings.SplitN(ret, "     ", 2)
	return strings.TrimRight(split[len(split)-1], "\n")
}

// Equal returns true if this fact equal the given fact
func (f *ImpliedFact) Equal(otherfact Fact) bool {
	other, ok := otherfact.(*ImpliedFact)
	if !ok {
		return false
	}
	return f.factptr == other.factptr
}

// Slots returns a function that can be called to get the next slot for this fact. Will return nil when no more slots remain
func (f *ImpliedFact) Slots() (map[string]interface{}, error) {
	data, err := slotValue(f.env, f.factptr, "")
	if err != nil {
		return nil, err
	}
	defer data.Delete()
	ret := make(map[string]interface{}, 1)
	ret[""] = data.Value()
	return ret, nil
}

// Slot returns the value of the given slot. For Implied Facts, the only valid slot name is ""
func (f *ImpliedFact) Slot(slotname string) (interface{}, error) {
	if slotname != "" {
		return nil, fmt.Errorf(`Invalid slot name "%s"`, slotname)
	}
	data, err := slotValue(f.env, f.factptr, "")
	if err != nil {
		return nil, err
	}
	defer data.Delete()
	return data.Value(), nil
}

// ExtractSlot unmarshals the value of the given slot into the user provided object. For Implied Facts, the only valid slot name is ""
func (f *ImpliedFact) ExtractSlot(retval interface{}, slotname string) error {
	if slotname != "" {
		return fmt.Errorf(`Invalid slot name "%s"`, slotname)
	}
	data, err := slotValue(f.env, f.factptr, "")
	if err != nil {
		return err
	}
	defer data.Delete()
	return data.ExtractValue(retval, false)
}

// Set alters the item at a specific in the multifield
func (f *ImpliedFact) Set(index int, value interface{}) error {
	if f.Asserted() {
		return fmt.Errorf("Unable to change asserted fact")
	}
	if f.multifield == nil || index >= len(f.multifield) {
		return fmt.Errorf("Invalid multifield index %d", index)
	}
	f.multifield[index] = value
	return nil
}

// Append an element to the fact
func (f *ImpliedFact) Append(value interface{}) error {
	if f.Asserted() {
		return fmt.Errorf("Unable to change asserted fact")
	}
	if f.multifield == nil {
		f.multifield = make([]interface{}, 0, 10)
	}
	f.multifield = append(f.multifield, value)
	return nil
}

// Extend Appends the contents of a slice to the fact
func (f *ImpliedFact) Extend(values []interface{}) error {
	if f.Asserted() {
		return fmt.Errorf("Unable to change asserted fact")
	}
	if values == nil {
		return nil
	}
	if f.multifield == nil {
		f.multifield = make([]interface{}, 0, len(values))
	}
	for _, val := range values {
		f.multifield = append(f.multifield, val)
	}
	return nil
}

// Extract unmarshals this fact into the user provided object
func (f *ImpliedFact) Extract(retval interface{}) error {
	return f.ExtractSlot(retval, "")
}
