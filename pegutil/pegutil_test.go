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

func TestAddress(t *testing.T) {
	data := []fullMatchTestData{
		// Hardware addresses.
		{"00:00:00:00:ff", false, MAC},
		{"00:00:00:00:ff:ff", true, MAC},
		{"00-00-00-00-ff-ff", true, MAC},
		{"0000.0000.ffff", true, MAC},
		{"00:00:00:00:ff:ff", false, EUI64},
		{"00:00:00:00:ff:ff:a0:a0", true, EUI64},
		{"00-00-00-00-ff-ff-a0-a0", true, EUI64},
		{"0000.0000.ffff.a0a0", true, EUI64},

		// IP addresses.
		{"192.168", false, IPv4},
		{"192.168.0.1", true, IPv4},
		{"192.256.0.1", false, IPv4},
		{"192.0168.0.1", false, IPv4},
		{"0xc0.0xa8.0x0.0x1", false, IPv4},
		{"0xc0.0xa8.0.1", false, IPv4},
		{"0xc0a80001", false, IPv4},
		{"192.168.0.1/24", true, CIDRv4},
		{"192.168.0.1/128", false, CIDRv4},
		{":", false, IPv6},
		{"::", true, IPv6},
		{"::1", true, IPv6},
		{"ffff:0000:0000:0000:0000:0000:0000:0000", true, IPv6},
		{"ffff:0:0:0:0:0:0:0", true, IPv6},
		{"ff:0:0:0:0:0:0:0", true, IPv6},
		{"ffff::0:0:0:0:0:0", true, IPv6},
		{"ffff::0:0:0:0:0:0:0", false, IPv6},
		{"ffff::0:0:0:0:0::0", false, IPv6},
		{"ffff:0:0:0:0:0::0", true, IPv6},
		{"ffff:0:0:0:0:0:0::0", false, IPv6},
		{"ffff:0:0:0:0:0:0::", true, IPv6},
		{"ffff:0:0:0:0:0:0:0::", false, IPv6},
		{"ffff:0:0:0:0:0:0:192.168.0.1", false, IPv6},
		{"ffff:0:0:0:0:0:192.168.0.1", true, IPv6},
		{"ffff:0:0:0:0::192.168.0.1", true, IPv6},
		{"ffff::192.168.0.1", true, IPv6},
		{"::192.168.0.1", true, IPv6},
		{"ffff::0:192.168.0.1", true, IPv6},
		{"ffff::0:0:0:0:192.168.0.1", true, IPv6},
		{"ffff::0:0:0:0:0:192.168.0.1", false, IPv6},
		{"::1%8", true, IPv6WithZone},
		{"ffff::%10", true, IPv6WithZone},
		{"ffff:0:0:0:0:0:0:0%10", true, IPv6WithZone},
		{"ffff:0:0:0:0:0:192.168.0.1%10", true, IPv6WithZone},
		{"::1/112", true, CIDRv6},
		{"ffff::/112", true, CIDRv6},
		{"ffff:0:0:0:0:0:0:0/112", true, CIDRv6},
		{"ffff:0:0:0:0:0:192.168.0.1/112", true, CIDRv6},
		{"ffff:0:0:0:0:0:192.168.0.1/129", false, CIDRv6},
		{"192.168", false, IP},
		{"192.168.0.1", true, IP},
		{"::", true, IP},
		{"::1", true, IP},
		{"ffff:0:0:0:0:0:0:0", true, IP},
		{"ff:0:0:0:0:0:0:0", true, IP},
		{"ffff::0:0:0:0:0:0", true, IP},
		{"ffff::0:0:0:0:0:0:0", false, IP},
		{"ffff::0:0:0:0:0::0", false, IP},
		{"ffff:0:0:0:0:0::0", true, IP},
		{"ffff:0:0:0:0:0:0::0", false, IP},
		{"ffff:0:0:0:0:0:0::", true, IP},
		{"ffff:0:0:0:0:0:0:0::", false, IP},
		{"ffff:0:0:0:0:0:192.168.0.1", true, IP},
		{"ffff:0:0:0:0::192.168.0.1", true, IP},
		{"192.168.0.1/24", true, CIDR},
		{"192.168.0.1/128", false, CIDR},
		{"ffff:0:0:0:0:0:192.168.0.1/112", true, CIDR},
		{"ffff:0:0:0:0:0:192.168.0.1/129", false, CIDR},

		// URI addresses.
		{"google.com", true, Host},
		{"google.com.", true, Host},
		{"google.com:80", false, Host},
		{"user:pass@google.com:80", false, Host},
		{"user:pass@google.com", false, Host},
		{"localhost", true, Host},
		{"localhost.localdomain", true, Host},
		{"192.168.0.1", true, Host},
		{"[ffff:0:0:0:0:0:0:0]", true, Host},
		{"[ffff::192.168.0.1]", true, Host},
		{"[vff.blabla]", true, Host},
		{"http://google.com", true, URI},
		{"http://google.com/", true, URI},
		{"http://google.com:80/", true, URI},
		{"http://user:pass@google.com/", true, URI},
		{"http://user:pass@google.com:80/", true, URI},
		{"http://google.com/?q=blabla", true, URI},
		{"http://google.com?q=blabla", true, URI},
		{"http://google.com/path?q=blabla", true, URI},
		{"http://google.com/path#fragment", true, URI},
		{"http://google.com/path?q=blabla#fragment", true, URI},
		{"file:///bin/sh", true, URI},
		{"scheme:", true, URI},
		{"scheme:/", true, URI},
		{"scheme:/a", true, URI},
		{"scheme://", true, URI},
		{"magnet:?xt=urn:btih:blabla", true, URI},
		{"urn:foo:bar", true, URI},

		// E-Mail address.
		{`prettyandsimple@example.com`, true, EMail},
		{`very.common@example.com`, true, EMail},
		{`disposable.style.email.with+symbol@example.com`, true, EMail},
		{`other.email-with-dash@example.com`, true, EMail},
		{`fully-qualified-domain@example.com.`, true, EMail},
		{`x@example.com`, true, EMail},
		{`"very.unusual.@.unusual.com"@example.com`, true, EMail},
		{`example-indeed@strange-example.com`, true, EMail},
		{`admin@mailserver1`, true, EMail},
		{`"()<>[]:,;@\\\"!#$%&'-/=?^_{}| ~.a"@example.org`, true, EMail},
		{`" "@example.org`, true, EMail},
		{`example@s.solutions`, true, EMail},
		{`user@localserver`, true, EMail},
		{`user@[IPv6:2001:DB8::1]`, true, EMail},
		{`Abc.example.com`, false, EMail},
		{`A@b@c@example.com`, false, EMail},
		{`a"b(c)d,e:f;g<h>i[j\k]l@example.com`, false, EMail},
		{`just"not"right@example.com`, false, EMail},
		{`this is"not\allowed@example.com`, false, EMail},
		{`this\ still\"not\\allowed@example.com`, false, EMail},
		{`1234567890123456789012345678901234567890123456789012345678901234+x@example.com`, false, EMail},
		{`john..doe@example.com`, false, EMail},
		{`john.doe@example..com`, false, EMail},
	}

	for _, d := range data {
		runFullMatchTestData(t, d)
	}
}
