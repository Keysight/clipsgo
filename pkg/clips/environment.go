package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
//
// void goFunction(void *env, DATA_OBJECT *data);
//
// static inline void callGoFunction(void * env, DATA_OBJECT *data) {
//	 goFunction(env,data);
// }
//
// int define_function(void *environment)
// {
//     return EnvDefineFunction(
//         environment, "go-function", 'u',
//         PTIEF callGoFunction, "callGoFunction");
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
	"reflect"
	"runtime"
	"strings"
	"unsafe"
)

const defFunction = `
(deffunction %[1]s (%[2]s)
  (go-function %[1]s %[3]s))
`

// Environment stores a CLIPS environment
type Environment struct {
	env      unsafe.Pointer
	callback map[string]reflect.Value
	router   map[string]Router
	errRtr   *ErrorRouter
}

var environmentObj = make(map[unsafe.Pointer]*Environment)

// CreateEnvironment creates a new instance of a CLIPS environment
func CreateEnvironment() *Environment {
	ret := &Environment{
		env:      C.CreateEnvironment(),
		callback: make(map[string]reflect.Value),
		router:   make(map[string]Router),
	}
	ret.errRtr = CreateErrorRouter(ret)
	runtime.SetFinalizer(ret, func(env *Environment) {
		env.Delete()
	})
	C.define_function(ret.env)
	environmentObj[ret.env] = ret

	return ret
}

// Delete destroys the CLIPS environment
func (env *Environment) Delete() {
	if env.env != nil {
		delete(environmentObj, env.env)
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
		return EnvError(env, "Unable to load file \"%s\"", path)
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
		return EnvError(env, "Unable to save to file \"%s\"", path)
	}
	return nil
}

// BatchStar executes the CLIPS code found in path. Equivalent to CLIPS (batch*)
func (env *Environment) BatchStar(path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	if C.EnvBatchStar(env.env, cpath) != 1 {
		return EnvError(env, "Unable to open file \"%s\"", path)
	}
	return nil
}

// Build builds a single construct within the CLIPS environment
func (env *Environment) Build(construct string) error {
	cconstruct := C.CString(construct)
	defer C.free(unsafe.Pointer(cconstruct))
	if C.EnvBuild(env.env, cconstruct) != 1 {
		return EnvError(env, "Unable to parse construct \"%s\"", construct)
	}
	return nil
}

// Eval evaluates an expression returning its value
func (env *Environment) Eval(construct string) (interface{}, error) {
	cconstruct := C.CString(construct)
	defer C.free(unsafe.Pointer(cconstruct))

	data := createDataObject(env)
	defer data.Delete()
	errint := int(C.EnvEval(env.env, cconstruct, data.byRef()))

	if errint != 1 {
		return nil, EnvError(env, "Unable to parse construct \"%s\"", construct)
	}
	return data.Value(), nil
}

// ExtractEval evaluates an expression, storing its return value into the object passed by the user
func (env *Environment) ExtractEval(retval interface{}, construct string) error {
	cconstruct := C.CString(construct)
	defer C.free(unsafe.Pointer(cconstruct))

	data := createDataObject(env)
	defer data.Delete()
	errint := int(C.EnvEval(env.env, cconstruct, data.byRef()))

	if errint != 1 {
		return EnvError(env, "Unable to parse construct \"%s\"", construct)
	}
	return data.ExtractValue(retval, false)
}

// Reset resets the CLIPS environment
func (env *Environment) Reset() {
	C.EnvReset(env.env)
}

// Clear clears the CLIPS environment
func (env *Environment) Clear() {
	C.EnvClear(env.env)
}

// DefineFunction defines a Go function within the CLIPS environment. If the given name is "", the name of the go funciton will be used
func (env *Environment) DefineFunction(name string, callback interface{}) error {
	val := reflect.ValueOf(callback)
	if val.Kind() != reflect.Func {
		return fmt.Errorf(`Invalid function pointer %v"`, callback)
	}
	if name == "" {
		name = runtime.FuncForPC(val.Pointer()).Name()
	}
	typ := val.Type()
	fixedArgs := typ.NumIn()
	argslist := make([]string, fixedArgs)
	for i := 0; i < fixedArgs; i++ {
		argslist[i] = fmt.Sprintf("?arg%d", i)
	}
	var usage string
	var declaration string
	if !typ.IsVariadic() {
		usage = strings.Join(argslist, " ")
		declaration = strings.Join(argslist, " ")
	} else {
		fixedArgs--
		argslist[fixedArgs] = fmt.Sprintf("$?arg%d", fixedArgs)
		declaration = strings.Join(argslist, " ")
		argslist[fixedArgs] = fmt.Sprintf("(expand$ ?arg%d)", fixedArgs)
		usage = strings.Join(argslist, " ")
	}
	env.callback[name] = val
	return env.Build(fmt.Sprintf(defFunction, name, declaration, usage))
}

// CompleteCommand checks the string to see if it is a complete command yet
func (env *Environment) CompleteCommand(cmd string) (bool, error) {
	ccmd := C.CString(cmd + "\n")
	defer C.free(unsafe.Pointer(ccmd))

	ret := int(C.CompleteCommand(ccmd))
	if ret == 1 {
		return true, nil
	}
	if ret == -1 {
		return false, fmt.Errorf(`Invalid command: "%s"`, cmd)
	}
	return false, nil
}

// SendCommand evaluates a command as if it were typed in the CLIPS shell
func (env *Environment) SendCommand(cmd string) error {
	ccmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(ccmd))

	// Commands cribbed from the CLIPS shell, and inspired by PyCLIPS
	C.FlushPPBuffer(env.env)
	C.SetPPBufferStatus(env.env, 0)
	ret := C.RouteCommand(env.env, ccmd, 1)
	res := C.GetEvaluationError(env.env)
	C.FlushPPBuffer(env.env)
	C.SetHaltExecution(env.env, 0)
	C.SetEvaluationError(env.env, 0)
	C.CleanCurrentGarbageFrame(env.env, nil)
	C.CallPeriodicTasks(env.env)
	if ret == 0 || res != 0 {
		return EnvError(env, `Unable to execute command "%s"`, cmd)
	}
	return nil
}
