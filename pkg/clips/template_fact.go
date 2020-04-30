package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
import "C"
import (
	"runtime"
	"unsafe"
)

// TemplateFact is an unordered fact
type TemplateFact struct {
	env     *Environment
	factptr unsafe.Pointer
}

func createTemplateFact(env *Environment, factptr unsafe.Pointer) *TemplateFact {
	ret := &TemplateFact{
		env:     env,
		factptr: factptr,
	}
	runtime.SetFinalizer(ret, func(*TemplateFact) {
		ret.Delete()
	})
	return ret
}

// Delete drops the reference to the fact in CLIPS. should be called when done with the fact
func (f *TemplateFact) Delete() {
	if f.factptr != nil {
		C.EnvDecrementFactCount(f.env.env, f.factptr)
		f.factptr = nil
	}
}

// Index returns the index number of this fact within CLIPS
func (f *TemplateFact) Index() int {
	return 0
}

// Asserted returns true if the fact has been asserted.
func (f *TemplateFact) Asserted() bool {
	return false
}

// Assert asserts the fact
func (f *TemplateFact) Assert() error {
	return nil
}

// Retract retracts the fact from CLIPS
func (f *TemplateFact) Retract() error {
	return nil
}

// Template returns the template defining this fact
func (f *TemplateFact) Template() *Template {
	return nil
}

// String returns a string representation of the fact
func (f *TemplateFact) String() string {
	return ""
}

// Equals returns true if this fact equals the given fact
func (f *TemplateFact) Equals(Fact) bool {
	return false
}

// Slots returns a function that can be called to get the next slot for this fact. Will return nil when no more slots remain
func (f *TemplateFact) Slots() (map[string]interface{}, error) {
	/*
		data := createDataObject(f.env)
		defer data.Delete()
		   lib.EnvDeftemplateSlotNames(env, tpl, data.byref)

		   return ((s, slot_value(env, fact, s.encode())) for s in data.value)
	*/
	return nil, nil
}

// Slot returns the value stored in the given slot
func (f *TemplateFact) Slot(string) (interface{}, error) {
	return nil, nil
}

/*

class Fact(object):
    """CLIPS Fact base class."""

    __slots__ = '_env', '_fact'

    def __init__(self, env, fact):
        self._env = env
        self._fact = fact
        lib.EnvIncrementFactCount(self._env, self._fact)

    def __del__(self):
        try:
            lib.EnvDecrementFactCount(self._env, self._fact)
        except (AttributeError, TypeError):
            pass  # mostly happening during interpreter shutdown

    def __hash__(self):
        return hash(self._fact)

    def __eq__(self, fact):
        return self._fact == fact._fact

    def __str__(self):
        string = fact_pp_string(self._env, self._fact)

        return string.split('     ', 1)[-1]

    def __repr__(self):
        return "%s: %s" % (
            self.__class__.__name__, fact_pp_string(self._env, self._fact))

    @property
    def index(self):
        """The fact index."""
        return lib.EnvFactIndex(self._env, self._fact)

    @property
    def asserted(self):
        """True if the fact has been asserted within CLIPS."""
        # https://sourceforge.net/p/clipsrules/discussion/776945/thread/4f04bb9e/
        if self.index == 0:
            return False

        return bool(lib.EnvFactExistp(self._env, self._fact))

    @property
    def template(self):
        """The associated Template."""
        return Template(
            self._env, lib.EnvFactDeftemplate(self._env, self._fact))

    def assertit(self):
        """Assert the fact within the CLIPS environment."""
        if self.asserted:
            raise RuntimeError("Fact already asserted")

        lib.EnvAssignFactSlotDefaults(self._env, self._fact)

        if lib.EnvAssert(self._env, self._fact) == ffi.NULL:
            raise CLIPSError(self._env)

    def retract(self):
        """Retract the fact from the CLIPS environment."""
        if lib.EnvRetract(self._env, self._fact) != 1:
            raise CLIPSError(self._env)

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

*/
