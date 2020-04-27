package clips

import "C"

// CLIPSType is an enumeration CLIPS uses to describe data types
type CLIPSType C.short

const (
	FLOAT = iota
	INTEGER
	SYMBOL
	STRING
	MULTIFIELD
	EXTERNAL_ADDRESS
	FACT_ADDRESS
	INSTANCE_ADDRESS
	INSTANCE_NAME
)

var clipstypes = [...]string{
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

func (typ CLIPSType) String() string {
	return clipstypes[typ]
}

/*
if sys.version_info.major == 3:
    class Symbol(str):
        """Python equivalent of a CLIPS SYMBOL."""
        def __new__(cls, symbol):
            return str.__new__(cls, sys.intern(symbol))
elif sys.version_info.major == 2:
    class Symbol(str):
        """Python equivalent of a CLIPS SYMBOL."""
        def __new__(cls, symbol):
            # pylint: disable=E0602
            return str.__new__(cls, intern(str(symbol)))


class InstanceName(Symbol):
    """Instance names are CLIPS SYMBOLS."""
    pass


class SaveMode(IntEnum):
    LOCAL_SAVE = 0
    VISIBLE_SAVE = 1


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


class TemplateSlotDefaultType(IntEnum):
    NO_DEFAULT = 0
    STATIC_DEFAULT = 1
    DYNAMIC_DEFAULT = 2


# Assign functions and routers per Environment
ENVIRONMENT_DATA = {}
EnvData = namedtuple('EnvData', ('user_functions', 'routers'))
*/
