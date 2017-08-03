package peg

import "unicode/utf8"

// Running state of pattern matching.
type context struct {
	// Configuration
	config Config

	// Text
	text  string // current matched text is text[at-n:at])
	at    int
	n     int
	pcalc positionCalculator

	// Current stack frame
	pat    Pattern
	locals localValues
	isret  bool
	ret    returnValues // allow accessing from pat.match(ctx)

	// Groups
	groups      []string
	namedGroups map[string]string

	// Call stack
	levels    int // execute(pat) won't push callstack, use additional counter instead
	callstack []stackFrame

	// Grammar tree construction
	scopes   []map[string]Pattern
	capstack []captureThunk
}

// Local values of running pattern.
type localValues struct {
	i int // loop counter

	// to be extended
}

// Return values of pattern match
type returnValues struct {
	ok          bool
	n           int
	groups      []string
	namedGroups map[string]string
}

// Callstack frame.
type stackFrame struct {
	pat         Pattern
	at          int
	n           int
	locals      localValues
	levels      int
	groups      []string
	namedGroups map[string]string
}

// Incomplete grammar tree construction.
type captureThunk struct {
	cons NonTerminalConstructor
	args []Capture
}

func newContext(pat Pattern, text string, config Config) *context {
	ctx := &context{}
	ctx.reset(pat, text, config)
	return ctx
}

func (ctx *context) reset(pat Pattern, text string, config Config) {
	ctx.config = config

	ctx.text = text
	ctx.at = 0
	ctx.n = 0
	ctx.pcalc = positionCalculator{text: text}

	ctx.pat = pat
	ctx.locals = localValues{}
	ctx.isret = false
	ctx.ret = returnValues{}

	ctx.levels = 0
	ctx.callstack = nil

	ctx.groups = nil
	ctx.namedGroups = nil

	ctx.scopes = nil
	ctx.capstack = []captureThunk{{cons: nil, args: nil}}
}

