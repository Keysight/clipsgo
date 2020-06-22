package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
//
// struct multifield *multifield_ptr(void *mallocval) {
// 	 return (struct multifield *)mallocval;
// }
//
// short get_data_type(struct dataObject *data )
// {
//	 return GetpType(data);
// }
//
// short set_data_type(struct dataObject *data, int type)
// {
//   return SetpType(data, type);
// }
//
// void *get_data_value(struct dataObject *data)
// {
//   return GetpValue(data);
// }
//
// void *set_data_value(struct dataObject *data, void *value)
// {
//   return SetpValue(data, value);
// }
//
// long get_data_begin(struct dataObject *data)
// {
//   return GetpDOBegin(data);
// }
//
// long set_data_begin(struct dataObject *data, long begin)
// {
//   return SetpDOBegin(data, begin);
// }
//
// long get_data_end(struct dataObject *data)
// {
//   return GetpDOEnd(data);
// }
//
// long set_data_end(struct dataObject *data, long end)
// {
//   return SetpDOEnd(data, end);
// }
//
// long get_data_length(struct dataObject *data)
// {
//   return GetpDOLength(data);
// }
//
// short get_multifield_type(struct multifield *mf, long index)
// {
//   return GetMFType(mf, index);
// }
//
// short set_multifield_type(struct multifield *mf, long index, short type)
// {
//   return SetMFType(mf, index, type);
// }
//
// void *get_multifield_value(struct multifield *mf, long index)
// {
//   return GetMFValue(mf, index);
// }
//
// void *set_multifield_value(struct multifield *mf, long index, void *value)
// {
//   return SetMFValue(mf, index, value);
// }
//
// long get_multifield_length(struct multifield *mf)
// {
//   return GetMFLength(mf);
// }
//
// char *to_string(void *data)
// {
//   return (char *) ValueToString(data);
// }
//
// long long to_integer(void *data)
// {
//   return ValueToLong(data);
// }
//
// double to_double(void *data)
// {
//   return ValueToDouble(data);
// }
//
// void *to_pointer(void *data)
// {
//   return ValueToPointer(data);
// }
//
// void *to_external_address(void *data)
// {
//   return ValueToExternalAddress(data);
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
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unsafe"
)

// Symbol represents a CLIPS SYMBOL value
type Symbol string

// InstanceName represents a CLIPS INSTANCE_NAME value
type InstanceName Symbol

// DataObject wraps a CLIPS data object
type DataObject struct {
	env  *Environment
	typ  Type
	data *C.struct_dataObject
}

func createDataObjectInitialized(env *Environment, data *C.struct_dataObject) *DataObject {
	ret := &DataObject{
		env:  env,
		typ:  -1,
		data: data,
	}
	return ret
}

func createDataObject(env *Environment) *DataObject {
	datamem := C.malloc(C.sizeof_struct_dataObject)
	data := (*C.struct_dataObject)(datamem)
	ret := createDataObjectInitialized(env, data)
	runtime.SetFinalizer(ret, func(data *DataObject) {
		data.Delete()
	})
	return ret
}

// Delete frees up associated memory
func (do *DataObject) Delete() {
	if do.data != nil {
		C.free(unsafe.Pointer(do.data))
		do.data = nil
	}
}

func (do *DataObject) byRef() *C.struct_dataObject {
	return do.data
}

// Value returns the Go value for this data object
func (do *DataObject) Value() interface{} {
	dtype := Type(C.get_data_type(do.data))
	dvalue := C.get_data_value(do.data)

	if dvalue == C.NULL {
		return nil
	}

	return do.goValue(dtype, dvalue)
}

func clipsTypeFor(typ reflect.Type) Type {
	if typ == nil {
		return SYMBOL
	}
	switch typ.Kind() {
	case reflect.Bool:
		return SYMBOL
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return INTEGER
	case reflect.Float32, reflect.Float64:
		return FLOAT
	case reflect.Array, reflect.Slice:
		return MULTIFIELD
	case reflect.Struct:
		return INSTANCE_NAME
	default:
		switch typ {
		case reflect.TypeOf(""):
			return STRING
		case reflect.TypeOf((*Symbol)(nil)).Elem():
			return SYMBOL
		case reflect.TypeOf(unsafe.Pointer(nil)):
			return EXTERNAL_ADDRESS
		case reflect.TypeOf((*InstanceName)(nil)).Elem():
			return INSTANCE_NAME
		case reflect.TypeOf((*ImpliedFact)(nil)), reflect.TypeOf((*TemplateFact)(nil)):
			return FACT_ADDRESS
		case reflect.TypeOf((*Instance)(nil)):
			return INSTANCE_ADDRESS
		}
	}
	if typ.Kind() == reflect.String {
		// If we got here, it's a string but not one of our special cases (Symbol, InstanceName)
		return STRING
	}
	return SYMBOL
}

