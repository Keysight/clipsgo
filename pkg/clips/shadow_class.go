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
	"reflect"
	"strings"
)

// InsertClassOption tweaks how the inserted class is instructed
type InsertClassOption string

const (
	// DoNotRestrictAllowedClasses prevents the class insertion from using an allowed-class constraint for instance-name slots. Primarily useful if [nil] must be allowed
	DoNotRestrictAllowedClasses InsertClassOption = "DoNotRestrictAllowedClasses"
)

// InsertClass creates a representation of a Go struct as a CLIPS defclass
func (env *Environment) InsertClass(basis interface{}, opts ...InsertClassOption) (*Class, error) {
	typ := reflect.TypeOf(basis)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	classname, err := classNameFor(typ)
	if err != nil {
		return nil, err
	}
	cls, err := env.FindClass(classname)
	if err == nil {
		return cls, fmt.Errorf("Class %s already exists", classname)
	}
	switch typ.Kind() {
	case reflect.Struct, reflect.Interface:
		// that's what we want
	default:
		return nil, fmt.Errorf(`Unable to insert defclass for type "%s"`, typ.String())
	}
	if err := env.insertShadowClass(classname, typ, opts...); err != nil {
		return nil, err
	}

	/*
		if err := env.insertShadowMessages(typ); err != nil {
			return nil, err
		}
	*/

	return env.FindClass(classname)
}

func classNameFor(typ reflect.Type) (string, error) {
	classname := clipsSymbolEscape(typ.Name())
	if classname == "" {
		classname = clipsSymbolEscape(typ.String())
	}
	if classname == "" {
		return "", fmt.Errorf("Unable to insert unnamed class %v", typ)
	}
	return classname, nil
}

func (env *Environment) insertShadowClass(classname string, typ reflect.Type, opts ...InsertClassOption) error {
	// first, build effectively a forward declaration, so we don't get into
	// infinite recursion if some field references this class. That lookup
	// will succeed.
	if err := env.Build(fmt.Sprintf(`(defclass %s (is-a USER))`, classname)); err != nil {
		return err
	}
	// Now, we'll override with a full definition
	var defclass strings.Builder
	fmt.Fprintf(&defclass, "(defclass %s (is-a USER)\n", classname)
	for ii := 0; ii < typ.NumField(); ii++ {
		field := typ.Field(ii)

		if err := env.defclassSlots(&defclass, field, opts...); err != nil {
			return err
		}
	}
	fmt.Fprint(&defclass, ")")
	buildcmd := defclass.String()
	return env.Build(buildcmd)
}

func (env *Environment) defclassSlots(defclass *strings.Builder, field reflect.StructField, opts ...InsertClassOption) error {
	if field.Anonymous {
		// treat fields of the anonymous class just like they are native
		for ii := 0; ii < field.Type.NumField(); ii++ {
			if err := env.defclassSlots(defclass, field.Type.Field(ii)); err != nil {
				return err
			}
		}
		return nil
	}
	fieldtype := field.Type
	if fieldtype.Kind() == reflect.Ptr {
		fieldtype = fieldtype.Elem()
	}
	switch fieldtype.Kind() {
	case reflect.Interface:
		fmt.Fprintf(defclass, "    (slot %s (type ?VARIABLE))\n", slotNameFor(field))
		return nil
	case reflect.Struct:
		classname, err := classNameFor(fieldtype)
		if err != nil {
			return err
		}
		if _, err = env.checkRecurseClass(classname, fieldtype); err != nil {
			return err
		}
		allowed := fmt.Sprintf(" (allowed-classes %s)", classname)
		for _, v := range opts {
			switch v {
			case DoNotRestrictAllowedClasses:
				allowed = ""
			}
		}
		fmt.Fprintf(defclass, "    (slot %s (type INSTANCE-NAME)%s)\n", slotNameFor(field), allowed)
		return nil
	case reflect.Array, reflect.Slice:
		// don't handle here
	default:
		fmt.Fprintf(defclass, "    (slot %s (type %s))\n", slotNameFor(field), clipsTypeFor(fieldtype))
		return nil
	}

	subtype := fieldtype.Elem()
	if subtype.Kind() == reflect.Ptr {
		subtype = subtype.Elem()
	}
	var clipsSubtype string
	switch subtype.Kind() {
	case reflect.Array, reflect.Slice:
		return fmt.Errorf(`Unable to represent type for field "%s"`, field.Name)
	case reflect.Interface:
		clipsSubtype = "?VARIABLE"
	case reflect.Struct:
		classname, err := classNameFor(subtype)
		if err != nil {
			return err
		}
		if _, err = env.checkRecurseClass(classname, subtype); err != nil {
			return err
		}
		fmt.Fprintf(defclass, "    (multislot %s (type INSTANCE-NAME) (allowed-classes %s))\n", slotNameFor(field), subtype.Name())
		return nil
	default:
		clipsSubtype = clipsTypeFor(subtype).String()
	}
	fmt.Fprintf(defclass, "    (multislot %s (type %s))\n", slotNameFor(field), clipsSubtype)
	return nil
}

func (env *Environment) checkRecurseClass(classname string, fieldtype reflect.Type, opts ...InsertClassOption) (*Class, error) {
	cls, err := env.FindClass(classname)
	if err != nil {
		if _, ok := err.(NotFoundError); !ok {
			return nil, err
		}
		// need to recurse
		if cls, err = env.InsertClass(reflect.Zero(reflect.PtrTo(fieldtype)).Interface(), opts...); err != nil {
			return nil, err
		}
	}
	return cls, nil
}
