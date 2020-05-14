package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips/clips.h>
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

// Module represents a CLIPS module
type Module struct {
	env    *Environment
	modptr unsafe.Pointer
}

// CurrentModule returns the current module of the env
func (env *Environment) CurrentModule() *Module {
	modptr := C.EnvGetCurrentModule(env.env)
	return createModule(env, modptr)
}

// SetModule sets the current module for the CLIPS env
func (env *Environment) SetModule(module *Module) {
	C.EnvSetCurrentModule(env.env, module.modptr)
}

// Modules returns the list of modulesb
func (env *Environment) Modules() []*Module {
	modptr := C.EnvGetNextDefmodule(env.env, nil)

	ret := make([]*Module, 0, 10)
	for modptr != nil {
		ret = append(ret, createModule(env, modptr))
		modptr = C.EnvGetNextDefmodule(env.env, modptr)
	}
	return ret
}

// FindModule returns the module with the given name
func (env *Environment) FindModule(name string) (*Module, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	modptr := C.EnvFindDefmodule(env.env, cname)
	if modptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Module "%s" not found`, name))
	}
	return createModule(env, modptr), nil
}

func createModule(env *Environment, modptr unsafe.Pointer) *Module {
	return &Module{
		env:    env,
		modptr: modptr,
	}
}

// Equal returns true if the other module references the same CLIPS module
func (m *Module) Equal(other *Module) bool {
	return m.modptr == other.modptr
}

func (m *Module) String() string {
	module := C.EnvGetDefmodulePPForm(m.env.env, m.modptr)
	return strings.TrimRight(C.GoString(module), "\n")
}

// Name returns the name of this module
func (m *Module) Name() string {
	name := C.EnvGetDefmoduleName(m.env.env, m.modptr)
	return C.GoString(name)
}
