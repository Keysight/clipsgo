package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
//
// struct dataObject *data_object_ptr(void *mallocval) {
// 	 return (struct dataObject *)mallocval;
// }
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
// short set_data_type(struct dataObject *data, short type)
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

// TYPES maps from Go datatypes to corresponding CLIPS representation
var TYPES = map[reflect.Kind]C.short{
	reflect.Bool:    SYMBOL,
	reflect.Int:     INTEGER,
	reflect.Float32: FLOAT,
	reflect.Float64: FLOAT,
	reflect.String:  STRING,
	reflect.Array:   MULTIFIELD,
	reflect.Slice:   MULTIFIELD,
	/*
		clips.common.Symbol: clips.common.CLIPSType.SYMBOL,
		clips.facts.ImpliedFact: clips.common.CLIPSType.FACT_ADDRESS,
		clips.facts.TemplateFact: clips.common.CLIPSType.FACT_ADDRESS,
		clips.classes.Instance: clips.common.CLIPSType.INSTANCE_ADDRESS,
		clips.common.InstanceName: clips.common.CLIPSType.INSTANCE_NAME}
	*/
}

// Symbol represents a CLIPS SYMBOL value
type Symbol string

// InstanceName represents a CLIPS INSTANCE_NAME value
type InstanceName Symbol

// DataObject wraps a CLIPS data object
type DataObject struct {
	env  *Environment
	typ  C.short
	data *C.struct_dataObject
}

func createDataObject(env *Environment) *DataObject {
	datamem := C.malloc(C.sizeof_struct_dataObject)
	data := C.data_object_ptr(datamem)
	ret := &DataObject{
		env:  env,
		typ:  -1,
		data: data,
	}
	runtime.SetFinalizer(ret, func(data *DataObject) {
		data.Close()
	})
	return ret
}

// Close frees up associated memory
func (do *DataObject) Close() {
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
	dtype := C.get_data_type(do.data)
	dvalue := C.get_data_value(do.data)

	if dvalue == C.NULL {
		return nil
	}

	return do.goValue(dtype, dvalue)
}

func (do *DataObject) setValue(value interface{}) {
	var dtype C.short
	if do.typ < 0 {
		dtype = TYPES[reflect.TypeOf(value).Kind()]
	} else {
		dtype = do.typ
	}

	C.set_data_type(do.data, dtype)
	C.set_data_value(do.data, do.clipsValue(value))
}

// goValue converts a CLIPS data value into a Go data structure
func (do *DataObject) goValue(dtype C.short, dvalue unsafe.Pointer) interface{} {
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
		return Symbol(C.GoString(C.to_string(dvalue)))
	case INSTANCE_NAME:
		return InstanceName(C.GoString(C.to_string(dvalue)))
	case MULTIFIELD:
		return do.multifieldToList()
		/*
			case FACT_ADDRESS:
				return clips.facts.new_fact(self._env, lib.to_pointer(dvalue))
			case INSTANCE_ADDRESS:
				return clips.classes.Instance(self._env, lib.to_pointer(dvalue))
		*/
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
	if v, ok := dvalue.(int64); ok {
		return C.EnvAddLong(do.env.env, C.longlong(v))
	}
	if v, ok := dvalue.(float64); ok {
		return C.EnvAddDouble(do.env.env, C.double(v))
	}
	// ffi.CData?
	if v, ok := dvalue.(unsafe.Pointer); ok {
		return v
	}
	if v, ok := dvalue.(string); ok {
		vstr := C.CString(v)
		defer C.free(unsafe.Pointer(vstr))
		return C.EnvAddSymbol(do.env.env, vstr)
	}
	if v, ok := dvalue.([]interface{}); ok {
		return do.listToMultifield(v)
	}
	/*
		if isinstance(dvalue, (clips.facts.Fact)):
			return dvalue._fact
		if isinstance(dvalue, (clips.classes.Instance)):
			return dvalue._ist
	*/
	return nil
}

func (do *DataObject) multifieldToList() []interface{} {
	end := C.get_data_end(do.data)
	begin := C.get_data_begin(do.data)
	multifield := C.multifield_ptr(C.get_data_value(do.data))

	ret := make([]interface{}, 0, end-begin+1)
	for i := begin; i <= end; i++ {
		dtype := C.get_multifield_type(multifield, i)
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
		C.set_multifield_type(multifield, C.long(i+1), TYPES[reflect.TypeOf(v).Kind()])
		C.set_multifield_value(multifield, C.long(i+1), do.clipsValue(v))
	}
	C.set_data_begin(do.data, 1)
	C.set_data_end(do.data, size)
	return ret
}
