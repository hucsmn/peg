package pegutil

import (
	"math"
	"strconv"
	"strings"

	"github.com/hucsmn/peg"
)

// Literals
var (
	// Bare integers.
	DecInteger = peg.Q1(DecDigit)
	DecUint8   = DecIntegerBetween(0, math.MaxUint8)
	DecUint16  = DecIntegerBetween(0, math.MaxUint16)
	DecUint32  = DecIntegerBetween(0, math.MaxUint32)
	DecUint64  = DecIntegerBetween(0, math.MaxUint64)
	HexInteger = peg.Q1(HexDigit)
	HexUint8   = HexIntegerBetween(0, math.MaxUint8)
	HexUint16  = HexIntegerBetween(0, math.MaxUint16)
	HexUint32  = HexIntegerBetween(0, math.MaxUint32)
	HexUint64  = HexIntegerBetween(0, math.MaxUint64)
	OctInteger = peg.Q1(OctDigit)
	OctUint8   = OctIntegerBetween(0, math.MaxUint8)
	OctUint16  = OctIntegerBetween(0, math.MaxUint16)
	OctUint32  = OctIntegerBetween(0, math.MaxUint32)
	OctUint64  = OctIntegerBetween(0, math.MaxUint64)

	// Numbers.
	Integer = peg.Alt(
		peg.Seq(peg.TI("0x"), HexInteger),
		DecInteger,
		peg.Seq(peg.T("0"), OctInteger))
	Decimal = peg.Check(func(s string) bool { return s != "." },
		peg.Alt(
			peg.Seq(peg.Q0(DecDigit), peg.T("."), peg.Q0(DecDigit)),
			DecInteger))
	Float = peg.Seq(
		Decimal,
		peg.Q01(
			peg.Seq(peg.TI("e"), peg.Q01(peg.S("+-")), DecInteger)))
	Number = peg.Alt(Float, Integer)

	// Identifer.
	Identifier = peg.Seq(
		peg.Alt(Letter, peg.T("_")),
		peg.Q0(peg.Alt(LetterDigit, peg.T("_"))))

	// Spaces and newlines.
	AnySpaces = peg.Q0(Whitespace)
	Spaces    = peg.Q1(Whitespace)
	Newline   = peg.Alt(peg.T("\r\n"), peg.S("\r\n"))

	// Quoted string.
	String = peg.Seq(
		peg.T(`"`),
		peg.Q0(peg.Alt(
			peg.Seq(peg.T(`\U`), peg.Qnn(8, HexDigit)),
			peg.Seq(peg.T(`\u`), peg.Qnn(4, HexDigit)),
			peg.Seq(peg.T(`\x`), peg.Qnn(2, HexDigit)),
			peg.Seq(peg.T(`\`), peg.Qnn(3, OctDigit)),
			peg.Seq(peg.T(`\`), peg.S(`abfnrtv\'"`)),
			peg.NS(`"\n\r`))),
		peg.T(`"`))
)

// IntegerBetween matches an Integer in the range [m, n].
func IntegerBetween(m, n uint64) peg.Pattern {
	return peg.Check(func(s string) bool {
		x, err := parseInteger(s)
		if err != nil {
			return false
		}

		return x >= m && x <= n
	}, Integer)
}

// NoRedundantZeroes matches "0" or a bare integer without leading zeroes.
func NoRedundantZeroes(bareinteger peg.Pattern) peg.Pattern {
	return peg.Alt(peg.When(peg.Not(peg.T("0")), bareinteger), peg.T("0"))
}

// DecIntegerBetween mathces a DecInteger in the range [m, n].
func DecIntegerBetween(m, n uint64) peg.Pattern {
	num, max := countBareIntegerDigits(n, 10)
	if m == 0 && max {
		return peg.Qmn(1, num, DecInteger)
	}
	return peg.Check(newBareIntegerChecker(m, n, 10),
		peg.Qmn(1, num, DecInteger))
}

// HexIntegerBetween mathces a HexInteger in the range [m, n].
func HexIntegerBetween(m, n uint64) peg.Pattern {
	num, max := countBareIntegerDigits(n, 16)
	if m == 0 && max {
		return peg.Qmn(1, num, HexInteger)
	}
	return peg.Check(newBareIntegerChecker(m, n, 16),
		peg.Qmn(1, num, HexInteger))
}

// OctIntegerBetween mathces a OctInteger in the range [m, n].
func OctIntegerBetween(m, n uint64) peg.Pattern {
	num, max := countBareIntegerDigits(n, 8)
	if m == 0 && max {
		return peg.Qmn(1, num, OctInteger)
	}
	return peg.Check(newBareIntegerChecker(m, n, 8),
		peg.Qmn(1, num, OctInteger))
}

// helpers

func newBareIntegerChecker(m, n uint64, base int) func(s string) bool {
	return func(s string) bool {
		x, err := strconv.ParseUint(s, base, 64)
		if err != nil {
			return false
		}

		return x >= m && x <= n
	}
}

func countBareIntegerDigits(x uint64, base int) (n int, max bool) {
	b := uint64(base)
	n = 1
	max = true
	for x >= b {
		if x%b != b-1 {
			max = false
		}
		x /= b
		n++
	}
	if x != b-1 {
		max = false
	}
	return
}

func parseInteger(s string) (x uint64, err error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		x, err = strconv.ParseUint(s[2:], 16, 64)
		if err != nil {
			return 0, err
		}
		return x, nil
	}
	if strings.HasPrefix(s, "0") {
		x, err = strconv.ParseUint(s[1:], 8, 64)
		if err == nil {
			return x, nil
		}
	}

	x, err = strconv.ParseUint(s, 10, 64)
	if err == nil {
		return x, nil
	}
	return 0, err
}
