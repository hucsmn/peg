package peg

import (
	"fmt"
	"strings"
)

// Underlying types implemented Pattern interface.
type (
	patternSequence struct {
		pats []Pattern
	}

	patternAlternative struct {
		pats []Pattern
	}

	patternSkip struct {
		n int
	}

	patternAnyRuneUntil struct {
		without bool
		pat     Pattern
	}

	patternQualifierAtLeast struct {
		n   int
		pat Pattern
	}

	patternQualifierOptional struct {
		pat Pattern
	}

	patternQualifierRange struct {
		m, n int
		pat  Pattern
	}
)

// Seq tries to match patterns in given sequence, the Seq itself only matched
// when all of the patterns is successfully matched, the text is consumed in
// order. It dismatches if any dismatched pattern is encountered.
func Seq(sequence ...Pattern) Pattern {
	switch len(sequence) {
	case 0:
		return True
	case 1:
		return sequence[0]
	default:
		return &patternSequence{sequence}
	}
}

// Alt searches the first matched pattern in the given choices, theAlt itself
// only matches when any pattern is successfully matched, the then Alt consumes
// the searched pattern's number of bytes matched. It dismatches if all the
// choices is dismatched.
//
// It is recommended to place pattern that match more text in a prior order.
// For example, Alt(Seq(Q1(R('0', '9')), T("."), Q1(R('0', '9'))),
// Q1(R('0', '9'))) could match both "0.0" and "0", while Alt(Q1(R('0', '9')),
// Seq(Q1(R('0', '9')), T("."), Q1(R('0', '9')))) could only match "0".
func Alt(choices ...Pattern) Pattern {
	switch len(choices) {
	case 0:
		return False
	case 1:
		return choices[0]
	default:
		return &patternAlternative{choices}
	}
}

// Skip matches exactly n runes.
func Skip(n int) Pattern {
	if n < 0 {
		n = 0
	}
	if n == 0 {
		return True
	}
	return &patternSkip{n}
}

// Until matches any rune until the start of text piece
// searched by given pattern.
func Until(pat Pattern) Pattern {
	return &patternAnyRuneUntil{without: true, pat: pat}
}

// UntilB matches any rune until the end of text piece
// searched by given pattern.
func UntilB(pat Pattern) Pattern {
	return &patternAnyRuneUntil{without: false, pat: pat}
}

// Q0 matches any of the given choices repeated zero or more times.
func Q0(choices ...Pattern) Pattern {
	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Q0(False) <=> True
		return True
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	return &patternQualifierAtLeast{n: 0, pat: pat}
}

// Q1 matches any of the given choices repeated one or more times.
func Q1(choices ...Pattern) Pattern {
	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Q1(False) <=> False
		return False
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	return &patternQualifierAtLeast{n: 1, pat: pat}
}

// Qn matches any of the given choices repeated at least n times.
//
// The paramenter n is required to be non-negative,
// or the negative value would be treated as zero.
func Qn(n int, choices ...Pattern) Pattern {
	if n < 0 {
		n = 0
	}

	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Qn(n = 0, False) <=> True
		// Qn(n > 0, False) <=> False
		if n == 0 {
			return True
		}
		return False
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	return &patternQualifierAtLeast{n: n, pat: pat}
}

// Q01 matches any of the given choices zero to one time.
func Q01(choices ...Pattern) Pattern {
	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Q01(False) <=> True
		return True
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	return &patternQualifierOptional{pat}
}

// Q0n matches any of the given choices repeated at most n times.
//
// The paramenter n is required to be non-negative,
// or the negative value would be treated as zero.
func Q0n(n int, choices ...Pattern) Pattern {
	if n < 0 {
		n = 0
	}

	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Q0n(n, False) <=> True
		return True
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	switch n {
	case 0:
		return True
	case 1:
		return &patternQualifierOptional{pat}
	default:
		return &patternQualifierRange{m: 0, n: n, pat: pat}
	}
}

