package pegutil

import (
	"github.com/hucsmn/peg"
)

// Rune sets.
var (
	// Digits.
	OctDigit = peg.R('0', '7')
	DecDigit = peg.R('0', '9')
	HexDigit = peg.R('0', '9', 'a', 'f', 'A', 'F')

	// ASCII runes.
	ASCIIWhitespace    = peg.S(" \t\n\r\v\f")
	ASCIINotWhitespace = peg.NS(" \t\n\r\v\f")
	ASCIIDigit         = peg.R('0', '9')
	ASCIILetter        = peg.R('a', 'z', 'A', 'Z')
	ASCIILower         = peg.R('a', 'z')
	ASCIIUpper         = peg.R('A', 'Z')
	ASCIILetterDigit   = peg.R('a', 'z', 'A', 'Z', '0', '9')
	ASCIIControl       = peg.R('\x00', '\x1f', '\x7f', '\x7f')
	ASCIINotControl    = peg.R('\x20', '\x7e')

	// Unicode runes.
	Whitespace    = peg.U("White_Space")
	NotWhitespace = peg.U("-White_Space")
	Digit         = peg.U("Digit")
	Letter        = peg.U("Letter")
	Lower         = peg.U("Lower")
	Upper         = peg.U("Upper")
	Title         = peg.U("Title")
	LetterDigit   = peg.U("Letter", "Digit")
	Control       = peg.U("Control")
	NotControl    = peg.U("-Control")
	Printable     = peg.U("Print")
	NotPrintable  = peg.U("-Print")
	Graphic       = peg.U("Graphic")
	NotGraphic    = peg.U("-Graphic")

	// New line.
	NewlineRune    = peg.S("\n\r")
	NotNewlineRune = peg.NS("\n\r")
)
