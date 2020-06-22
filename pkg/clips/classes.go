package clips
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

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

// Class is a reference to a CLIPS class
type Class struct {
	env   *Environment
	clptr unsafe.Pointer
}

// MessageHandler is a reference to a messagehandler for a particular class
type MessageHandler struct {
	class *Class
	index C.int
}

// ClassDefaultsMode defines the mode for defaults in a class
type ClassDefaultsMode int

const (
	CONVENIENCE_MODE ClassDefaultsMode = iota
	CONSERVATION_MODE
)

var clipsClassDefaultsMode = [...]string{
	"CONVENIENCE_MODE",
	"CONSERVATION_MODE",
}

func (sm ClassDefaultsMode) String() string {
	return clipsClassDefaultsMode[int(sm)]
}

// CVal returns the value as appropriate for a C call
func (sm ClassDefaultsMode) CVal() C.ushort {
	return C.ushort(sm)
}

type MessageHandlerType Symbol

const (
	AROUND  MessageHandlerType = "around"
	BEFORE  MessageHandlerType = "before"
	PRIMARY MessageHandlerType = "primary"
	AFTER   MessageHandlerType = "after"
)

// ClassDefaultsMode returns the current class defaults mode. Equivalent to (get-class-defaults-mode)
func (env *Environment) ClassDefaultsMode() ClassDefaultsMode {
	ret := C.EnvGetClassDefaultsMode(env.env)
	return ClassDefaultsMode(ret)
}

// SetClassDefaultsMode sets the class defaults mode
func (env *Environment) SetClassDefaultsMode(mode ClassDefaultsMode) {
	C.EnvSetClassDefaultsMode(env.env, mode.CVal())
}

// Classes returns the set of defined classes
func (env *Environment) Classes() []*Class {
	clptr := C.EnvGetNextDefclass(env.env, nil)
	ret := make([]*Class, 0, 10)
	for clptr != nil {
		ret = append(ret, createClass(env, clptr))
		clptr = C.EnvGetNextDefclass(env.env, clptr)
	}
	return ret
}

