package peg

import "fmt"

type (
	patternGrouping struct {
		pat     Pattern
		grpname string
	}

	patternTrigger struct {
		pat     Pattern
		label   string
		trigger func(string, Position) error
	}

	patternInjector struct {
		pat    Pattern
		label  string
		inject func(string) (n int, ok bool)
	}
)

// G saves the the matched text to an anonymous group.
func G(pat Pattern) Pattern {
	return &patternGrouping{pat: pat, grpname: ""}
}

// NG saves the the matched text to a group named grpname.
func NG(grpname string, pat Pattern) Pattern {
	return &patternGrouping{pat: pat, grpname: grpname}
}

// Trigger invokes user defined hook with the text matched.
//
// Note that, the hook could be triggered when the inner pattern was matched
// while the parent pattern is later proved to be dismatched.
func Trigger(hook func(string, Position) error, pat Pattern) Pattern {
	return &patternTrigger{
		pat:     pat,
		label:   fmt.Sprintf("trigger_%p", hook),
		trigger: hook}
}

// Inject attaches an injector to given pattern which checks the matched text
// after matched, then determines whether anything should be matched and
// how many bytes to consume.
//
// Note that, minimum of the original consumed bytes and the injector's
// returned bytes number would be taken.
func Inject(fn func(string) (int, bool), pat Pattern) Pattern {
	if fn == nil {
		return pat
	}
	return &patternInjector{
		pat:    pat,
		label:  fmt.Sprintf("inject_%p", fn),
		inject: fn,
	}
}

// Check attaches a validator to given pattern which checks the matched text
// after matched, then determines whether anything should be matched.
//
// Note that, Check(fn, pat) is not a predicator, it may consume text.
func Check(fn func(string) bool, pat Pattern) Pattern {
	if fn == nil {
		return pat
	}
	return &patternInjector{
		pat:    pat,
		label:  fmt.Sprintf("check_%p", fn),
		inject: newCheckInjector(fn),
	}
}

func newCheckInjector(fn func(string) bool) func(string) (int, bool) {
	return func(s string) (n int, ok bool) {
		if fn(s) {
			return len(s), true
		}
		return 0, false
	}
}

// Trunc detects the matched text's length to keep the matched rune count
// is no more than maxrune.
func Trunc(maxrune int, pat Pattern) Pattern {
	return &patternInjector{
		pat:    pat,
		label:  fmt.Sprintf("trunc_%d", maxrune),
		inject: newTruncateInjector(maxrune),
	}
}

func newTruncateInjector(maxrune int) func(string) (int, bool) {
	return func(s string) (int, bool) {
		if maxrune < 0 {
			return 0, false
		} else if maxrune == 0 {
			return 0, true
		}

		// rune_count(s) < len(s) < maxrune
		if len(s) < maxrune {
			return len(s), true
		}

		n := 0
		for i := range s {
			if n >= maxrune {
				return i, true
			}
			n++
		}
		return len(s), true
	}
}

// Captures text to a group.
func (pat *patternGrouping) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if !ret.ok {
		return ctx.predicates(false)
	}
	ctx.consume(ret.n)
	ctx.group(pat.grpname)
	return ctx.commit()
}

// Captures text to trigger a hook.
func (pat *patternTrigger) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if !ret.ok {
		return ctx.predicates(false)
	}

	head := ctx.tell()
	ctx.consume(ret.n)
	err := pat.trigger(ctx.span(), head)
	if err != nil {
		return err
	}
	return ctx.commit()
}

// Further validate matched text, determines how many bytes to consume.
func (pat *patternInjector) match(ctx *context) error {
	if !ctx.justReturned() {
		return ctx.call(pat.pat)
	}

	ret := ctx.ret
	if ret.ok {
		if n, ok := pat.inject(ctx.next(ret.n)); ok {
			ctx.consume(n)
			return ctx.commit()
		}
	}
	return ctx.predicates(false)
}

func (pat *patternGrouping) String() string {
	if pat.grpname == "" {
		return fmt.Sprintf("{%s}", pat.pat)
	}
	return fmt.Sprintf("%%%s%%{%s}", pat.grpname, pat.pat)
}

func (pat *patternTrigger) String() string {
	return fmt.Sprintf("%s(%s)", pat.label, pat.pat)
}

func (pat *patternInjector) String() string {
	return fmt.Sprintf("%s(%s)", pat.label, pat.pat)
}
