package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
//
// int queryFunction(void* env, const char* name);
// int printFunction(void* env, const char* name, const char* message);
// int getcFunction(void* env, const char* name);
// int ungetcFunction(void* env, int ch, const char* name);
// int exitFunction(void* env, int exitcode);
//
// int addRouter(void* env, char* routerName, int priority, char* userData) {
//	return EnvAddRouterWithContext(env, routerName, priority,
//		queryFunction,
//		printFunction,
//		getcFunction,
//		ungetcFunction,
//		exitFunction,
//		userData);
// }
//
// char* getNameFromContext(void* env) {
//   return (char*)GetEnvironmentRouterContext(env);
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
	"bytes"
	"fmt"
	"log"
	"strings"
	"unsafe"
)

// Router defines an object responsible for handling IO with CLIPS
type Router interface {
	// Name returns a name for this router
	Name() string

	// Query should return true if the router handles the given logical IO name
	Query(name string) bool

	// Print is called with a message if Query has returned true
	Print(name string, message string)

	// Getc is called by CLIPS to obtain a character from input
	Getc(name string) byte

	// Ungetc is called by CLIPS to push a character back into the input queue
	Ungetc(name string, ch byte) error

	// Exit is called by CLIPS before CLIPS itself exits
	Exit(exitcode int)

	// Activate activates this router with the Env
	Activate() error

	// Deactivate deactivates this router with the Env
	Deactivate() error

	// Delete removes this router from the Env
	Delete() error
}

// RouterCore is an implementation of common Router guts which can be used inside other Router implementations
type RouterCore struct {
	env        *Environment
	routerimpl Router
	name       string
	handled    map[string]interface{}
	priority   int
	routername *C.char
}

// LoggingRouter is a router that puts messages into go logging
type LoggingRouter struct {
	core    *RouterCore
	logger  *log.Logger
	linebuf bytes.Buffer
}

// CreateRouterCore creates an instance of the RouterCore which can be used to easily create a full Router
func CreateRouterCore(env *Environment, routerimpl Router, name string, handled []string, priority int) *RouterCore {
	ret := &RouterCore{
		env:        env,
		routerimpl: routerimpl,
		name:       name,
		handled:    make(map[string]interface{}, len(handled)),
		priority:   priority,
		routername: C.CString(name),
	}
	for _, v := range handled {
		ret.handled[v] = nil
	}
	env.router[name] = routerimpl
	C.addRouter(env.env, ret.routername, C.int(priority), ret.routername)
	return ret
}

// Name returns the name of the router
func (r *RouterCore) Name() string {
	return r.name
}

// Query returns true for handled logical io types
func (r *RouterCore) Query(name string) bool {
	_, ok := r.handled[name]
	return ok
}

// Print outputs message
func (r *RouterCore) Print(name string, message string) {
	fmt.Println(message)
}

// Activate activates the router in the Environment
func (r *RouterCore) Activate() error {
	errcode := int(C.EnvActivateRouter(r.env.env, r.routername))
	if errcode != 1 {
		return EnvError(r.env, "Failed to activate router")
	}
	return nil
}

// Deactivate deactives the router in the environment
func (r *RouterCore) Deactivate() error {
	errcode := int(C.EnvDeactivateRouter(r.env.env, r.routername))
	if errcode != 1 {
		return EnvError(r.env, "Failed to deactivate router")
	}
	return nil
}

// Delete deletes the router from the environment
func (r *RouterCore) Delete() error {
	defer C.free(unsafe.Pointer(r.routername))
	errcode := int(C.EnvDeleteRouter(r.env.env, r.routername))
	if errcode != 1 {
		return EnvError(r.env, "Failed to delete router")
	}
	return nil
}

var loggingHandlers = []string{
	"wtrace",
	"stdout",
	"wclips",
	"wdialog",
	"wdisplay",
	"wwarning",
	"werror",
}

// CreateLoggingRouter returns a new logging router
func CreateLoggingRouter(env *Environment, logger *log.Logger) *LoggingRouter {
	ret := &LoggingRouter{
		logger: logger,
	}
	ret.core = CreateRouterCore(env, ret, "go-logging-router", loggingHandlers, 30)
	return ret
}

// Name of this router
func (r *LoggingRouter) Name() string {
	return r.core.Name()
}

// Query should return true if the router handles the given logical IO name
func (r *LoggingRouter) Query(name string) bool {
	return r.core.Query(name)
}

// Print is called with a message if Query has returned true
func (r *LoggingRouter) Print(name string, message string) {
	r.linebuf.WriteString(message)
	for strings.Contains(r.linebuf.String(), "\n") {
		// logger.Print always adds a newline, so we don't print till we want one
		line, _ := r.linebuf.ReadString('\n')
		r.logger.Print(line)
	}
}

// Getc is called by CLIPS to obtain a character from input
func (r *LoggingRouter) Getc(name string) byte {
	return 0
}

// Ungetc is called by CLIPS to push a character back into the input queue
func (r *LoggingRouter) Ungetc(name string, ch byte) error {
	return fmt.Errorf("Not implemented")
}

// Exit is called by CLIPS before CLIPS itself exits
func (r *LoggingRouter) Exit(exitcode int) {
	log.Println("CLIPS will exit")
}

// Activate activates this router with the Env
func (r *LoggingRouter) Activate() error {
	return r.core.Activate()
}

// Deactivate deactivates this router with the Env
func (r *LoggingRouter) Deactivate() error {
	return r.core.Deactivate()
}

// Delete removes this router from the Env
func (r *LoggingRouter) Delete() error {
	return r.core.Delete()
}
