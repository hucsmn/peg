package peg

import (
	"fmt"
	"sort"
	"strings"
)

// TODO: make TI, TSI, BackI works on all the UTF-8 strings.

// Underlying types implemented Pattern interface.
type (
	patternText struct {
		insensitive bool
		text        string
	}

	patternBackwardPredicate struct {
		insensitive bool
		text        string
	}

	patternTextSet struct {
		insensitive bool
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
// Panics if case insensitive is not implemented for text,
// see CouldBeCaseInsensitive.
func TI(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	if !CouldBeCaseInsensitive(text) {
		panic(errorCaseInsensitive(text))
	}
	return &patternText{insensitive: true, text: strings.ToLower(text)}
}

// Back predicates if text matches in backward.
func Back(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	return &patternBackwardPredicate{
		insensitive: false,
		text:        text,
	}
}

// BackI predicates if text matches in backward case insensitively.
// Panics if case insensitive is not implemented for text,
// see CouldBeCaseInsensitive.
func BackI(text string) Pattern {
	if len(text) == 0 {
		return True
	}
	if !CouldBeCaseInsensitive(text) {
		panic(errorCaseInsensitive(text))
	}
	return &patternBackwardPredicate{
		insensitive: true,
		text:        text,
	}
}

// TS matches texts in set.
func TS(textset ...string) Pattern {
	pat := &patternTextSet{insensitive: false}
	pat.set(textset)
	return pat
}

// TSI matches texts in set case insensitively.
// Panics if case insensitive is not implemented for text,
// see CouldBeCaseInsensitive.
func TSI(textset ...string) Pattern {
	pat := &patternTextSet{insensitive: true}
	err := pat.set(textset)
	if err != nil {
		panic(err)
	}
	return pat
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
		text = strings.ToLower(text)
	}

	if text == pat.text {
		ctx.consume(len(text))
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

// Predicates backward text.
func (pat *patternBackwardPredicate) match(ctx *context) error {
	text := ctx.readPrev(len(pat.text))
	if pat.insensitive {
		text = strings.ToLower(text)
	}
	return ctx.returnsPredication(text == pat.text)
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
			s = strings.ToLower(s)
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

func (pat *patternTextSet) set(textset []string) error {
	pat.sorted = make([]string, len(textset))
	if pat.insensitive {
		for i := range textset {
			lower := strings.ToLower(textset[i])
			if len(textset[i]) != len(lower) {
				return errorCaseInsensitive(textset[i])
			}
			pat.sorted[i] = lower
		}
	} else {
		copy(pat.sorted, textset)
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
	pre := ""
	if pat.insensitive {
		pre = "I"
	}
	return fmt.Sprintf("back? %s%q", pre, pat.text)
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
	return strings.Join(strs, "|")
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

// CouldBeCaseInsensitive detects if the given string is safe for
// case insensitive text matching.
// Considering the implementations of TI, B, TSI, etc assume that
// len(text) == len(strings.ToLower(text)), but it was not guaranteed by the
// UTF-8 encoding.
// In fact, only 24 uncode letters break this rule (including "İȺȾẞΩKÅⱢⱤ").
// For example, "İ" (2 bytes) => "i" (one byte).
func CouldBeCaseInsensitive(s string) bool {
	return len(s) == len(strings.ToLower(s))
}
