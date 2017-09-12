package peg

import (
	"fmt"
	"strings"
)

var (
	// True always matches, but consumes no text.
	True Pattern = &patternBoolean{true}

	// False always dismatch.
	False Pattern = &patternBoolean{false}

	// SOL predicates start of line.
	SOL Pattern = &patternLineAnchorPredicate{true}

	// EOL predicates end of line.
	EOL Pattern = &patternLineAnchorPredicate{false}

	// EOF predicates end of file.
	EOF Pattern = patternEOFPredicate{}
)

// Underlying types implemented Pattern interface.
type (
	patternBoolean struct {
		ok bool
	}

	patternAbort struct {
		msg string
	}

	patternLineAnchorPredicate struct {
		linestart bool
	}

	patternEOFPredicate struct{}

	patternPredicate struct {
		not bool
		pat Pattern
	}

	patternAndPredicate struct {
		pats []Pattern
	}

	patternOrPredicate struct {
		pats []Pattern
	}

	patternIf struct {
		cond Pattern
		yes  Pattern
		no   Pattern
	}

	patternSwitch struct {
		cases []struct {
			cond Pattern
			then Pattern
		}
		otherwise Pattern
	}
)

// Test predicates if pattern is matched, consuming no text.
//
// Note that, if predicates true, groups/parse captures won't be discarded.
func Test(pat Pattern) Pattern {
	return &patternPredicate{not: false, pat: pat}
}

// Not predicates if pattern is dismatched, consuming no text.
//
// Note that, if predicates true, groups/parse captures won't be discarded.
func Not(pat Pattern) Pattern {
	return &patternPredicate{not: true, pat: pat}
}

// And searches the patterns in order to predicate if all the pattern
// is matched at current position. It consumes no text.
//
// Note that, if predicates true, groups/parse captures won't be discarded.
func And(pats ...Pattern) Pattern {
	if len(pats) == 0 {
		return True
	}
	return &patternAndPredicate{pats}
}

// Or searches the patterns in order to predicate if any the pattern
// is matched at current position. It consumes no text.
//
// Note that, if predicates true, groups/parse captures won't be discarded.
func Or(pats ...Pattern) Pattern {
	if len(pats) == 0 {
		return False
	}
	return &patternOrPredicate{pats}
}

// Abort quit the normal matching process, reporting an error message.
func Abort(msg string) Pattern {
	return &patternAbort{msg}
}

// When tests the given condition but consumes no text, then determines to
// execute then-branch or just predicates false.
//
// Note that, if cond predicates true, groups/parse captures won't be
// discarded.
func When(cond, then Pattern) Pattern {
	return &patternIf{cond: cond, yes: then, no: False}
}

// If tests the given condition but consumes no text, then determines to
// execute yes-branch or no-branch.
//
// Note that, if cond predicates true, groups/parse captures won't be
// discarded.
func If(cond, yes, no Pattern) Pattern {
	return &patternIf{cond: cond, yes: yes, no: no}
}

// Switch tests cond-then pairs in order, executes then-branch if cond is true.
// If there is no true cond, executes the optional otherwise-branch
// (the default pattern is True).
//
// Note that, if cond predicates true, groups/parse captures won't be
// discarded.
func Switch(cond, then Pattern, rest ...Pattern) Pattern {
	pat := &patternSwitch{}
	pat.cases = append(pat.cases, struct {
		cond Pattern
		then Pattern
	}{cond, then})

	if len(rest)%2 != 0 {
		pat.otherwise = rest[len(rest)-1]
		rest = rest[:len(rest)-1]
	} else {
		pat.otherwise = True
	}

	for len(rest) > 0 {
		cond, then = rest[0], rest[1]
		rest = rest[2:]
		pat.cases = append(pat.cases, struct {
			cond Pattern
			then Pattern
		}{cond, then})
	}
	return pat
}

// Matches empty string if true, dismatches if false.
func (pat *patternBoolean) match(ctx *context) error {
	return ctx.predicates(pat.ok)
}

