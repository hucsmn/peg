// Package peg implements the Parsing Expression Grammars inspired by LPeg.
//
// This package implements the Parsing Expression Grammars (PEGs),
// a powerful tool for pattern matching and writing top-down parsers.
// PEGs were designed to focus on expressing the match or parsing progress,
// rather than to describe what text should be matched as regexps do.
// The package was strongly influenced by LPeg for lua, see:
// http://www.inf.puc-rio.br/~roberto/lpeg/.
// Take a look at it for further readings.
//
// Overlook
//
// There were four methods for PEGs pattern matching:
//
//     MatchedPrefix(pat, text) (prefix, ok)
//     IsFullMatched(pat, text) ok
//     Parse(pat, text) (captures, err)
//     Match(pat, text) (result, err)
//
// The most general one is `config.Match(pat, text)`, which returns a `*Result`
// typed match result and an error if any error occured.
//
// The config tells the max recursion level, the max repeatition times and
// whether grouping or capturing is enabled. The default config enables
// both grouping and capturing, while limits for recursion and repeat are
// setup to DefaultCallstackLimit and DefaultRepeatLimit.
//
// The result of `config.Match(pat, text)` contains:
// whether pattern was matched, how many bytes were matched,
// the saved groups and the parser captures.
// Saved groups are text pieces captured with an optional name.
// Parser captures are parse trees or user defined structures constructed
// during the parsing process.
//
// Note that, both `MatchedPrefix` and `IsFullMatched` disables capturing.
// That is, the side effects of user defined constructors won't be triggered.
//
// Categories of patterns
//
// Basic patterns, which matches a single rune or a piece of text,
// are listed below:
//
//     T(text), TI(insensitivetext), TS(text, ...), TSI(insensitivetext, ...)
//     Dot, S(runes), NS(excluderunes), R(low, high, ...), NR(low, high, ...)
//     U(unicoderangename)
//
// Patterns are combined by sequence or alternation:
//
//     Seq(pat, ...), Alt(pat, ...)
//
// Predicators test if pattern would be matched, but consume no text:
//
//     True, False, SOL, EOL, EOF
//     B(text), Test(pat), Not(pat), And(pat...), Or(pat...)
//     When(cond, pat), If(cond, yes, no), Switch(cond, pat, ..., [otherwise])
//
// Available pattern qualifiers are:
//
//     Skip(n), Until(pat), UntilEndOf(pat)
//     Q0(pat), Q1(pat), Qn(atleast, pat)
//     Q01(pat), Q0n(atmost, pat), Qnn(exact, pat), Qmn(from, to, pat)
//
// Pattern where item separated by sep could be expressed using:
//
//     J0(item, sep), J1(item, sep), Jn(atleast, item, sep)
//     J0n(atmost, item, sep), Jnn(exact, item, sep), Jmn(from, to, item, sep)
//
// Functionalities for groups, references, triggers and injectors:
//
//     G(pat), NG(groupname, pat)
//     Ref(groupname), RefB(groupname)
//     Trigger(hook, pat), Save(pointer, pat),
//     Send(channel, pat), SendToken(channel, tokentype, pat)
//     Inject(injector, pat), Check(checker, pat), Trunc(maxrune, pat)
//
// Functionalities for grammars and parsing captures:
//
//     Let(scope, pat), V(varname), CV(varname), CK(tokentype, pat)
//     CC(nontermcons, pat), CT(termcons, pat)
//
// Common mistakes
//
// Greedy qualifiers:
//
// The qualifiers are designed to be greedy. Thus, considering the pattern
// `Seq(Q0(A), B)`, text supposed to be matched by `B` could be swallowed
// ahead of time by the preceding `A`, which is usually unexpected.
// It is recommended to wrap `A` with an additional assertion to avoid this.
//
// For example, `Seq(Q0(R('0', '9')), S("02468"), T(" is even"))` is incorrect,
// because the greedy `Q0(R('0', '9'))` would consume the last digit, thus the
// following `S("02468")` would always dismatch. To make everything right,
// `Q0(R('0', '9'))` should be replaced by a pattern like
// `Q0(Seq(R('0', '9'), Test(R('0', '9'))))` (assert one digit follow it),
// which won't consume the last digit.
//
// Unreachable branches:
//
// Branch of `Seq` or `Alt` could be unreachable, considering that Seq searches
// the first dismatch in the sequence, while Alt searches the first match in the
// choices. Thus, a pattern like `Alt(T("match"), T("match more"))` would get an
// unexpected match result, becuase longer patterns are not in prior order.
//
// Infinite loops:
//
// Any pattern which could macth an empty string should not be nested inside
// qualifiers like `Q0`, `Q1`, `Qn`, for this would cause infinite loops.
//
// For example, `Q1(True)` or `Q0(Q0(T("not empty")))` would loop until
// `config.RepeatLimit` is reached.
//
// Left recursion:
//
// PEG parsers are top-down, that is, the grammar rules would be expanded
// immediately, thus a left recursion never terminates until
// `config.CallstackLimit` is reached.
//
// For example, `Let(map[string]Pattern{"var": Seq(T("A"), V("var"))}, V("var"))`
// terminates, while
// `Let(map[string]Pattern{"var": Seq(V("var"), T("A"))}, V("var"))` won't
// terminate until `CallstackLimit` is reached.
package peg // import "github.com/hucsmn/peg"

