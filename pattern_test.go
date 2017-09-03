package peg

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
)

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
	grouping0 := showGrouping(r.Groups, r.NamedGroups)
	grouping1 := showGrouping(grps, ngrps)
	if grouping0 != grouping1 {
		t.Errorf("GROUPS DISMATCH: match(%s, %q) => {%s} != {%s}",
			data.pat, data.text, grouping0, grouping1)
		return
	}

	caps := showCapList(r.Captures)
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

func showGrouping(grps []string, ngrps map[string]string) string {
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

func showCapList(caps []Capture) string {
	strs := make([]string, len(caps))
	for i := range caps {
		strs[i] = showCap(caps[i])
	}
	return strings.Join(strs, ", ")
}

func showCap(cap Capture) string {
	if cap.IsTerminal() {
		if tok, ok := cap.(*Token); ok {
			return fmt.Sprintf("<%d%q>", tok.Type, tok.Value)
		}
		return fmt.Sprintf("<%v>", cap)
	} else if v, ok := cap.(*Variable); ok {

		return fmt.Sprintf("%s(%s)", v.Name, showCapList(v.Subs))
	}
	return fmt.Sprintf("%v", cap)
}

// Tests T, TI, TS, TSI, True, False, Dot, R, NR, S, NS, U.
func TestBasicPatterns(t *testing.T) {
	ASC := "Ec|X,^X]ibC)$[S42\\]J`.rMe_@a2mEiri=>\\Dt0xaG^qc<C0qNnqzf_GVt^=SrMostRm5_$\x7flO&VZ'Q&a6h0;@[\\/O$oq/NFDK"
	ARB := "ؿۼؗ٠خڥۊٛۉڭڏ٥ۧٽڞ۫ڤؗل؀ږڊؘؙۨ۰پڅچؙ٠ۖؗ٤ِؔڷۮٴؒڇەڂٸ۳ؚدۺ١َ۵إ۽۾٤صګڦش٤۠ۂؽبضٝ۽ؽۑٕ۶ۧ؃ٖ؞ٙٮێۉا؟ث؈٬ۜھڃو؉ڈٍ؄ۄۻۥ٪ۭٱه"
	CHS := "以仇价亓仕仡仯亅什亞亟仇亦亊亣仹任产仨仱亿仾仌事仔仐亽亓仨亇仯仉仔云亄仅从亊争亹仑从亾亊亟仢二亩亁仺令亏亶交仰仫仃亵仸亂京予亙亘亷仩他亗亍亷从二仰亏亝亇仴亜仭仍亘仪亹仆亹亘亳亙人亜仓亚亂亀产仼亀仔仯亍"
	JPN := "うすせちぎそづぁぉぅだわりょめぽゐぷぷぶのらづうをざもうためそおどそをごんぁぃへにぞゑびこかまゅょゑすぢけをだづきぱびほきおとごぴつさひにろのわたるっぷぴつくかぼきふむぐゆよぷしみじぷきゆぇぼぽぬぽむ"
	KRN := "픀쉀붏뙘췒삹늭쏐뵢쾙푃멥묶쿔봼즱끇섭챊힜싰윂묷뢥롹잌룈샳껪꼤홬퀩쁇뾩뤽눿아땧뜚됂퐭걻큼텿끝꾘븏뉿촂빅뗵홌굛찄햂굟흍뎍즠쬫쌂얝뮂긌뱍쉸경뺀횞찚쿤퀃핿샶패풄쐘뇙쉼쾕썤똵슻묌흥댍풲힆볎떄끙뷟큺쬔뛣곺떡뫀쓷씂"

	data := []patternTestData{
		{"", true, 0, false, ``, ``, T("")},
		{"abcdefg", true, 3, false, ``, ``, T("abc")},
		{"中国人", true, 6, false, ``, ``, T("中国")},
		{"A", true, 1, false, ``, ``, TI("a")},
		{"a", true, 1, false, ``, ``, TI("A")},
		{"Ё", true, 2, false, ``, ``, TI("ё")},
		{"ſK", true, 5, false, ``, ``, TI("ſK")},
		{"ß", true, 2, false, ``, ``, TI("ẞ")},
		{"ѣ", true, 2, false, ``, ``, TI("Ѣ")},
		{"aAåÅA", true, 8, false, ``, ``, TI("aaÅåa")},

		{"01", true, 2, false, ``, ``, TS(strings.Split("0/01/011/012/021/123/1234", "/")...)},
		{"013", true, 2, false, ``, ``, TS(strings.Split("0/01/011/012/021/123/1234", "/")...)},
		{"022", true, 1, false, ``, ``, TS(strings.Split("0/01/011/012/021/123/1234", "/")...)},
		{"1234", true, 4, false, ``, ``, TS(strings.Split("0/01/011/012/021/123/1234", "/")...)},
		{"true", true, 4, false, ``, ``, TS("true", "false")},
		{"false", true, 5, false, ``, ``, TS("true", "false")},
		{"True", true, 4, false, ``, ``, TSI("true", "false")},
		{"False", true, 5, false, ``, ``, TSI("true", "false")},
		{"åÅA", true, 6, false, ``, ``, TSI("Ååa", "Bbb")},
		{"åÅA", true, 6, false, ``, ``, TSI("Ååa", "Bbb")},
		{"bBb", true, 3, false, ``, ``, TSI("Ååa", "Bbb")},
		{"å", true, 2, false, ``, ``, TSI("Ååa", "Å")},
		{"a", true, 1, false, ``, ``, TSI("aaa", "A")},

		{"", true, 0, false, ``, ``, True},
		{"a", true, 0, false, ``, ``, True},
		{"", false, 0, false, ``, ``, False},
		{"a", false, 0, false, ``, ``, False},

		{"", false, 0, false, ``, ``, Dot},
		{"a", true, 1, false, ``, ``, Dot},
		{"你好", true, 3, false, ``, ``, Dot},

		// "\xe4\xbd\x20" is a bad utf-8 sequence,
		// only one byte (unicode rune replacement) would be matched.
		{"\xe4\xbd\x20", true, 1, false, ``, ``, Dot},

		{"", false, 0, false, ``, ``, R('a', 'z')},
		{"word", true, 1, false, ``, ``, R('a', 'z')},
		{"h", true, 1, false, ``, ``, R('a', 'z', '0', '9')},
		{"5", true, 1, false, ``, ``, R('a', 'z', '0', '9')},
		{"H", false, 0, false, ``, ``, R('a', 'z', '0', '9')},
		{"你好", false, 0, false, ``, ``, R('a', 'z')},
		{"你好", true, 3, false, ``, ``, NR('a', 'z')},
		{"", false, 0, false, ``, ``, NR('a', 'z')},
		{"H", true, 1, false, ``, ``, NR('a', 'z', '0', '9')},
		{"h", false, 0, false, ``, ``, NR('a', 'z', '0', '9')},
		{"5", false, 0, false, ``, ``, NR('a', 'z', '0', '9')},

		{"", false, 0, false, ``, ``, S("abc")},
		{"a", true, 1, false, ``, ``, S("abc")},
		{"c", true, 1, false, ``, ``, S("abc")},
		{"好", true, 3, false, ``, ``, S("你好")},
		{"你", true, 3, false, ``, ``, S("你好")},
		{"x", true, 1, false, ``, ``, NS("abc")},
		{"你", true, 3, false, ``, ``, NS("abc")},
		{"你", false, 0, false, ``, ``, NS("你好")},
		{"中国", true, 3, false, ``, ``, NS("你好")},
		{"a", true, 1, false, ``, ``, S(ASC)},
		{"۵", true, 2, false, ``, ``, S(ARB)},
		{"京", true, 3, false, ``, ``, S(CHS)},
		{"め", true, 3, false, ``, ``, S(JPN)},
		{"아", true, 3, false, ``, ``, S(KRN)},

		{"你好", true, 3, false, ``, ``, U("Letter")},
		{"你好", true, 3, false, ``, ``, U("-Punct")},
		{"。", false, 0, false, ``, ``, U("-Punct")},
		{"、", true, 3, false, ``, ``, U("Punct")},
		{" ", true, 1, false, ``, ``, U("White_Space")},
		{"你好", false, 0, false, ``, ``, U("Han", "-Letter")},
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}

// Tests Seq, Alt.
func TestSeqAlt(t *testing.T) {
	data := []patternTestData{
		{"", true, 0, false, ``, ``, Seq(True, True)},
		{"", false, 0, false, ``, ``, Seq(True, False)},
		{"", false, 0, false, ``, ``, Seq(False, True)},
		{"", false, 0, false, ``, ``, Seq(False, False)},
		{"AB", true, 2, false, ``, ``, Seq(T("A"), T("B"))},
		{"AC", false, 0, false, ``, ``, Seq(T("A"), T("B"))},

		{"", true, 0, false, ``, ``, Alt(True, True)},
		{"", true, 0, false, ``, ``, Alt(True, False)},
		{"", true, 0, false, ``, ``, Alt(False, True)},
		{"", false, 0, false, ``, ``, Alt(False, False)},
		{"AB", true, 1, false, ``, ``, Alt(T("A"), T("B"))},
		{"AC", true, 1, false, ``, ``, Alt(T("A"), T("B"))},
		{"BC", true, 1, false, ``, ``, Alt(T("A"), T("B"))},
		{"CA", false, 0, false, ``, ``, Alt(T("A"), T("B"))},
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}

// Tests Test, Not, SOL, EOL, EOF, B,
//       And, Or, When, If, Switch
func TestPredicators(t *testing.T) {
	data := []patternTestData{
		{"", false, 0, false, ``, ``, Test(False)},
		{"", true, 0, false, ``, ``, Test(True)},
		{"", false, 0, false, ``, ``, Test(T("A"))},
		{"A", true, 0, false, ``, ``, Test(T("A"))},
		{"", true, 0, false, ``, ``, Not(False)},
		{"", false, 0, false, ``, ``, Not(True)},
		{"", true, 0, false, ``, ``, Not(T("A"))},
		{"A", false, 0, false, ``, ``, Not(T("A"))},

		{"", true, 0, false, ``, ``, SOL},
		{"A", true, 0, false, ``, ``, SOL},
		{"A", false, 0, false, ``, ``, Seq(Dot, SOL)},
		{"AB", false, 0, false, ``, ``, Seq(Dot, SOL)},
		{"A\nB", true, 2, false, ``, ``, Seq(Dot, Dot, SOL)},
		{"A\rB", true, 2, false, ``, ``, Seq(Dot, Dot, SOL)},
		{"A\r\nB", false, 0, false, ``, ``, Seq(Dot, Dot, SOL)},
		{"", true, 0, false, ``, ``, EOL},
		{"A", true, 1, false, ``, ``, Seq(Dot, EOL)},
		{"AB", false, 0, false, ``, ``, Seq(Dot, EOL)},
		{"A\nB", true, 1, false, ``, ``, Seq(Dot, EOL)},
		{"A\rB", true, 1, false, ``, ``, Seq(Dot, EOL)},
		{"A\r\nB", true, 1, false, ``, ``, Seq(Dot, EOL)},
		{"A\r\nB", false, 0, false, ``, ``, Seq(Dot, Dot, EOL)},
		{"", true, 0, false, ``, ``, EOF},
		{"A", false, 0, false, ``, ``, EOF},
		{"A", true, 1, false, ``, ``, Seq(Dot, EOF)},

		{"A", true, 1, false, ``, ``, Seq(Dot, B("A"))},
		{"A", false, 0, false, ``, ``, Seq(Dot, B("B"))},
		{"A", false, 0, false, ``, ``, Seq(Dot, B("AB"))},
		{"A", false, 0, false, ``, ``, Seq(Dot, B("BA"))},

		{"", true, 0, false, ``, ``, And(True, True)},
		{"", false, 0, false, ``, ``, And(True, False)},
		{"", false, 0, false, ``, ``, And(False, True)},
		{"", false, 0, false, ``, ``, And(False, False)},
		{"AB", false, 0, false, ``, ``, And(T("A"), T("B"))},
		{"AB", true, 0, false, ``, ``, And(T("AB"), Dot)},
		{"AB", false, 0, false, ``, ``, And(T("AB"), T("ABC"))},
		{"ABC", true, 0, false, ``, ``, And(T("AB"), T("ABC"))},

		{"", true, 0, false, ``, ``, Or(True, True)},
		{"", true, 0, false, ``, ``, Or(True, False)},
		{"", true, 0, false, ``, ``, Or(False, True)},
		{"", false, 0, false, ``, ``, Or(False, False)},
		{"AB", true, 0, false, ``, ``, Or(T("A"), T("B"))},
		{"AC", true, 0, false, ``, ``, Or(T("A"), T("B"))},
		{"BC", true, 0, false, ``, ``, Or(T("A"), T("B"))},
		{"CA", false, 0, false, ``, ``, Or(T("A"), T("B"))},
		{"AB", true, 0, false, ``, ``, Or(T("AB"), T("ABC"))},

		{"ABC", true, 2, false, ``, ``, Seq(Dot, When(B("A"), T("B")))},
		{"AB", true, 2, false, ``, ``, When(Not(T("0")), Seq(U("-White_Space"), U("-White_Space")))},
		{"0A", false, 0, false, ``, ``, When(Not(T("0")), Seq(U("-White_Space"), U("-White_Space")))},
		{"A0", true, 2, false, ``, ``, When(Not(T("0")), Seq(U("-White_Space"), U("-White_Space")))},
		{"0num", false, 0, false, ``, ``, If(R('0', '9'), T("num"), T("nan"))},
		{"0nan", false, 0, false, ``, ``, If(R('0', '9'), T("num"), T("nan"))},
		{"0num", true, 4, false, ``, ``, If(R('0', '9'), Seq(Dot, T("num")), Seq(Dot, T("nan")))},
		{"0nan", false, 0, false, ``, ``, If(R('0', '9'), Seq(Dot, T("num")), Seq(Dot, T("nan")))},
		{"anum", false, 0, false, ``, ``, If(R('0', '9'), Seq(Dot, T("num")), Seq(Dot, T("nan")))},
		{"anan", true, 4, false, ``, ``, If(R('0', '9'), Seq(Dot, T("num")), Seq(Dot, T("nan")))},
		{"0a", true, 2, false, ``, ``, Seq(Dot, Switch(B("0"), T("a"), B("1"), T("b"), T("c")))},
		{"0b", false, 0, false, ``, ``, Seq(Dot, Switch(B("0"), T("a"), B("1"), T("b"), T("c")))},
		{"1a", false, 0, false, ``, ``, Seq(Dot, Switch(B("0"), T("a"), B("1"), T("b"), T("c")))},
		{"1b", true, 2, false, ``, ``, Seq(Dot, Switch(B("0"), T("a"), B("1"), T("b"), T("c")))},
		{"2c", true, 2, false, ``, ``, Seq(Dot, Switch(B("0"), T("a"), B("1"), T("b"), T("c")))},
		{"2a", false, 0, false, ``, ``, Seq(Dot, Switch(B("0"), T("a"), B("1"), T("b"), T("c")))},
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}

// Tests Count, Until, UntilB, Q0, Q1, Qn, Q01, Qnn, Qmn, J0, Jn, Jnn, Jmn.
func TestQualifierAndJoin(t *testing.T) {
	data := []patternTestData{
		{"", false, 0, false, ``, ``, Skip(2)},
		{"A", false, 0, false, ``, ``, Skip(2)},
		{"AA", true, 2, false, ``, ``, Skip(2)},
		{"AAA", true, 2, false, ``, ``, Skip(2)},

		{"", false, 0, false, ``, ``, Until(T("."))},
		{"A", false, 0, false, ``, ``, Until(T("."))},
		{".", true, 0, false, ``, ``, Until(T("."))},
		{".B", true, 0, false, ``, ``, Until(T("."))},
		{"A.", true, 1, false, ``, ``, Until(T("."))},
		{"A.B", true, 1, false, ``, ``, Until(T("."))},
		{"AA.", true, 2, false, ``, ``, Until(T("."))},
		{"AA.B", true, 2, false, ``, ``, Until(T("."))},

		{"", false, 0, false, ``, ``, UntilB(T("."))},
		{"A", false, 0, false, ``, ``, UntilB(T("."))},
		{".", true, 1, false, ``, ``, UntilB(T("."))},
		{".B", true, 1, false, ``, ``, UntilB(T("."))},
		{"A.", true, 2, false, ``, ``, UntilB(T("."))},
		{"A.B", true, 2, false, ``, ``, UntilB(T("."))},
		{"AA.", true, 3, false, ``, ``, UntilB(T("."))},
		{"AA.B", true, 3, false, ``, ``, UntilB(T("."))},

		{"", true, 0, false, ``, ``, Q0(T("A"))},
		{"A", true, 1, false, ``, ``, Q0(T("A"))},
		{"B", true, 0, false, ``, ``, Q0(T("A"))},
		{"AA", true, 2, false, ``, ``, Q0(T("A"))},
		{"AB", true, 1, false, ``, ``, Q0(T("A"))},

		{"", false, 0, false, ``, ``, Q1(T("A"))},
		{"A", true, 1, false, ``, ``, Q1(T("A"))},
		{"B", false, 0, false, ``, ``, Q1(T("A"))},
		{"AA", true, 2, false, ``, ``, Q1(T("A"))},
		{"AB", true, 1, false, ``, ``, Q1(T("A"))},
		{"A", true, 1, false, ``, ``, Q1(T("A"))},
		{"AAB", true, 2, false, ``, ``, Q1(T("A"))},

		{"", false, 0, false, ``, ``, Qn(2, T("A"))},
		{"A", false, 0, false, ``, ``, Qn(2, T("A"))},
		{"AB", false, 0, false, ``, ``, Qn(2, T("A"))},
		{"AA", true, 2, false, ``, ``, Qn(2, T("A"))},
		{"AAB", true, 2, false, ``, ``, Qn(2, T("A"))},
		{"AAA", true, 3, false, ``, ``, Qn(2, T("A"))},
		{"AAAB", true, 3, false, ``, ``, Qn(2, T("A"))},

		{"", true, 0, false, ``, ``, Q01(T("A"))},
		{"A", true, 1, false, ``, ``, Q01(T("A"))},
		{"B", true, 0, false, ``, ``, Q01(T("A"))},

		{"", true, 0, false, ``, ``, Q0n(2, T("A"))},
		{"A", true, 1, false, ``, ``, Q0n(2, T("A"))},
		{"AB", true, 1, false, ``, ``, Q0n(2, T("A"))},
		{"AA", true, 2, false, ``, ``, Q0n(2, T("A"))},
		{"AAB", true, 2, false, ``, ``, Q0n(2, T("A"))},
		{"AAA", true, 2, false, ``, ``, Q0n(2, T("A"))},
		{"AAAB", true, 2, false, ``, ``, Q0n(2, T("A"))},

		{"", false, 0, false, ``, ``, Qnn(2, T("A"))},
		{"A", false, 0, false, ``, ``, Qnn(2, T("A"))},
		{"AB", false, 0, false, ``, ``, Qnn(2, T("A"))},
		{"AA", true, 2, false, ``, ``, Qnn(2, T("A"))},
		{"AAB", true, 2, false, ``, ``, Qnn(2, T("A"))},
		{"AAA", true, 2, false, ``, ``, Qnn(2, T("A"))},
		{"AAAB", true, 2, false, ``, ``, Qnn(2, T("A"))},

		{"", false, 0, false, ``, ``, Qmn(1, 3, T("A"))},
		{"A", true, 1, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AB", true, 1, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AA", true, 2, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AAB", true, 2, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AAA", true, 3, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AAAB", true, 3, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AAAA", true, 3, false, ``, ``, Qmn(1, 3, T("A"))},
		{"AAAAB", true, 3, false, ``, ``, Qmn(1, 3, T("A"))},

		{"", true, 0, false, ``, ``, J0(Q1(R('0', '9')), T("."))},
		{"192", true, 3, false, ``, ``, J0(Q1(R('0', '9')), T("."))},
		{"192.168.0.1", true, 11, false, ``, ``, J0(Q1(R('0', '9')), T("."))},
		{"192.168.0.a", true, 9, false, ``, ``, J0(Q1(R('0', '9')), T("."))},

		{"", false, 0, false, ``, ``, Jn(2, Q1(R('0', '9')), T("."))},
		{"1", false, 1, false, ``, ``, Jn(2, Q1(R('0', '9')), T("."))},
		{"1.1", true, 3, false, ``, ``, Jn(2, Q1(R('0', '9')), T("."))},
		{"1.1.1", true, 5, false, ``, ``, Jn(2, Q1(R('0', '9')), T("."))},
		{"1.1.1.1", true, 7, false, ``, ``, Jn(2, Q1(R('0', '9')), T("."))},

		{"", true, 0, false, ``, ``, J0n(4, Q1(R('0', '9')), T("."))},
		{"192", true, 3, false, ``, ``, J0n(4, Q1(R('0', '9')), T("."))},
		{"192.168.0.1", true, 11, false, ``, ``, J0n(4, Q1(R('0', '9')), T("."))},
		{"192.168.0.1.0", true, 11, false, ``, ``, J0n(4, Q1(R('0', '9')), T("."))},

		{"192.168", false, 0, false, ``, ``, Jnn(4, Q1(R('0', '9')), T("."))},
		{"192.168.0.1", true, 11, false, ``, ``, Jnn(4, Q1(R('0', '9')), T("."))},
		{"192.168.0.a", false, 0, false, ``, ``, Jnn(4, Q1(R('0', '9')), T("."))},

		{"", true, 0, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},
		{"1", true, 1, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},
		{"1.1", true, 3, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},
		{"1.1.1", true, 3, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},
		{".1", true, 2, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},
		{"1.", true, 2, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},
		{"1", true, 1, false, ``, ``, Jmn(1, 2, Q0(R('0', '9')), T("."))},

		// any Qn(n, pat) where pat could match ""
		// will leads to an infinite loop
		{"", false, 0, true, ``, ``, Q0(True)},
		{"", false, 0, true, ``, ``, Q0(Q0(T("A")))},
		{"", false, 0, true, ``, ``, Qn(100, Q01(T("A")))},
		{"", false, 0, true, ``, ``, J0(T(""), T(""))},
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}