// FindClass returns a reference to the given class
func (env *Environment) FindClass(name string) (*Class, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	clptr := C.EnvFindDefclass(env.env, cname)
	if clptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Class "%s" not found`, name))
	}
	return createClass(env, clptr), nil
}

func createClass(env *Environment, clptr unsafe.Pointer) *Class {
	return &Class{
		env:   env,
		clptr: clptr,
	}
}

// Name returns the name of this class
func (cl *Class) Name() string {
	ret := C.EnvGetDefclassName(cl.env.env, cl.clptr)
	return C.GoString(ret)
}

func (cl *Class) String() string {
	ret := C.EnvGetDefclassPPForm(cl.env.env, cl.clptr)
	if ret == nil {
		return ""
	}
	return strings.TrimRight(C.GoString(ret), "\n")
}

// Equal returns true if other class represents the same CLIPS class as this one
func (cl *Class) Equal(other *Class) bool {
	return cl.clptr == other.clptr
}

// Abstract returns true if the class is abstract
func (cl *Class) Abstract() bool {
	ret := C.EnvClassAbstractP(cl.env.env, cl.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// Reactive returns true if the class is reactive
func (cl *Class) Reactive() bool {
	ret := C.EnvClassReactiveP(cl.env.env, cl.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// Module returns the module in which this class is defined
func (cl *Class) Module() *Module {
	modname := C.EnvDefclassModule(cl.env.env, cl.clptr)
	modptr := C.EnvFindDefmodule(cl.env.env, modname)
	return createModule(cl.env, modptr)
}

// Deletable returns true if the class is unreferenced and therefore deletable
func (cl *Class) Deletable() bool {
	ret := C.EnvIsDefclassDeletable(cl.env.env, cl.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// WatchedInstances returns true if the class instances are being watched
func (cl *Class) WatchedInstances() bool {
	ret := C.EnvGetDefclassWatchInstances(cl.env.env, cl.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// WatchInstances sets whether instances of this class should be watched
func (cl *Class) WatchInstances(val bool) {
	var flag C.uint
	if val {
		flag = 1
	}
	C.EnvSetDefclassWatchInstances(cl.env.env, flag, cl.clptr)
}

// WatchedSlots returns true if the class slots are being watched
func (cl *Class) WatchedSlots() bool {
	ret := C.EnvGetDefclassWatchSlots(cl.env.env, cl.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// WatchSlots sets whether instances of this class should be watched
func (cl *Class) WatchSlots(val bool) {
	var flag C.uint
	if val {
		flag = 1
	}
	C.EnvSetDefclassWatchSlots(cl.env.env, flag, cl.clptr)
}

// NewInstance creates an instance of this class. If skipInit is true, a new,
// uninitialized instance of this class. Slots will be unset until the caller
// calls SetSlot on each one, or calls (initialize-instance [instname])
func (cl *Class) NewInstance(name string, skipInit bool) (*Instance, error) {
	if !skipInit {
		var cmd string
		if name == "" {
			cmd = fmt.Sprintf("(of %s)", cl.Name())
		} else {
			cmd = fmt.Sprintf("(%s of %s)", name, cl.Name())
		}
		return cl.env.MakeInstance(cmd)
	}
	if name == "" {
		if err := cl.env.ExtractEval(&name, "(gensym)"); err != nil {
			return nil, err
		}
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	instptr := C.EnvCreateRawInstance(cl.env.env, cl.clptr, cname)
	if instptr == nil {
		return nil, EnvError(cl.env, "Unable to create instance")
	}
	return createInstance(cl.env, instptr), nil
}

// MessageHandlers returns a list of all message handlers for this class
func (cl *Class) MessageHandlers() []*MessageHandler {
	index := C.EnvGetNextDefmessageHandler(cl.env.env, cl.clptr, 0)

	ret := make([]*MessageHandler, 0, 10)
	for index != 0 {
		ret = append(ret, createMessageHandler(cl, index))
		index = C.EnvGetNextDefmessageHandler(cl.env.env, cl.clptr, index)
	}
	return ret
}

// FindMessageHandler returns a reference to the named message handler
func (cl *Class) FindMessageHandler(name string, handlerType MessageHandlerType) (*MessageHandler, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	chandler := C.CString(string(handlerType))
	defer C.free(unsafe.Pointer(chandler))
	index := C.EnvFindDefmessageHandler(cl.env.env, cl.clptr, cname, chandler)
	if index == 0 {
		return nil, EnvError(cl.env, `MessageHandler "%s" of type "%s" not found`, name, handlerType)
	}
	return createMessageHandler(cl, C.int(index)), nil
}

// Subclass returns true if this class is a subclass of the given one
func (cl *Class) Subclass(other *Class) bool {
	if cl.env != other.env {
		return false
	}
	ret := C.EnvSubclassP(cl.env.env, cl.clptr, other.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// Superclass returns true if this class is a superclass of the given one
func (cl *Class) Superclass(other *Class) bool {
	if cl.env != other.env {
		return false
	}
	ret := C.EnvSuperclassP(cl.env.env, cl.clptr, other.clptr)
	if ret == 1 {
		return true
	}
	return false
}

// Slots returns a list of all slots for this class. inhereted determines whether inhereted slots are included
func (cl *Class) Slots(inherited bool) []*ClassSlot {
	data := createDataObject(cl.env)
	defer data.Delete()

	var flag C.int
	if inherited {
		flag = 1
	}

	C.EnvClassSlots(cl.env.env, cl.clptr, data.byRef(), flag)
	dv := data.Value()
	slots, ok := dv.([]interface{})
	if !ok {
		panic("Unexpected response from CLIPS")
	}
	ret := make([]*ClassSlot, len(slots))
	var ii int
	for _, v := range slots {
		slotname, ok := v.(Symbol)
		if !ok {
			panic("Unexpected response from clips")
		}
		ret[ii] = createClassSlot(cl, string(slotname))
		ii++
	}
	return ret
}

// Slot returns the given slot by name
func (cl *Class) Slot(name string) (*ClassSlot, error) {
	slots := cl.Slots(true)
	for _, slot := range slots {
		if slot.Name() == name {
			return slot, nil
		}
	}
	return nil, NotFoundError(fmt.Errorf(`Slot "%s" not found`, name))
}

// Instances returns the list of instances of this class
func (cl *Class) Instances() []*Instance {
	instptr := C.EnvGetNextInstanceInClass(cl.env.env, cl.clptr, nil)

	ret := make([]*Instance, 0, 10)
	for instptr != nil {
		ret = append(ret, createInstance(cl.env, instptr))
		instptr = C.EnvGetNextInstanceInClass(cl.env.env, cl.clptr, instptr)
	}
	return ret
}

// Subclasses returns the list of subclasses of this class
func (cl *Class) Subclasses(inherited bool) ([]*Class, error) {
	data := createDataObject(cl.env)
	defer data.Delete()

	var flag C.int
	if inherited {
		flag = 1
	}

	C.EnvClassSubclasses(cl.env.env, cl.clptr, data.byRef(), flag)
	return classes(cl.env, data.Value())
}

// Superclasses returns the list of superclasses of this class
func (cl *Class) Superclasses(inherited bool) ([]*Class, error) {
	data := createDataObject(cl.env)
	defer data.Delete()

	var flag C.int
	if inherited {
		flag = 1
	}

	C.EnvClassSuperclasses(cl.env.env, cl.clptr, data.byRef(), flag)
	return classes(cl.env, data.Value())
}

// Undefine undefines the class within CLIPS. Equivalent to undefclass
func (cl *Class) Undefine() error {
	ret := C.EnvUndefclass(cl.env.env, cl.clptr)
	if ret != 1 {
		return EnvError(cl.env, "Unable to undefine class")
	}
	cl.clptr = nil
	return nil
}

func createMessageHandler(class *Class, index C.int) *MessageHandler {
	return &MessageHandler{
		class: class,
		index: index,
	}
}

// Name returns the name of this message handler
func (mh *MessageHandler) Name() string {
	ret := C.EnvGetDefmessageHandlerName(mh.class.env.env, mh.class.clptr, mh.index)
	return C.GoString(ret)
}

func (mh *MessageHandler) String() string {
	ret := C.EnvGetDefmessageHandlerPPForm(mh.class.env.env, mh.class.clptr, mh.index)
	return strings.TrimRight(C.GoString(ret), "\n")
}

// Equal returns true if this messagehandler represents the same CLIPS handler as the other one
func (mh *MessageHandler) Equal(other *MessageHandler) bool {
	return mh.class.Equal(other.class) && mh.index == other.index
}

// Type returns the messagehandler type
func (mh *MessageHandler) Type() MessageHandlerType {
	ret := C.EnvGetDefmessageHandlerType(mh.class.env.env, mh.class.clptr, mh.index)
	return MessageHandlerType(C.GoString(ret))
}

// Watched returns true if this messagehandler is being watched
func (mh *MessageHandler) Watched() bool {
	ret := C.EnvGetDefmessageHandlerWatch(mh.class.env.env, mh.class.clptr, mh.index)
	if ret == 1 {
		return true
	}
	return false
}

// Watch sets whether this messagehandler should be watched
func (mh *MessageHandler) Watch(val bool) {
	var flag C.int
	if val {
		flag = 1
	}
	C.EnvSetDefmessageHandlerWatch(mh.class.env.env, flag, mh.class.clptr, mh.index)
}

// Deletable returns true if this messagehandler can be deleted
func (mh *MessageHandler) Deletable() bool {
	ret := C.EnvIsDefmessageHandlerDeletable(mh.class.env.env, mh.class.clptr, mh.index)
	if ret == 1 {
		return true
	}
	return false
}

// Undefine undefines the message handler. Equivalent to undefmessage-handler
func (mh *MessageHandler) Undefine() error {
	ret := C.EnvUndefmessageHandler(mh.class.env.env, mh.class.clptr, mh.index)
	if ret != 1 {
		return EnvError(mh.class.env, "Unable to undef message handler")
	}
	mh.index = 0
	return nil
}

func classes(env *Environment, classlist interface{}) ([]*Class, error) {
	classes, ok := classlist.([]interface{})
	if !ok {
		panic("Unexpected response from CLIPS")
	}
	ret := make([]*Class, len(classes))
	var ii int
	for _, v := range classes {
		classname, ok := v.(Symbol)
		if !ok {
			panic("Unexpected response from clips")
		}
		cname := C.CString(string(classname))
		defer C.free(unsafe.Pointer(cname))
		clptr := C.EnvFindDefclass(env.env, cname)
		if clptr == nil {
			return nil, EnvError(env, `Class "%s" not found`, classname)
		}
		ret[ii] = createClass(env, clptr)
		ii++
	}
	return ret, nil
}