// Qnn matches any of the given choices repeated exactly n times.
//
// The paramenter n is required to be non-negative,
// or the negative value would be treated as zero.
func Qnn(n int, choices ...Pattern) Pattern {
	if n < 0 {
		n = 0
	}

	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Qnn(n = 0, False) <=> True
		// Qnn(n > 0, False) <=> False
		if n == 0 {
			return True
		}
		return False
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	switch n {
	case 0:
		return True
	case 1:
		return pat
	default:
		return &patternQualifierRange{m: n, n: n, pat: pat}
	}
}

// Qmn matches any of the given choices repeated from m to n times.
//
// The paramenter m and n are required to be non-negative,
// or the negative value would be treated as zero.
func Qmn(m, n int, choices ...Pattern) Pattern {
	if n < 0 {
		n = 0
	}
	if m < 0 {
		m = 0
	}
	if m > n {
		m, n = n, m
	}

	var pat Pattern

	switch len(choices) {
	case 0:
		// Alt() <=> False
		// Qmn(m = 0, n, False) <=> True
		// Qmn(m > 0, n, False) <=> False
		if m == 0 {
			return True
		}
		return False
	case 1:
		pat = choices[0]
	default:
		pat = Alt(choices...)
	}

	if n == 0 {
		return True
	}
	if m == 0 && n == 1 {
		return &patternQualifierOptional{pat}
	}
	return &patternQualifierRange{m: m, n: n, pat: pat}
}

// J0 matches zero or more items separated by sep.
func J0(item, sep Pattern) Pattern {
	return Jn(0, item, sep)
}

// J1 matches one or more items separated by sep.
func J1(item, sep Pattern) Pattern {
	return Jn(1, item, sep)
}

// Jn matches at least n items separated by sep.
func Jn(n int, item, sep Pattern) Pattern {
	if n < 0 {
		n = 0
	}

	if n == 0 {
		return Alt(
			Seq(item, Q0(Seq(sep, item))),
			True)
	}
	return Seq(item, Qn(n-1, Seq(sep, item)))
}

// J0n matches at most n items separated by sep.
func J0n(n int, item, sep Pattern) Pattern {
	if n < 0 {
		n = 0
	}

	return Jmn(0, n, item, sep)
}

// Jnn matches exactly n items separated by sep.
func Jnn(n int, item, sep Pattern) Pattern {
	if n < 0 {
		n = 0
	}

	switch n {
	case 0:
		return True
	case 1:
		return item
	default:
		return Seq(item, Qnn(n-1, Seq(sep, item)))
	}
}

// Jmn matches m to n items separated by sep.
func Jmn(m, n int, item, sep Pattern) Pattern {
	if m < 0 {
		m = 0
	}
	if n < 0 {
		n = 0
	}
	if m > n {
		m, n = n, m
	}

	switch {
	case n == 0:
		return True
	case m == 0:
		return Alt(
			Seq(item, Qmn(0, n-1, Seq(sep, item))),
			True)
	default:
		return Seq(item, Qmn(m-1, n-1, Seq(sep, item)))
	}
}

// Matches if all the sub-patterns match in order.
func (pat *patternSequence) match(ctx *context) error {
	for ctx.locals.i < len(pat.pats) {
		if !ctx.justReturned() {
			return ctx.call(pat.pats[ctx.locals.i])
		}

		ret := ctx.ret
		if !ret.ok {
			return ctx.predicates(false)
		}
		ctx.consume(ret.n)
		ctx.locals.i++
	}
	return ctx.commit()
}

// Matches if any sub-pattern matches, searches in order.
func (pat *patternAlternative) match(ctx *context) error {
	for ctx.locals.i < len(pat.pats) {
		if !ctx.justReturned() {
			// optimize for the last choice
			if ctx.locals.i == len(pat.pats)-1 {
				return ctx.execute(pat.pats[ctx.locals.i])
			}
			return ctx.call(pat.pats[ctx.locals.i])
		}

		ret := ctx.ret
		if ret.ok {
			ctx.consume(ret.n)
			return ctx.commit()
		}
		ctx.locals.i++
	}
	return ctx.predicates(false)
}

