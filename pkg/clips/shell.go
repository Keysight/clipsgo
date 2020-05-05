package clips

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/styles"
	"github.com/c-bata/go-prompt"
)

// ShellContext stores the context of the shell environment
type ShellContext struct {
	cmd       strings.Builder
	env       *Environment
	lexer     chroma.Lexer
	style     *chroma.Style
	formatter chroma.Formatter
}

var shellContext *ShellContext

var keywords = []string{
	"SYMBOL",
	"STRING",
	"INTEGER",
	"FLOAT",
	"crlf",
	"object",
	"deftemplate",
	"deffunction",
	"defmodule",
	"defrule",
	"defclass",
	"defglobal",
	"deffacts",
	"test",
	"and",
	"or",
	"eq",
	"neq",
	"name",
	"is-a",
	"type",
	"slot",
	"multislot",
	"t ",
}

var builtins = []string{
	"!=",
	"*",
	"**",
	"+",
	"-",
	"/",
	"<",
	"<=",
	"<>",
	"=",
	">",
	">=",
	"abs",
	"acos",
	"acosh",
	"acot",
	"acoth",
	"acsc",
	"acsch",
	"active-duplicate-instance",
	"active-initialize-instance",
	"active-make-instance",
	"active-message-duplicate-instance",
	"active-message-modify-instance",
	"active-modify-instance",
	"agenda",
	"and",
	"any-instancep",
	"apropos",
	"asec",
	"asech",
	"asin",
	"assert",
	"cot",
	"coth",
	"create$",
	"csc",
	"csch",
	"defclass-module",
	"deffacts-module",
	"deffunction-module",
	"defgeneric-module",
	"defglobal-module",
	"definstances-module",
	"defrule-module",
	"deftemplate-module",
	"deg-grad",
	"deg-rad",
	"delayed-do-for-all-instances",
	"delete$",
	"delete-instance",
	"dependencies",
	"dependents",
	"describe-class",
	"direct-mv-delete",
	"direct-mv-insert",
	"direct-mv-replace",
	"div",
	"do-for-all-instances",
	"do-for-instance",
	"dribble-off",
	"dribble-on",
	"duplicate",
	"duplicate-instance",
	"duplicate-instance",
	"dynamic-get",
	"dynamic-put",
	"edit",
	"eq",
	"eval",
	"evenp",
	"exit",
	"exp",
	"expand$",
	"explode$",
	"facts",
	"fact-existp",
	"halt",
	"if",
	"implode$",
	"init-slots",
	"initialize-instance",
	"initialize-instance",
	"insert$",
	"instance-address",
	"instance-addressp",
	"instance-existp",
	"instance-name",
	"instance-name-to-symbol",
	"instance-namep",
	"instancep",
	"instances",
	"integer",
	"integerp",
	"length",
	"length$",
	"lexemep",
	"list-defclasses",
	"list-deffacts",
	"list-deffunctions",
	"list-defgenerics",
	"list-defglobals",
	"list-definstances",
	"list-defmessage-handlers",
	"list-defmethods",
	"list-defmodules",
	"list-defrules",
	"list-deftemplates",
	"list-focus-stack",
	"list-watch-items",
	"load",
	"load*",
	"load-facts",
	"load-instances",
	"log",
	"log10",
	"loop-for-count",
	"lowcase",
	"make-instance",
	"ppdefclass",
	"ppdeffacts",
	"ppdeffunction",
	"ppdefgeneric",
	"ppdefglobal",
	"ppdefinstances",
	"ppdefmessage-handler",
	"ppdefmethod",
	"ppdefmodule",
	"ppdefrule",
	"ppdeftemplate",
	"ppinstance",
	"preview-generic",
	"preview-send",
	"primitives-info",
	"print-region",
	"printout",
	"progn",
	"progn$",
	"put",
	"rad-deg",
	"random",
	"read",
	"readline",
	"refresh",
	"refresh-agenda",
	"release-mem",
	"remove",
	"remove-break",
	"rename",
	"replace$",
	"reset",
	"rest$",
	"restore-instances",
	"retract",
	"return",
	"round",
	"rule-complexity",
	"rules",
	"run",
	"save",
	"save-facts",
	"save-instances",
	"str-assert",
	"str-cat",
	"str-compare",
	"str-explode",
	"str-implode",
	"str-index",
	"str-length",
	"stringp",
	"sub-string",
	"subclassp",
	"subseq$",
	"subset",
	"subsetp",
	"superclassp",
	"switch",
	"sym-cat",
	"symbol-to-instance-name",
	"symbolp",
	"system",
	"tan",
	"tanh",
	"time",
	"toss",
	"type",
	"type",
	"undefclass",
	"undeffacts",
	"undeffunction",
	"undefgeneric",
	"undefglobal",
	"undefinstances",
	"undefmessage-handler",
	"undefmethod",
	"undefrule",
	"undeftemplate",
	"unmake-instance",
	"unwatch",
	"upcase",
	"watch",
	"while",
	"wordp",
}

