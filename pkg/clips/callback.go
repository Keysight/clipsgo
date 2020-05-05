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

//export goFunction
func goFunction(envptr unsafe.Pointer, dataObject *C.struct_dataObject) {
	env, ok := environmentObj[envptr]
	if !ok {
		panic("Got a callback from an unknown environment")
	}
	arguments := make([]interface{}, 0)
	temp := createDataObject(env)
	data := createDataObjectInitialized(env, dataObject)
	argnum := int(C.EnvRtnArgCount(envptr))

	fname := C.CString("go-function")
	defer C.free(unsafe.Pointer(fname))
	if C.EnvArgTypeCheck(envptr, fname, 1, SYMBOL.CVal(), temp.byRef()) != 1 {
		panic("Error: Invalid argument count")
	}

	funcval := temp.Value()
	funcname, ok := funcval.(Symbol)
	if !ok {
		panic("Error: Unexpected argument type in callback")
	}
	fn, ok := env.callback[string(funcname)]
	if !ok {
		panic(fmt.Sprintf("Error: Unknown callback name %s", funcname))
	}

	for index := 2; index <= argnum; index++ {
		C.EnvRtnUnknown(envptr, C.int(index), temp.byRef())
		arguments = append(arguments, temp.Value())
	}
	ret, err := fn(arguments)
	if err != nil {
		ret = fmt.Sprintf("%v: %s", reflect.TypeOf(err).String(), err.Error())
	}
	// return value is set into pass-by-reference argument
	data.SetValue(ret)
}