// The main loop.
func (ctx *context) match() error {
	for ctx.pat != nil {
		// ctx.pat.match(ctx) yields when:
		//   1) return ctx.call(callee)
		//      or return ctx.execute(callee)
		//   2) return ctx.returns(ret)
		//      or return ctx.returnsPredication(ok)
		//      or return ctx.returnsMatched()
		//   3) return any_error
		err := ctx.pat.match(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Invoke callee, and backups stack frame and matching state.
func (ctx *context) call(callee Pattern) error {
	// backup stack frame
	if ctx.config.CallstackLimit > 0 &&
		ctx.levels >= ctx.config.CallstackLimit {
		return errorCallstackOverflow
	}
	ctx.callstack = append(ctx.callstack, stackFrame{
		pat:         ctx.pat,
		at:          ctx.at,
		n:           ctx.n,
		locals:      ctx.locals,
		levels:      ctx.levels,
		groups:      ctx.groups,
		namedGroups: ctx.namedGroups,
	})
	ctx.levels++

	// skip the matched span.
	ctx.n = 0

	// setup stack frame
	ctx.pat = callee
	ctx.locals = localValues{}
	ctx.isret = false
	ctx.ret = returnValues{}
	ctx.groups = nil
	ctx.namedGroups = nil

	return nil
}

// Invoke callee, but do not backups stack frame.
// No text should be already consumed before execute.
func (ctx *context) execute(callee Pattern) error {
	// assert no text already consumed
	if ctx.n != 0 {
		return errorExecuteWhenConsumed
	}

	// increase call level counter
	if ctx.config.CallstackLimit > 0 &&
		ctx.levels >= ctx.config.CallstackLimit {
		return errorCallstackOverflow
	}
	ctx.levels++

	// setup stack frame
	ctx.pat = callee
	ctx.locals = localValues{}
	ctx.isret = false
	ctx.ret = returnValues{}

	return nil
}

// Returns to uplevel, predicates if matched, empty text is matched text.
func (ctx *context) returnsPredication(ok bool) error {
	return ctx.returns(returnValues{
		ok:          ok,
		n:           0,
		groups:      ctx.groups,
		namedGroups: ctx.namedGroups,
	})
}

// Returns to uplevel, the consumed text is matched.
func (ctx *context) returnsMatched() error {
	return ctx.returns(returnValues{
		ok:          true,
		n:           ctx.n,
		groups:      ctx.groups,
		namedGroups: ctx.namedGroups,
	})
}

// Returns to uplevel.
func (ctx *context) returns(ret returnValues) error {
	ctx.isret = true
	ctx.ret = ret

	if len(ctx.callstack) > 0 {
		// pop callstack
		if len(ctx.callstack) < 1 || ctx.levels < 1 {
			return errorCornerCase
		}
		frame := ctx.callstack[len(ctx.callstack)-1]
		ctx.callstack = ctx.callstack[:len(ctx.callstack)-1]
		ctx.levels--

		// recover stack frame
		ctx.pat = frame.pat
		ctx.at = frame.at
		ctx.n = frame.n
		ctx.locals = frame.locals
		ctx.levels = frame.levels
		ctx.groups = frame.groups
		ctx.namedGroups = frame.namedGroups

		// update groups
		if ret.ok {
			if len(ctx.groups) == 0 {
				ctx.groups = ret.groups
			} else {
				ctx.groups = append(ctx.groups, ret.groups...)
			}
			if len(ctx.namedGroups) == 0 {
				ctx.namedGroups = ret.namedGroups
			} else {
				for n, g := range ret.namedGroups {
					ctx.namedGroups[n] = g
				}
			}
		}
	} else {
		// terminate pattern matching normally
		ctx.pat = nil
	}
	return nil
}

// Tests if just returned from a callee.
func (ctx *context) justReturned() bool {
	isret := ctx.isret
	ctx.isret = false
	return isret
}

// Tests if the looping counter reached loop limit.
func (ctx *context) reachedLoopLimit() bool {
	return ctx.config.LoopLimit > 0 && ctx.locals.i >= ctx.config.LoopLimit
}

// Moves cursor forward.
func (ctx *context) consume(n int) {
	ctx.n += n
	ctx.at += n
}

// Tell the position of cursor.
func (ctx *context) tell() Position {
	if ctx.config.DisableLineColumnCounting {
		return Position{Offest: ctx.at}
	}
	return ctx.pcalc.calculate(ctx.at)
}

// Tell the matched text.
func (ctx *context) span() string {
	return ctx.text[ctx.at-ctx.n : ctx.at]
}

// Reads next n bytes.
func (ctx *context) readNext(n int) string {
	tail := ctx.text[ctx.at:]
	if len(tail) < n {
		return tail
	}
	return tail[:n]
}

// Reads previous n bytes.
func (ctx *context) readPrev(n int) string {
	if ctx.at < n {
		return ctx.text[:ctx.at]
	}
	return ctx.text[ctx.at-n : ctx.at]
}

// Reads next rune.
func (ctx *context) readRune() (r rune, n int) {
	return utf8.DecodeRuneInString(ctx.text[ctx.at:])
}

// Enter the given namespace, overriding uplevel definitions.
func (ctx *context) enter(namespace map[string]Pattern) {
	ctx.scopes = append(ctx.scopes, namespace)
}

// Leave current namespace.
func (ctx *context) leave() {
	ctx.scopes = ctx.scopes[:len(ctx.scopes)-1]
}

// Looks up variable definition, gets nil if undefined.
func (ctx *context) lookup(name string) Pattern {
	for i := len(ctx.scopes) - 1; i >= 0; i-- {
		namespace := ctx.scopes[i]
		if pat, ok := namespace[name]; ok {
			return pat
		}
	}
	return nil
}

// Stores matched text to named group if grpname is non-empty,
// or push the text to groups.
func (ctx *context) group(grpname string) {
	if ctx.config.DisableGrouping {
		return
	}

	span := ctx.span()
	if grpname != "" {
		if ctx.namedGroups == nil {
			ctx.namedGroups = map[string]string{grpname: span}
		} else {
			ctx.namedGroups[grpname] = span
		}
	} else {
		ctx.groups = append(ctx.groups, span)
	}
}

// Gets the text stored in named groups if grpname is non-empty,
// or gets the the lastest text in groups ("" if nothing in groups).
func (ctx *context) refer(grpname string) string {
	if ctx.config.DisableGrouping {
		return ""
	}

	if grpname != "" {
		g, ok := ctx.namedGroups[grpname]
		if ok {
			return g
		}
		for i := len(ctx.callstack) - 1; i >= 0; i-- {
			g, ok := ctx.callstack[i].namedGroups[grpname]
			if ok {
				return g
			}
		}
		return ""
	}

	if len(ctx.groups) != 0 {
		return ctx.groups[len(ctx.groups)-1]
	}
	for i := len(ctx.callstack) - 1; i >= 0; i-- {
		groups := ctx.callstack[i].groups
		if len(groups) != 0 {
			return groups[len(groups)-1]
		}
	}
	return ""
}

// Pushes a constructed capture (terminal or non-terminal).
func (ctx *context) push(cap Capture) error {
	if ctx.config.DisableCapturing {
		return nil
	}

	if len(ctx.capstack) < 1 {
		return errorCornerCase
	}

	argsp := &ctx.capstack[len(ctx.capstack)-1].args
	*argsp = append(*argsp, cap)
	return nil
}

// Begins non-terminal construction.
func (ctx *context) begin(cons NonTerminalConstructor) {
	if ctx.config.DisableCapturing {
		return
	}

	ctx.capstack = append(ctx.capstack, captureThunk{
		cons: cons,
		args: nil,
	})
}

// Ends up current non-terminal construction.
func (ctx *context) end(matched bool) error {
	if ctx.config.DisableCapturing {
		return nil
	}

	if len(ctx.capstack) < 2 {
		return errorCornerCase
	}

	thunk := ctx.capstack[len(ctx.capstack)-1]
	ctx.capstack = ctx.capstack[:len(ctx.capstack)-1]

	if !matched {
		return nil
	}

	if thunk.cons == nil {
		return errorNilConstructor
	}
	cap, err := thunk.cons(thunk.args)
	if err != nil {
		return err
	}
	return ctx.push(cap)
}
