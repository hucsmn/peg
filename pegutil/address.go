package pegutil

import (
	"strings"

	"github.com/hucsmn/peg"
)

// Hardware addresses.
var (
	// 48-bit hardware address (i.e. EUI-48).
	MAC = peg.Alt(
		peg.Jnn(6, peg.Qnn(2, HexDigit), peg.T(":")),
		peg.Jnn(6, peg.Qnn(2, HexDigit), peg.T("-")),
		peg.Jnn(3, peg.Qnn(4, HexDigit), peg.T(".")))

	// 64-bit hardware address.
	EUI64 = peg.Alt(
		peg.Jnn(8, peg.Qnn(2, HexDigit), peg.T(":")),
		peg.Jnn(8, peg.Qnn(2, HexDigit), peg.T("-")),
		peg.Jnn(4, peg.Qnn(4, HexDigit), peg.T(".")))
)

// IP addresses.
var (
	// IPv4 dot-decimal address.
	IPv4 = peg.Jnn(4, NoRedundantZeroes(DecUint8), peg.T("."))

	// IPv4 address with subnetwork mask.
	CIDRv4 = peg.Seq(IPv4, peg.T("/"), DecIntegerBetween(0, 32))

	// IPv6 address.
	IPv6 = peg.Alt(
		// ellipsis with trailing 32-bit dot-decimals
		// (e.g. ::192.168.0.1).
		peg.Inject(ipv6EllipsisIPv4,
			peg.Seq(
				peg.Jmn(0, 5, SimpleHexUint16, peg.T(":")),
				peg.T("::"),
				peg.Alt(
					peg.Seq(
						peg.Jmn(1, 5, SimpleHexUint16, peg.T(":")),
						peg.Seq(peg.T(":"), IPv4)),
					IPv4))),
		// ellipsis without trailing 32-bit dot-decimals
		// (e.g. ::1, ::, ffff::, ffff::1).
		peg.Inject(ipv6EllipsisNoIPv4,
			peg.Seq(
				peg.Jmn(0, 7, SimpleHexUint16, peg.T(":")),
				peg.T("::"),
				peg.Jmn(0, 7, SimpleHexUint16, peg.T(":")))),
		// 8 uint16 groups (e.g. ffff:0:0:0:0:0:0:0).
		peg.Jnn(8, HexUint16, peg.T(":")),
		// 6 uint16 groups with trailing 32-bit dot-decimals
		// (e.g. ffff:0:0:0:0:0:192.168.0.1).
		peg.Seq(
			peg.Jnn(6, SimpleHexUint16, peg.T(":")),
			peg.Seq(peg.T(":"), IPv4)))

	// IPv6 address with zone identifer.
	IPv6WithZone = peg.Seq(IPv6, peg.T("%"), DecInteger)

	// IPv6 address with subnetwork mask.
	CIDRv6 = peg.Seq(IPv6, peg.T("/"), DecIntegerBetween(0, 128))

	// IPv4 or IPv6 address.
	IP = peg.Alt(IPv6, IPv4)

	// IP address with subnetwork mask.
	CIDR = peg.Alt(CIDRv6, CIDRv4)
)

// URI definitions described in RFC 3986.
var (
	// helpers.
	uriEncodedByte       = peg.Seq(peg.T("%"), HexDigit, HexDigit)
	uriSchemeRune        = peg.Alt(ASCIILetterDigit, peg.S("+-."))
	uriUserInfoRune      = peg.Alt(ASCIILetterDigit, peg.S(":-_.~!$&'()*+,;="), uriEncodedByte)
	uriIPvFutureRune     = peg.Alt(ASCIILetterDigit, peg.S(":-_.~!$&'()*+,;="))
	uriRegNameRune       = peg.Alt(ASCIILetterDigit, peg.S("-_.~!$&'()*+,;="), uriEncodedByte)
	uriPathRune          = peg.Alt(ASCIILetterDigit, peg.S(":@-_.~!$&'()*+,;="), uriEncodedByte)
	uriPathNoSchemeRune  = peg.Alt(ASCIILetterDigit, peg.S("@-_.~!$&'()*+,;="), uriEncodedByte)
	uriQueryFragmentRune = peg.Alt(ASCIILetterDigit, peg.S("?/:@-_.~!$&'()*+,;="), uriEncodedByte)

	// URI host part.
	//     `ipv4 | regname | "[" (ipv6|"v"version"."data) "]"`.
	URIHost = peg.Alt(
		peg.Seq(peg.TI("[v"), peg.Q1(HexDigit), peg.T("."), peg.Q1(uriIPvFutureRune), peg.T("]")),
		peg.Seq(peg.T("["), IPv6, peg.T("]")),
		IPv4,
		peg.Q1(uriRegNameRune))

	// URI authority and path part:
	//     `//[userinfo@]host[:port]/path`.
	uriAuthorityAndPath = peg.Seq(
		peg.T("//"),
		peg.Seq(
			peg.Q01(peg.Seq(peg.NG("userinfo", peg.Q1(uriUserInfoRune)), peg.T("@"))),
			peg.NG("host", URIHost),
			peg.Q01(peg.Seq(peg.T(":"), peg.NG("port", DecUint16)))),
		peg.NG("path",
			peg.Q0(peg.Seq(peg.T("/"), peg.Q0(uriPathRune)))))

	// URI without fragment.
	URIAbsolute = peg.Seq(
		peg.NG("scheme", peg.Seq(ASCIILetter, peg.Q0(uriSchemeRune))),
		peg.Alt(
			uriAuthorityAndPath,
			// or use bare path.
			peg.NG("path", peg.Alt(
				peg.Seq(
					peg.Q01(peg.T("/")),
					peg.Q1(uriPathRune),
					peg.Q0(peg.Seq(peg.T("/"), peg.Q0(uriPathRune)))),
				peg.T("/"),
				peg.True))),
		// query part.
		peg.Q01(peg.Seq(peg.T("?"), peg.NG("query", peg.Q0(uriQueryFragmentRune)))))

	// URI.
	URI = peg.Seq(
		// uri without fragment.
		URIAbsolute,
		// fragment.
		peg.Q01(peg.Seq(peg.T("#"), peg.NG("fragment", peg.Q0(uriQueryFragmentRune)))))

	// URI as reference.
	URIReference = peg.Alt(
		// has scheme.
		URI,
		// no scheme.
		peg.Seq(
			peg.Alt(
				uriAuthorityAndPath,
				// or use bare path, but avoided conflicting with scheme.
				// the first segment of relative path do not use ':'.
				peg.NG("path", peg.Alt(
					peg.Seq(
						peg.T("/"),
						peg.Q1(uriPathRune),
						peg.Q0(peg.Seq(peg.T("/"), peg.Q0(uriPathRune)))),
					peg.Seq(
						peg.Q1(uriPathNoSchemeRune),
						peg.Q0(peg.Seq(peg.T("/"), peg.Q0(uriPathRune)))),
					peg.T("/"),
					peg.True))),
			// query
			peg.Q01(peg.Seq(peg.T("?"), peg.NG("query", peg.Q0(uriQueryFragmentRune)))),
			// fragment
			peg.Q01(peg.Seq(peg.T("#"), peg.NG("fragment", peg.Q0(uriQueryFragmentRune))))))
)