// SetValue copies the go value into the dataobject
func (do *DataObject) SetValue(value interface{}) {
	var dtype Type
	if do.typ < 0 {
		dtype = clipsTypeFor(reflect.TypeOf(value))
	} else {
		dtype = do.typ
	}

	C.set_data_type(do.data, dtype.CVal())
	C.set_data_value(do.data, do.clipsValue(value))
}

// goValue converts a CLIPS data value into a Go data structure
func (do *DataObject) goValue(dtype Type, dvalue unsafe.Pointer) interface{} {
	switch dtype {
	case FLOAT:
		return float64(C.to_double(dvalue))
	case INTEGER:
		return int64(C.to_integer(dvalue))
	case STRING:
		return C.GoString(C.to_string(dvalue))
	case EXTERNAL_ADDRESS:
		return C.to_external_address(dvalue)
	case SYMBOL:
		cstr := C.to_string(dvalue)
		gstr := C.GoString(cstr)
		if gstr == "nil" {
			return nil
		}
		if gstr == "TRUE" {
			return true
		}
		if gstr == "FALSE" {
			return false
		}
		return Symbol(gstr)
	case INSTANCE_NAME:
		return InstanceName(C.GoString(C.to_string(dvalue)))
	case MULTIFIELD:
		return do.multifieldToList()
	case FACT_ADDRESS:
		return do.env.newFact(C.to_pointer(dvalue))
	case INSTANCE_ADDRESS:
		return createInstance(do.env, C.to_pointer(dvalue))
	}
	return nil
}

