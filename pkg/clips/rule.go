package clips

// #cgo CFLAGS: -I ../../clips_source
// #cgo LDFLAGS: -L ../../clips_source -l clips -lm
// #include <clips/clips.h>
import "C"
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
	"strings"
	"unsafe"
)

// Rule represents a rule within CLIPS
type Rule struct {
	env  *Environment
	rptr unsafe.Pointer
}

// Activation represents an activation from the agenda
type Activation struct {
	env    *Environment
	actptr unsafe.Pointer
}

// Strategy is used to specify the conflict resolution strategy
type Strategy int

const (
	DEPTH Strategy = iota
	BREADTH
	LEX
	MEA
	COMPLEXITY
	RANDOM
)

var clipsStrategies = [...]string{
	"DEPTH",
	"BREADTH",
	"LEX",
	"MEA",
	"COMPLEXITY",
	"RANDOM",
}

func (sm Strategy) String() string {
	return clipsStrategies[int(sm)]
}

// CVal returns the value as appropriate for a C call
func (sm Strategy) CVal() C.int {
	return C.int(sm)
}

// SalienceEvaluation is used to specify the salience evaluation behavior
type SalienceEvaluation int

const (
	WHEN_DEFINED SalienceEvaluation = iota
	WHEN_ACTIVATED
	EVERY_CYCLE
)

var clipsSalienceEvaluations = [...]string{
	"WHEN_DEFINED",
	"WHEN_ACTIVATED",
	"EVERY_CYCLE",
}

func (sm SalienceEvaluation) String() string {
	return clipsSalienceEvaluations[int(sm)]
}

// CVal returns the value as appropriate for a C call
func (sm SalienceEvaluation) CVal() C.int {
	return C.int(sm)
}

// Verbosity controls how much is output to stdout for rule matches
type Verbosity int

const (
	VERBOSE Verbosity = iota
	SUCCINCT
	TERSE
)

var clipsVerbosity = [...]string{
	"VERBOSE",
	"SUCCINCT",
	"TERSE",
}

func (sm Verbosity) String() string {
	return clipsVerbosity[int(sm)]
}

// CVal returns the value as appropriate for a C call
func (sm Verbosity) CVal() C.int {
	return C.int(sm)
}

// AgendaChanged returns true if any rule activation changes have occurred since last call
func (env *Environment) AgendaChanged() bool {
	ret := C.EnvGetAgendaChanged(env.env)
	C.EnvSetAgendaChanged(env.env, 0)
	if ret == 1 {
		return true
	}
	return false
}

// Focus returns the module associated with the current focus
func (env *Environment) Focus() *Module {
	modptr := C.EnvGetFocus(env.env)
	return createModule(env, modptr)
}

// SetFocus sets the current focus to the given module
func (env *Environment) SetFocus(module *Module) {
	if env != module.env {
		panic("SetFocus to module from another environment")
	}
	C.EnvFocus(env.env, module.modptr)
}

// Strategy returns the current conflict resolution strategy
func (env *Environment) Strategy() Strategy {
	ret := C.EnvGetStrategy(env.env)
	return Strategy(ret)
}

// SetStrategy sets the conflict resolution strategy
func (env *Environment) SetStrategy(strategy Strategy) {
	C.EnvSetStrategy(env.env, strategy.CVal())
}

// SalienceEvaluation returns the salience evaulation behavior
func (env *Environment) SalienceEvaluation() SalienceEvaluation {
	ret := C.EnvGetSalienceEvaluation(env.env)
	return SalienceEvaluation(ret)
}

// SetSalienceEvaluation sets the salience evaluation behavior
func (env *Environment) SetSalienceEvaluation(val SalienceEvaluation) {
	C.EnvSetSalienceEvaluation(env.env, val.CVal())
}

