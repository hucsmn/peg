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

	// Simplified bare integer.
	SimpleHexUint8  = peg.Qmn(1, 2, HexDigit)
	SimpleHexUint16 = peg.Qmn(1, 4, HexDigit)
	SimpleHexUint32 = peg.Qmn(1, 8, HexDigit)
	SimpleHexUint64 = peg.Qmn(1, 16, HexDigit)

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
	return peg.Inject(newIntegerInjector(m, n), Integer)
}

// NoRedundantZeroes matches "0" or a bare integer without leading zeroes.
func NoRedundantZeroes(bareinteger peg.Pattern) peg.Pattern {
	return peg.Alt(peg.When(peg.Not(peg.T("0")), bareinteger), peg.T("0"))
}

// DecIntegerBetween mathces a DecInteger in the range [m, n].
func DecIntegerBetween(m, n uint64) peg.Pattern {
	return peg.Inject(newBareIntegerInjector(m, n, 10), DecInteger)
}

// HexIntegerBetween mathces a HexInteger in the range [m, n].
func HexIntegerBetween(m, n uint64) peg.Pattern {
	return peg.Inject(newBareIntegerInjector(m, n, 16), HexInteger)
}

// OctIntegerBetween mathces a OctInteger in the range [m, n].
func OctIntegerBetween(m, n uint64) peg.Pattern {
	return peg.Inject(newBareIntegerInjector(m, n, 8), OctInteger)
}

// helpers

func newIntegerInjector(m, n uint64) func(s string) (int, bool) {
	// assumes: len(s) > 0, matches Integer.
	return func(s string) (int, bool) {
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			n, ok := newBareIntegerInjector(m, n, 16)(s[2:])
			if !ok {
				return 0, false
			}
			return n + 2, true
		}
		if s[0] == '0' {
			oct := true
			for _, r := range s {
				if !strings.ContainsRune("01234567", r) {
					oct = false
					break
				}
			}
			if oct {
				return newBareIntegerInjector(m, n, 8)(s)
			}
		}
		return newBareIntegerInjector(m, n, 10)(s)
	}
}

func newBareIntegerInjector(m, n uint64, base int) func(s string) (int, bool) {
	if m > n {
		m, n = n, m
	}
	dm := countDigits(m, base)
	dn := countDigits(n, base)

	// assumes: len(s) > 0, 2<=base<=36, all digits are [0-9a-zA-Z].
	return func(s string) (int, bool) {
		var zeroes int
		var r rune
		for zeroes, r = range s {
			if r != '0' {
				break
			}
		}
		if s[zeroes:] == "" {
			s = s[zeroes-1:]
		} else {
			s = s[zeroes:]
		}

		if len(s) > dn {
			s = s[:dn]
		}
		for len(s) >= dm {
			x, _ := strconv.ParseUint(s, base, 64)
			if x >= m && x <= n {
				return zeroes + len(s), true
			}
			s = s[:len(s)-1]
		}
		return 0, false
	}
}

func countDigits(x uint64, base int) (n int) {
	b := uint64(base)
	n = 1
	for x >= b {
		x /= b
		n++
	}
	return
}