// clipsValue convers a Go data structure into a CLIPS data value
func (do *DataObject) clipsValue(dvalue interface{}) unsafe.Pointer {
	if dvalue == nil {
		vstr := C.CString("nil")
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	switch v := dvalue.(type) {
	case unsafe.Pointer:
		return C.EnvAddExternalAddress(do.env.env, v, C.C_POINTER_EXTERNAL_ADDRESS)
	case Symbol:
		vstr := C.CString(string(v))
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	case InstanceName:
		vstr := C.CString(string(v))
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	case []interface{}:
		return do.listToMultifield(v)
	case *ImpliedFact:
		return v.factptr
	case *TemplateFact:
		return v.factptr
	case *Instance:
		return v.instptr
	}
	val := reflect.ValueOf(dvalue)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// We use Kind() in case of user subtypes, so they are mapped to their base type
	switch val.Kind() {
	case reflect.Bool:
		v := val.Bool()
		if v {
			vstr := C.CString("TRUE")
			defer C.free(unsafe.Pointer(vstr))
			return C.EnvAddSymbol(do.env.env, vstr)
		}
		vstr := C.CString("FALSE")
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := val.Int()
		return C.EnvAddLong(do.env.env, C.longlong(v))
	case reflect.Float32, reflect.Float64:
		v := val.Float()
		return C.EnvAddDouble(do.env.env, C.double(v))
	case reflect.Slice, reflect.Array:
		mvalue := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			mvalue[i] = val.Index(i).Interface()
		}
		return do.listToMultifield(mvalue)
	case reflect.String:
		v := val.String()
		vstr := C.CString(v)
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	case reflect.Struct:
		// need to insert the struct first, then store its INSTANCE-NAME
		instname := "nil"
		if dvalue != nil {
			subinst, err := do.env.Insert("", dvalue)
			if err != nil {
				panic(err)
			}
			instname = string(subinst.Name())
		}
		vstr := C.CString(instname)
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	// Fall back to FALSE in typical CLIPS style
	vstr := C.CString("FALSE")
	defer C.free(unsafe.Pointer(vstr))
	return C.EnvAddSymbol(do.env.env, vstr)
}

func (do *DataObject) multifieldToList() []interface{} {
	end := C.get_data_end(do.data)
	begin := C.get_data_begin(do.data)
	multifield := C.multifield_ptr(C.get_data_value(do.data))

	ret := make([]interface{}, 0, end-begin+1)
	for i := begin; i <= end; i++ {
		dtype := Type(C.get_multifield_type(multifield, i))
		dvalue := C.get_multifield_value(multifield, i)
		ret = append(ret, do.goValue(dtype, dvalue))
	}
	return ret
}

func (do *DataObject) listToMultifield(values []interface{}) unsafe.Pointer {
	size := C.long(len(values))
	ret := C.EnvCreateMultifield(do.env.env, size)
	multifield := C.multifield_ptr(ret)
	for i, v := range values {
		C.set_multifield_type(multifield, C.long(i+1), C.short(clipsTypeFor(reflect.TypeOf(v))))
		C.set_multifield_value(multifield, C.long(i+1), do.clipsValue(v))
	}
	C.set_data_begin(do.data, 1)
	C.set_data_end(do.data, size)
	return ret
}

// ExtractValue attempts to put the represented data value into the item provided by the user.
func (do *DataObject) ExtractValue(retval interface{}, extractClasses bool) error {
	directType := reflect.TypeOf(retval)
	if directType.Kind() != reflect.Ptr {
		return fmt.Errorf("retval must be a pointer to the value to be filled in")
	}
	val := do.Value()
	knownInstances := make(map[InstanceName]interface{})
	return do.env.convertArg(reflect.ValueOf(retval), reflect.ValueOf(val), extractClasses, knownInstances)
}

// MustExtractValue attempts to put the represented data value into the item provided by the user, and panics if it can't
func (do *DataObject) MustExtractValue(retval interface{}, extractClasses bool) {
	if err := do.ExtractValue(retval, extractClasses); err != nil {
		panic(err)
	}
}

func safeIndirect(output reflect.Value) reflect.Value {
	switch output.Type() {
	// pointers to our own Instance, ImpliedFact, or TemplateFact don't need to use Indirect()
	case reflect.TypeOf((*Instance)(nil)),
		reflect.TypeOf((*ImpliedFact)(nil)),
		reflect.TypeOf((*TemplateFact)(nil)):
		return output
	}
	// pointers to user data need Indirect, to get a valid value
	val := reflect.Indirect(output)
	if !val.IsValid() {
		val = reflect.New(output.Type().Elem())
		output.Set(val)
		val = val.Elem()
	}
	return val
}

func (env *Environment) convertArg(output reflect.Value, data reflect.Value, extractClasses bool, knownInstances map[InstanceName]interface{}) error {
	val := safeIndirect(output)

	if extractClasses && data.IsValid() {
		if output.Type() != reflect.TypeOf(InstanceName("")) && output.Type() != reflect.TypeOf((*Instance)(nil)) {
			dif := data.Interface()
			var subinst *Instance
			var err error
			instname, ok := dif.(InstanceName)
			if ok {
				if instname == "nil" {
					// it's not an instance we can look up, it's just nil
					subinst = nil
				} else {
					subinst, err = env.FindInstance(instname, "")
					if err != nil {
						return err
					}
				}
			} else {
				subinst, ok = dif.(*Instance)
			}
			if ok {
				if subinst == nil {
					data = reflect.ValueOf(nil)
				} else {
					if knownVal, ok := knownInstances[subinst.Name()]; ok {
						// This implies a circular recursive reference
						val.Set(reflect.ValueOf(knownVal).Elem())
						return nil
					} else {
						// extract the instance
						slots := subinst.Slots(true)
						knownInstances[subinst.Name()] = val.Addr().Interface()
						return subinst.env.structuredExtract(val.Addr().Interface(), slots, true, knownInstances)
					}
				}
			}
		}
	}

	if !data.IsValid() {
		switch val.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice:
			val.Set(reflect.Zero(val.Type()))
			return nil
		default:
			return fmt.Errorf("Unable to convert nil value to non-pointer type %v", val.Type())
		}
	}

	checktype := val.Type()
	if val.Kind() == reflect.Ptr && data.Kind() != reflect.Ptr {
		checktype = checktype.Elem()
	}

	if data.Type().AssignableTo(checktype) {
		if val.Kind() == reflect.Ptr && data.Kind() != reflect.Ptr {
			val = safeIndirect(val)
		}
		val.Set(data)
		return nil
	}

	if data.Kind() == reflect.Int64 {
		// Make an exception when it's just loss of scale, and make it work
		val = safeIndirect(val)
		intval := data.Int()
		var checkval int64
		switch val.Type().Kind() {
		case reflect.Int64:
			// no check needed
		case reflect.Int:
			checkval = int64(int(intval))
		case reflect.Int32:
			checkval = int64(int32(intval))
		case reflect.Int16:
			checkval = int64(int16(intval))
		case reflect.Int8:
			checkval = int64(int8(intval))
		case reflect.Uint:
			checkval = int64(uint(intval))
		case reflect.Uint64:
			checkval = int64(uint64(intval))
		case reflect.Uint32:
			checkval = int64(uint32(intval))
		case reflect.Uint16:
			checkval = int64(uint16(intval))
		case reflect.Uint8:
			checkval = int64(uint8(intval))
		default:
			return fmt.Errorf(`Invalid type "%v", expected "%v"`, data.Type(), val.Type())
		}
		if checkval != intval {
			return fmt.Errorf(`Integer %d too large`, intval)
		}
		val.SetInt(checkval)
		return nil
	} else if data.Kind() == reflect.Float64 {
		val = safeIndirect(val)
		floatval := data.Float()
		switch val.Type().Kind() {
		case reflect.Float64:
			// no check needed
		case reflect.Float32:
			if float64(float32(floatval)) != floatval {
				return fmt.Errorf(`Floating point %f too precise to represent`, floatval)
			}
		default:
			return fmt.Errorf(`Invalid type "%v", expected "%v"`, data.Type(), val.Type())
		}
		val.SetFloat(floatval)
		return nil
	} else if data.Kind() == reflect.Slice {
		sliceval := val
		slicetype := val.Type()
		var mustSet bool
		if slicetype.Kind() == reflect.Ptr {
			// if we were handed a pointer, make sure it points to something and then set that thing
			sliceval = sliceval.Elem()
			slicetype = slicetype.Elem()
		}
		if slicetype.Kind() != reflect.Slice {
			return fmt.Errorf(`Invalid type "%v", expected "%v"`, data.Type(), val.Type())
		}
		if !sliceval.IsValid() || sliceval.Cap() < data.Len() {
			mustSet = true
			sliceval = reflect.MakeSlice(reflect.SliceOf(slicetype.Elem()), data.Len(), data.Len())
		}
		// see if we can translate to right kind of slice
		for i := 0; i < data.Len(); i++ {
			// what we get from CLIPS is always []interface{}, so there's no check for that here
			indexval := data.Index(i).Elem()
			if err := env.convertArg(sliceval.Index(i), indexval, extractClasses, knownInstances); err != nil {
				return err
			}
		}
		if data.Len() > 0 && mustSet {
			if val.Kind() == reflect.Ptr {
				val.Set(reflect.New(sliceval.Type()))
				val.Elem().Set(sliceval)
			} else {
				val.Set(sliceval)
			}
		}
		return nil
	} else if data.Type().ConvertibleTo(checktype) {
		if val.Kind() == reflect.Ptr && data.Kind() != reflect.Ptr {
			val = safeIndirect(val)
		}
		// This could actually handle ints and floats, too, except it hides wraparound and loss of precision
		converted := data.Convert(val.Type())
		val.Set(converted)
		return nil
	}
	return fmt.Errorf(`Invalid type "%v", expected "%v"`, data.Type(), val.Type())
}

