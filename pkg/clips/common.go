package clips

import "C"

// Type is an enumeration CLIPS uses to describe data types
type Type C.int

const (
	FLOAT Type = iota
	INTEGER
	SYMBOL
	STRING
	MULTIFIELD
	EXTERNAL_ADDRESS
	FACT_ADDRESS
	INSTANCE_ADDRESS
	INSTANCE_NAME
)

var clipsTypes = [...]string{
	"FLOAT",
	"INTEGER",
	"SYMBOL",
	"STRING",
	"MULTIFIELD",
	"EXTERNAL_ADDRESS",
	"FACT_ADDRESS",
	"INSTANCE_ADDRESS",
	"INSTANCE_NAME",
}

func (typ Type) String() string {
	return clipsTypes[int(typ)]
}

// CVal returns the value as appropriate for a C call
func (typ Type) CVal() C.int {
	return C.int(typ)
}

// SaveMode is used to specify the type of save when saving objects to a file
type SaveMode C.short

const (
	LOCAL_SAVE SaveMode = iota
	VISIBLE_SAVE
)

var clipsSaveModes = [...]string{
	"LOCAL_SAVE",
	"VISIBLE_SAVE",
}

func (sm SaveMode) String() string {
	return clipsSaveModes[int(sm)]
}

// CVal returns the value as appropriate for a C call
func (sm SaveMode) CVal() C.int {
	return C.int(sm)
}

/* TODO
class ClassDefaultMode(IntEnum):
    CONVENIENCE_MODE = 0
    CONSERVATION_MODE = 1


class Strategy(IntEnum):
    DEPTH = 0
    BREADTH = 1
    LEX = 2
    MEA = 3
    COMPLEXITY = 4
    SIMPLICITY = 5
    RANDOM = 6


class SalienceEvaluation(IntEnum):
    WHEN_DEFINED = 0
    WHEN_ACTIVATED = 1
    EVERY_CYCLE = 2


class Verbosity(IntEnum):
    VERBOSE = 0
    SUCCINT = 1
    TERSE = 2

*/
