package pegutil

import (
	"github.com/hucsmn/peg"
)

// Digits.
var (
	octDigit = peg.R('0', '7')
	decDigit = peg.R('0', '9')
	hexDigit = peg.R('0', '9', 'a', 'f', 'A', 'F')

	OctDigit = octDigit
	DecDigit = decDigit
	HexDigit = hexDigit
)

// ASCII runes.
var (
	asciiWhitespace    = peg.S(" \t\n\r\v\f")
	asciiNotWhitespace = peg.NS(" \t\n\r\v\f")
	asciiDigit         = peg.R('0', '9')
	asciiLetter        = peg.R('a', 'z', 'A', 'Z')
	asciiLower         = peg.R('a', 'z')
	asciiUpper         = peg.R('A', 'Z')
	asciiLetterDigit   = peg.R('a', 'z', 'A', 'Z', '0', '9')
	asciiControl       = peg.R('\x00', '\x1f', '\x7f', '\x7f')
	asciiNotControl    = peg.R('\x20', '\x7e')

	ASCIIWhitespace    = asciiWhitespace
	ASCIINotWhitespace = asciiNotWhitespace
	ASCIIDigit         = asciiDigit
	ASCIILetter        = asciiLetter
	ASCIILower         = asciiLower
	ASCIIUpper         = asciiUpper
	ASCIILetterDigit   = asciiLetterDigit
	ASCIIControl       = asciiControl
	ASCIINotControl    = asciiNotControl
)

// Unicode runes.
var (
	unicodeWhitespace    = peg.U("White_Space")
	unicodeNotWhitespace = peg.U("-White_Space")
	unicodeDigit         = peg.U("Digit")
	unicodeLetter        = peg.U("Letter")
	unicodeLower         = peg.U("Lower")
	unicodeUpper         = peg.U("Upper")
	unicodeTitle         = peg.U("Title")
	unicodeLetterDigit   = peg.U("Letter", "Digit")
	unicodeControl       = peg.U("Control")
	unicodeNotControl    = peg.U("-Control")
	unicodePrintable     = peg.U("Print")
	unicodeNotPrintable  = peg.U("-Print")
	unicodeGraphic       = peg.U("Graphic")
	unicodeNotGraphic    = peg.U("-Graphic")

	Whitespace    = unicodeWhitespace
	NotWhitespace = unicodeNotWhitespace
	Digit         = unicodeDigit
	Letter        = unicodeLetter
	Lower         = unicodeLower
	Upper         = unicodeUpper
	Title         = unicodeTitle
	LetterDigit   = unicodeLetterDigit
	Control       = unicodeControl
	NotControl    = unicodeNotControl
	Printable     = unicodePrintable
	NotPrintable  = unicodeNotPrintable
	Graphic       = unicodeGraphic
	NotGraphic    = unicodeNotGraphic
)

// New line.
var (
	newlineRune    = peg.S("\n\r")
	notNewlineRune = peg.NS("\n\r")

	NewlineRune    = newlineRune
	NotNewlineRune = notNewlineRune
)
