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
	"reflect"
	"unsafe"
)

func printError(env *Environment, err string) {
	werror := C.CString(C.WERROR)
	// because this is a const, free is neither necessary nor allowed
	//defer C.free(unsafe.Pointer(werror))
	fullerr := fmt.Sprintf("\nERROR: \n%s\n", err)
	cerr := C.CString(fullerr)
	defer C.free(unsafe.Pointer(werror))
	C.EnvPrintRouter(env.env, werror, cerr)
	C.SetEvaluationError(env.env, 1)
}

//export goFunction
func goFunction(envptr unsafe.Pointer, dataObject *C.struct_dataObject) {
	env, ok := environmentObj[envptr]
	if !ok {
		panic("Got a callback from an unknown environment")
	}
	temp := createDataObject(env)
	returnData := createDataObjectInitialized(env, dataObject)
	argnum := int(C.EnvRtnArgCount(envptr)) - 1
	arguments := make([]reflect.Value, 0, argnum)

	fname := C.CString("go-function")
	defer C.free(unsafe.Pointer(fname))
	if C.EnvArgTypeCheck(envptr, fname, 1, SYMBOL.CVal(), temp.byRef()) != 1 {
		printError(env, "Improper Go function call, function name missing")
		return
	}

	funcval := temp.Value()
	funcname, ok := funcval.(Symbol)
	if !ok {
		printError(env, "Unexpected argument type in callback")
		return
	}
	fn, ok := env.callback[string(funcname)]
	if !ok {
		printError(env, fmt.Sprintf(`Unknown callback name "%s"`, funcname))
		return
	}

	typ := fn.Type()
	if !typ.IsVariadic() {
		if argnum < typ.NumIn() {
			printError(env, fmt.Sprintf(`Not enough arguments to "%s"`, funcname))
			return
		}
		if argnum > typ.NumIn() {
			printError(env, fmt.Sprintf(`Too many arguments to "%s"`, funcname))
			return
		}
	} else {
		if argnum < typ.NumIn()-1 {
			printError(env, fmt.Sprintf(`Not enough arguments to "%s"`, funcname))
			return
		}
	}

	fixedArgs := typ.NumIn()
	if typ.IsVariadic() {
		fixedArgs--
	}
	knownInstances := make(map[InstanceName]interface{})
	for index := 0; index < argnum; index++ {
		// CLIPS is 1-based plus we prefixed args with function name
		C.EnvRtnUnknown(envptr, C.int(index+2), temp.byRef())

		var needType reflect.Type
		if index >= fixedArgs {
			// variadic arguments
			needType = typ.In(fixedArgs).Elem()
		} else {
			needType = typ.In(index)
		}
		paramVal := reflect.New(needType).Elem()
		arg := temp.Value()
		err := env.convertArg(paramVal, reflect.ValueOf(arg), true, knownInstances)
		if err != nil {
			printError(env, fmt.Sprintf("error calling function %s: %v", funcname, err.Error()))
			return
		}
		arguments = append(arguments, paramVal)
	}
	ret := fn.Call(arguments)
	if ret == nil {
		returnData.SetValue(false)
		return
	}
	// see if the final return value is an error type
	errVal := ret[len(ret)-1]
	if errVal.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		// if it is, treat it as an error not a return
		if !errVal.IsNil() {
			err := errVal.MethodByName("Error").Call([]reflect.Value{})
			printError(env, fmt.Sprintf(`Error from user function: %s: %s`,
				errVal.Type().String(), err))
			return
		}
		// remove the error argument
		ret = ret[:len(ret)-1]
	}
	retlist := make([]interface{}, len(ret))
	for i, retval := range ret {
		retlist[i] = retval.Interface()
	}

	if len(retlist) > 1 {
		returnData.SetValue(retlist)
	} else if len(retlist) == 1 {
		returnData.SetValue(retlist[0])
	} else {
		returnData.SetValue(false)
	}
}