// Rules returns the list of all rules in the CLIPS environment
func (env *Environment) Rules() []*Rule {
	rptr := C.EnvGetNextDefrule(env.env, nil)
	ret := make([]*Rule, 0, 10)
	for rptr != nil {
		ret = append(ret, createRule(env, rptr))
		rptr = C.EnvGetNextDefrule(env.env, rptr)
	}
	return ret
}

// FindRule returns the rule of the given name
func (env *Environment) FindRule(name string) (*Rule, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	rptr := C.EnvFindDefrule(env.env, cname)
	if rptr == nil {
		return nil, NotFoundError(fmt.Errorf(`Rule "%s" not found`, name))
	}
	return createRule(env, rptr), nil
}

// Reorder reorders the activations in the agenda. If module is nil, the current module is used. To be called after changing the conflict resoution strategy
func (env *Environment) Reorder(module *Module) {
	var modptr unsafe.Pointer
	if module != nil {
		modptr = module.modptr
	}
	C.EnvReorderAgenda(env.env, modptr)
}

// Refresh recomputes the salience values of the Activations on the Agenda. If module is nil, the current module is used. To be called after changing the conflict resoution strategy
func (env *Environment) Refresh(module *Module) {
	var modptr unsafe.Pointer
	if module != nil {
		modptr = module.modptr
	}
	C.EnvRefreshAgenda(env.env, modptr)
}

// Activations returns the list of activations in the agenda
func (env *Environment) Activations() []*Activation {
	actptr := C.EnvGetNextActivation(env.env, nil)
	ret := make([]*Activation, 0, 10)
	for actptr != nil {
		ret = append(ret, createActivation(env, actptr))
		actptr = C.EnvGetNextActivation(env.env, actptr)
	}
	return ret
}

// ClearAgenda deletes all activations in the agenda
func (env *Environment) ClearAgenda() error {
	ret := C.EnvDeleteActivation(env.env, nil)
	if ret != 1 {
		return EnvError(env, "Unable to clear agenda")
	}
	return nil
}

// ClearFocus removes all modules from the focus stack
func (env *Environment) ClearFocus() {
	C.EnvClearFocusStack(env.env)
}

// Run runs the activations in the agenda. If limit is not negative, only the first activations up to the limit will be run
func (env *Environment) Run(limit int64) int64 {
	if limit < 0 {
		limit = -1
	}
	ret := C.EnvRun(env.env, C.longlong(limit))
	return int64(ret)
}

func createRule(env *Environment, rptr unsafe.Pointer) *Rule {
	return &Rule{
		env:  env,
		rptr: rptr,
	}
}

// Equal returns true if the other rule represents the same CLIPS rule as this one
func (r *Rule) Equal(other *Rule) bool {
	return r.rptr == other.rptr
}

func (r *Rule) String() string {
	cstr := C.EnvGetDefrulePPForm(r.env.env, r.rptr)
	return strings.TrimRight(C.GoString(cstr), "\n")
}

// Name returns the name of this rule
func (r *Rule) Name() string {
	cname := C.EnvGetDefruleName(r.env.env, r.rptr)
	return C.GoString(cname)
}

// Module returns the module in which the rule is defined
func (r *Rule) Module() *Module {
	cmodname := C.EnvDefruleModule(r.env.env, r.rptr)
	modptr := C.EnvFindDefmodule(r.env.env, cmodname)
	return createModule(r.env, modptr)
}

// Deletable returns true if the rule is unreferenced and can be deleted
func (r *Rule) Deletable() bool {
	ret := C.EnvIsDefruleDeletable(r.env.env, r.rptr)
	if ret == 1 {
		return true
	}
	return false
}

// WatchedFirings returns true if rule firings are being watched
func (r *Rule) WatchedFirings() bool {
	ret := C.EnvGetDefruleWatchFirings(r.env.env, r.rptr)
	if ret == 1 {
		return true
	}
	return false
}