import (
	"fmt"
	"strings"
)

// Default limits of pattern matching.
const (
	DefaultCallstackLimit = 500
	DefaultRepeatLimit    = 500
)

var (
	defaultConfig = Config{
		CallstackLimit:            DefaultCallstackLimit,
		RepeatLimit:               DefaultRepeatLimit,
		DisableLineColumnCounting: false,
		DisableGrouping:           false,
		DisableCapturing:          false,
	}
)

type (
	// Pattern is the tree representation for Parse Grammar Expression.
	Pattern interface {
		match(ctx *context) error
		String() string
	}

	// Config contains configration for pattern matching.
	Config struct {
		// Maximum callstack size, zero or negative for unlimited.
		CallstackLimit int

		// Maximum qualifier repeatition times, zero or negative for unlimited.
		RepeatLimit int

		// Determines if the position calculation is disabled.
		DisableLineColumnCounting bool

		// Determines if grouping is disabled.
		DisableGrouping bool

		// Determines if parse tree capturing is disabled.
		DisableCapturing bool
	}

	// Result stores the results from pattern matching.
	Result struct {
		// Is pattern matched and how many bytes matched.
		Ok bool
		N  int

		// Grouped text pieces with optional names.
		Groups      []string
		NamedGroups map[string]string

		// Parse captures.
		Captures []Capture
	}

	// Capture stores structures from parse capturing.
	// User defined structures (the types implemented Capture interface other
	// than the predefined Variable type and Token type) are constructed by
	// customed TerminalConstructor or NonTerminalConstructor.
	Capture interface {
		// IsTerminal tells if it is a terminal type.
		IsTerminal() bool
	}

	// Variable is a predefined non-terminal type for PEG variable capturing.
	Variable struct {
		Name string
		Subs []Capture
	}

	// Token is a predefined terminal type stores a piece of typed text
	// and its position in the source text.
	Token struct {
		Type     int
		Value    string
		Position Position
	}

	// TerminalConstructor is customed terminal type constructor.
	TerminalConstructor func(string, Position) (Capture, error)

	// NonTerminalConstructor is customed non-terminal type constructor.
	NonTerminalConstructor func([]Capture) (Capture, error)
)

// MatchedPrefix returns the matched prefix of text when successfully matched.
func MatchedPrefix(pat Pattern, text string) (prefix string, ok bool) {
	return defaultConfig.MatchedPrefix(pat, text)
}

