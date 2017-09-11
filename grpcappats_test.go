package peg

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// Tests for grouping and capturing.

// Tests Cgrp, Cname, Check, Trunc, Inject, Ref, RefB.
func TestGroupInjectAndRefer(t *testing.T) {
	zeroes := func(s string) (int, bool) {
		var i int
		var r rune
		for i, r = range s {
			if r != '0' {
				break
			}
		}
		if i > 0 {
			return i, true
		}
		return 0, false
	}
	patDigit := R('0', '9')
	patSpaces := Q0(U("White_Space"))
	patFloating := Check(func(s string) bool {
		s = strings.ToLower(s)
		s = strings.SplitN(s, "e", 2)[0] // fraction part
		s = strings.TrimLeft(s, "+-")    // fraction part without leading +/-
		return s != "."
	},
		Seq(
			Q01(S("+-")),
			Alt(
				Seq(Q0(patDigit), T("."), Q0(patDigit)),
				Q1(patDigit)),
			Q01(Seq(S("eE"), Q01(S("+-")), Q1(patDigit)))))

	patJudge := Seq(
		NG("num", patFloating),
		patSpaces,
		Alt(
			Seq(T("=="), patSpaces, Ref("num")),
			Seq(T("!="), patSpaces, Not(Ref("num")), patFloating)))

	data := []patternTestData{
		{"", true, 0, false, `""`, ``, G(True)},
		{"", false, 0, false, ``, ``, G(False)},
		{"", false, 0, false, ``, ``, Seq(True, Seq(G(True), False))},
		{"", false, 0, false, ``, ``, Seq(G(True), False)},
		{"AAA", true, 3, false, `"AAA"`, ``, G(Q0(T("A")))},
		{"AAA", true, 3, false, `"A","A","A"`, ``, Q0(G(T("A")))},

		{"", true, 0, false, `"cap"=""`, ``, NG("cap", True)},
		{"", false, 0, false, ``, ``, NG("cap", False)},
		{"AAA", true, 3, false, `"cap"="AAA"`, ``, NG("cap", Q0(T("A")))},
		{"AAA", true, 3, false, `"cap"="A"`, ``, Q0(NG("cap", T("A")))},
		{"ABC", true, 3, false, `"cap"="C"`, ``, Q0(NG("cap", S("ABC")))}, // overwrites
		{"+1.e-2", true, 6, false, `"num"="+1.e-2"`, ``, NG("num", patFloating)},

		{"0", true, 1, false, ``, ``, Check(func(s string) bool { return s == "0" }, Dot)},
		{"1", false, 0, false, ``, ``, Check(func(s string) bool { return s == "0" }, Dot)},
		{"", true, 0, false, ``, ``, Trunc(2, Q0(Dot))},
		{"A", true, 1, false, ``, ``, Trunc(2, Q0(Dot))},
		{"AA", true, 2, false, ``, ``, Trunc(2, Q0(Dot))},
		{"AAA", true, 2, false, ``, ``, Trunc(2, Q0(Dot))},
		{"AAA", true, 0, false, ``, ``, Trunc(0, Q0(Dot))},
		{"AAA", false, 0, false, ``, ``, Trunc(-1, Q0(Dot))},
		{"0246", true, 1, false, ``, ``, Inject(zeroes, Seq(Dot, Dot, Dot, Dot))},
		{"0046", true, 2, false, ``, ``, Inject(zeroes, Seq(Dot, Dot, Dot, Dot))},
		{"1246", false, 0, false, ``, ``, Inject(zeroes, Seq(Dot, Dot, Dot, Dot))},

		{"+1.e-2 == +1.e-2", true, 16, false, `"num"="+1.e-2"`, ``, patJudge},
		{"+1.e-2 == .1e2", false, 0, false, ``, ``, patJudge},
		{"+1.e-2 != .1e2", true, 14, false, `"num"="+1.e-2"`, ``, patJudge},
		{"+1.e-2 != +1.e-2", false, 0, false, ``, ``, patJudge},

		{"ABC", false, 0, false, ``, ``, Seq(G(Dot), Q0(Dot), RefB(""))},
		{"ABA", true, 3, false, `"A"`, ``, Seq(G(Dot), Q0(Dot), RefB(""))},
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}

// Tests Let, Var, Cvar, Ctoken.
func TestGrammar(t *testing.T) {
	scope := map[string]Pattern{
		"T":  True,
		"F":  False,
		"D":  Dot,
		"v0": T("A"),
		"vt": CK(0, T("A")),
		"vq": Q1(CK(0, T("A"))),
		"U":  V("undefined_var"),
	}
	recurscope := map[string]Pattern{
		"R0": V("R0"),
		"R1": When(True, V("R1")),
		"Ra": V("Rb"),
		"Rb": V("Ra"),
	}
	data := []patternTestData{
		{"", true, 0, false, ``, ``, Let(scope, V("T"))},
		{"A", true, 1, false, ``, ``, Let(scope, V("v0"))},
		{"", false, 0, true, ``, ``, Let(scope, V("undef"))}, // error variable undefined
		{"", false, 0, true, ``, ``, Let(scope, V("U"))},     // error variable undefined
		{"", true, 0, false, ``, ``,
			Let(scope, Let(map[string]Pattern{ // override the definition of U
				"U": True,
			}, V("U")))}, // no error occurs

		{"", false, 0, false, ``, ``, Let(scope,
			CV("v0"))},
		{"A", true, 1, false, ``, `v0()`, Let(scope,
			CV("v0"))},
		{"A", true, 1, false, ``, `vt(<0"A">)`, Let(scope,
			CV("vt"))},
		{"A", true, 1, false, ``, `vq(<0"A">)`, Let(scope,
			CV("vq"))},
		{"A", true, 1, false, ``, `v0()`, Let(scope,
			Q0(CV("v0")))},
		{"A", true, 1, false, ``, `vt(<0"A">)`, Let(scope,
			Q0(CV("vt")))},
		{"A", true, 1, false, ``, `vq(<0"A">)`, Let(scope,
			Q0(CV("vq")))},
		{"AA", true, 2, false, ``, `v0(), v0()`, Let(scope,
			Q0(CV("v0")))},
		{"AA", true, 2, false, ``, `vt(<0"A">), vt(<0"A">)`, Let(scope,
			Q0(CV("vt")))},
		{"AA", true, 2, false, ``, `vq(<0"A">, <0"A">)`, Let(scope,
			Q0(CV("vq")))},

		// Tests max recursion
		{"", false, 0, true, ``, ``, Let(recurscope, V("R0"))},
		{"", false, 0, true, ``, ``, Let(recurscope, V("R1"))},
		{"", false, 0, true, ``, ``, Let(recurscope, V("Ra"))},
		{"", false, 0, true, ``, ``, Let(recurscope, V("Rb"))},
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}

// Tests Cterm, Ccons.

type termInt int32

func (t termInt) IsTerminal() bool {
	return true
}

type termOp string

func (t termOp) IsTerminal() bool {
	return true
}

func TestCustomedConstructor(t *testing.T) {
	intcons := func(text string, pos Position) (Capture, error) {
		i, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return nil, err
		}
		return termInt(int32(i)), nil
	}
	opcons := func(text string, pos Position) (Capture, error) {
		return termOp(text), nil
	}
	eval := func(caps []Capture) (Capture, error) {
		if len(caps) <= 0 || len(caps)%2 != 1 {
			return nil, fmt.Errorf("eval: invalid argument number %d", len(caps))
		}

		x, ok := caps[0].(termInt)
		if !ok {
			return nil, fmt.Errorf("eval: expect number but get %v", caps[0])
		}
		caps = caps[1:]
		for len(caps) > 0 {
			op, ok := caps[0].(termOp)
			if !ok {
				return nil, fmt.Errorf("eval: expect op but get %v", caps[0])
			}
			y, ok := caps[1].(termInt)
			if !ok {
				return nil, fmt.Errorf("eval: expect number but get %v", caps[1])
			}
			caps = caps[2:]
			switch string(op) {
			case "+":
				x = termInt(int32(x) + int32(y))
			case "-":
				x = termInt(int32(x) - int32(y))
			case "*":
				x = termInt(int32(x) * int32(y))
			case "/":
				if int32(y) == 0 {
					return nil, fmt.Errorf("eval: division by zero")
				}
				x = termInt(int32(x) / int32(y))
			default:
				return nil, fmt.Errorf("eval: unknown op: %q", string(op))
			}
		}
		return x, nil
	}
	patDigit := R('0', '9')
	patSpace := Q0(U("White_Space"))
	patNumber := Seq(CT(intcons, Q1(patDigit)), patSpace)
	patTermOp := Seq(CT(opcons, S("+-")), patSpace)
	patFactorOp := Seq(CT(opcons, S("*/")), patSpace)
	patOpen := Seq(T("("), patSpace)
	patClose := Seq(T(")"), patSpace)
	patExpr := Let(map[string]Pattern{
		"factor": Alt(
			patNumber,
			Seq(patOpen, V("expr"), patClose)),
		"term": CC(eval, Seq(V("factor"), Q0(Seq(patFactorOp, V("factor"))))),
		"expr": CC(eval, Seq(V("term"), Q0(Seq(patTermOp, V("term"))))),
	}, V("expr"))
	data := []patternTestData{
		{"", false, 0, false, ``, ``, patExpr},
		{"A", false, 0, false, ``, ``, patExpr},
		{"0", true, 1, false, ``, `<0>`, patExpr},
		{"1", true, 1, false, ``, `<1>`, patExpr},
		{"1+2", true, 3, false, ``, `<3>`, patExpr},
		{"1-2", true, 3, false, ``, `<-1>`, patExpr},
		{"3*(1+2)", true, 7, false, ``, `<9>`, patExpr},
		{"1-2*((3+4)/5+6*(7-8))/9", true, 23, false, ``, `<2>`, patExpr},

		{"10000000000", false, 0, true, ``, ``, patExpr}, // bigger than max int 32
		{"1/0", false, 0, true, ``, ``, patExpr},         // division by zero
	}

	for _, d := range data {
		runPatternTestData(t, d)
	}
}
