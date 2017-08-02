package peg

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	// If use binary search for patternRuneSet.
	runeSetSizeThreshold = 16
)

var (
	// Dot matches any rune.
	Dot Pattern = patternAnyRune{}
)

// Underlying types implemented Pattern interface.
type (
	patternAnyRune struct{}

	patternRuneSet struct {
		not     bool
		charset []rune
	}

	patternRuneRange struct {
		not    bool
		ranges []struct {
			low, high rune
		}
	}

	patternUnicodeRanges struct {
		not    bool
		names  []string
		ranges []*unicode.RangeTable
	}

	patternUnicodeRangesWithExcluding struct {
		include patternUnicodeRanges // include.not == false
		exclude patternUnicodeRanges // include.not == true
	}
)

// S matches a rune existed in given rune set.
func S(set string) Pattern {
	pat := &patternRuneSet{not: false}
	pat.set(set)
	return pat
}

// NS matches a rune not existed in given rune set.
func NS(exclude string) Pattern {
	pat := &patternRuneSet{not: true}
	pat.set(exclude)
	return pat
}

// R matches a rune in any given range pairs [low, high].
func R(low, high rune, rest ...rune) Pattern {
	pat := &patternRuneRange{
		not:    false,
		ranges: make([]struct{ low, high rune }, 1+len(rest)/2),
	}
	pat.ranges[0].low = low
	pat.ranges[0].high = high
	for i := 1; i < len(pat.ranges); i++ {
		pat.ranges[i].low = rest[(i-1)*2]
		pat.ranges[i].high = rest[(i-1)*2+1]
	}
	return pat
}

// NR matches a rune out of all given range pairs [low, high].
func NR(low, high rune, rest ...rune) Pattern {
	pat := &patternRuneRange{
		not:    true,
		ranges: make([]struct{ low, high rune }, len(rest)/2+1),
	}
	pat.ranges[0].low = low
	pat.ranges[0].high = high
	for i := 1; i < len(pat.ranges); i++ {
		pat.ranges[i].low = rest[(i-1)*2]
		pat.ranges[i].high = rest[(i-1)*2+1]
	}
	return pat
}

// U matches a rune in the given unicode ranges (see IsUnicodeRangeName).
// Range name started with an optional prefix "-" indicates excluding.
//
// It would dismatch any rune when given no arguments.
//
// Panics if any range name is undefined.
func U(ranges ...string) Pattern {
	// preprocessing names
	iset := make(map[string]bool)
	eset := make(map[string]bool)
	for _, name := range ranges {
		if strings.HasPrefix(name, "-") {
			eset[name[1:]] = true
		} else {
			iset[name] = true
		}
	}
	inames := make([]string, 0, len(iset))
	for name := range iset {
		inames = append(inames, name)
	}
	enames := make([]string, 0, len(eset))
	for name := range eset {
		enames = append(enames, name)
	}

	// choose underlying type
	switch {
	case len(inames) == 0 && len(enames) == 0:
		return False
	case len(enames) == 0:
		pat := &patternUnicodeRanges{not: false}
		err := pat.set(inames)
		if err != nil {
			panic(err)
		}
		return pat
	case len(inames) == 0:
		pat := &patternUnicodeRanges{not: true}
		err := pat.set(enames)
		if err != nil {
			panic(err)
		}
		return pat
	default:
		pat := &patternUnicodeRangesWithExcluding{}
		pat.include.not = false
		err := pat.include.set(inames)
		if err != nil {
			panic(err)
		}
		pat.exclude.not = true
		err = pat.exclude.set(enames)
		if err != nil {
			panic(err)
		}
		return pat
	}
}

// Matches any rune.
func (patternAnyRune) match(ctx *context) error {
	_, n := ctx.readRune()
	if n == 0 {
		return ctx.returnsPredication(false)
	}
	ctx.consume(n)
	return ctx.returnsMatched()
}

