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
	"strings"
	"unsafe"
)

// Generic represents a CLIPS genneric
type Generic struct {
	env    *Environment
	genptr unsafe.Pointer
}

// Method represents one method of a CLIPS generic
type Method struct {
	gen   *Generic
	index C.long
}

// Generics returns a list of all generics in CLIPS
func (env *Environment) Generics() []*Generic {
	genptr := C.EnvGetNextDefgeneric(env.env, nil)

	ret := make([]*Generic, 0, 10)
	for genptr != nil {
		ret = append(ret, createGeneric(env, genptr))
		genptr = C.EnvGetNextDefgeneric(env.env, genptr)
	}
	return ret
}

// FindGeneric returns the generic identified by name
func (env *Environment) FindGeneric(name string) (*Generic, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	genptr := C.EnvFindDefgeneric(env.env, cname)
	if genptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Generic "%s" not found`, name))
	}
	return createGeneric(env, genptr), nil
}

func createGeneric(env *Environment, genptr unsafe.Pointer) *Generic {
	return &Generic{
		env:    env,
		genptr: genptr,
	}
}

// Equal returns true if the other generic represents the same CLIPS generic
func (g *Generic) Equal(other *Generic) bool {
	return g.genptr == other.genptr
}

func (g *Generic) String() string {
	cstr := C.EnvGetDefgenericPPForm(g.env.env, g.genptr)
	return strings.TrimRight(C.GoString(cstr), "\n")
}

// Name returns the name of this generic
func (g *Generic) Name() string {
	cstr := C.EnvGetDefgenericName(g.env.env, g.genptr)
	return C.GoString(cstr)
}

// Call calls the CLIPS generic function. Arguments must be passed as a string
func (g *Generic) Call(arguments string) (interface{}, error) {
	cname := C.EnvGetDefgenericName(g.env.env, g.genptr)
	data := createDataObject(g.env)
	defer data.Delete()

	var cargs *C.char
	if arguments != "" {
		cargs = C.CString(arguments)
		defer C.free(unsafe.Pointer(cargs))
	}

	ret := C.EnvFunctionCall(g.env.env, cname, cargs, data.byRef())
	// the sense of this return is backwards from the usual convention
	if ret == 1 {
		return nil, EnvError(g.env, `Unable to call generic function "%s"`, g.Name())
	}
	return data.Value(), nil
}

// Module returns a reference to the module of this generic
func (g *Generic) Module() *Module {
	cmodname := C.EnvDefgenericModule(g.env.env, g.genptr)
	modptr := C.EnvFindDefmodule(g.env.env, cmodname)
	return createModule(g.env, modptr)
}

// Deletable returns true if the generic is unreferenced and can be deleted
func (g *Generic) Deletable() bool {
	ret := C.EnvIsDefgenericDeletable(g.env.env, g.genptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watched returns true if the generic is watched
func (g *Generic) Watched() bool {
	ret := C.EnvGetDefgenericWatch(g.env.env, g.genptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watch sets whether this generic is watched
func (g *Generic) Watch(val bool) {
	var flag C.uint
	if val {
		flag = C.uint(1)
	}
	C.EnvSetDefgenericWatch(g.env.env, flag, g.genptr)
}

// Methods returns a list of all methods for this generic
func (g *Generic) Methods() []*Method {
	index := C.EnvGetNextDefmethod(g.env.env, g.genptr, 0)
	ret := make([]*Method, 0, 10)
	for index != 0 {
		ret = append(ret, createMethod(g, index))
		index = C.EnvGetNextDefmethod(g.env.env, g.genptr, index)
	}
	return ret
}

// Undefine undefines the Generic
func (g *Generic) Undefine() error {
	ret := C.EnvUndefgeneric(g.env.env, g.genptr)
	if ret != 1 {
		return EnvError(g.env, `Unable to undefine generic "%s"`, g.Name())
	}
	g.genptr = nil
	return nil
}

func createMethod(gen *Generic, index C.long) *Method {
	return &Method{
		gen:   gen,
		index: index,
	}
}

// Equal returns true of other represents the same CLIPS method as this
func (m *Method) Equal(other *Method) bool {
	return m.gen.genptr == other.gen.genptr && m.index == other.index
}

func (m *Method) String() string {
	cstr := C.EnvGetDefmethodPPForm(m.gen.env.env, m.gen.genptr, m.index)
	return strings.TrimRight(C.GoString(cstr), "\n")
}

// Watched returns true if watch is enabled on this method
func (m *Method) Watched() bool {
	ret := C.EnvGetDefmethodWatch(m.gen.env.env, m.gen.genptr, m.index)
	if ret == 1 {
		return true
	}
	return false
}

// Watch sets whether this method is watched
func (m *Method) Watch(val bool) {
	var flag C.uint
	if val {
		flag = C.uint(1)
	}
	C.EnvSetDefmethodWatch(m.gen.env.env, flag, m.gen.genptr, m.index)
}

// Deletable returns true if this method is unreferenced and deletable
func (m *Method) Deletable() bool {
	ret := C.EnvIsDefmethodDeletable(m.gen.env.env, m.gen.genptr, m.index)
	if ret == 1 {
		return true
	}
	return false
}

// Restrictions returns the method restrictions for this method
func (m *Method) Restrictions() interface{} {
	data := createDataObject(m.gen.env)
	defer data.Delete()
	C.EnvGetMethodRestrictions(m.gen.env.env, m.gen.genptr, m.index, data.byRef())
	return data.Value()
}

// Description returns the description of this method
func (m *Method) Description() string {
	// TODO grow buf if we fill the 1k buffer, and try again
	var bufsize C.ulong = 1024
	buf := (*C.char)(C.malloc(C.sizeof_char * bufsize))
	defer C.free(unsafe.Pointer(buf))
	C.EnvGetDefmethodDescription(m.gen.env.env, buf, bufsize-1, m.gen.genptr, m.index)

	return C.GoString(buf)
}

// Undefine undefines the method
func (m *Method) Undefine() error {
	ret := C.EnvUndefmethod(m.gen.env.env, m.gen.genptr, m.index)
	if ret != 1 {
		return EnvError(m.gen.env, "Unable to undefine method")
	}
	return nil
}
