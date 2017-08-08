// Package pegutil provides extra utils for the Parsing Expression Grammars.
//
// Following categories of utils are provided by this package:
//     Rune sets (including *Digit, ASCII*, <common unicode rune sets>)
//     Bare integers (including *Integer, *Uint*, SimpleHexUint*)
//     Integer helpers (*IntegerBetweenm, NoRedundantZeroes, ...).
//     Simple literals (including Float, Identifier, String, Newline, ...)
//     TCP/IP addresses (including MAC, EUI64, IPv4, IPv6, CIDR, ...)
//     URI addresses (including URI, EMail, ...)
// Ths package API is currently volatile.
package pegutil //import "github.com/hucsmn/peg/pegutil"

import (
	"github.com/hucsmn/peg"
)

// Scope contains all the variables defined in this package.
var Scope = map[string]peg.Pattern{
	"OctDigit": OctDigit,
	"DecDigit": DecDigit,
	"HexDigit": HexDigit,

	"ASCIIWhitespace":    ASCIIWhitespace,
	"ASCIINotWhitespace": ASCIINotWhitespace,
	"ASCIIDigit":         ASCIIDigit,
	"ASCIILetter":        ASCIILetter,
	"ASCIILower":         ASCIILower,
	"ASCIIUpper":         ASCIIUpper,
	"ASCIILetterDigit":   ASCIILetterDigit,
	"ASCIIControl":       ASCIIControl,
	"ASCIINotControl":    ASCIINotControl,

	"Whitespace":    Whitespace,
	"NotWhitespace": NotWhitespace,
	"Digit":         Digit,
	"Letter":        Letter,
	"Lower":         Lower,
	"Upper":         Upper,
	"Title":         Title,
	"LetterDigit":   LetterDigit,
	"Control":       Control,
	"NotControl":    NotControl,
	"Printable":     Printable,
	"NotPrintable":  NotPrintable,
	"Graphic":       Graphic,
	"NotGraphic":    NotGraphic,

	"NewlineRune":    NewlineRune,
	"NotNewlineRune": NotNewlineRune,

	"DecInteger": DecInteger,
	"DecUint8":   DecUint8,
	"DecUint16":  DecUint16,
	"DecUint32":  DecUint32,
	"DecUint64":  DecUint64,

	"HexInteger": HexInteger,
	"HexUint8":   HexUint8,
	"HexUint16":  HexUint16,
	"HexUint32":  HexUint32,
	"HexUint64":  HexUint64,

	"OctInteger": OctInteger,
	"OctUint8":   OctUint8,
	"OctUint16":  OctUint16,
	"OctUint32":  OctUint32,
	"OctUint64":  OctUint64,

	"SimpleHexUint8":  SimpleHexUint8,
	"SimpleHexUint16": SimpleHexUint16,
	"SimpleHexUint32": SimpleHexUint32,
	"SimpleHexUint64": SimpleHexUint64,

	"Integer":    Integer,
	"Decimal":    Decimal,
	"Float":      Float,
	"Number":     Number,
	"Identifier": Identifier,
	"AnySpaces":  AnySpaces,
	"Spaces":     Spaces,
	"Newline":    Newline,
	"String":     String,

	"MAC":          MAC,
	"EUI64":        EUI64,
	"IP":           IP,
	"CIDR":         CIDR,
	"IPv4":         IPv4,
	"CIDRv4":       CIDRv4,
	"IPv6":         IPv6,
	"CIDRv6":       CIDRv6,
	"IPv6WithZone": IPv6WithZone,

	"URI":          URI,
	"URIHost":      URIHost,
	"URIAbsolute":  URIAbsolute,
	"URIReference": URIReference,
	"Slug":         Slug,
	"Domain":       Domain,
	"EMail":        EMail,
}
