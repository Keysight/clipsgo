package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
import "C"
import (
	"unsafe"
)

// TemplateFact is an unordered fact
type TemplateFact struct {
	env     *Environment
	factptr unsafe.Pointer
}

// Template is an unordered fact
type Template interface {
}

// Index returns the index number of this fact within CLIPS
func (f *TemplateFact) Index() int {
	return 0
}

// Inserted returns true if the fact has been asserted. (We adopt here the DROOLS terminology to avoid confusing with Go assert)
func (f *TemplateFact) Inserted() bool {
	return false
}

// Insert asserts the fact (We adopt here the DROOLS convention to avoid confusion with the Go assert)
func (f *TemplateFact) Insert() {

}

// Retract retracts the fact from CLIPS
func (f *TemplateFact) Retract() {

}

// Template returns the template defining this fact
func (f *TemplateFact) Template() Template {
	return nil
}

// String returns a string representation of the fact
func (f *TemplateFact) String() string {
	return ""
}

// Delete drops the reference to the fact in CLIPS. should be called when done with the fact
func (f *TemplateFact) Delete() {

}

// Equals returns true if this fact equals the given fact
func (f *TemplateFact) Equals(Fact) bool {
	return false
}

// Iterator returns a function that can be called to get the next slot for this fact. Will return nil when no more slots remain
func (f *TemplateFact) Iterator() func() (string, interface{}) {
	return nil
}

// Slot returns the value stored in the given slot
func (f *TemplateFact) Slot(string) interface{} {
	return nil
}

/*
class TemplateFact(Fact):
    """An Template Fact or Unordered Fact is a dictionary
    where each slot name is a key.

    """

    def __iter__(self):
        return chain(slot_values(self._env, self._fact, self.template._tpl))

    def __len__(self):
        slots = slot_values(self._env, self._fact, self.template._tpl)

        return len(tuple(slots))

    def __getitem__(self, key):
        slot = slot_value(self._env, self._fact, str(key).encode())

        if slot is not None:
            return slot

        raise KeyError(
            "'%s' fact has not slot '%s'" % (self.template.name, key))

    def __setitem__(self, key, value):
        if self.asserted:
            raise RuntimeError("Fact already asserted")

        data = clips.data.DataObject(self._env)
        data.value = value

        ret = lib.EnvPutFactSlot(
            self._env, self._fact, str(key).encode(), data.byref)
        if ret != 1:
            if key not in (s.name for s in self.template.slots()):
                raise KeyError(
                    "'%s' fact has not slot '%s'" % (self.template.name, key))

            raise CLIPSError(self._env)

    def update(self, sequence=None, **mapping):
        """Add multiple elements to the fact."""
        if sequence is not None:
            if isinstance(sequence, dict):
                for slot in sequence:
                    self[slot] = sequence[slot]
            else:
                for slot, value in sequence:
                    self[slot] = value
        if mapping:
            for slot in sequence:
				self[slot] = sequence[slot]

    """A Fact Template is a formal representation of the fact data structure.

    In CLIPS, Templates are defined via the (deftemplate) statement.

    Templates allow to create new facts
    to be asserted within the CLIPS environment.

    Implied facts are associated to implied templates. Implied templates
    have a limited set of features. For example, they do not support slots.

    """

    __slots__ = '_env', '_tpl'

    def __init__(self, env, tpl):
        self._env = env
        self._tpl = tpl

    def __hash__(self):
        return hash(self._tpl)

    def __eq__(self, tpl):
        return self._tpl == tpl._tpl

    def __str__(self):
        return template_pp_string(self._env, self._tpl)

    def __repr__(self):
        string = template_pp_string(self._env, self._tpl)

        return "%s: %s" % (self.__class__.__name__, string)

    @property
    def name(self):
        """Template name."""
        return ffi.string(
            lib.EnvGetDeftemplateName(self._env, self._tpl)).decode()

    @property
    def module(self):
        """The module in which the Template is defined.

        Python equivalent of the CLIPS deftemplate-module command.

        """
        modname = ffi.string(lib.EnvDeftemplateModule(self._env, self._tpl))
        defmodule = lib.EnvFindDefmodule(self._env, modname)

        return Module(self._env, defmodule)

    @property
    def implied(self):
        """True if the Template is implied."""
        return bool(lib.implied_deftemplate(self._tpl))

    @property
    def watch(self):
        """Whether or not the Template is being watched."""
        return bool(lib.EnvGetDeftemplateWatch(self._env, self._tpl))

    @watch.setter
    def watch(self, flag):
        """Whether or not the Template is being watched."""
        lib.EnvSetDeftemplateWatch(self._env, int(flag), self._tpl)

    @property
    def deletable(self):
        """True if the Template can be deleted."""
        return bool(lib.EnvIsDeftemplateDeletable(self._env, self._tpl))

    def slots(self):
        """Iterate over the Slots of the Template."""
        if self.implied:
            return ()

        data = clips.data.DataObject(self._env)

        lib.EnvDeftemplateSlotNames(self._env, self._tpl, data.byref)

        return tuple(
            TemplateSlot(self._env, self._tpl, n.encode()) for n in data.value)

    def new_fact(self):
        """Create a new Fact from this template."""
        fact = lib.EnvCreateFact(self._env, self._tpl)
        if fact == ffi.NULL:
            raise CLIPSError(self._env)

        return new_fact(self._env, fact)

    def undefine(self):
        """Undefine the Template.

        Python equivalent of the CLIPS undeftemplate command.

        The object becomes unusable after this method has been called.

        """
        if lib.EnvUndeftemplate(self._env, self._tpl) != 1:
            raise CLIPSError(self._env)


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
