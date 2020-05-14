package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips/clips.h>
import "C"
import (
	"fmt"
	"reflect"
	"strings"
)

// InsertClass creates a representation of a Go struct as a CLIPS defclass
func (env *Environment) InsertClass(basis interface{}) (*Class, error) {
	typ := reflect.TypeOf(basis)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	classname := typ.Name()
	switch typ.Kind() {
	case reflect.Struct, reflect.Interface:
		// that's what we want
	default:
		return nil, fmt.Errorf(`Unable to insert defclass for type "%s"`, classname)
	}
	if err := env.insertShadowClass(typ); err != nil {
		return nil, err
	}

	if err := env.insertShadowMessages(typ); err != nil {
		return nil, err
	}

	return env.FindClass(classname)
}

func (env *Environment) insertShadowClass(typ reflect.Type) error {
	var defclass strings.Builder
	fmt.Fprintf(&defclass, "(defclass %s (is-a USER)\n", typ.Name())
	for ii := 0; ii < typ.NumField(); ii++ {
		field := typ.Field(ii)

		if err := env.defclassSlots(&defclass, field); err != nil {
			return err
		}
	}
	fmt.Fprint(&defclass, ")")
	return env.Build(defclass.String())
}

func (env *Environment) defclassSlots(defclass *strings.Builder, field reflect.StructField) error {
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
		if err := env.checkRecurseClass(fieldtype); err != nil {
			return err
		}
		fmt.Fprintf(defclass, "    (slot %s (type INSTANCE-NAME) (allowed-classes %s))\n", slotNameFor(field), fieldtype.Name())
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
		if err := env.checkRecurseClass(subtype); err != nil {
			return err
		}
		fmt.Fprintf(defclass, "    (multislot %s (type INSTANCE-NAME) (allowed-classes %s))\n", slotNameFor(field), subtype.Name())
		return nil
	default:
		clipsSubtype = clipsTypeFor(field.Type.Elem()).String()
	}
	fmt.Fprintf(defclass, "    (multislot %s (type %s))\n", slotNameFor(field), clipsSubtype)
	return nil
}

func (env *Environment) checkRecurseClass(fieldtype reflect.Type) error {
	_, err := env.FindClass(fieldtype.Name())
	if err != nil {
		// need to recurse
		if _, err := env.InsertClass(reflect.Zero(reflect.PtrTo(fieldtype)).Interface()); err != nil {
			return err
		}
	}
	return nil
}

func (env *Environment) insertShadowMessages(typ reflect.Type) error {
	return nil
}