// WatchFirings sets whether rule firigns are watched
func (r *Rule) WatchFirings(val bool) {
	var cflag C.uint
	if val {
		cflag = 1
	}
	C.EnvSetDefruleWatchFirings(r.env.env, cflag, r.rptr)
}

// WatchedActivations returns true if rule activations are being watched
func (r *Rule) WatchedActivations() bool {
	ret := C.EnvGetDefruleWatchActivations(r.env.env, r.rptr)
	if ret == 1 {
		return true
	}
	return false
}

// WatchActivations sets whether rule activations should be watched
func (r *Rule) WatchActivations(val bool) {
	var cflag C.uint
	if val {
		cflag = 1
	}
	C.EnvSetDefruleWatchActivations(r.env.env, cflag, r.rptr)
}

// Matches shows partial matches and activations for the rule. Returns a list containing the
// combined sum of the matches, the combined sum of partial matches, then the total activations.
// Verbosity determines how much to output to stdout
func (r *Rule) Matches(verbosity Verbosity) ([]interface{}, error) {
	data := createDataObject(r.env)
	defer data.Delete()
	C.EnvMatches(r.env.env, r.rptr, verbosity.CVal(), data.byRef())
	retval := data.Value()
	ret, ok := retval.([]interface{})
	if !ok {
		panic("Unexpected return value from CLIPS")
	}
	return ret, nil
}

// Refresh refreshes the rule
func (r *Rule) Refresh() error {
	ret := C.EnvRefresh(r.env.env, r.rptr)
	if ret != 1 {
		return EnvError(r.env, "Unable to refresh rule")
	}
	return nil
}

// AddBreakpoint adds a breakpoint for the rule
func (r *Rule) AddBreakpoint() {
	C.EnvSetBreak(r.env.env, r.rptr)
}

// RemoveBreakpoint removes a breakpoint for the rule
func (r *Rule) RemoveBreakpoint() error {
	ret := C.EnvRemoveBreak(r.env.env, r.rptr)
	if ret != 1 {
		return EnvError(r.env, "Unable to remove breakpoint")
	}
	return nil
}

// Undefine undefines a rule
func (r *Rule) Undefine() error {
	ret := C.EnvUndefrule(r.env.env, r.rptr)
	if ret != 1 {
		return EnvError(r.env, "Unable to undef rule")
	}
	return nil
}

func createActivation(env *Environment, actptr unsafe.Pointer) *Activation {
	return &Activation{
		env:    env,
		actptr: actptr,
	}
}

// Equal returns true if other activation represents the same CLIPS activation as this one
func (a *Activation) Equal(other *Activation) bool {
	return a.actptr == other.actptr
}

func (a *Activation) String() string {
	// TODO grow buf if we fill the 1k buffer, and try again
	var bufsize C.ulong = 1024
	buf := (*C.char)(C.malloc(C.sizeof_char * bufsize))
	defer C.free(unsafe.Pointer(buf))
	C.EnvGetActivationPPForm(a.env.env, buf, bufsize-1, a.actptr)

	return C.GoString(buf)
}

// Name returns the name of the rule of this activation
func (a *Activation) Name() string {
	ret := C.EnvGetActivationName(a.env.env, a.actptr)
	return C.GoString(ret)
}

// Salience returns the salience value for this activation
func (a *Activation) Salience() int {
	ret := C.EnvGetActivationSalience(a.env.env, a.actptr)
	return int(ret)
}

// SetSalience modifies the salience of this activation
func (a *Activation) SetSalience(salience int) {
	C.EnvSetActivationSalience(a.env.env, a.actptr, C.int(salience))
}

// Remove removes this activation from the agenda. Renamed from "delete" to avoid confusion with other Deletes which always only drop references to CLIPS
func (a *Activation) Remove() error {
	ret := C.EnvDeleteActivation(a.env.env, a.actptr)
	if ret != 1 {
		return EnvError(a.env, "Unable to remove activation from the agenda")
	}
	a.actptr = nil
	return nil
}
