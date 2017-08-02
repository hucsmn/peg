package pegutil

import (
	"testing"

	"github.com/hucsmn/peg"
)

type fullMatchTestData struct {
	text string
	full bool
	pat  peg.Pattern
}

func runFullMatchTestData(t *testing.T, data fullMatchTestData) {
	full := peg.IsFullMatched(data.pat, data.text)
	if full != data.full {
		t.Errorf("RESULT DISMATCH: IsFullMatched(%s, %q) => %t != %t\n",
			data.pat, data.text, full, data.full)
	}
}

func TestLiteralFullMatch(t *testing.T) {
	data := []fullMatchTestData{
		// Bare integers.
		{"0", true, DecUint8},
		{"1", true, DecUint8},
		{"255", true, DecUint8},
		{"256", false, DecUint8},
		{"0", true, DecInteger},
		{"18446744073709551616", true, DecInteger}, // 1<<64
		{"18446744073709551616", false, DecUint64}, // 1<<64
		{"0", true, SimpleHexUint8},
		{"0f", true, SimpleHexUint8},
		{"0ff", false, SimpleHexUint8},

		// Integer.
		{"0123", true, Integer},
		{"0X0123", true, Integer},
		{"0x0123", true, Integer},
		{"0x01fe", true, Integer},
		{"0x01fx", false, Integer},
		{"0777", true, Integer},
		{"0778", true, Integer},

		// Integer between.
		{"000775", true, IntegerBetween(0, 999)},
		{"123", true, IntegerBetween(23, 223)},
		{"123", true, IntegerBetween(23, 123)},
		{"123", true, IntegerBetween(123, 223)},
		{"0x7b", true, IntegerBetween(23, 223)},
		{"0x7b", true, IntegerBetween(23, 123)},
		{"0x7b", true, IntegerBetween(123, 223)},
		{"0173", true, IntegerBetween(23, 223)},
		{"0173", true, IntegerBetween(23, 123)},
		{"0173", true, IntegerBetween(123, 223)},
		{"00123", true, DecIntegerBetween(23, 223)},
		{"00123", true, DecIntegerBetween(23, 123)},
		{"00123", true, DecIntegerBetween(123, 223)},

		// No redundant zeroes.
		{"0", true, NoRedundantZeroes(DecInteger)},
		{"00123", false, NoRedundantZeroes(DecInteger)},
		{"10023", true, NoRedundantZeroes(DecInteger)},

		// Decimal, Float and Number.
		{"1.1", true, Decimal},
		{"1", true, Decimal},
		{".1", true, Decimal},
		{"1.", true, Decimal},
		{".", false, Decimal},
		{"1.e-3", true, Float},
		{"1E-3", true, Float},
		{"0.1E3", true, Float},
		{"0.E+3", true, Float},
		{"1", true, Number},
		{"0xff", true, Number},
		{"0777", true, Number},
		{"1.1", true, Number},
		{"1.1e3", true, Number},

		// Identifer.
		{"Id", true, Identifier},
		{"_Id", true, Identifier},
		{"Id_", true, Identifier},
		{"Id0", true, Identifier},
		{"0Id", false, Identifier},
		{"标识符", true, Identifier},

		// Space and newline.
		{"", false, Spaces},
		{" \t\n \r ", true, Spaces},
		{"", true, AnySpaces},
		{" \t\n \r ", true, AnySpaces},
		{"\n", true, Newline},
		{"\r", true, Newline},
		{"\r\n", true, Newline},
		{"\n\r", false, Newline},

		// String.
		{`""`, true, String},
		{`"a"`, true, String},
		{`"\n\r"`, true, String},
		{`"title\r\nparagraph"`, true, String},
		{`"\uFFFD\U0000FFFD\xef\xbf\xbd\a\b\f\n\r\t\v\\\'\""`, true, String},
	}

	for _, d := range data {
		runFullMatchTestData(t, d)
	}
}
