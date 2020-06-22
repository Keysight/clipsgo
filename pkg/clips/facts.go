package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
//
// int implied_deftemplate(void *template)
// {
//   return ((struct deftemplate*)template)->implied;
// }
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
	"unsafe"
)

// Fact represents a fact within CLIPS
type Fact interface {
	// Index returns the index number of this fact within CLIPS
	Index() int

	// Asserted returns true if the fact has been asserted.
	Asserted() bool

	// Assert asserts the fact
	Assert() error

	// Retract retracts the fact from CLIPS
	Retract() error

	// Template returns the template defining this fact
	Template() *Template

	// String returns a string representation of the fact
	String() string

	// Drop drops the reference to the fact in CLIPS. should be called when done with the fact
	Drop()

	// Equal returns true if this fact equal the given fact
	Equal(Fact) bool

	// Slots returns a *copy* of slot values for each slot in this fact
	Slots() (map[string]interface{}, error)

	// Slot returns the value of a given slot. For Implied Facts, "" is the only valid slot name
	Slot(slotname string) (interface{}, error)

	// ExtractSlot unmarshals the given slot into the user provided object
	ExtractSlot(retval interface{}, slotname string) error

	// Extract unmarshals the full fact into the user provided object
	Extract(retval interface{}) error
}

// Facts returns a slice of all facts known to CLIPS
func (env *Environment) Facts() []Fact {
	ret := make([]Fact, 0, 10)
	factptr := C.EnvGetNextFact(env.env, nil)
	for factptr != nil {
		ret = append(ret, env.newFact(factptr))
		factptr = C.EnvGetNextFact(env.env, factptr)
	}
	return ret
}

// AssertString asserts a fact as a string.
func (env *Environment) AssertString(factstr string) (Fact, error) {
	cfactstr := C.CString(factstr)
	defer C.free(unsafe.Pointer(cfactstr))
	factptr := C.EnvAssertString(env.env, cfactstr)
	if factptr == nil {
		return nil, EnvError(env, `Error asserting fact "%s"`, factstr)
	}
	return env.newFact(factptr), nil
}

// LoadFacts loads facts from the given file
func (env *Environment) LoadFacts(filename string) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	retcode := C.EnvLoadFacts(env.env, cfilename)
	if retcode == -1 {
		return EnvError(env, `Error loading facts from "%s"`, filename)
	}
	return nil
}

// LoadFactsFromString loads facts from the given string
func (env *Environment) LoadFactsFromString(factstr string) error {
	cfactstr := C.CString(factstr)
	defer C.free(unsafe.Pointer(cfactstr))

	retcode := C.EnvLoadFactsFromString(env.env, cfactstr, -1)
	if retcode == -1 {
		return EnvError(env, `Error loading facts from string`)
	}
	return nil
}

// SaveFacts saves facts to the given file
func (env *Environment) SaveFacts(filename string, savemode SaveMode) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	retcode := C.EnvSaveFacts(env.env, cfilename, savemode.CVal())
	if retcode == -1 {
		return EnvError(env, `Error saving facts to "%s"`, filename)
	}
	return nil
}

// Templates returns a slice of all defined templates
func (env *Environment) Templates() []*Template {
	ret := make([]*Template, 0, 10)
	for tplptr := C.EnvGetNextDeftemplate(env.env, nil); tplptr != nil; tplptr = C.EnvGetNextDeftemplate(env.env, tplptr) {
		ret = append(ret, createTemplate(env, tplptr))
	}
	return ret
}

// FindTemplate returns an object representing the given template name
func (env *Environment) FindTemplate(name string) (*Template, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	tplptr := C.EnvFindDeftemplate(env.env, cname)
	if tplptr == nil {
		return nil, fmt.Errorf(`Template "%s" not found`, name)
	}
	return createTemplate(env, tplptr), nil
}

func (env *Environment) newFact(fact unsafe.Pointer) Fact {
	templ := C.EnvFactDeftemplate(env.env, fact)
	if C.implied_deftemplate(templ) == 1 {
		return createImpliedFact(env, fact)
	}
	return createTemplateFact(env, fact)
}

func factPPString(env *Environment, factptr unsafe.Pointer) string {
	// TODO grow buf if we fill the 1k buffer, and try again
	var bufsize C.ulong = 1024
	buf := (*C.char)(C.malloc(C.sizeof_char * bufsize))
	defer C.free(unsafe.Pointer(buf))
	C.EnvGetFactPPForm(env.env, buf, bufsize-1, factptr)
	return C.GoString(buf)
}

func slotValue(env *Environment, factptr unsafe.Pointer, slot Symbol) (*DataObject, error) {
	implied := C.implied_deftemplate(C.EnvFactDeftemplate(env.env, factptr))

	if implied == 1 && slot != "" {
		return nil, fmt.Errorf("Invalid call to slotValue")
	}

	var cslot *C.char
	if slot != Symbol("") {
		cslot = C.CString(string(slot))
		defer C.free(unsafe.Pointer(cslot))
	}
	data := createDataObject(env)
	ret := C.EnvGetFactSlot(env.env, factptr, cslot, data.byRef())
	if ret != 1 {
		data.Delete()
		return nil, EnvError(env, "Unable to get slot value")
	}
	return data, nil
}
