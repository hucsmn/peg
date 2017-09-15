package pegutil

import (
	"math"
	"strconv"
	"strings"

	"github.com/hucsmn/peg"
)

// Bare integers.
var (
	decInteger = peg.Q1(DecDigit)
	decUint8   = DecIntegerBetween(0, math.MaxUint8)
	decUint16  = DecIntegerBetween(0, math.MaxUint16)
	decUint32  = DecIntegerBetween(0, math.MaxUint32)
	decUint64  = DecIntegerBetween(0, math.MaxUint64)
	hexInteger = peg.Q1(HexDigit)
	hexUint8   = HexIntegerBetween(0, math.MaxUint8)
	hexUint16  = HexIntegerBetween(0, math.MaxUint16)
	hexUint32  = HexIntegerBetween(0, math.MaxUint32)
	hexUint64  = HexIntegerBetween(0, math.MaxUint64)
	octInteger = peg.Q1(OctDigit)
	octUint8   = OctIntegerBetween(0, math.MaxUint8)
	octUint16  = OctIntegerBetween(0, math.MaxUint16)
	octUint32  = OctIntegerBetween(0, math.MaxUint32)
	octUint64  = OctIntegerBetween(0, math.MaxUint64)

	DecInteger = decInteger
	DecUint8   = decUint8
	DecUint16  = decUint16
	DecUint32  = decUint32
	DecUint64  = decUint64
	HexInteger = hexInteger
	HexUint8   = hexUint8
	HexUint16  = hexUint16
	HexUint32  = hexUint32
	HexUint64  = hexUint64
	OctInteger = octInteger
	OctUint8   = octUint8
	OctUint16  = octUint16
	OctUint32  = octUint32
	OctUint64  = octUint64
)

// Hexadecimal digit sequences.
var (
	literalHexDigits8   = peg.Qnn(2, HexDigit)
	literalHexDigits16  = peg.Qnn(4, HexDigit)
	literalHexDigits32  = peg.Qnn(8, HexDigit)
	literalHexDigits64  = peg.Qnn(16, HexDigit)
	literalHexDigits128 = peg.Qnn(32, HexDigit)
	literalHexDigits256 = peg.Qnn(64, HexDigit)
	literalHexDigits512 = peg.Qnn(128, HexDigit)

	literalVaryHexDigits8   = peg.Qmn(1, 2, HexDigit)
	literalVaryHexDigits16  = peg.Qmn(1, 4, HexDigit)
	literalVaryHexDigits32  = peg.Qmn(1, 8, HexDigit)
	literalVaryHexDigits64  = peg.Qmn(1, 16, HexDigit)
	literalVaryHexDigits128 = peg.Qmn(1, 32, HexDigit)
	literalVaryHexDigits256 = peg.Qmn(1, 64, HexDigit)
	literalVaryHexDigits512 = peg.Qmn(1, 128, HexDigit)

	HexDigits8   = literalHexDigits8
	HexDigits16  = literalHexDigits16
	HexDigits32  = literalHexDigits32
	HexDigits64  = literalHexDigits64
	HexDigits128 = literalHexDigits128
	HexDigits256 = literalHexDigits256
	HexDigits512 = literalHexDigits512

	VaryHexDigits8   = literalVaryHexDigits8
	VaryHexDigits16  = literalVaryHexDigits16
	VaryHexDigits32  = literalVaryHexDigits32
	VaryHexDigits64  = literalVaryHexDigits64
	VaryHexDigits128 = literalVaryHexDigits128
	VaryHexDigits256 = literalVaryHexDigits256
	VaryHexDigits512 = literalVaryHexDigits512
)

// Numbers.
var (
	literalInteger = peg.Alt(
		peg.Seq(peg.TI("0x"), HexInteger),
		DecInteger,
		peg.Seq(peg.T("0"), OctInteger))
	literalDecimal = peg.Check(func(s string) bool { return s != "." },
		peg.Alt(
			peg.Seq(peg.Q0(DecDigit), peg.T("."), peg.Q0(DecDigit)),
			DecInteger))
	literalFloat = peg.Seq(
		literalDecimal,
		peg.Q01(
			peg.Seq(peg.TI("e"), peg.Q01(peg.S("+-")), DecInteger)))
	literalNumber = peg.Alt(
		peg.Seq(peg.TI("0x"), HexInteger),
		literalFloat,
		peg.Seq(peg.T("0"), OctInteger))

	Integer = literalInteger
	Decimal = literalDecimal
	Float   = literalFloat
	Number  = literalNumber
)

// Identifer.
var (
	literalIdentifier = peg.Seq(
		peg.Alt(Letter, peg.T("_")),
		peg.Q0(LetterDigit, peg.T("_")))

	Identifier = literalIdentifier
)

// Spaces and newlines.
var (
	lexAnySpaces = peg.Q0(Whitespace)
	lexSpaces    = peg.Q1(Whitespace)
	lexNewline   = peg.Alt(peg.T("\r\n"), peg.S("\r\n"))

	AnySpaces = lexAnySpaces
	Spaces    = lexSpaces
	Newline   = lexNewline
)

// Quoted string.
var (
	literalString = peg.Seq(
		peg.T(`"`),
		peg.Q0(
			peg.Seq(peg.T(`\U`), peg.Qnn(8, HexDigit)),
			peg.Seq(peg.T(`\u`), peg.Qnn(4, HexDigit)),
			peg.Seq(peg.T(`\x`), peg.Qnn(2, HexDigit)),
			peg.Seq(peg.T(`\`), peg.Qnn(3, OctDigit)),
			peg.Seq(peg.T(`\`), peg.S(`abfnrtv\'"`)),
			peg.NS("\"\n\r")),
		peg.T(`"`))

	String = literalString
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

// Helpers for integer literal validation.

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
			x, err := strconv.ParseUint(s, base, 64)
			if err == nil && x >= m && x <= n {
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