var primaryPrompt = "» "
var secondaryPrompt = "+ "

func completer(d prompt.Document) []prompt.Suggest {
	/*
		iterator, err := shellContext.lexer.Tokenise(nil, shellContext.cmd.String())
		if err != nil {
			return []prompt.Suggest{}
		}
		tokens := iterator.Tokens()
		lasttoken := tokens[len(tokens)-1]
	*/
	s := []prompt.Suggest{
		/*
			{Text: "agenda", Description: "list agenda"},
			{Text: "defclass", Description: "define a class"},
			{Text: "deffunction", Description: "define a function"},
			{Text: "defmessage-handler", Description: "define a function"},
			{Text: "defrule", Description: "define a rule"},
			{Text: "deftemplate", Description: "define a template fact"},
			{Text: "facts", Description: "list current facts"},
			{Text: "instances", Description: "list current instances"},
			{Text: "matches", Description: "list matches for a rule"},
			{Text: "watch", Description: "enable watch"},
			{Text: "unwatch", Description: "disable watch"},
			{Text: "rules", Description: "disable watch"},
			{Text: "exit", Description: "exit the shell"},
		*/
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), false)
}

func changePrefix() (string, bool) {
	if shellContext.cmd.Len() > 0 {
		return secondaryPrompt, true
	}
	return primaryPrompt, false
}

func executor(in string) {
	shellContext.cmd.WriteString(fmt.Sprintf("%s\n", in))
	cmdstr := shellContext.cmd.String()
	iterator, err := shellContext.lexer.Tokenise(nil, cmdstr)
	if err == nil {
		err = shellContext.formatter.Format(os.Stdout, shellContext.style, iterator)
	}
	complete, err := shellContext.env.CompleteCommand(cmdstr)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[SHELL]: %s\n", err.Error()))
		shellContext.cmd.Reset()
		return
	}
	if complete {
		err := shellContext.env.SendCommand(strings.TrimRight(cmdstr, "\n"))
		shellContext.cmd.Reset()
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("[SHELL]: %s\n", err.Error()))
		}
	}
}

func initContext(env *Environment) {
	rules := []chroma.Rule{
		{Pattern: `;.*$`, Type: chroma.Comment, Mutator: nil},
		{Pattern: `\s+`, Type: chroma.Text, Mutator: nil},
		{Pattern: `-?\d+\.\d+`, Type: chroma.NumberFloat, Mutator: nil},
		{Pattern: `-?\d+`, Type: chroma.NumberInteger, Mutator: nil},
		{Pattern: `"(\\\\|\\"|[^"])*"`, Type: chroma.String, Mutator: nil},
		{Pattern: `(TRUE|FALSE|nil)`, Type: chroma.NameConstant, Mutator: nil},
		{Pattern: "('|#|`|,@|,|\\.)", Type: chroma.Operator, Mutator: nil},
	}
	for _, kw := range keywords {
		rules = append(rules, chroma.Rule{
			Pattern: fmt.Sprintf(`(%s)`, regexp.QuoteMeta(kw)),
			Type:    chroma.Keyword,
			Mutator: nil,
		})
	}
	for _, bi := range builtins {
		rules = append(rules, chroma.Rule{
			Pattern: fmt.Sprintf(`(?<=\(\s*)(%s)`, regexp.QuoteMeta(bi)),
			Type:    chroma.NameBuiltin,
			Mutator: nil,
		})
	}
	for _, rl := range []chroma.Rule{
		{Pattern: `\?`, Type: chroma.NameLabel, Mutator: nil},
		{Pattern: `(\(|\))`, Type: chroma.Punctuation, Mutator: nil},
		{Pattern: `[\w!$%*+,/:<=>@^~|-]+`, Type: chroma.Text, Mutator: nil},
	} {
		rules = append(rules, rl)
	}
	shellContext = &ShellContext{
		cmd: strings.Builder{},
		env: env,
		lexer: chroma.MustNewLexer(&chroma.Config{
			Name:      "CLIPS",
			Aliases:   []string{"clips", "clp"},
			Filenames: []string{"*.clp"},
			MimeTypes: []string{"text/x-clips", "application/x-clips"},
		}, chroma.Rules{
			"root": rules,
		}),
		style:     styles.Get("native"),
		formatter: formatters.Get("terminal256"),
	}
}

// Shell sets up an interactive CLIPS shell within the given environment
func (env *Environment) Shell() {
	initContext(env)
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(primaryPrompt),
		prompt.OptionLivePrefix(changePrefix),
	)
	p.Run()
}
