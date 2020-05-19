package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <string.h>
// #include <clips/clips.h>
//
// char* getNameFromContext(void* env);
import "C"
import "unsafe"

func lookupRouter(envptr unsafe.Pointer) Router {
	env := environmentObj[envptr]
	routername := C.GoString(C.getNameFromContext(envptr))
	return env.router[routername]
}

//export queryFunction
func queryFunction(envptr unsafe.Pointer, name *C.char) C.int {
	ret := lookupRouter(envptr).Query(C.GoString(name))
	if ret {
		return 1
	}
	return 0
}

//export printFunction
func printFunction(envptr unsafe.Pointer, name *C.char, message *C.char) C.int {
	length := C.strlen(message)
	lookupRouter(envptr).Print(C.GoString(name), C.GoStringN(message, C.int(length)))
	return 0
}

//export getcFunction
func getcFunction(envptr unsafe.Pointer, name *C.char) C.int {
	return C.int(lookupRouter(envptr).Getc(C.GoString(name)))
}

//export ungetcFunction
func ungetcFunction(envptr unsafe.Pointer, ch C.int, name *C.char) C.int {
	err := lookupRouter(envptr).Ungetc(C.GoString(name), byte(ch))
	if err != nil {
		return ch
	}
	return C.EOF
}

//export exitFunction
func exitFunction(envptr unsafe.Pointer, exitcode C.int) C.int {
	lookupRouter(envptr).Exit(int(exitcode))
	return 0
}
