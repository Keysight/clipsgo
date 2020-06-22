package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
//
// int implied_deftemplate(void*);
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
	"strings"
	"unsafe"
)

// Template is a formal representation of the fact data structure, defined by deftemplate in CLIPS
type Template struct {
	env    *Environment
	tplptr unsafe.Pointer
}

// TemplateSlot defines one slot within a template
type TemplateSlot struct {
	tpl  *Template
	name string
}

// TemplateSlotDefaultType is used to specify how default is specifified for a slot within a template
type TemplateSlotDefaultType int

const (
	NO_DEFAULT TemplateSlotDefaultType = iota
	STATIC_DEFAULT
	DYNAMIC_DEFAULT
)

var templateSlotDefaultTypes = [...]string{
	"NO_DEFAULT",
	"STATIC_DEFAULT",
	"DYNAMIC_DEFAULT",
}

func (tsdt TemplateSlotDefaultType) String() string {
	return templateSlotDefaultTypes[tsdt]
}

// CVal returns the value as appropriate for a C call
func (tsdt TemplateSlotDefaultType) CVal() C.int {
	return C.int(tsdt)
}

func createTemplate(env *Environment, tplptr unsafe.Pointer) *Template {
	return &Template{
		env:    env,
		tplptr: tplptr,
	}
}

// Equal returns true if this template represents the same template as the given one
func (t *Template) Equal(other *Template) bool {
	return t.tplptr == other.tplptr
}

// String returns a string representation of the template
func (t *Template) String() string {
	cstr := C.EnvGetDeftemplatePPForm(t.env.env, t.tplptr)
	if cstr != nil {
		return strings.TrimRight(C.GoString(cstr), "\n")
	}
	cmodule := C.EnvDeftemplateModule(t.env.env, t.tplptr)
	name := t.Name()
	return fmt.Sprintf("(deftemplate %s::%s", C.GoString(cmodule), name)
}

// Name returns the name of this template
func (t *Template) Name() string {
	cname := C.EnvGetDeftemplateName(t.env.env, t.tplptr)
	return C.GoString(cname)
}

// Module returns the module in which the template is defined. Equivalent to (deftempalte-module)
func (t *Template) Module() *Module {
	cmodname := C.EnvDeftemplateModule(t.env.env, t.tplptr)
	modptr := C.EnvFindDefmodule(t.env.env, cmodname)
	return createModule(t.env, modptr)
}

// Implied returns whether the template is implied
func (t *Template) Implied() bool {
	if C.implied_deftemplate(t.tplptr) == 1 {
		return true
	}
	return false
}

// Watched returns whether or not the template is being watched
func (t *Template) Watched() bool {
	ret := C.EnvGetDeftemplateWatch(t.env.env, t.tplptr)
	if ret == 1 {
		return true
	}
	return false
}

// Watch sets whether or not the template should be watched
func (t *Template) Watch(val bool) {
	var cval C.uint = 0
	if val {
		cval = 1
	}
	C.EnvSetDeftemplateWatch(t.env.env, cval, t.tplptr)
}

// Deletable returns true if the Template can be deleted from CLIPS
func (t *Template) Deletable() bool {
	ret := C.EnvIsDeftemplateDeletable(t.env.env, t.tplptr)
	if ret == 1 {
		return true
	}
	return false
}

// Slots returns the slot definitions contained in this template
func (t *Template) Slots() map[string]*TemplateSlot {
	if t.Implied() {
		return make(map[string]*TemplateSlot)
	}

	data := createDataObject(t.env)
	defer data.Delete()

	C.EnvDeftemplateSlotNames(t.env.env, t.tplptr, data.byRef())
	namesblob := data.Value()
	names, ok := namesblob.([]interface{})
	if !ok {
		panic("Unexpected data returned from CLIPS for slot names")
	}
	ret := make(map[string]*TemplateSlot, len(names))
	for _, name := range names {
		namestr, ok := name.(Symbol)
		if !ok {
			panic("Unexpected data returned from CLIPS for slot names")
		}
		ret[string(namestr)] = t.createTemplateSlot(string(namestr))
	}
	return ret
}

// NewFact creates a new fact from this template
func (t *Template) NewFact() (Fact, error) {
	factptr := C.EnvCreateFact(t.env.env, t.tplptr)
	if factptr == nil {
		return nil, EnvError(t.env, "Unable to create fact from template %s", t.Name())
	}
	return t.env.newFact(unsafe.Pointer(factptr)), nil
}

