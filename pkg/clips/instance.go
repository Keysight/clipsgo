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
	"runtime"
	"unsafe"
)

// Instance represents an instance of a class from CLIPS
type Instance struct {
	env     *Environment
	instptr unsafe.Pointer
}

// InstancesChanged returns true if any instance has changed
func (env *Environment) InstancesChanged() bool {
	ret := C.EnvGetInstancesChanged(env.env)
	C.EnvSetInstancesChanged(env.env, 0)
	if ret == 1 {
		return true
	}
	return false
}

// Instances returns all defined instances
func (env *Environment) Instances() []*Instance {
	instptr := C.EnvGetNextInstance(env.env, nil)
	ret := make([]*Instance, 0, 10)
	for instptr != nil {
		ret = append(ret, createInstance(env, instptr))
		instptr = C.EnvGetNextInstance(env.env, instptr)
	}
	return ret
}

// FindInstance returns the instance of the given name. module may be the empty string to use the current module
func (env *Environment) FindInstance(name InstanceName, module string) (*Instance, error) {
	var modptr unsafe.Pointer
	if module != "" {
		cmod := C.CString(module)
		defer C.free(unsafe.Pointer(cmod))
		modptr = C.EnvFindDefmodule(env.env, cmod)
		if modptr == nil {
			return nil, NotFoundError(fmt.Errorf(`Module "%s" not found`, module))
		}
	}
	cname := C.CString(string(name))
	defer C.free(unsafe.Pointer(cname))
	instptr := C.EnvFindInstance(env.env, modptr, cname, 1)
	if instptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Instance "%s" not found`, name))
	}
	return createInstance(env, instptr), nil
}

// LoadInstancesFromString loads a set of instances into the CLIPS database. Equivalent to the load-instances command
func (env *Environment) LoadInstancesFromString(instances string) error {
	cstr := C.CString(instances)
	defer C.free(unsafe.Pointer(cstr))
	ret := int(C.EnvLoadInstancesFromString(env.env, cstr, -1))
	if ret == -1 {
		return EnvError(env, "Unable to load instances")
	}
	return nil
}

// LoadInstances loads a set of instances into the CLIPS database. Equivalent to the load-instances command
func (env *Environment) LoadInstances(filename string) error {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))
	ret := C.EnvBinaryLoadInstances(env.env, cstr)
	if ret != -1 {
		return nil
	}
	ret = C.EnvLoadInstances(env.env, cstr)
	if ret == -1 {
		return EnvError(env, "Unable to load instances")
	}
	return nil
}

// RestoreInstancesFromString loads a set of instances into CLIPS, bypassing message handling. Intended for use with save. Equivalent to restore-isntances command
func (env *Environment) RestoreInstancesFromString(instances string) error {
	cstr := C.CString(instances)
	defer C.free(unsafe.Pointer(cstr))
	ret := C.EnvRestoreInstancesFromString(env.env, cstr, -1)
	if ret == -1 {
		return EnvError(env, "Unable to restore instances")
	}
	return nil
}

// RestoreInstances loads a set of instances into CLIPS, bypassing message handling. Intended for use with save. Equivalent to restore-isntances command
func (env *Environment) RestoreInstances(filename string) error {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))
	ret := C.EnvRestoreInstances(env.env, cstr)
	if ret == -1 {
		return EnvError(env, "Unable to restore instances")
	}
	return nil
}

// SaveInstances saves the instances in the system to the specified file. If binary is true, instances will be aaved in binary format. Equivalent to save-instances
func (env *Environment) SaveInstances(path string, binary bool, mode SaveMode) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var ret C.long
	if binary {
		ret = C.EnvBinarySaveInstances(env.env, cpath, mode.CVal())
	} else {
		ret = C.EnvSaveInstances(env.env, cpath, mode.CVal())
	}
	if ret == 0 {
		return EnvError(env, "Unable to save instances")
	}
	return nil
}

// MakeInstance creates and initializes an instance of a user-defined class. Equivalent to make-instance Command must be a string in the form
// ([<instance-name>] of <class-name> <slot-override>*)
// <slot-override> :== (<slot-name> <constant>*)
func (env *Environment) MakeInstance(command string) (*Instance, error) {
	ccmd := C.CString(command)
	defer C.free(unsafe.Pointer(ccmd))
	instptr := C.EnvMakeInstance(env.env, ccmd)
	if instptr == nil {
		return nil, EnvError(env, "Unable to create instance")
	}
	return createInstance(env, instptr), nil
}

func createInstance(env *Environment, instptr unsafe.Pointer) *Instance {
	ret := &Instance{
		env:     env,
		instptr: instptr,
	}
	C.EnvIncrementInstanceCount(env.env, instptr)
	runtime.SetFinalizer(ret, func(*Instance) {
		ret.Drop()
	})
	return ret

}

// Drop drops the reference to the instance in CLIPS. should be called when done with the instance
func (inst *Instance) Drop() {
	if inst.instptr != nil {
		C.EnvDecrementInstanceCount(inst.env.env, inst.instptr)
		inst.instptr = nil
	}
}

// Equal returns true if the other instance represents the same CLIPS inst as this one
func (inst *Instance) Equal(other *Instance) bool {
	return inst.instptr == other.instptr
}

func (inst *Instance) String() string {
	var bufsize C.ulong = 1024
	buf := (*C.char)(C.malloc(C.sizeof_char * bufsize))
	defer C.free(unsafe.Pointer(buf))
	C.EnvGetInstancePPForm(inst.env.env, buf, bufsize-1, inst.instptr)
	return C.GoString(buf)
}

// Name returns the name of this instance
func (inst *Instance) Name() InstanceName {
	ret := C.EnvGetInstanceName(inst.env.env, inst.instptr)
	return InstanceName(C.GoString(ret))
}

// Class returns a reference to the class of this instance
func (inst *Instance) Class() *Class {
	clptr := C.EnvGetInstanceClass(inst.env.env, inst.instptr)
	return createClass(inst.env, clptr)
}

// Slots returns a map of values for each slot by name
func (inst *Instance) Slots(inherited bool) map[string]interface{} {
	cl := inst.Class()
	slots := cl.Slots(inherited)
	ret := make(map[string]interface{}, len(slots))
	for _, slot := range slots {
		name := slot.Name()
		ret[name] = inst.slotValue(name)
	}
	return ret
}

// Slot returns the value of the given slot. Warning, this function bypasses message-passing
func (inst *Instance) Slot(name string) (interface{}, error) {
	cl := inst.Class()
	_, err := cl.Slot(name)
	if err != nil {
		return nil, err
	}
	return inst.slotValue(name), nil
}

func (inst *Instance) slotValue(name string) interface{} {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	data := createDataObject(inst.env)
	defer data.Delete()
	C.EnvDirectGetSlot(inst.env.env, inst.instptr, cname, data.byRef())
	return data.Value()
}

// SetSlot sets the slot to the given value. Warning, this function bypasses message-passing
func (inst *Instance) SetSlot(name string, value interface{}) error {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Struct {
		// need to insert the struct first, then store its INSTANCE-NAME
		subinst, err := inst.env.Insert("", value)
		if err != nil {
			return err
		}
		value = subinst.Name()
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	data := createDataObject(inst.env)
	defer data.Delete()

	data.SetValue(value)

	ret := C.EnvDirectPutSlot(inst.env.env, inst.instptr, cname, data.byRef())
	if ret == 0 {
		return EnvError(inst.env, `Unable to set slot "%s"`, name)
	}
	return nil
}

// Send sends a message tot his instance. Message arguments must be provided as a string
func (inst *Instance) Send(message string, arguments string) interface{} {
	data := createDataObject(inst.env)
	defer data.Delete()

	instaddr := createDataObject(inst.env)
	defer instaddr.Delete()
	instaddr.SetValue(inst)

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	var cargs *C.char
	if arguments != "" {
		cargs = C.CString(arguments)
		defer C.free(unsafe.Pointer(cargs))
	}
	C.EnvSend(inst.env.env, instaddr.byRef(), cmsg, cargs, data.byRef())
	return data.Value()
}

// Delete unmakes the instance within CLIPS, bypassing message passing
func (inst *Instance) Delete() error {
	ret := C.EnvDeleteInstance(inst.env.env, inst.instptr)
	if ret != 1 {
		return EnvError(inst.env, "Unable to delete instance")
	}
	return nil
}

// Unmake unmakes the instance within CLIPS, using message passing
func (inst *Instance) Unmake() error {
	ret := C.EnvUnmakeInstance(inst.env.env, inst.instptr)
	if ret != 1 {
		return EnvError(inst.env, "Unable to unmake instance")
	}
	return nil
}

// ExtractSlot obtains the given slot value into the user-provided object
func (inst *Instance) ExtractSlot(retval interface{}, name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	data := createDataObject(inst.env)
	defer data.Delete()
	C.EnvDirectGetSlot(inst.env.env, inst.instptr, cname, data.byRef())
	return data.ExtractValue(retval, true)
}

// Extract attempts to marshall the CLIPS instance data into the user-provided or pointer
// The return value can be a struct or a map of string to another datatype. If retval points
// to a valid object, that object will be populated. If it is not, one will be created
func (inst *Instance) Extract(retval interface{}) error {
	slots := inst.Slots(true)
	knownInstances := make(map[InstanceName]interface{})
	knownInstances[inst.Name()] = retval
	return inst.env.structuredExtract(retval, slots, true, knownInstances)
}
