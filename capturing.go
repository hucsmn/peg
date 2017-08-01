package peg

import (
	"fmt"
	"strings"
)

// Underlying types implemented Pattern interface.
type (
	patternLet struct {
		pat  Pattern
		vars map[string]Pattern
	}

	patternCaptureVariable struct {
		varname string
		cons    NonTerminalConstructor
	}

	patternCaptureToken struct {
		pat     Pattern
		toktype int
		cons    TerminalConstructor
	}

	patternCaptureCons struct {
		pat  Pattern
		cons NonTerminalConstructor
	}

	patternCaptureTerm struct {
		pat  Pattern
		cons TerminalConstructor
	}
)

// Let binds variable dfinitions to a flatten namescope,
// then invokes the entry pattern.
// Panics if any variable is nil.
func Let(vars map[string]Pattern, entry Pattern) Pattern {
	for name, pat := range vars {
		if pat == nil {
			panic(errorUndefinedVar(name))
		}
	}
	return &patternLet{pat: entry, vars: vars}
}

// V invokes a defined variable without capturing.
func V(varname string) Pattern {
	return &patternCaptureVariable{
		varname: varname,
		cons:    nil,
	}
}

// CV invokes a defined variable with capturing enabled.
func CV(varname string) Pattern {
	return &patternCaptureVariable{
		varname: varname,
		cons:    newVariableConstructor(varname),
	}
}

// CK constructs Token-typed terminals from matched text.
func CK(toktype int, pat Pattern) Pattern {
	return &patternCaptureToken{
		pat:     pat,
		toktype: toktype,
		cons:    newTokenConstructor(toktype),
	}
}

// CC constructs non-terminal using user defined constructor.
func CC(cons NonTerminalConstructor, pat Pattern) Pattern {
	return &patternCaptureCons{pat: pat, cons: cons}
}

// CT constructs terminal using user defined constructor.
func CT(cons TerminalConstructor, pat Pattern) Pattern {
	return &patternCaptureTerm{pat: pat, cons: cons}
}

func newVariableConstructor(name string) NonTerminalConstructor {
	return func(subs []Capture) (Capture, error) {
		return &Variable{Name: name, Subs: subs}, nil
	}
}

func newTokenConstructor(toktype int) TerminalConstructor {
	return func(span string, pos Position) (Capture, error) {
		return &Token{Type: toktype, Value: span, Position: pos}, nil
	}
}

// Setups variables.
func (pat *patternLet) match(ctx *context) error {
	if !ctx.justReturned() {
		// enter namespace
		ctx.enter(pat.vars)
		return ctx.call(pat.pat)
	}

	// leave namespace.
	ret := ctx.ret
	ctx.leave()
	return ctx.returns(ret)
}

// Invokes vaiable and optionally captures it.
func (pat *patternCaptureVariable) match(ctx *context) error {
	// lookup and invoke
	if !ctx.justReturned() {
		callee := ctx.lookup(pat.varname)
		if callee == nil {
			return errorUndefinedVar(pat.varname)
		}

		if pat.cons == nil {
			// won't capture the variable
			return ctx.execute(callee)
		}
		ctx.begin(pat.cons)
		return ctx.call(callee)
	}

	// finish capturing
	ret := ctx.ret
	err := ctx.end(ret.ok)
	if err != nil {
		return err
	}
	return ctx.returns(ret)
}

// Captures text to construct a token.
func (pat *patternCaptureToken) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if !ret.ok {
		return ctx.returnsPredication(false)
	}

	head := ctx.tell()
	ctx.consume(ret.n)
	term, err := pat.cons(ctx.span(), head)
	if err != nil {
		return err
	}
	err = ctx.push(term)
	if err != nil {
		return err
	}
	return ctx.returnsMatched()
}

// Captures using customed non-terminal constructor.
func (pat *patternCaptureCons) match(ctx *context) error {
	if !ctx.justReturned() {
		ctx.begin(pat.cons)
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	err := ctx.end(ret.ok)
	if err != nil {
		return err
	}
	return ctx.returns(ret)
}

// Captures text to construct terminal using customed constructor.
func (pat *patternCaptureTerm) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if !ret.ok {
		return ctx.returnsPredication(false)
	}

	head := ctx.tell()
	ctx.consume(ret.n)
	term, err := pat.cons(ctx.span(), head)
	if err != nil {
		return err
	}
	err = ctx.push(term)
	if err != nil {
		return err
	}
	return ctx.returnsMatched()
}

func (pat *patternLet) String() string {
	strs := make([]string, 0, len(pat.vars))
	for name, value := range pat.vars {
		strs = append(strs, fmt.Sprintf("$%s := %s", name, value))
	}
	return fmt.Sprintf("let (%s) in %s", strings.Join(strs, "; "), pat.pat)
}

func (pat *patternCaptureVariable) String() string {
	if pat.cons == nil {
		return fmt.Sprintf("$%s", pat.varname)
	}
	return fmt.Sprintf("${%s}", pat.varname)
}

func (pat *patternCaptureCons) String() string {
	return fmt.Sprintf("cons_%p{%s}", pat.cons, pat.pat)
}

func (pat *patternCaptureToken) String() string {
	return fmt.Sprintf("token_%d{%s}", pat.toktype, pat.pat)
}

func (pat *patternCaptureTerm) String() string {
	return fmt.Sprintf("term_%p{%s}", pat.cons, pat.pat)
}
