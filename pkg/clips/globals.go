package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
import "C"
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
		return nil, fmt.Errorf(`Global "%s" not found`, name)
	}
	return createGlobal(env, glbptr), nil
}

func createGlobal(env *Environment, glbptr unsafe.Pointer) *Global {
	return &Global{
		env:    env,
		glbptr: glbptr,
	}
}

// Delete frees any reference to CLIPS data
func (g *Global) Delete() {
	// nothing to do really, just being consistent with API
	g.glbptr = nil
}

// Equals returns true if the other object represents the same global in CLIPS
func (g *Global) Equals(other *Global) bool {
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

/*
   @property
   def module(self):
       """The module in which the Global is defined.

       Python equivalent of the CLIPS defglobal-module command.

       """
       modname = ffi.string(lib.EnvDefglobalModule(self._env, self._glb))
       defmodule = lib.EnvFindDefmodule(self._env, modname)

       return Module(self._env, defmodule)

*/

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
