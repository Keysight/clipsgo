package clips

/*
   Copyright 2020 Keysight Technologies

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*/

import (
	"fmt"
	"log"
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
	"NUMBER",
	"OBJECT",
	"USER",
	"INITIAL-OBJECT",
	"PRIMITIVE",
	"INSTANCE",
	"INSTANCE-NAME",
	"INSTANCE-ADDRESS",
	"ADDRESS",
	"FACT-ADDRESS",
	"EXTERNAL-ADDRESS",
	"MULTIFIELD",
	"LEXEME",
	"SYMBOL",
	"STRING",
	"crlf",
	"object",
	"deftemplate",
	"deffunction",
	"defmodule",
	"defrule",
	"defclass",
	"defglobal",
	"deffacts",
	"defmessage-handler",
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
	"batch",
	"batch*",
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
	"fact-slot-value",
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
	"matches",
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
	"send",
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

// HighlightedWriter tries to catch selected writes and use the syntax highlighting on them
type HighlightedWriter struct {
	delegate     prompt.ConsoleWriter
	writingInput bool
}

// WriteRaw to write raw byte array.
func (hw *HighlightedWriter) WriteRaw(data []byte) {
	hw.delegate.WriteRaw(data)
}

// Write to write safety byte array by removing control sequences.
func (hw *HighlightedWriter) Write(data []byte) {
	hw.delegate.Write(data)
}

// WriteRawStr to write raw string.
func (hw *HighlightedWriter) WriteRawStr(data string) {
	hw.delegate.WriteRawStr(data)
}

// WriteStr to write safety string by removing control sequences.
func (hw *HighlightedWriter) WriteStr(data string) {
	if hw.writingInput && data != "" {
		// If prompt is trying to write input text, intercept and replace the write with syntax highlighted text
		partial := shellContext.cmd.String() + data
		iterator, err := shellContext.lexer.Tokenise(nil, partial)
		if err != nil {
			hw.delegate.WriteStr(data)
			return
		}
		modified := strings.Builder{}
		shellContext.formatter.Format(&modified, shellContext.style, iterator)
		lines := strings.Split(modified.String(), "\n")
		if strings.Contains(data, "\n") {
			// data was a complete line, so we should output the next-to-last
			hw.delegate.WriteRawStr(lines[len(lines)-2] + "\n")
		} else {
			hw.delegate.WriteRawStr(lines[len(lines)-1])
		}
		return
	}
	hw.delegate.WriteStr(data)
}

// Flush to flush buffer.
func (hw *HighlightedWriter) Flush() error {
	return hw.delegate.Flush()
}

// EraseScreen erases the screen with the background colour and moves the cursor to home.
func (hw *HighlightedWriter) EraseScreen() {
	hw.delegate.EraseScreen()
}

// EraseUp erases the screen from the current line up to the top of the screen.
func (hw *HighlightedWriter) EraseUp() {
	hw.delegate.EraseUp()
}

// EraseDown erases the screen from the current line down to the bottom of the screen.
func (hw *HighlightedWriter) EraseDown() {
	hw.delegate.EraseDown()
}

// EraseStartOfLine erases from the current cursor position to the start of the current line.
func (hw *HighlightedWriter) EraseStartOfLine() {
	hw.delegate.EraseStartOfLine()
}

// EraseEndOfLine erases from the current cursor position to the end of the current line.
func (hw *HighlightedWriter) EraseEndOfLine() {
	hw.delegate.EraseEndOfLine()
}

// EraseLine erases the entire current line.
func (hw *HighlightedWriter) EraseLine() {
	hw.delegate.EraseLine()
}

// ShowCursor stops blinking cursor and show.
func (hw *HighlightedWriter) ShowCursor() {
	hw.delegate.ShowCursor()
}

// HideCursor hides cursor.
func (hw *HighlightedWriter) HideCursor() {
	hw.delegate.HideCursor()
}

// CursorGoTo sets the cursor position where subsequent text will begin.
func (hw *HighlightedWriter) CursorGoTo(row, col int) {
	hw.delegate.CursorGoTo(row, col)
}

// CursorUp moves the cursor up by 'n' rows; the default count is 1.
func (hw *HighlightedWriter) CursorUp(n int) {
	hw.delegate.CursorUp(n)
}

// CursorDown moves the cursor down by 'n' rows; the default count is 1.
func (hw *HighlightedWriter) CursorDown(n int) {
	hw.delegate.CursorDown(n)
}

// CursorForward moves the cursor forward by 'n' columns; the default count is 1.
func (hw *HighlightedWriter) CursorForward(n int) {
	hw.delegate.CursorForward(n)
}

// CursorBackward moves the cursor backward by 'n' columns; the default count is 1.
func (hw *HighlightedWriter) CursorBackward(n int) {
	hw.delegate.CursorBackward(n)
}

// AskForCPR asks for a cursor position report (CPR).
func (hw *HighlightedWriter) AskForCPR() {
	hw.delegate.AskForCPR()
}

// SaveCursor saves current cursor position.
func (hw *HighlightedWriter) SaveCursor() {
	hw.delegate.SaveCursor()
}

// UnSaveCursor restores cursor position after a Save Cursor.
func (hw *HighlightedWriter) UnSaveCursor() {
	hw.delegate.UnSaveCursor()
}

// ScrollDown scrolls display down one line.
func (hw *HighlightedWriter) ScrollDown() {
	hw.delegate.ScrollDown()
}

// ScrollUp scroll display up one line.
func (hw *HighlightedWriter) ScrollUp() {
	hw.delegate.ScrollUp()
}

// SetTitle sets a title of terminal window.
func (hw *HighlightedWriter) SetTitle(title string) {
	hw.delegate.SetTitle(title)
}

// ClearTitle clears a title of terminal window.
func (hw *HighlightedWriter) ClearTitle() {
	hw.delegate.ClearTitle()
}

// SetColor sets text and background colors. and specify whether text is bold.
func (hw *HighlightedWriter) SetColor(fg, bg prompt.Color, bold bool) {
	// We use unreasonable color settings to "flag" when prompt is about to write input text
	if fg == prompt.Red && bg == prompt.Red {
		hw.writingInput = true
		hw.delegate.SetColor(prompt.DefaultColor, prompt.DefaultColor, bold)
	} else {
		hw.writingInput = false
		hw.delegate.SetColor(fg, bg, bold)
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		/* Just leave it to syntax highlighting for now
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
	/*
		iterator, err := shellContext.lexer.Tokenise(nil, cmdstr)
		if err == nil {
			err = shellContext.formatter.Format(os.Stdout, shellContext.style, iterator)
		}
	*/
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
		{Pattern: `\[([^\]]+)\]`, Type: chroma.LiteralStringOther, Mutator: nil},
		{Pattern: `<([^>]+)>`, Type: chroma.LiteralStringOther, Mutator: nil},
		{Pattern: `(TRUE|FALSE|nil)`, Type: chroma.NameVariableInstance, Mutator: nil},
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

	writer := &HighlightedWriter{delegate: prompt.NewStandardOutputWriter()}
	// go-prompt screws with the log destination, so make sure it gets put back
	logout := log.Writer()
	defer log.SetOutput(logout)
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(primaryPrompt),
		prompt.OptionWriter(writer),
		prompt.OptionLivePrefix(changePrefix),
		prompt.OptionInputTextColor(prompt.Red),
		prompt.OptionInputBGColor(prompt.Red),
	)
	p.Run()
}
