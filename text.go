package peg

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// TODO: make TI, TSI, BackI works on all the UTF-8 strings.

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

	patternTextRefered struct {
		grpname string
	}

	patternBackwardPredicateRefered struct {
		grpname string
	}

	// Internal structure for text set.
	prefixTree struct {
		term  bool     // if searching could be terminated here
		width int      // length of each key
		keys  []string // sorted keys
		subs  []prefixTree
	}
)

// T matches text literally.
func T(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	return &patternText{insensitive: false, text: text}
}

// TI matches text case insensitively.
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

// Back predicates if text matches in backward.
func Back(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	return &patternBackwardPredicate{text}
}

// TS matches texts in set.
func TS(textset ...string) Pattern {
	pat := &patternTextSet{insensitive: false}
	copied := make([]string, len(textset))
	copy(copied, textset)
	pat.set(copied)
	return pat
}

// TSI matches texts in set case insensitively.
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

// Ref matches the text in groups.
func Ref(grpname string) Pattern {
	return &patternTextRefered{grpname}
}

// RefBack predicates if the text in groups matches in backward.
func RefBack(grpname string) Pattern {
	return &patternBackwardPredicateRefered{grpname}
}

// Matches text.
func (pat *patternText) match(ctx *context) error {
	text := ctx.readNext(len(pat.text))
	if pat.insensitive {
		text = foldCase(text)
	}

	if text == pat.text {
		ctx.consume(len(text))
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

// Predicates backward text.
func (pat *patternBackwardPredicate) match(ctx *context) error {
	return ctx.returnsPredication(ctx.readPrev(len(pat.text)) == pat.text)
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
				return ctx.returnsMatched()
			}
			continue
		}

		s := ctx.readNext(state.n + state.width)[state.n:]
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
	return ctx.returnsPredication(false)
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

func buildPrefixTree(sorted []string) prefixTree {
	tree := prefixTree{}
	var i int
	for ; i < len(sorted) && sorted[i] == ""; i++ {
		tree.term = true
	}
	sorted = sorted[i:]
	if len(sorted) == 0 {
		return tree
	}

	tree.width = len(sorted[0])
	for _, s := range sorted {
		if len(s) < tree.width {
			tree.width = len(s)
		}
	}

	var lastprefix = sorted[0][:tree.width]
	var lasttail = sorted[0][tree.width:]
	var tails = []string{lasttail}
	for _, s := range sorted[1:] {
		prefix, tail := s[:tree.width], s[tree.width:]
		if prefix == lastprefix {
			if tail != lasttail {
				tails = append(tails, tail)
				lasttail = tail
			}
		} else {
			tree.keys = append(tree.keys, lastprefix)
			tree.subs = append(tree.subs, buildPrefixTree(tails))
			lastprefix = prefix
			lasttail = tail
			tails = []string{lasttail}
		}
	}
	tree.keys = append(tree.keys, lastprefix)
	tree.subs = append(tree.subs, buildPrefixTree(tails))
	return tree
}

func (tree prefixTree) search(s string) (int, bool) {
	if len(s) != tree.width {
		return 0, false
	}

	i, j := 0, len(tree.keys)
	for i < j {
		m := i + (j-i)/2
		if s == tree.keys[m] {
			return m, true
		} else if s > tree.keys[m] {
			i = m + 1
		} else {
			j = m
		}
	}
	return 0, false
}

// Matches refered text from groups.
func (pat *patternTextRefered) match(ctx *context) error {
	text := ctx.refer(pat.grpname)
	if ctx.readNext(len(text)) == text {
		ctx.consume(len(text))
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

// Predicates refered text from groups in backward.
func (pat *patternBackwardPredicateRefered) match(ctx *context) error {
	text := ctx.refer(pat.grpname)
	return ctx.returnsPredication(ctx.readPrev(len(text)) == text)
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

func (pat *patternTextRefered) String() string {
	if pat.grpname == "" {
		return "%%"
	}
	return fmt.Sprintf("%%%q%%", pat.grpname)
}

func (pat *patternBackwardPredicateRefered) String() string {
	if pat.grpname == "" {
		return "back? %%"
	}
	return fmt.Sprintf("back? %%%q%%", pat.grpname)
}
