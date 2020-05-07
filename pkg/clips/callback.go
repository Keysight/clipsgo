package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips/clips.h>
import "C"
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

func convertArg(funcname Symbol, haveType reflect.Type, needType reflect.Type, arg interface{}) (interface{}, error) {
	if haveType.AssignableTo(needType) {
		return arg, nil
	}
	if haveType.Kind() == reflect.Int64 {
		// Make an exception when it's just loss of scale, and make it work
		intval := arg.(int64)
		switch needType.Kind() {
		case reflect.Int:
			ret := int(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Int32:
			ret := int32(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Int16:
			ret := int16(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Int8:
			ret := int8(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Uint:
			ret := uint(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Uint64:
			ret := uint64(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Uint32:
			ret := uint32(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Uint16:
			ret := uint16(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		case reflect.Uint8:
			ret := uint8(intval)
			if int64(ret) != intval {
				return nil, fmt.Errorf(`Integer %d too large calling function "%s"`, intval, funcname)
			}
			return ret, nil
		}
	} else if haveType.Kind() == reflect.Float64 {
		floatval := arg.(float64)
		if needType.Kind() == reflect.Float32 {
			ret := float32(floatval)
			if float64(ret) != floatval {
				return nil, fmt.Errorf(`Floating point %f too precise calling function "%s"`, floatval, funcname)
			}
			return ret, nil
		}
	} else if haveType.Kind() == reflect.Slice && needType.Kind() == reflect.Slice {
		// see if we can translate to right kind of slice
		haveArr := reflect.ValueOf(arg)
		eNeedType := needType.Elem()
		slice := reflect.MakeSlice(reflect.SliceOf(eNeedType), haveArr.Len(), haveArr.Len())
		for i := 0; i < haveArr.Len(); i++ {
			// what we get from CLIPS is always []interface{}, so there's no check for that here
			val := haveArr.Index(i).Elem()
			valif, err := convertArg(funcname, val.Type(), eNeedType, val.Interface())
			if err != nil {
				return nil, err
			}
			//slice.SetLen(i)
			slice.Index(i).Set(reflect.ValueOf(valif))
		}
		return slice.Interface(), nil
	}
	return nil, fmt.Errorf(`Invalid type "%v" passed to function "%s", expected "%v"`, haveType, funcname, needType)
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
		arg := temp.Value()
		haveType := reflect.TypeOf(arg)
		arg, err := convertArg(funcname, haveType, needType, arg)
		if err != nil {
			printError(env, err.Error())
			return
		}
		arguments = append(arguments, reflect.ValueOf(arg))
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
