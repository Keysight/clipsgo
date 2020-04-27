package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// CLIPSError error returned from CLIPS
type CLIPSError struct {
	Err error
}

func (e *CLIPSError) Error() string {
	return e.Err.Error()
}

// Environment stores a CLIPS environment
type Environment struct {
	env unsafe.Pointer
}

// CreateEnvironment creates a new instance of a CLIPS environment
func CreateEnvironment() *Environment {
	ret := &Environment{
		env: C.CreateEnvironment(),
	}
	runtime.SetFinalizer(ret, func(env *Environment) {
		env.Close()
	})
	return ret
}

// Close destroys the CLIPS environment
func (env *Environment) Close() {
	if env.env != nil {
		C.DestroyEnvironment(env.env)
		env.env = nil
	}
}

// Load loads a set of constructs into the CLIPS data base. Constructs can be in text or binary format. Equivalent to CLIPS (load)
func (env *Environment) Load(path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	errint := int(C.EnvBload(env.env, cpath))
	if errint != 1 {
		errint = int(C.EnvLoad(env.env, cpath))
	}
	if errint != 1 {
		return &CLIPSError{
			Err: fmt.Errorf("Unable to load file %s", path),
		}
	}
	return nil
}

// Save saves the current state of the environment
func (env *Environment) Save(path string, binary bool) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var errint int
	if binary {
		errint = int(C.EnvBsave(env.env, cpath))
	} else {
		errint = int(C.EnvSave(env.env, cpath))
	}
	if errint != 1 {
		return &CLIPSError{
			Err: fmt.Errorf("Unable to save to file %s", path),
		}
	}
	return nil
}

// BatchStar executes the CLIPS code found in path. Equivalent to CLIPS (batch*)
func (env *Environment) BatchStar(path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	if C.EnvBatchStar(env.env, cpath) != 1 {
		return &CLIPSError{
			Err: fmt.Errorf("Unable to open file %s", path),
		}
	}
	return nil
}

// Build builds a single construct within the CLIPS environment
func (env *Environment) Build(construct string) error {
	cconstruct := C.CString(construct)
	defer C.free(unsafe.Pointer(cconstruct))
	if C.EnvBuild(env.env, cconstruct) != 1 {
		return &CLIPSError{
			Err: fmt.Errorf("Unable to parse construct %s", construct),
		}
	}
	return nil
}

// Eval evaluates an expression returning its value
func (env *Environment) Eval(construct string) (interface{}, error) {
	cconstruct := C.CString(construct)
	defer C.free(unsafe.Pointer(cconstruct))

	data := createDataObject(env)
	errint := int(C.EnvEval(env.env, cconstruct, data.byRef()))

	if errint != 1 {
		return nil, &CLIPSError{
			Err: fmt.Errorf("Unable to parse construct %s", construct),
		}
	}
	return data.Value(), nil
}

// Reset resets the CLIPS environment
func (env *Environment) Reset() {
	C.EnvReset(env.env)
}

// Clear clears the CLIPS environment
func (env *Environment) Clear() {
	C.EnvClear(env.env)
}

// DefineFunction defines a Go function within the CLIPS environment
/*
func (env *Environment) DefineFunction(func) {
}
*/
