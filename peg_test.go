package peg

import "testing"

type matchTestData struct {
	text   string
	ok     bool
	prefix string
	full   bool
	pat    Pattern
}

func runMatchTestData(t *testing.T, data matchTestData) {
	prefix, ok := MatchedPrefix(data.pat, data.text)
	full := IsFullMatched(data.pat, data.text)
	if ok != data.ok {
		t.Errorf("RESULT DISMACTH: MatchedPrefix(%s, %q) => (%q, %t != %t)\n",
			data.pat, data.text, prefix, ok, data.ok)
		return
	}
	if prefix != data.prefix {
		t.Errorf("RESULT DISMACTH: MatchedPrefix(%s, %q) => (%q != %q, %t)\n",
			data.pat, data.text, prefix, data.prefix, ok)
		return
	}
	if full != data.full {
		t.Errorf("RESULT DISMACTH: IsFullMatched(%s, %q) => %t != %t\n",
			data.pat, data.text, full, data.full)
		return
	}
}

// Tests MatchedPrefix, IsFullMatched.
func TestMatchedPrefixAndIsFullMatched(t *testing.T) {
	balance := Let(
		map[string]Pattern{
			"S": Alt(Seq(T("A"), CV("B")), Seq(T("B"), CV("A")), T("")),
			"A": Alt(Seq(T("A"), CV("S")), Seq(T("B"), CV("A"), CV("A"))),
			"B": Alt(Seq(T("B"), CV("S")), Seq(T("A"), CV("B"), CV("B"))),
		},
		CV("S"))
	data := []matchTestData{
		{"", true, "", true, True},
		{"A", true, "", false, True},
		{"", false, "", false, Qmn(1, 3, T("A"))},
		{"A", true, "A", true, Qmn(1, 3, T("A"))},
		{"AA", true, "AA", true, Qmn(1, 3, T("A"))},
		{"AAA", true, "AAA", true, Qmn(1, 3, T("A"))},
		{"AAAA", true, "AAA", false, Qmn(1, 3, T("A"))},

		// Infinite loop/recursion causes error.
		{"", false, "", false, Q0(True)},
		{"", false, "", false, Let(map[string]Pattern{"var": V("var")}, V("var"))},

		// test if grouping and capturing fails when configured DisableCapturing = true.
		{"AABA", true, "AA", false, Seq(NG("first", Dot), Q0(Ref("first")))},
		{"AABA", true, "", false, balance},
		{"ABBAB", true, "ABBA", false, balance},
		{"ABBABA", true, "ABBABA", true, balance},
	}

	for _, d := range data {
		runMatchTestData(t, d)
	}
}