// EMail address described in RFC 5322.
//
// This is a rewrite of the regex taking from http://emailregex.com/ in PEG.
var (
	emailLocalRune   = peg.Alt(ASCIILetterDigit, peg.S("!#$%&'*+/=?^_`{|}~-"))
	emailLocalQuoted = peg.Seq(
		peg.T(`"`),
		peg.Q0(peg.Alt(
			peg.Seq(peg.T(`\`),
				peg.R(
					'\x01', '\x09',
					'\x0b', '\x0c',
					'\x0e', '\x7f')),
			peg.R(
				'\x01', '\x08',
				'\x0b', '\x0c',
				'\x0e', '\x1f',
				'\x21', '\x21',
				'\x23', '\x5b',
				'\x5d', '\x7f'))),
		peg.T(`"`))
	emailDomainData = peg.Seq(
		peg.T(`[`),
		peg.Alt(
			peg.Seq(
				peg.Q0(peg.Alt(ASCIILetterDigit, peg.T("-"))),
				ASCIILetterDigit,
				peg.T(":"),
				peg.Q1(peg.Alt(
					peg.Seq(peg.T(`\`),
						peg.R(
							'\x01', '\x09',
							'\x0b', '\x0c',
							'\x0e', '\x7f')),
					peg.R(
						'\x01', '\x08',
						'\x0b', '\x0c',
						'\x0e', '\x1f',
						'\x21', '\x5a',
						'\x53', '\x7f')))),
			IPv4),
		peg.T(`]`))

	EMail = peg.Seq(
		peg.NG("local", peg.Alt(
			emailLocalQuoted,
			peg.J1(peg.Q1(emailLocalRune), peg.T(".")))),
		peg.T("@"),
		peg.NG("domain", peg.Alt(
			emailDomainData,
			Domain)))
)

// Useful web things
var (
	webLetterHypen = peg.R('a', 'z', 'A', 'Z', '-', '-')

	// URL slug.
	Slug = peg.Q1(webLetterHypen)

	// DNS domain name.
	Domain = peg.Trunc(253,
		peg.Seq(
			peg.Qmn(1, 63, webLetterHypen),
			peg.Qmn(0, 126, peg.Seq(peg.T("."), peg.Qmn(1, 63, webLetterHypen))),
			peg.Q01(peg.T("."))))
)

// helpers.

func ipv6EllipsisIPv4(s string) (n int, ok bool) {
	split := strings.Split(s, "::")
	left := strings.Split(split[0], ":")
	noipv4 := strings.TrimSuffix(strings.TrimRight(split[1], ".0123456789"), ":")
	right := strings.Split(noipv4, ":")

	if len(left)+len(right) <= 5 {
		return len(s), true
	}
	if len(left)+len(right) <= 7 {
		return len(split[0]) + 2 + len(noipv4), true
	}
	right = right[:7-len(left)]
	rightlen := len(strings.Join(right, ":"))
	return len(split[0]) + 2 + rightlen, true
}

func ipv6EllipsisNoIPv4(s string) (n int, ok bool) {
	split := strings.Split(s, "::")
	left := strings.Split(split[0], ":")
	right := strings.Split(split[1], ":")

	if len(left)+len(right) <= 7 {
		return len(s), true
	}
	right = right[:7-len(left)]
	rightlen := len(strings.Join(right, ":"))
	return len(split[0]) + 2 + rightlen, true
}
