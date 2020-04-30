package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
//
// int implied_deftemplate(void*);
import "C"
import (
	"fmt"
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

func createTemplate(env *Environment, tplptr unsafe.Pointer) *Template {
	return &Template{
		env:    env,
		tplptr: tplptr,
	}
}

// Delete deletes this template object
func (t *Template) Delete() {
	// no refcounting for templates
}

// Equals returns true if this template represents the same template as the given one
func (t *Template) Equals(other *Template) bool {
	return t.tplptr == other.tplptr
}

// String returns a string representation of the template
func (t *Template) String() string {
	cstr := C.EnvGetDeftemplatePPForm(t.env.env, t.tplptr)
	if cstr != nil {
		return C.GoString(cstr)
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
func (t *Template) Module() {
	/* TODO
	   @property
	   def module(self):
	       """The module in which the Template is defined.

	       Python equivalent of the CLIPS deftemplate-module command.

	       """
	       modname = ffi.string(lib.EnvDeftemplateModule(self._env, self._tpl))
	       defmodule = lib.EnvFindDefmodule(self._env, modname)

	       return Module(self._env, defmodule)

	*/
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
		namestr, ok := name.(string)
		if !ok {
			panic("Unexpected data returned from CLIPS for slot names")
		}
		ret[namestr] = t.createTemplateSlot(namestr)
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

/*
class TemplateSlot(object):
    """Template Facts organize the information within Slots.

    Slots might restrict the type or amount of data they store.

    """

    __slots__ = '_env', '_tpl', '_name'

    def __init__(self, env, tpl, name):
        self._env = env
        self._tpl = tpl
        self._name = name

    def __hash__(self):
        return hash(self._tpl) + hash(self._name)

    def __eq__(self, slot):
        return self._tpl == slot._tpl and self._name == slot._name

    def __str__(self):
        return self.name

    def __repr__(self):
        return "%s: %s" % (self.__class__.__name__, self.name)

    @property
    def name(self):
        """The slot name."""
        return self._name.decode()

    @property
    def multifield(self):
        """True if the slot is a multifield slot."""
        return bool(lib.EnvDeftemplateSlotMultiP(
            self._env, self._tpl, self._name))

    @property
    def types(self):
        """A tuple containing the value types for this Slot.

        The Python equivalent of the CLIPS deftemplate-slot-types function.

        """
        data = clips.data.DataObject(self._env)

        lib.EnvDeftemplateSlotTypes(
            self._env, self._tpl, self._name, data.byref)

        return tuple(data.value) if isinstance(data.value, list) else ()

    @property
    def range(self):
        """A tuple containing the numeric range for this Slot.

        The Python equivalent of the CLIPS deftemplate-slot-range function.

        """
        data = clips.data.DataObject(self._env)

        lib.EnvDeftemplateSlotRange(
            self._env, self._tpl, self._name, data.byref)

        return tuple(data.value) if isinstance(data.value, list) else ()

    @property
    def cardinality(self):
        """A tuple containing the cardinality for this Slot.

        The Python equivalent
        of the CLIPS deftemplate-slot-cardinality function.

        """
        data = clips.data.DataObject(self._env)

        lib.EnvDeftemplateSlotCardinality(
            self._env, self._tpl, self._name, data.byref)

        return tuple(data.value) if isinstance(data.value, list) else ()

    @property
    def default_type(self):
        """The default value type for this Slot.

        The Python equivalent of the CLIPS deftemplate-slot-defaultp function.

        """
        return TemplateSlotDefaultType(
            lib.EnvDeftemplateSlotDefaultP(self._env, self._tpl, self._name))

    @property
    def default_value(self):
        """The default value for this Slot.

        The Python equivalent
        of the CLIPS deftemplate-slot-default-value function.

        """
        data = clips.data.DataObject(self._env)

        lib.EnvDeftemplateSlotDefaultValue(
            self._env, self._tpl, self._name, data.byref)

        return data.value

    @property
    def allowed_values(self):
        """A tuple containing the allowed values for this Slot.

        The Python equivalent of the CLIPS slot-allowed-values function.

        """
        data = clips.data.DataObject(self._env)

        lib.EnvDeftemplateSlotAllowedValues(
            self._env, self._tpl, self._name, data.byref)

        return tuple(data.value) if isinstance(data.value, list) else ()
*/
