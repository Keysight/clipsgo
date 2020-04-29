package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips
// #include <clips.h>
//
// int implied_deftemplate(void *template)
// {
//   return ((struct deftemplate*)template)->implied;
// }
import "C"
import (
	"runtime"
	"unsafe"
)

// Fact represents a fact within CLIPS
type Fact interface {
	// Index returns the index number of this fact within CLIPS
	Index() int

	// Inserted returns true if the fact has been asserted. (We adopt here the DROOLS terminology to avoid confusing with Go assert)
	Inserted() bool

	// Insert asserts the fact (We adopt here the DROOLS convention to avoid confusion with the Go assert)
	Insert()

	// Retract retracts the fact from CLIPS
	Retract()

	// Template returns the template defining this fact
	Template() Template

	// String returns a string representation of the fact
	String() string

	// Delete drops the reference to the fact in CLIPS. should be called when done with the fact
	Delete()

	// Equals returns true if this fact equals the given fact
	Equals(Fact) bool
}

// FactIterator returns a function that can be called to iterate over all facts known to CLIPS
func (env *Environment) FactIterator() func() Fact {
	var started bool
	var factptr unsafe.Pointer

	retfun := func() Fact {
		if factptr != nil {
			C.EnvDecrementFactCount(env.env, factptr)
		}
		if factptr != nil || !started {
			factptr = C.EnvGetNextFact(env.env, factptr)
			started = true
		}
		if factptr != nil {
			C.EnvIncrementFactCount(env.env, factptr)
			return env.newFact(factptr)
		}
		return nil
	}
	runtime.SetFinalizer(&retfun, func(fun *func() Fact) {
		if factptr != nil {
			C.EnvDecrementFactCount(env.env, factptr)
			factptr = nil
		}
	})
	return retfun
}

// TemplateIterator returns a function that can be called to iterate over all facts known to CLIPS
func (env *Environment) TemplateIterator() func() Template {
	return nil
}

// InsertString asserts a fact as a string. (We adopt here the insert terminology from DROOLS to avoid confusion with the Go assert)
func (env *Environment) InsertString(factstr string) (Fact, error) {
	cfactstr := C.CString(factstr)
	defer C.free(unsafe.Pointer(cfactstr))
	factptr := C.EnvAssertString(env.env, cfactstr)
	if factptr == nil {
		return nil, EnvError(env, `Error asserting fact "%s"`, factstr)
	}
	return env.newFact(factptr), nil
}

// LoadFactsFromFile loads facts from the given file
func (env *Environment) LoadFactsFromFile(filename string) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	retcode := C.EnvLoadFacts(env.env, cfilename)
	if retcode == -1 {
		return EnvError(env, `Error loading facts from "%s"`, filename)
	}
	return nil
}

// LoadFacts loads facts from the given string
func (env *Environment) LoadFacts(factstr string) error {
	cfactstr := C.CString(factstr)
	defer C.free(unsafe.Pointer(cfactstr))

	retcode := C.EnvLoadFactsFromString(env.env, cfactstr, -1)
	if retcode == -1 {
		return EnvError(env, `Error loading facts from string`)
	}
	return nil
}

// SaveFacts saves facts to the given file
func (env *Environment) SaveFacts(filename string, savemode SaveMode) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	retcode := C.EnvSaveFacts(env.env, cfilename, savemode.CVal())
	if retcode == -1 {
		return EnvError(env, `Error saving facts to "%s"`, filename)
	}
	return nil
}

func (env *Environment) newFact(fact unsafe.Pointer) Fact {
	templ := C.EnvFactDeftemplate(env.env, fact)
	if C.implied_deftemplate(templ) == 0 {
		return &ImpliedFact{
			env:     env,
			factptr: fact,
		}
	}
	return &TemplateFact{
		env:     env,
		factptr: fact,
	}
}

/*
   def templates(self):
        """Iterate over the defined Templates."""
        template = lib.EnvGetNextDeftemplate(self._env, ffi.NULL)

        while template != ffi.NULL:
            yield Template(self._env, template)

            template = lib.EnvGetNextDeftemplate(self._env, template)

    def find_template(self, name):
        """Find the Template by its name."""
        deftemplate = lib.EnvFindDeftemplate(self._env, name.encode())
        if deftemplate == ffi.NULL:
            raise LookupError("Template '%s' not found" % name)

        return Template(self._env, deftemplate)






def slot_value(env, fact, slot):
    data = clips.data.DataObject(env)
    slot = slot if slot is not None else ffi.NULL
    implied = lib.implied_deftemplate(lib.EnvFactDeftemplate(env, fact))

    if not implied and slot == ffi.NULL:
        raise ValueError()

    if bool(lib.EnvGetFactSlot(env, fact, slot, data.byref)):
        return data.value


def slot_values(env, fact, tpl):
    data = clips.data.DataObject(env)
    lib.EnvDeftemplateSlotNames(env, tpl, data.byref)

    return ((s, slot_value(env, fact, s.encode())) for s in data.value)


def fact_pp_string(env, fact):
    buf = ffi.new('char[1024]')
    lib.EnvGetFactPPForm(env, buf, 1024, fact)

    return ffi.string(buf).decode()


def template_pp_string(env, template):
    strn = lib.EnvGetDeftemplatePPForm(env, template)

    if strn != ffi.NULL:
        return ffi.string(strn).decode().strip()
    else:
        module = ffi.string(lib.EnvDeftemplateModule(env, template)).decode()
        name = ffi.string(lib.EnvGetDeftemplateName(env, template)).decode()

        return '(deftemplate %s::%s)' % (module, name)
*/
