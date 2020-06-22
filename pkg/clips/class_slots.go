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
	"unsafe"
)

// ClassSlot is a reference to a slot within a particular class
type ClassSlot struct {
	class *Class
	name  string
}

func createClassSlot(class *Class, name string) *ClassSlot {
	return &ClassSlot{
		class: class,
		name:  name,
	}
}

// Name returns the name of the slot
func (cs *ClassSlot) Name() string {
	return cs.name
}

func (cs *ClassSlot) String() string {
	return cs.name
}

// Equal returns true if this slot represents the same CLIPS slot as the other slot
func (cs *ClassSlot) Equal(other *ClassSlot) bool {
	return cs.class.Equal(other.class) && cs.name == other.name
}

// Public returns true if the slot is public
func (cs *ClassSlot) Public() bool {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvSlotPublicP(cs.class.env.env, cs.class.clptr, cname)
	if ret == 1 {
		return true
	}
	return false
}

// Initable returns true if the slot is initable
func (cs *ClassSlot) Initable() bool {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvSlotInitableP(cs.class.env.env, cs.class.clptr, cname)
	if ret == 1 {
		return true
	}
	return false
}

// Writable returns true if the slot is writable
func (cs *ClassSlot) Writable() bool {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvSlotWritableP(cs.class.env.env, cs.class.clptr, cname)
	if ret == 1 {
		return true
	}
	return false
}

// Accessible returns true if the slot is accessible
func (cs *ClassSlot) Accessible() bool {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvSlotDirectAccessP(cs.class.env.env, cs.class.clptr, cname)
	if ret == 1 {
		return true
	}
	return false
}

// Types returns a list of value types for this slot. Equivalent to slot-types
func (cs *ClassSlot) Types() []Symbol {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	data := createDataObject(cs.class.env)
	defer data.Delete()
	C.EnvSlotTypes(cs.class.env.env, cs.class.clptr, cname, data.byRef())

	types, ok := data.Value().([]interface{})
	if !ok {
		return make([]Symbol, 0)
	}
	ret := make([]Symbol, 0, len(types))
	for _, v := range types {
		s, ok := v.(Symbol)
		if !ok {
			panic("Unexpected response from CLIPS")
		}
		ret = append(ret, s)
	}
	return ret
}

// Sources returns a list of names of class sources for this slot. Equivalent to slot-sources
func (cs *ClassSlot) Sources() []Symbol {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	data := createDataObject(cs.class.env)
	defer data.Delete()
	C.EnvSlotSources(cs.class.env.env, cs.class.clptr, cname, data.byRef())

	sources, ok := data.Value().([]interface{})
	if !ok {
		return make([]Symbol, 0)
	}
	ret := make([]Symbol, 0, len(sources))
	for _, v := range sources {
		s, ok := v.(Symbol)
		if !ok {
			panic("Unexpected response from CLIPS")
		}
		ret = append(ret, s)
	}
	return ret
}

// IntRange returns the numeric range for the slot for integer values - e.g. low, haslow, high, hashigh := ts.Range()
func (cs *ClassSlot) IntRange() (low int64, hasLow bool, high int64, hasHigh bool) {
	data := createDataObject(cs.class.env)
	defer data.Delete()
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvSlotRange(cs.class.env.env, cs.class.clptr, cname, data.byRef())
	dv := data.Value()
	ilist, ok := dv.([]interface{})
	if !ok {
		return 0, false, 0, false
	}
	if len(ilist) != 2 {
		panic("Unexpected response from CLIPS for range")
	}

	// fmt.Printf("%v / %v\n", reflect.TypeOf(ilist[0]), reflect.TypeOf(ilist[1]))
	// A Symbol represents infinity
	low, hasLow = ilist[0].(int64)
	high, hasHigh = ilist[1].(int64)
	return
}

// FloatRange returns the numeric range for the slot for floating point values - e.g. low, haslow, high, hashigh := ts.Range()
func (cs *ClassSlot) FloatRange() (low float64, hasLow bool, high float64, hasHigh bool) {
	data := createDataObject(cs.class.env)
	defer data.Delete()
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvSlotRange(cs.class.env.env, cs.class.clptr, cname, data.byRef())
	dv := data.Value()
	ilist, ok := dv.([]interface{})
	if !ok {
		return 0, false, 0, false
	}
	if len(ilist) != 2 {
		panic("Unexpected response from CLIPS for range")
	}

	// fmt.Printf("%v / %v\n", reflect.TypeOf(ilist[0]), reflect.TypeOf(ilist[1]))
	// A Symbol represents infinity
	low, hasLow = ilist[0].(float64)
	high, hasHigh = ilist[1].(float64)
	return
}

// Facets returns a list of facets for this slot
func (cs *ClassSlot) Facets() []Symbol {
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))
	data := createDataObject(cs.class.env)
	defer data.Delete()
	C.EnvSlotFacets(cs.class.env.env, cs.class.clptr, cname, data.byRef())

	facets, ok := data.Value().([]interface{})
	if !ok {
		return make([]Symbol, 0)
	}
	ret := make([]Symbol, 0, len(facets))
	for _, v := range facets {
		s, ok := v.(Symbol)
		if !ok {
			panic("Unexpected response from CLIPS")
		}
		ret = append(ret, s)
	}
	return ret
}

// Cardinality returns the cardinality for the slot
func (cs *ClassSlot) Cardinality() (low int64, high int64, hasHigh bool) {
	data := createDataObject(cs.class.env)
	defer data.Delete()
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvSlotCardinality(cs.class.env.env, cs.class.clptr, cname, data.byRef())
	dv := data.Value()
	ilist, ok := dv.([]interface{})
	if !ok || len(ilist) != 2 {
		return 0, 0, false
	}
	low, _ = ilist[0].(int64)
	high, hasHigh = ilist[1].(int64)
	return
}

// DefaultValue returns a default value for the slot.  (This might be a new, unique value for DYNAMIC_DEFAULT defaults)
func (cs *ClassSlot) DefaultValue() interface{} {
	data := createDataObject(cs.class.env)
	defer data.Delete()
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvSlotDefaultValue(cs.class.env.env, cs.class.clptr, cname, data.byRef())
	return data.Value()
}

// AllowedValues returns the set of allowed values for this slot, if specified
func (cs *ClassSlot) AllowedValues() (values []interface{}, ok bool) {
	data := createDataObject(cs.class.env)
	defer data.Delete()
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvSlotAllowedValues(cs.class.env.env, cs.class.clptr, cname, data.byRef())
	dv := data.Value()
	values, ok = dv.([]interface{})
	return
}

// AllowedClasses returns the names of allowed classes for this slot, if specified. Equivalent to slot-allowed-classes
func (cs *ClassSlot) AllowedClasses() (values []Symbol, ok bool) {
	data := createDataObject(cs.class.env)
	defer data.Delete()
	cname := C.CString(cs.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvSlotAllowedClasses(cs.class.env.env, cs.class.clptr, cname, data.byRef())
	dv := data.Value()
	ret, ok := dv.([]interface{})
	if !ok {
		values = make([]Symbol, 0)
		return
	}
	values = make([]Symbol, 0, len(ret))
	for _, v := range ret {
		s, ok := v.(Symbol)
		if !ok {
			panic("Unexpected response from CLIPS")
		}
		values = append(values, s)
	}
	ok = true
	return
}