// Undefine the template. Equivalent to (undeftemplate). This object is unusable after this call
func (t *Template) Undefine() error {
	ret := C.EnvUndeftemplate(t.env.env, t.tplptr)
	if ret != 1 {
		return EnvError(t.env, "Unable to undefine template %s", t.Name())
	}
	return nil
}

func (t *Template) createTemplateSlot(name string) *TemplateSlot {
	return &TemplateSlot{
		tpl:  t,
		name: name,
	}
}

// Equal checks if the other templateslot represents the same slot
func (ts *TemplateSlot) Equal(other *TemplateSlot) bool {
	if other == nil {
		return false
	}
	return ts.tpl.Equal(other.tpl) && ts.name == other.name
}

func (ts *TemplateSlot) String() string {
	return ts.name
}

// Name returns the name of this slot
func (ts *TemplateSlot) Name() string {
	return ts.name
}

// Multifield returns true if the slot is a multifield slot
func (ts *TemplateSlot) Multifield() bool {
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvDeftemplateSlotMultiP(ts.tpl.env.env, ts.tpl.tplptr, cname)
	if ret == 1 {
		return true
	}
	return false
}

// Types returns the set of value types for this slot
func (ts *TemplateSlot) Types() []Symbol {
	data := createDataObject(ts.tpl.env)
	defer data.Delete()
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvDeftemplateSlotTypes(ts.tpl.env.env, ts.tpl.tplptr, cname, data.byRef())
	dv := data.Value()
	ilist, ok := dv.([]interface{})
	if !ok {
		panic("Unexpected response from CLIPS for response types")
	}
	ret := make([]Symbol, len(ilist))
	i := 0
	for _, v := range ilist {
		ret[i], ok = v.(Symbol)
		if !ok {
			panic("Unexpected response from CLIPS for a response type")
		}
		i++
	}
	return ret
}

// IntRange returns the numeric range for the slot for integer values - e.g. low, haslow, high, hashigh := ts.Range()
func (ts *TemplateSlot) IntRange() (low int64, hasLow bool, high int64, hasHigh bool) {
	data := createDataObject(ts.tpl.env)
	defer data.Delete()
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvDeftemplateSlotRange(ts.tpl.env.env, ts.tpl.tplptr, cname, data.byRef())
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
func (ts *TemplateSlot) FloatRange() (low float64, hasLow bool, high float64, hasHigh bool) {
	data := createDataObject(ts.tpl.env)
	defer data.Delete()
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvDeftemplateSlotRange(ts.tpl.env.env, ts.tpl.tplptr, cname, data.byRef())
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

// Cardinality returns the cardinality for the slot
func (ts *TemplateSlot) Cardinality() (low int64, high int64, hasHigh bool) {
	data := createDataObject(ts.tpl.env)
	defer data.Delete()
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvDeftemplateSlotCardinality(ts.tpl.env.env, ts.tpl.tplptr, cname, data.byRef())
	dv := data.Value()
	ilist, ok := dv.([]interface{})
	if !ok || len(ilist) != 2 {
		return 0, 0, false
	}
	low, _ = ilist[0].(int64)
	high, hasHigh = ilist[1].(int64)
	return
}

// DefaultType returns the type of default value for this slot
func (ts *TemplateSlot) DefaultType() TemplateSlotDefaultType {
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.EnvDeftemplateSlotDefaultP(ts.tpl.env.env, ts.tpl.tplptr, cname)
	return TemplateSlotDefaultType(ret)
}

// DefaultValue returns a default value for the slot.  (This might be a new, unique value for DYNAMIC_DEFAULT defaults)
func (ts *TemplateSlot) DefaultValue() interface{} {
	data := createDataObject(ts.tpl.env)
	defer data.Delete()
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvDeftemplateSlotDefaultValue(ts.tpl.env.env, ts.tpl.tplptr, cname, data.byRef())
	return data.Value()
}

// AllowedValues returns the set of allowed values for this slot, if specified
func (ts *TemplateSlot) AllowedValues() (values []interface{}, ok bool) {
	data := createDataObject(ts.tpl.env)
	defer data.Delete()
	cname := C.CString(ts.name)
	defer C.free(unsafe.Pointer(cname))

	C.EnvDeftemplateSlotAllowedValues(ts.tpl.env.env, ts.tpl.tplptr, cname, data.byRef())
	dv := data.Value()
	values, ok = dv.([]interface{})
	return
}
