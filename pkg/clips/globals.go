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

// Global represents a global variable within CLIPS
type Global struct {
	env    *Environment
	glbptr unsafe.Pointer
}

// GlobalsChanged returns true if any global has changed since last call
func (env *Environment) GlobalsChanged() bool {
	ret := C.EnvGetGlobalsChanged(env.env)
	C.EnvSetGlobalsChanged(env.env, 0)
	if ret == 1 {
		return true
	}
	return false
}

// Globals returns a slice containing references to all globals
func (env *Environment) Globals() []*Global {
	glbptr := C.EnvGetNextDefglobal(env.env, nil)

	ret := make([]*Global, 0, 10)
	for glbptr != nil {
		ret = append(ret, createGlobal(env, glbptr))
		glbptr = C.EnvGetNextDefglobal(env.env, glbptr)
	}
	return ret
}

// FindGlobal finds the global by name
func (env *Environment) FindGlobal(name string) (*Global, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	glbptr := C.EnvFindDefglobal(env.env, cname)
	if glbptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Global "%s" not found`, name))
	}
	return createGlobal(env, glbptr), nil
}

func createGlobal(env *Environment, glbptr unsafe.Pointer) *Global {
	return &Global{
		env:    env,
		glbptr: glbptr,
	}
}

// Equal returns true if the other object represents the same global in CLIPS
func (g *Global) Equal(other *Global) bool {
	return g.glbptr == other.glbptr
}

func (g *Global) String() string {
	ret := ""
	cstr := C.EnvGetDefglobalPPForm(g.env.env, g.glbptr)
	if cstr != nil {
		ret = C.GoString(cstr)
	}
	return strings.TrimRight(ret, "\n")
}

// Name returns the name of this global
func (g *Global) Name() string {
	cstr := C.EnvGetDefglobalName(g.env.env, g.glbptr)
	return C.GoString(cstr)
}

// Value returns the value of this global
func (g *Global) Value() (interface{}, error) {
	data := createDataObject(g.env)
	defer data.Delete()
	name := g.Name()
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvGetDefglobalValue(g.env.env, cname, data.byRef())
	if ret != 1 {
		return nil, EnvError(g.env, `Unable to get value for global "%s"`, name)
	}
	return data.Value(), nil
}

// SetValue sets the value of this global
func (g *Global) SetValue(value interface{}) error {
	name := g.Name()
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	data := createDataObject(g.env)
	defer data.Delete()
	data.SetValue(value)

	ret := C.EnvSetDefglobalValue(g.env.env, cname, data.byRef())
	if ret != 1 {
		return EnvError(g.env, `Unable to set value for global "%s"`, name)
	}
	return nil
}

// Module returns a referece to the module of this global
func (g *Global) Module() *Module {
	modname := C.EnvDefglobalModule(g.env.env, g.glbptr)
	modptr := C.EnvFindDefmodule(g.env.env, modname)
	return createModule(g.env, modptr)
}

// Deletable returns true if the global can be deleted
func (g *Global) Deletable() bool {
	ret := C.EnvIsDefglobalDeletable(g.env.env, g.glbptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watched returns true if the global can be deleted
func (g *Global) Watched() bool {
	ret := C.EnvGetDefglobalWatch(g.env.env, g.glbptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watch sets whether the global is watched
func (g *Global) Watch(val bool) {
	var flag C.uint
	if val {
		flag = C.uint(1)
	}
	C.EnvSetDefglobalWatch(g.env.env, flag, g.glbptr)
}

// Undefine undefines the global
func (g *Global) Undefine() error {
	ret := C.EnvUndefglobal(g.env.env, g.glbptr)
	if ret != 1 {
		return EnvError(g.env, `Unable to undefine global "%s"`, g.Name())
	}
	return nil
}
