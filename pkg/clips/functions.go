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

// Function references a CLIPS function
type Function struct {
	env  *Environment
	fptr unsafe.Pointer
}

// Functions returns the set of all functions in CLIPS
func (env *Environment) Functions() []*Function {
	fptr := C.EnvGetNextDeffunction(env.env, nil)
	ret := make([]*Function, 0, 10)
	for fptr != nil {
		ret = append(ret, createFunction(env, fptr))
		fptr = C.EnvGetNextDeffunction(env.env, fptr)
	}
	return ret
}

// FindFunction returns the function of the given name
func (env *Environment) FindFunction(name string) (*Function, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	fptr := C.EnvFindDeffunction(env.env, cname)
	if fptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Function "%s" not found`, name))
	}
	return createFunction(env, fptr), nil
}

func createFunction(env *Environment, fptr unsafe.Pointer) *Function {
	return &Function{
		env:  env,
		fptr: fptr,
	}
}

// Equal returns true if the other function represents the same CLIPS function as this one
func (f *Function) Equal(other *Function) bool {
	return f.fptr == other.fptr
}

func (f *Function) String() string {
	cstr := C.EnvGetDeffunctionPPForm(f.env.env, f.fptr)
	return strings.TrimRight(C.GoString(cstr), "\n")
}

// Name returns the name of this function
func (f *Function) Name() string {
	cstr := C.EnvGetDeffunctionName(f.env.env, f.fptr)
	return C.GoString(cstr)
}

// Call calls the CLIPS function with the given arguments (must be a space-delimited string)
func (f *Function) Call(arguments string) (interface{}, error) {
	cname := C.EnvGetDeffunctionName(f.env.env, f.fptr)
	data := createDataObject(f.env)
	defer data.Delete()
	var cargs *C.char
	if arguments != "" {
		cargs = C.CString(arguments)
		defer C.free(unsafe.Pointer(cargs))
	}

	ret := C.EnvFunctionCall(f.env.env, cname, cargs, data.byRef())
	if ret == 1 {
		return nil, EnvError(f.env, `Unable to call function "%s"`, f.Name())
	}
	return data.Value(), nil
}

// Module returns the module in which this function is defined
func (f *Function) Module() *Module {
	cmodname := C.EnvDeffunctionModule(f.env.env, f.fptr)
	modptr := C.EnvFindDefmodule(f.env.env, cmodname)

	return createModule(f.env, modptr)
}

// Deletable returns true if function is unreferenced and deletable
func (f *Function) Deletable() bool {
	ret := C.EnvIsDeffunctionDeletable(f.env.env, f.fptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watched returns true if function is being watched
func (f *Function) Watched() bool {
	ret := C.EnvGetDeffunctionWatch(f.env.env, f.fptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watch sets whether the function is being watched
func (f *Function) Watch(val bool) {
	var flag C.uint
	if val {
		flag = 1
	}
	C.EnvSetDeffunctionWatch(f.env.env, flag, f.fptr)
}

// Undefine undefines the function within CLIPS
func (f *Function) Undefine() error {
	ret := C.EnvUndeffunction(f.env.env, f.fptr)
	if ret != 1 {
		return EnvError(f.env, `Unable to undef function "%s"`, f.Name())
	}
	f.fptr = nil
	return nil
}
