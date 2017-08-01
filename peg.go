// Package peg implements the Parsing Expression Grammars inspired by LPeg.
package peg

import (
	"fmt"
	"strings"
)

// Default limits of pattern matching.
const (
	DefaultCallstackLimit = 500
	DefaultLoopLimit      = 500
)

var (
	defaultConfig = Config{
		CallstackLimit:   DefaultCallstackLimit,
		LoopLimit:        DefaultLoopLimit,
		DisableGrouping:  false,
		DisableCapturing: false,
	}
)

type (
	// Pattern is tree of Parse Grammar Expression.
	Pattern interface {
		match(ctx *context) error
	}

	// Config is configration for pattern matching.
	Config struct {
		// Maximum nesting number, <= 0 for unlimited.
		CallstackLimit int

		// Maximum looping number, <= 0 for unlimited.
		LoopLimit int

		// Disable grouping.
		DisableGrouping bool

		// Disable capturing.
		DisableCapturing bool
	}

	// Result is the match result.
	Result struct {
		Ok bool // Whether the whole pattern is matched.
		N  int  // N bytes matched.

		// Groups
		Groups      []string
		NamedGroups map[string]string

		// Captures
		Captures []Capture
	}

	// Capture is PEG capturing tree.
	Capture interface {
		// IsTerminal tells if it was a terminal.
		IsTerminal() bool
	}

	// Variable is a non-terminal type, capturing a PEG variable.
	Variable struct {
		Name string
		Subs []Capture
	}

	// Token is a terminal type.
	Token struct {
		Type     int
		Value    string
		Position Position
	}

	// TerminalConstructor is customed terminal constructor.
	TerminalConstructor func(string, Position) (Capture, error)

	// NonTerminalConstructor is customed non-terminal constructor.
	NonTerminalConstructor func([]Capture) (Capture, error)
)

// MatchedPrefix returns the matched prefix of text.
func MatchedPrefix(pat Pattern, text string) (prefix string, ok bool) {
	config := defaultConfig
	config.DisableCapturing = true
	r, err := ConfiguredMatch(config, pat, text)
	if err != nil || !r.Ok {
		return "", false
	}
	return text[:r.N], true
}

// IsFullMatched tests if pattern matches full text.
func IsFullMatched(pat Pattern, text string) bool {
	config := defaultConfig
	config.DisableCapturing = true
	r, err := ConfiguredMatch(config, pat, text)
	return err == nil && r.Ok && r.N == len(text)
}

// Match matches pattern against text, then returns the result of matching.
func Match(pat Pattern, text string) (result *Result, err error) {
	return ConfiguredMatch(defaultConfig, pat, text)
}

// ConfiguredMatch matches pattern against text with given configuration,
// then returns the result of matching.
func ConfiguredMatch(config Config, pat Pattern, text string) (result *Result, err error) {
	if pat == nil {
		return nil, errorNilMainPattern
	}

	ctx := newContext(pat, text, config)
	err = ctx.match()
	if err != nil {
		return nil, err
	}

	return &Result{
		Ok:          ctx.ret.ok,
		N:           ctx.ret.n,
		Groups:      ctx.groups,
		NamedGroups: ctx.namedGroups,
		Captures:    ctx.capstack[0].args,
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
	return fmt.Sprintf("token_%d%q@%s", tok.Type, tok.Value, tok.Position)
}