// IsFullMatched tells if given pattern matches the full text.
// It is recommended to use Seq(Alt(...), EOF) rather than use Alt(...) when
// testing IsFullMatched.
// For example, IsFullMatched(Alt(T("match"), T("match more")), "match more")
// returns false rather than true counter-intuitively.
func IsFullMatched(pat Pattern, text string) bool {
	return defaultConfig.IsFullMatched(pat, text)
}

// Parse runs pattern matching on given text, guaranteeing that the text must
// only be full-matched when success.
func Parse(pat Pattern, text string) (caps []Capture, err error) {
	return defaultConfig.Parse(pat, text)
}

// Match runs pattern matching on given text, using the default configuration.
// The default configuration uses DefaultCallstackLimit and DefaultLoopLimit,
// while line-column counting, grouping and parse capturing is enabled.
// Returns nil result if any error occurs.
func Match(pat Pattern, text string) (result *Result, err error) {
	return defaultConfig.Match(pat, text)
}

// MatchedPrefix returns the matched prefix of text when successfully matched.
func (cfg Config) MatchedPrefix(pat Pattern, text string) (prefix string, ok bool) {
	// disable capturing.
	config := cfg
	config.DisableLineColumnCounting = true
	config.DisableCapturing = true
	r, err := config.Match(pat, text)
	if err != nil || !r.Ok {
		return "", false
	}
	return text[:r.N], true
}

// IsFullMatched tells if given pattern matches the full text.
// It is recommended to use Seq(Alt(...), EOF) rather than use Alt(...) when
// testing IsFullMatched.
// For example, IsFullMatched(Alt(T("match"), T("match more")), "match more")
// returns false rather than true counter-intuitively.
func (cfg Config) IsFullMatched(pat Pattern, text string) bool {
	// disable capturing.
	config := cfg
	config.DisableLineColumnCounting = true
	config.DisableCapturing = true
	r, err := config.Match(pat, text)
	return err == nil && r.Ok && r.N == len(text)
}

// Parse runs pattern matching on given text, guaranteeing that the text must
// only be full-matched when success.
func (cfg Config) Parse(pat Pattern, text string) (caps []Capture, err error) {
	// enable capturing.
	config := cfg
	config.DisableLineColumnCounting = false
	config.DisableCapturing = false
	r, err := config.Match(pat, text)
	if err != nil {
		return nil, err
	}
	if !r.Ok {
		return nil, errorDismatch
	}
	if r.N != len(text) {
		return nil, errorNotFullMatched
	}
	return r.Captures, nil
}

// Match runs pattern matching on given text, using the default configuration.
// The default configuration uses DefaultCallstackLimit and DefaultLoopLimit,
// while line-column counting, grouping and parse capturing is enabled.
// Returns nil result if any error occurs.
func (cfg Config) Match(pat Pattern, text string) (result *Result, err error) {
	if pat == nil {
		return nil, errorNilMainPattern
	}

	ctx := newContext(pat, text, cfg)
	err = ctx.match()
	if err != nil {
		return nil, err
	}

	if ctx.ret.ok {
		return &Result{
			Ok:          true,
			N:           ctx.ret.n,
			Groups:      ctx.groups,
			NamedGroups: ctx.namedGroups,
			Captures:    ctx.capstack[0].args,
		}, nil
	}
	return &Result{
		Ok:          false,
		N:           0,
		Groups:      nil,
		NamedGroups: nil,
		Captures:    nil,
	}, nil
}

// IsTerminal method of the Variable type always returns false.
func (v *Variable) IsTerminal() bool {
	return false
}

// IsTerminal method of the Token type always returns true.
func (tok *Token) IsTerminal() bool {
	return true
}

func (v *Variable) String() string {
	strs := make([]string, len(v.Subs))
	for i := range v.Subs {
		strs[i] = fmt.Sprint(v.Subs[i])
	}
	return fmt.Sprintf("%s(%s)", v.Name, strings.Join(strs, ", "))
}

func (tok *Token) String() string {
	return fmt.Sprintf("token_%d%q@%s",
		tok.Type, tok.Value, tok.Position.String())
}
