package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
import "C"
import (
	"fmt"
	"log"
	"strings"
	"unsafe"
)

/*
   Copyright 2020 Keysight Technologies

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Error error returned from CLIPS
type Error struct {
	Err  error
	Code string
}

// ErrorRouter is a router that puts messages into go logging
type ErrorRouter struct {
	core        *RouterCore
	lastMessage strings.Builder
}

func (e *Error) Error() string {
	return e.Err.Error()
}

// EnvError return an error that came from CLIPS
func EnvError(env *Environment, msg string, args ...interface{}) *Error {
	var shellmsg string
	if env.errRtr != nil {
		shellmsg = strings.Trim(env.errRtr.LastMessage(), "\n")
	}
	codestart := strings.Index(shellmsg, "[")
	codeend := strings.Index(shellmsg, "]")
	code := "Error"
	if codestart >= 0 && codeend >= 0 {
		code = shellmsg[codestart+1 : codeend]
	}
	msg = fmt.Sprintf(msg, args...)
	msg = fmt.Sprintf("%s: %s", msg, shellmsg)
	return &Error{
		Err:  fmt.Errorf(msg),
		Code: code,
	}
}

// CreateErrorRouter returns a new error accumulation router
func CreateErrorRouter(env *Environment) *ErrorRouter {
	ret := &ErrorRouter{
		lastMessage: strings.Builder{},
	}
	ret.core = CreateRouterCore(env, ret, "go-error-router", []string{"werror"}, 40)
	return ret
}

// Name of this router
func (r *ErrorRouter) Name() string {
	return r.core.Name()
}

// Query should return true if the router handles the given logical IO name
func (r *ErrorRouter) Query(name string) bool {
	return r.core.Query(name)
}

// Print is called with a message if Query has returned true
func (r *ErrorRouter) Print(name string, message string) {
	r.lastMessage.WriteString(message)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cmessage := C.CString(message)
	defer C.free(unsafe.Pointer(cmessage))
	r.Deactivate()
	defer r.Activate()
	C.EnvPrintRouter(r.core.env.env, cname, cmessage)
}

// LastMessage returns the accumulated error message and resets it
func (r *ErrorRouter) LastMessage() string {
	ret := r.lastMessage.String()
	r.lastMessage.Reset()
	return ret
}

// Getc is called by CLIPS to obtain a character from input
func (r *ErrorRouter) Getc(name string) byte {
	return 0
}

// Ungetc is called by CLIPS to push a character back into the input queue
func (r *ErrorRouter) Ungetc(name string, ch byte) error {
	return fmt.Errorf("Not implemented")
}

// Exit is called by CLIPS before CLIPS itself exits
func (r *ErrorRouter) Exit(exitcode int) {
	log.Println("CLIPS will exit")
}

// Activate activates this router with the Env
func (r *ErrorRouter) Activate() error {
	return r.core.Activate()
}

// Deactivate deactivates this router with the Env
func (r *ErrorRouter) Deactivate() error {
	return r.core.Deactivate()
}

// Delete removes this router from the Env
func (r *ErrorRouter) Delete() error {
	return r.core.Delete()
}