func (env *Environment) structuredExtract(retval interface{}, slots map[string]interface{}, extractClasses bool, knownInstances map[InstanceName]interface{}) error {
	ptr := reflect.ValueOf(retval)
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("Unable to store data to non-pointer type")
	}
	if ptr.IsNil() || !ptr.IsValid() {
		return fmt.Errorf("Unable to store data to nil value")
	}

	val := reflect.Indirect(ptr.Elem())
	if !val.IsValid() {
		val = reflect.New(ptr.Type().Elem().Elem())
		ptr.Elem().Set(val)
		val = val.Elem()
	}
	typ := val.Type()

	switch val.Kind() {
	case reflect.Interface:
		val = reflect.ValueOf(make(map[string]interface{}))
		ptr.Elem().Set(val)
		typ = val.Type()
		fallthrough
	case reflect.Map:
		if typ.Key().Kind() != reflect.String {
			return fmt.Errorf("Key type must be type string")
		}
		if val.IsNil() {
			val = reflect.MakeMap(reflect.MapOf(typ.Key(), typ.Elem()))
			ptr.Elem().Set(val)
		}
		for k, v := range slots {
			newval := reflect.Indirect(reflect.New(typ.Elem()))
			if err := env.convertArg(newval, reflect.ValueOf(v), extractClasses, knownInstances); err != nil {
				return err
			}
			val.SetMapIndex(reflect.ValueOf(k), newval)
		}
	case reflect.Struct:
		for ii := 0; ii < typ.NumField(); ii++ {
			field := typ.Field(ii)
			fieldval := val.Field(ii)

			if err := env.fillStruct(fieldval, field, slots, extractClasses, knownInstances); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unable to extract CLIPS instance to %v", typ.String())
	}

	return nil
}

func (env *Environment) fillStruct(fieldval reflect.Value, field reflect.StructField, slots map[string]interface{}, extractClasses bool, knownInstances map[InstanceName]interface{}) error {
	if field.Anonymous {
		embedType := fieldval.Type()
		// treat fields of the anonymous class just like they are native
		for ii := 0; ii < fieldval.NumField(); ii++ {
			if err := env.fillStruct(fieldval.Field(ii), embedType.Field(ii), slots, extractClasses, knownInstances); err != nil {
				return err
			}
		}
	}

	fielddata, ok := slots[slotNameFor(field)]
	if !ok {
		return nil
	}
	return env.convertArg(fieldval.Addr(), reflect.ValueOf(fielddata), extractClasses, knownInstances)
}

// decide the CLIPS slot name based on tag
func slotNameFor(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup("clips"); ok {
		return tag
	}
	var ret = field.Name
	if tag, ok := field.Tag.Lookup("json"); ok {
		ret = strings.Split(tag, ",")[0]
	}
	if ret == "name" {
		ret = "_name"
	}
	return ret
}
