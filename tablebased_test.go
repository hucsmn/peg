package peg

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// Test pattern match results, errors, groups and captures.
type patternTestData struct {
	text   string
	ok     bool
	n      int
	fail   bool
	groups string
	caps   string
	pat    Pattern
}

func runPatternTestData(t *testing.T, data patternTestData) {
	r, err := Match(data.pat, data.text)
	if err != nil {
		if data.fail {
			t.Logf("INFO: the expected failure `%s` occurs when match(%s, %q)", err, data.pat, data.text)
		} else {
			t.Errorf("UNEXPECTED ERROR `%s` occurs when match(%s, %q)", err, data.pat, data.text)
		}
		return
	} else if data.fail {
		t.Errorf("EXPECTED BUT NO ERROR occurs when match(%s, %q)", data.pat, data.text)
		return
	}

	if r.Ok != data.ok {
		t.Errorf("RESULT DISMATCH: match(%s, %q) => (%v != %v, %d)",
			data.pat, data.text, r.Ok, data.ok, r.N)
		return
	}
	if data.ok && r.N != data.n {
		t.Errorf("RESULT DISMATCH: match(%s, %q) => (%v, %d != %d)",
			data.pat, data.text, r.Ok, r.N, data.n)
		return
	}

	grps, ngrps := parseGroupsString(data.groups)
	grouping0 := formatGrouping(r.Groups, r.NamedGroups)
	grouping1 := formatGrouping(grps, ngrps)
	if grouping0 != grouping1 {
		t.Errorf("GROUPS DISMATCH: match(%s, %q) => {%s} != {%s}",
			data.pat, data.text, grouping0, grouping1)
		return
	}

	caps := formatCapList(r.Captures)
	if caps != data.caps {
		t.Errorf("CAPTURES DISMATCH: match(%s, %q) => %q != %q",
			data.pat, data.text, caps, data.caps)
		return
	}
}

func parseGroupsString(groups string) (grps []string, ngrps map[string]string) {
	if len(groups) == 0 {
		return
	}
	ngrps = make(map[string]string)
	for _, item := range strings.Split(groups, ",") {
		fields := strings.SplitN(item, "=", 2)
		if len(fields) < 2 {
			s, _ := strconv.Unquote(item)
			grps = append(grps, s)
		} else {
			k, _ := strconv.Unquote(fields[0])
			v, _ := strconv.Unquote(fields[1])
			ngrps[k] = v
		}
	}
	return grps, ngrps
}

func formatGrouping(grps []string, ngrps map[string]string) string {
	gstrs := make([]string, 0, len(grps))
	for _, s := range grps {
		gstrs = append(gstrs, strconv.Quote(s))
	}
	nstrs := make([]string, 0, len(ngrps))
	for k, v := range ngrps {
		nstrs = append(nstrs, fmt.Sprintf("%q=%q", k, v))
	}
	sort.Strings(nstrs)
	gstr := strings.Join(gstrs, ",")
	nstr := strings.Join(nstrs, ",")
	switch {
	case len(gstr) == 0 && len(nstr) == 0:
		return ""
	case len(gstr) == 0 && len(nstr) != 0:
		return nstr
	case len(gstr) != 0 && len(nstr) == 0:
		return gstr
	default:
		return nstr + "," + gstr
	}
}

func formatCapList(caps []Capture) string {
	strs := make([]string, len(caps))
	for i := range caps {
		strs[i] = formatCap(caps[i])
	}
	return strings.Join(strs, ", ")
}

func formatCap(cap Capture) string {
	if cap.IsTerminal() {
		if tok, ok := cap.(*Token); ok {
			return fmt.Sprintf("<%d%q>", tok.Type, tok.Value)
		}
		return fmt.Sprintf("<%v>", cap)
	} else if v, ok := cap.(*Variable); ok {

		return fmt.Sprintf("%s(%s)", v.Name, formatCapList(v.Subs))
	}
	return fmt.Sprintf("%v", cap)
}

// Test pattern match side effects.
type (
	sideEffectsTestContext struct {
		states []string
	}

	sideEffectsTestData struct {
		text string
		ok   bool
		n    int
		fail bool

		state   string
		metapat func(*sideEffectsTestContext) Pattern
	}
)

func runSideEffectsTestData(t *testing.T, ctx *sideEffectsTestContext, data sideEffectsTestData) {
	ctx.clear()
	pat := data.metapat(ctx)
	r, err := Match(pat, data.text)
	if err != nil {
		if data.fail {
			t.Logf("INFO: the expected failure `%s` occurs when match(%s, %q)", err, pat, data.text)
		} else {
			t.Errorf("UNEXPECTED ERROR `%s` occurs when match(%s, %q)", err, pat, data.text)
		}
		return
	} else if data.fail {
		t.Errorf("EXPECTED BUT NO ERROR occurs when match(%s, %q)", pat, data.text)
		return
	}

	if r.Ok != data.ok {
		t.Errorf("RESULT DISMATCH: match(%s, %q) => (%v != %v, %d)",
			pat, data.text, r.Ok, data.ok, r.N)
		return
	}
	if data.ok && r.N != data.n {
		t.Errorf("RESULT DISMATCH: match(%s, %q) => (%v, %d != %d)",
			pat, data.text, r.Ok, r.N, data.n)
		return
	}

	state := ctx.format()
	if state != data.state {
		t.Errorf("SIDE EFFECTS DISMATCH: match(%s, %q) => %q != %q",
			pat, data.text, state, data.state)
		return
	}
}

func newSideEffectsTestContext() *sideEffectsTestContext {
	return &sideEffectsTestContext{}
}

func (ctx *sideEffectsTestContext) clear() {
	if ctx.states != nil {
		ctx.states = ctx.states[:0]
	}
}

func (ctx *sideEffectsTestContext) store(value string) {
	ctx.states = append(ctx.states, value)
}

func (ctx *sideEffectsTestContext) format() string {
	strs := make([]string, len(ctx.states))
	for i := range ctx.states {
		strs[i] = quoteState(ctx.states[i])
	}
	return "{" + strings.Join(strs, ", ") + "}"
}

func quoteState(s string) string {
	const (
		specials = ",{} \\'"
		noescape = ",{} "
	)

	lit := ""
	for {
		i := strings.IndexAny(s, specials)

		if i < 0 {
			lit += s
			break
		}

		lit += s[:i]
		s = s[i:]
		r, n := utf8.DecodeRuneInString(s)
		if strings.ContainsRune(noescape, r) {
			lit += s[:n]
		} else {
			lit += "\\" + s[:n]
		}
		s = s[n:]
	}
	return "'" + lit + "'"
}