// Matches a rune in/not in rune set.
func (pat *patternRuneSet) match(ctx *context) error {
	r, n := ctx.readRune()
	if n != 0 && pat.has(r) {
		ctx.consume(n)
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

func (pat *patternRuneSet) set(charset string) {
	pat.charset = []rune(charset)
	if len(pat.charset) > runeSetSizeThreshold {
		// preprocessing for binary search
		sort.Sort(&runesSorter{pat.charset})
		if len(pat.charset) > 0 {
			newrunes := make([]rune, 1, len(pat.charset))
			last := pat.charset[0]
			newrunes[0] = last
			for _, r := range pat.charset[1:] {
				if r != last {
					newrunes = append(newrunes, r)
					last = r
				}
			}
		}
	}
}

func (pat *patternRuneSet) has(r rune) bool {
	ok := false
	if len(pat.charset) > runeSetSizeThreshold {
		// use binary search
		i, j := 0, len(pat.charset)
		for i < j {
			m := i + (j-i)/2
			if r == pat.charset[m] {
				ok = true
				break
			} else if r > pat.charset[m] {
				i = m + 1
			} else {
				j = m
			}
		}
	} else {
		// linear search
		for i := range pat.charset {
			if r == pat.charset[i] {
				ok = true
				break
			}
		}
	}

	if pat.not {
		ok = !ok
	}
	return ok
}

// rune set sorter
type runesSorter struct {
	data []rune
}

func (rs *runesSorter) Len() int {
	return len(rs.data)
}

func (rs *runesSorter) Less(i, j int) bool {
	return rs.data[i] < rs.data[j]
}

func (rs *runesSorter) Swap(i, j int) {
	rs.data[i], rs.data[j] = rs.data[j], rs.data[i]
}

// Matches a rune in/not in range.
func (pat *patternRuneRange) match(ctx *context) error {
	r, n := ctx.readRune()
	if n != 0 && pat.has(r) {
		ctx.consume(n)
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

func (pat *patternRuneRange) has(r rune) bool {
	ok := false
	for _, pair := range pat.ranges {
		if r >= pair.low && r <= pair.high {
			ok = true
			break
		}
	}

	if pat.not {
		ok = !ok
	}
	return ok
}

// Matches a rune in/not in unicode ranges.
func (pat *patternUnicodeRanges) match(ctx *context) error {
	r, n := ctx.readRune()

	if n != 0 && pat.has(r) {
		ctx.consume(n)
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

func (pat *patternUnicodeRanges) set(names []string) error {
	var ranges []*unicode.RangeTable
	for _, name := range names {
		var ok bool
		ranges, ok = appendUnicodeRanges(ranges, name)
		if !ok {
			return errorUndefinedUnicodeRanges(name)
		}
	}
	pat.names = names
	pat.ranges = ranges
	return nil
}

func (pat *patternUnicodeRanges) has(r rune) bool {
	ok := false
	if len(pat.ranges) > 0 {
		ok = unicode.In(r, pat.ranges...)
	}

	if pat.not {
		ok = !ok
	}
	return ok
}

// Matches a rune in some unicode ranges while not in some other ranges.
func (pat *patternUnicodeRangesWithExcluding) match(ctx *context) error {
	r, n := ctx.readRune()
	if n != 0 && pat.has(r) {
		ctx.consume(n)
		return ctx.returnsMatched()
	}
	return ctx.returnsPredication(false)
}

func (pat *patternUnicodeRangesWithExcluding) has(r rune) bool {
	return pat.include.has(r) && pat.exclude.has(r)
}

func (patternAnyRune) String() string {
	return "#."
}

func (pat *patternRuneRange) String() string {
	strs := make([]string, len(pat.ranges))
	for i := range pat.ranges {
		strs[i] = fmt.Sprintf("%q..%q",
			pat.ranges[i].low, pat.ranges[i].high)
	}

	if pat.not {
		return fmt.Sprintf("#<-%s>", strings.Join(strs, "-"))
	}
	return fmt.Sprintf("#<%s>", strings.Join(strs, "+"))
}

func (pat *patternRuneSet) String() string {
	if pat.not {
		return fmt.Sprintf("#-%q", string(pat.charset))
	}
	return fmt.Sprintf("#%q", string(pat.charset))
}

func (pat *patternUnicodeRanges) String() string {
	if pat.not {
		if len(pat.names) == 0 {
			return "#."
		}
		return fmt.Sprintf("#[-%s]", strings.Join(pat.names, "-"))
	}
	return fmt.Sprintf("#[%s]", strings.Join(pat.names, "+"))
}

func (pat *patternUnicodeRangesWithExcluding) String() string {
	switch {
	case len(pat.include.names) == 0 && len(pat.exclude.names) == 0:
		return "#[]"
	case len(pat.include.names) != 0 && len(pat.exclude.names) == 0:
		return fmt.Sprintf("#[%s]", strings.Join(pat.include.names, "+"))
	case len(pat.include.names) == 0 && len(pat.exclude.names) != 0:
		return fmt.Sprintf("#[-%s]", strings.Join(pat.exclude.names, "-"))
	default:
		return fmt.Sprintf("#[%s-%s]",
			strings.Join(pat.include.names, "+"), strings.Join(pat.exclude.names, "-"))
	}
}

var (
	// Custom unicode range names of *unicode.RangeTable
	unicodeRangeAliases = map[string]*unicode.RangeTable{
		"Upper":     unicode.Lu,
		"Lower":     unicode.Ll,
		"Title":     unicode.Lt,
		"Letter":    unicode.L,
		"Mark":      unicode.M,
		"Number":    unicode.N,
		"Digit":     unicode.Nd,
		"Punct":     unicode.P,
		"Symbol":    unicode.S,
		"Separator": unicode.Z,
		"Other":     unicode.C,
		"Control":   unicode.Cc,
	}

	// Custom unicode range names of []*unicode.RangeTable
	unicodeRangeSliceAliases = map[string][]*unicode.RangeTable{
		"Graphic": unicode.GraphicRanges,
		"Print":   unicode.GraphicRanges,
	}
)

// IsUnicodeRangeName checks if unicode range name is valid.
// Available ranges names (case sensitive) are:
//   Upper, Lower, Title, Letter, Mark, Number Digit, Punct, Symbol,
//   Separator, Other, Control, Graphic, Print,
//   <names in unicode.Properties>   // e.g. White_Space, Quotation_Mark
//   <names in unicode.Scripts>,     // e.g. Latin, Greek
//   <names in unicode.Categories>,  // e.g. L, Lm, N, Nd
func IsUnicodeRangeName(name string) bool {
	if _, ok := unicodeRangeAliases[name]; ok {
		return true
	}
	if _, ok := unicodeRangeSliceAliases[name]; ok {
		return true
	}
	if _, ok := unicode.Properties[name]; ok {
		return true
	}
	if _, ok := unicode.Scripts[name]; ok {
		return true
	}
	if _, ok := unicode.Categories[name]; ok {
		return true
	}
	return false
}

// Helper function looks up and appends unicode ranges.
func appendUnicodeRanges(ranges []*unicode.RangeTable, name string) ([]*unicode.RangeTable, bool) {
	if r, ok := unicodeRangeAliases[name]; ok {
		return append(ranges, r), true
	}
	if rs, ok := unicodeRangeSliceAliases[name]; ok {
		return append(ranges, rs...), true
	}
	if r, ok := unicode.Properties[name]; ok {
		return append(ranges, r), true
	}
	if r, ok := unicode.Scripts[name]; ok {
		return append(ranges, r), true
	}
	if r, ok := unicode.Categories[name]; ok {
		return append(ranges, r), true
	}
	return ranges, false
}
