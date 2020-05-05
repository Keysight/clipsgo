package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
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
import (
	"reflect"
	"runtime"
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
	}
}

func (do *DataObject) byRef() *C.struct_dataObject {
	return do.data
}

/*
func (do *DataObject) byVal() {
	return
}
*/

// Value returns the Go value for this data object
func (do *DataObject) Value() interface{} {
	dtype := Type(C.get_data_type(do.data))
	dvalue := C.get_data_value(do.data)

	if dvalue == C.NULL {
		return nil
	}

	return do.goValue(dtype, dvalue)
}

func (do *DataObject) clipsTypeFor(v interface{}) Type {
	if v == nil {
		return SYMBOL
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Bool:
		return SYMBOL
	case reflect.Int:
		return INTEGER
	case reflect.Int32:
		return INTEGER
	case reflect.Int64:
		return INTEGER
	case reflect.Float32:
		return FLOAT
	case reflect.Float64:
		return FLOAT
	case reflect.Array:
		return MULTIFIELD
	case reflect.Slice:
		return MULTIFIELD
	default:
		switch reflect.TypeOf(v).String() {
		case "unsafe.Pointer":
			return EXTERNAL_ADDRESS
		case "clips.Symbol":
			return SYMBOL
		case "clips.InstanceName":
			return INSTANCE_NAME
		case "string":
			return STRING
		case "*clips.ImpliedFact":
			return FACT_ADDRESS
		case "*clips.TemplateFact":
			return FACT_ADDRESS
		case "*clips.Instance":
			return INSTANCE_ADDRESS
		}
	}
	return SYMBOL
}

// SetValue copies the go value into the dataobject
func (do *DataObject) SetValue(value interface{}) {
	var dtype Type
	if do.typ < 0 {
		dtype = do.clipsTypeFor(value)
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
	if v, ok := dvalue.(bool); ok {
		if v {
			vstr := C.CString("TRUE")
			defer C.free(unsafe.Pointer(vstr))
			return C.EnvAddSymbol(do.env.env, vstr)
		}
		vstr := C.CString("FALSE")
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	if v, ok := dvalue.(int); ok {
		return C.EnvAddLong(do.env.env, C.longlong(v))
	}
	if v, ok := dvalue.(int32); ok {
		return C.EnvAddLong(do.env.env, C.longlong(v))
	}
	if v, ok := dvalue.(int64); ok {
		return C.EnvAddLong(do.env.env, C.longlong(v))
	}
	if v, ok := dvalue.(float32); ok {
		return C.EnvAddDouble(do.env.env, C.double(v))
	}
	if v, ok := dvalue.(float64); ok {
		return C.EnvAddDouble(do.env.env, C.double(v))
	}
	if v, ok := dvalue.(unsafe.Pointer); ok {
		return C.EnvAddExternalAddress(do.env.env, v, C.C_POINTER_EXTERNAL_ADDRESS)
	}
	if v, ok := dvalue.(string); ok {
		vstr := C.CString(v)
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	if v, ok := dvalue.(Symbol); ok {
		vstr := C.CString(string(v))
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	if v, ok := dvalue.(InstanceName); ok {
		vstr := C.CString(string(v))
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	if v, ok := dvalue.([]interface{}); ok {
		return do.listToMultifield(v)
	}
	if v, ok := dvalue.(*ImpliedFact); ok {
		return v.factptr
	}
	if v, ok := dvalue.(*TemplateFact); ok {
		return v.factptr
	}
	if v, ok := dvalue.(*Instance); ok {
		return v.instptr
	}
	return nil
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
		C.set_multifield_type(multifield, C.long(i+1), C.short(do.clipsTypeFor(v)))
		C.set_multifield_value(multifield, C.long(i+1), do.clipsValue(v))
	}
	C.set_data_begin(do.data, 1)
	C.set_data_end(do.data, size)
	return ret
}