// Matches exactly n runes.
func (pat *patternSkip) match(ctx *context) error {
	for i := 0; i < pat.n; i++ {
		_, n := ctx.nextRune()
		ctx.consume(n)

		// no enough runes
		if n == 0 {
			return ctx.predicates(false)
		}
	}
	return ctx.commit()
}

// Matches until pattern.
func (pat *patternAnyRuneUntil) match(ctx *context) error {
	for {
		if ctx.reachedRepeatLimit(ctx.locals.i) {
			return errorReachedRepeatLimit
		}

		if !ctx.justReturned() {
			return ctx.call(pat.pat)
		}

		ret := ctx.ret
		if ret.ok {
			// pattern searched
			if !pat.without {
				ctx.consume(ret.n)
			}
			return ctx.commit()
		}

		_, n := ctx.nextRune()
		ctx.consume(n)
		ctx.locals.i++

		// search failed
		if n == 0 {
			return ctx.predicates(false)
		}
	}
}

// Matches at least n times.
func (pat *patternQualifierAtLeast) match(ctx *context) error {
	for {
		if ctx.reachedRepeatLimit(ctx.locals.i) {
			return errorReachedRepeatLimit
		}

		if !ctx.justReturned() {
			return ctx.call(pat.pat)
		}

		ret := ctx.ret
		if !ret.ok {
			if ctx.locals.i < pat.n {
				return ctx.predicates(false)
			}
			return ctx.commit()
		}
		ctx.consume(ret.n)
		ctx.locals.i++
	}
}

// Matches zero or one times.
func (pat *patternQualifierOptional) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if !ret.ok {
		return ctx.predicates(true)
	}
	ctx.consume(ret.n)
	return ctx.commit()
}

// Matches m to n times.
func (pat *patternQualifierRange) match(ctx *context) error {
	for ctx.locals.i < pat.n {
		if ctx.reachedRepeatLimit(ctx.locals.i) {
			return errorReachedRepeatLimit
		}

		if !ctx.justReturned() {
			return ctx.call(pat.pat)
		}

		ret := ctx.ret
		if !ret.ok {
			if ctx.locals.i < pat.m {
				return ctx.predicates(false)
			}
			return ctx.commit()
		}
		ctx.consume(ret.n)
		ctx.locals.i++
	}
	return ctx.commit()
}

func (pat *patternSequence) String() string {
	strs := make([]string, len(pat.pats))
	for i, pat := range pat.pats {
		strs[i] = fmt.Sprint(pat)
	}
	return fmt.Sprintf("(%s)", strings.Join(strs, " "))
}

func (pat *patternAlternative) String() string {
	strs := make([]string, len(pat.pats))
	for i, pat := range pat.pats {
		strs[i] = fmt.Sprint(pat)
	}
	return fmt.Sprintf("(%s)", strings.Join(strs, " | "))
}

func (pat *patternSkip) String() string {
	return fmt.Sprintf("#dot <%d>", pat.n)
}

func (pat *patternAnyRuneUntil) String() string {
	if pat.without {
		return fmt.Sprintf("until(%s)", pat.pat)
	}
	return fmt.Sprintf("until_end_of(%s)", pat.pat)
}

func (pat *patternQualifierAtLeast) String() string {
	switch pat.n {
	case 0:
		return fmt.Sprintf("%s *", pat.pat)
	case 1:
		return fmt.Sprintf("%s +", pat.pat)
	default:
		return fmt.Sprintf("%s <%d..>", pat.pat, pat.n)
	}
}

func (pat *patternQualifierOptional) String() string {
	return fmt.Sprintf("[ %s ]", pat.pat)
}

func (pat *patternQualifierRange) String() string {
	if pat.m == pat.n {
		return fmt.Sprintf("%s <%d>", pat.pat, pat.m)
	}
	return fmt.Sprintf("%s <%d..%d>", pat.pat, pat.m, pat.n)
}
