package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
import "C"
import (
	"unsafe"
)

// ImpliedFact is an ordered fact having an implied definition
type ImpliedFact struct {
	env     *Environment
	factptr unsafe.Pointer
}

// Index returns the index number of this fact within CLIPS
func (f *ImpliedFact) Index() int {
	return 0
}

// Inserted returns true if the fact has been asserted. (We adopt here the DROOLS terminology to avoid confusing with Go assert)
func (f *ImpliedFact) Inserted() bool {
	return false
}

// Insert asserts the fact (We adopt here the DROOLS convention to avoid confusion with the Go assert)
func (f *ImpliedFact) Insert() {
}

// Retract retracts the fact from CLIPS
func (f *ImpliedFact) Retract() {

}

// Template returns the template defining this fact
func (f *ImpliedFact) Template() Template {
	return nil
}

// String returns a string representation of the fact
func (f *ImpliedFact) String() string {
	return ""
}

// Delete drops the reference to the fact in CLIPS. should be called when done with the fact
func (f *ImpliedFact) Delete() {

}

// Equals returns true if this fact equals the given fact
func (f *ImpliedFact) Equals(Fact) bool {
	return false
}

// Iterator returns a function that can be called to get the next slot for this fact. Will return nil when no more slots remain
func (f *ImpliedFact) Iterator() func() interface{} {
	return nil
}

// Slot returns the value stored in the given slot
func (f *ImpliedFact) Slot(int) interface{} {
	return nil
}

/*
class ImpliedFact(Fact):
	"""An Implied Fact or Ordered Fact represents its data as a list of elements
	    similarly as for a Multifield.

    """

    __slots__ = '_env', '_fact', '_multifield'

    def __init__(self, env, fact):
        super(ImpliedFact, self).__init__(env, fact)
        self._multifield = []

    def __iter__(self):
        return chain(slot_value(self._env, self._fact, None))

    def __len__(self):
        return len(slot_value(self._env, self._fact, None))

    def __getitem__(self, item):
        return tuple(self)[item]

    def append(self, value):
        """Append an element to the fact."""
        if self.asserted:
            raise RuntimeError("Fact already asserted")

        self._multifield.append(value)

    def extend(self, values):
        """Append multiple elements to the fact."""
        if self.asserted:
            raise RuntimeError("Fact already asserted")

        self._multifield.extend(values)

    def assertit(self):
        """Assert the fact within CLIPS."""
        data = clips.data.DataObject(self._env)
        data.value = list(self._multifield)

        if lib.EnvPutFactSlot(
                self._env, self._fact, ffi.NULL, data.byref) != 1:
            raise CLIPSError(self._env)

		super(ImpliedFact, self).assertit()
*/
