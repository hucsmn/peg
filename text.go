package peg

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Underlying types implemented Pattern interface.
type (
	patternText struct {
		insensitive bool // assumes couldSafelyFoldCase(text) if true
		text        string
	}

	patternBackwardPredicate struct {
		text string
	}

	patternTextSet struct {
		insensitive bool // assumes couldSafelyFoldCase(text) if true
		sorted      []string
		tree        prefixTree
	}

	patternTextReferring struct {
		grpname string
	}

	patternBackwardPredicateReferring struct {
		grpname string
	}
)

// T matches given text literally.
func T(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	return &patternText{insensitive: false, text: text}
}

// TI matches given text case-insensitively.
func TI(text string) Pattern {
	if len(text) == 0 {
		return True
	}

	// split text into pieces that satisfies couldSafelyFoldCase(piece).
	// build Seq(TI(safe), S("unsafe"))
	var pats []Pattern
	var at, span int
	for at < len(text) {
		r, n := utf8.DecodeRuneInString(text[at:])
		if pat, ok := lengthChangedAfterFoldCase[r]; ok {
			if span > 0 {
				pats = append(pats, &patternText{
					insensitive: true,
					text:        foldCase(text[at-span : at]),
				})
			}
			pats = append(pats, pat)
			at += n
			span = 0
		} else {
			at += n
			span += n
		}
	}
	if span > 0 {
		pats = append(pats, &patternText{
			insensitive: true,
			text:        foldCase(text[at-span : at]),
		})
	}

	switch len(pats) {
	case 0:
		return True
	case 1:
		return pats[0]
	default:
		return Seq(pats...)
	}
}

// B predicates if given text is matched in backward.
func B(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	return &patternBackwardPredicate{text}
}

// TS matches any text existed in given set.
func TS(textset ...string) Pattern {
	pat := &patternTextSet{insensitive: false}
	copied := make([]string, len(textset))
	copy(copied, textset)
	pat.set(copied)
	return pat
}

// TSI matches any text existed in given set case-insensitively.
func TSI(textset ...string) Pattern {
	// Find out strings changed length after foldCase.
	safe := make([]string, 0, len(textset))
	unsafe := []string{}
	for _, s := range textset {
		if couldSafelyFoldCase(s) {
			safe = append(safe, s)
		} else {
			unsafe = append(unsafe, s)
		}
	}

	// build Alt(TI(unsafe[*]), ..., TSI(safe)).
	var tail Pattern
	if len(safe) != 0 {
		pat := &patternTextSet{insensitive: true}
		pat.set(safe)
		tail = pat
	}
	if len(unsafe) == 0 {
		if tail == nil {
			return False
		}
		return tail
	}

	// append in reversed order, to make sure the longer text is prior.
	sort.Strings(unsafe)
	pats := make([]Pattern, 0, len(unsafe)+1)
	for i := len(unsafe) - 1; i >= 0; i-- {
		pats = append(pats, TI(unsafe[i]))
	}
	if tail != nil {
		pats = append(pats, tail)
	}
	return Alt(pats...)
}

// Ref matches the text in the group named grpname.
//
// Use the lastest anonymous group if grpname == "".
//
// Use "" if the name grpname does not exist.
func Ref(grpname string) Pattern {
	return &patternTextReferring{grpname}
}

// RefB predicates if the text in the group named grpname matches in backward.
//
// Use the lastest anonymous group if grpname == "".
//
// Use "" if the name grpname does not exist.
func RefB(grpname string) Pattern {
	return &patternBackwardPredicateReferring{grpname}
}

// Matches text.
func (pat *patternText) match(ctx *context) error {
	text := ctx.next(len(pat.text))
	if pat.insensitive {
		text = foldCase(text)
	}

	if text == pat.text {
		ctx.consume(len(text))
		return ctx.commit()
	}
	return ctx.predicates(false)
}

// Predicates backward text.
func (pat *patternBackwardPredicate) match(ctx *context) error {
	return ctx.predicates(ctx.previous(len(pat.text)) == pat.text)
}

// Matches text set.
func (pat *patternTextSet) match(ctx *context) error {
	type matchState struct {
		n int
		prefixTree
	}

	back := false
	stack := []matchState{{0, pat.tree}}
	for len(stack) > 0 {
		state := stack[len(stack)-1]
		if back {
			stack = stack[:len(stack)-1]
			if state.term {
				ctx.consume(state.n)
				return ctx.commit()
			}
			continue
		}

		s := ctx.next(state.n + state.width)[state.n:]
		if pat.insensitive {
			s = foldCase(s)
		}
		i, ok := state.search(s)
		if !ok {
			back = true
			continue
		}
		stack = append(stack, matchState{
			n:          state.n + state.width,
			prefixTree: state.subs[i],
		})
	}
	return ctx.predicates(false)
}

// assumes that textset is owned by set.
func (pat *patternTextSet) set(textset []string) error {
	pat.sorted = textset
	if pat.insensitive {
		for i := range pat.sorted {
			pat.sorted[i] = foldCase(pat.sorted[i])
		}
	}
	sort.Strings(pat.sorted)
	pat.tree = buildPrefixTree(pat.sorted)
	return nil
}

// Matches referring text from groups.
func (pat *patternTextReferring) match(ctx *context) error {
	if ctx.config.DisableGrouping {
		return errorReferDisabled
	}

	text := ctx.refer(pat.grpname)
	if ctx.next(len(text)) == text {
		ctx.consume(len(text))
		return ctx.commit()
	}
	return ctx.predicates(false)
}

// Predicates referring text from groups in backward.
func (pat *patternBackwardPredicateReferring) match(ctx *context) error {
	if ctx.config.DisableGrouping {
		return errorReferDisabled
	}

	text := ctx.refer(pat.grpname)
	return ctx.predicates(ctx.previous(len(text)) == text)
}

func (pat *patternText) String() string {
	if pat.insensitive {
		return fmt.Sprintf("I%q", pat.text)
	}
	return fmt.Sprintf("%q", pat.text)
}

func (pat *patternBackwardPredicate) String() string {
	return fmt.Sprintf("back? %q", pat.text)
}

func (pat *patternTextSet) String() string {
	strs := make([]string, len(pat.sorted))
	for i := range pat.sorted {
		if pat.insensitive {
			strs[i] = fmt.Sprintf("I%q", pat.sorted[i])
		} else {
			strs[i] = fmt.Sprintf("%q", pat.sorted[i])
		}
	}
	return fmt.Sprintf("(%s)", strings.Join(strs, "|"))
}

func (pat *patternTextReferring) String() string {
	if pat.grpname == "" {
		return "%%"
	}
	return fmt.Sprintf("%%%q%%", pat.grpname)
}

func (pat *patternBackwardPredicateReferring) String() string {
	if pat.grpname == "" {
		return "back? %%"
	}
	return fmt.Sprintf("back? %%%q%%", pat.grpname)
}