// Predicates SOL/EOL.
func (pat *patternLineAnchorPredicate) match(ctx *context) error {
	p, n := ctx.previous(1), ctx.next(1)

	if pat.linestart {
		// SOL matches start of file or just after a "\n"|"\r"|"\r\n".
		if p == "" {
			return ctx.predicates(true)
		}
		return ctx.predicates(
			p == "\n" || (p == "\r" && n != "\n"))
	}

	// EOL matches end of file or just before a "\n"|"\r"|"\r\n".
	if n == "" {
		return ctx.predicates(true)
	}
	return ctx.predicates(n == "\r" || (n == "\n" && p != "\r"))
}

// Predicates EOF.
func (patternEOFPredicate) match(ctx *context) error {
	return ctx.predicates(ctx.next(1) == "")
}

// Predicates if sub-pattern matches.
func (pat *patternPredicate) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if pat.not {
		ret.ok = !ret.ok
	}
	return ctx.predicates(ret.ok)
}

// Predicates if all the sub-patterns match.
func (pat *patternAndPredicate) match(ctx *context) error {
	for ctx.locals.i < len(pat.pats) {
		if !ctx.justReturned() {
			return ctx.call(pat.pats[ctx.locals.i])
		}

		if !ctx.ret.ok {
			return ctx.predicates(false)
		}
		ctx.locals.i++
	}
	return ctx.predicates(true)
}

// Predicates if any sub-pattern matches.
func (pat *patternOrPredicate) match(ctx *context) error {
	for ctx.locals.i < len(pat.pats) {
		if !ctx.justReturned() {
			return ctx.call(pat.pats[ctx.locals.i])
		}

		if ctx.ret.ok {
			return ctx.predicates(true)
		}
		ctx.locals.i++
	}
	return ctx.predicates(false)
}

// Abort directly.
func (pat *patternAbort) match(ctx *context) error {
	pos := ctx.tell()
	return errorf("abort:%s: %s", pos.String(), pat.msg)
}

// Branch `if'.
func (pat *patternIf) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.cond)
	}

	if ctx.ret.ok {
		return ctx.execute(pat.yes)
	}
	return ctx.execute(pat.no)
}

// Branch `switch'.
func (pat *patternSwitch) match(ctx *context) error {
	for ctx.locals.i < len(pat.cases) {
		if !ctx.justReturned() {
			return ctx.call(pat.cases[ctx.locals.i].cond)
		}

		if ctx.ret.ok {
			return ctx.execute(pat.cases[ctx.locals.i].then)
		}
		ctx.locals.i++
	}
	return ctx.execute(pat.otherwise)
}

func (pat *patternBoolean) String() string {
	if pat.ok {
		return "true"
	}
	return "false"
}

func (pat *patternLineAnchorPredicate) String() string {
	if pat.linestart {
		return "sol?"
	}
	return "eol?"
}

func (patternEOFPredicate) String() string {
	return "eof?"
}

func (pat *patternPredicate) String() string {
	if pat.not {
		return fmt.Sprintf("!%s", pat.pat)
	}
	return fmt.Sprintf("?%s", pat.pat)
}

func (pat *patternAndPredicate) String() string {
	strs := make([]string, len(pat.pats))
	for i, pat := range pat.pats {
		strs[i] = fmt.Sprint(pat)
	}
	return fmt.Sprintf("(%s)", strings.Join(strs, " && "))
}

func (pat *patternOrPredicate) String() string {
	strs := make([]string, len(pat.pats))
	for i, pat := range pat.pats {
		strs[i] = fmt.Sprint(pat)
	}
	return fmt.Sprintf("(%s)", strings.Join(strs, " || "))
}

func (pat *patternAbort) String() string {
	return fmt.Sprintf("abort(%q)", pat.msg)
}

func (pat *patternIf) String() string {
	if b, ok := pat.no.(*patternBoolean); ok && b.ok == false {
		return fmt.Sprintf("switch(%s: %s)", pat.cond, pat.yes)
	}
	return fmt.Sprintf("switch(%s: %s; %s)", pat.cond, pat.yes, pat.no)
}

func (pat *patternSwitch) String() string {
	strs := make([]string, len(pat.cases))
	for i := range pat.cases {
		strs[i] = fmt.Sprintf("%s: %s", pat.cases[i].cond, pat.cases[i].then)
	}
	if b, ok := pat.otherwise.(*patternBoolean); ok && b.ok == false {
		return fmt.Sprintf("switch(%s)", strings.Join(strs, "; "))
	}
	return fmt.Sprintf("switch(%s; %s)", strings.Join(strs, "; "), pat.otherwise)
}
