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
	"reflect"
)

// Insert inserts the given object as a shadow instance in CLIPS. A shadow class
// will be created if it does not already exist
func (env *Environment) Insert(name string, basis interface{}, opts ...InsertClassOption) (*Instance, error) {
	knownBases := make(map[reflect.Value]InstanceName)
	return env.insertInstance(name, basis, knownBases)
}

func (env *Environment) insertInstance(name string, basis interface{}, knownBases map[reflect.Value]InstanceName, opts ...InsertClassOption) (*Instance, error) {
	typ := reflect.TypeOf(basis)
	val := reflect.ValueOf(basis)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	classname, err := classNameFor(typ)
	if err != nil {
		return nil, err
	}
	cls, err := env.checkRecurseClass(classname, typ)
	if err != nil {
		return nil, err
	}
	inst, err := cls.NewInstance(name, true)
	if err != nil {
		return nil, err
	}
	knownBases[val] = inst.Name()
	for ii := 0; ii < typ.NumField(); ii++ {
		field := typ.Field(ii)
		fieldval := val.Field(ii)

		if err := inst.fillSlot(field, fieldval, knownBases); err != nil {
			return nil, err
		}
	}

	/* This resets everything; not sure how to keep settings and still do this. Should be unnecessary as long as class is a shadow
	if _, err := env.Eval(fmt.Sprintf(`(initialize-instance [%s])`, inst.Name())); err != nil {
		return nil, err
	}
	*/
	return inst, nil
}

func (inst *Instance) fillSlot(field reflect.StructField, fieldval reflect.Value, knownBases map[reflect.Value]InstanceName) error {
	if field.Anonymous {
		for ii := 0; ii < field.Type.NumField(); ii++ {
			subfield := field.Type.Field(ii)
			subval := fieldval.Field(ii)
			if err := inst.fillSlot(subfield, subval, knownBases); err != nil {
				return err
			}
		}
		return nil
	}

	fieldtype := field.Type
	fielddata := fieldval
	if fieldtype.Kind() == reflect.Ptr {
		if fieldval.IsNil() {
			return inst.SetSlot(slotNameFor(field), InstanceName("nil"))
		}
		fieldtype = fieldtype.Elem()
		fielddata = fielddata.Elem()
	}

	if fieldtype.Kind() != reflect.Struct {
		return inst.SetSlot(slotNameFor(field), fielddata.Interface())
	}
	// may need to recurse
	subinstName, ok := knownBases[fielddata]
	if ok {
		return inst.SetSlot(slotNameFor(field), subinstName)
	}
	subinst, err := inst.env.insertInstance("", fieldval.Interface(), knownBases)
	if err != nil {
		return err
	}
	return inst.SetSlot(slotNameFor(field), subinst.Name())
}
