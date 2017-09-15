package pegutil

import (
	"strings"

	"github.com/hucsmn/peg"
)

// Hardware addresses.
var (
	hwaddrMAC = peg.Alt(
		peg.Jnn(6, peg.Qnn(2, HexDigit), peg.T(":")),
		peg.Jnn(6, peg.Qnn(2, HexDigit), peg.T("-")),
		peg.Jnn(3, peg.Qnn(4, HexDigit), peg.T(".")))
	hwaddrEUI64 = peg.Alt(
		peg.Jnn(8, peg.Qnn(2, HexDigit), peg.T(":")),
		peg.Jnn(8, peg.Qnn(2, HexDigit), peg.T("-")),
		peg.Jnn(4, peg.Qnn(4, HexDigit), peg.T(".")))

	// Hardware addresses (EUI-48 and EUI-64).
	MAC   = hwaddrMAC
	EUI64 = hwaddrEUI64
)

// IP addresses.
var (
	ipaddrIPv4 = peg.Jnn(4, NoRedundantZeroes(DecUint8), peg.T("."))
	ipaddrIPv6 = peg.Alt(
		// ellipsis with trailing 32-bit dot-decimals
		// (e.g. ::192.168.0.1).
		peg.Inject(ipv6EllipsisIPv4,
			peg.Seq(
				peg.Jmn(0, 5, SimpleHexUint16, peg.T(":")),
				peg.T("::"),
				peg.Alt(
					peg.Seq(
						peg.Jmn(1, 5,
							peg.Seq(SimpleHexUint16, peg.Not(peg.T("."))), // avoid matches ipv4 part
							peg.T(":")),
						peg.Seq(peg.T(":"), ipaddrIPv4)),
					ipaddrIPv4))),
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
			peg.Seq(peg.T(":"), ipaddrIPv4)))
	ipaddrIP = peg.Alt(ipaddrIPv6, ipaddrIPv4)

	ipaddrCIDRv4 = peg.Seq(ipaddrIPv4, peg.T("/"), DecIntegerBetween(0, 32))
	ipaddrCIDRv6 = peg.Seq(ipaddrIPv6, peg.T("/"), DecIntegerBetween(0, 128))
	ipaddrCIDR   = peg.Alt(ipaddrCIDRv6, ipaddrCIDRv4)

	ipaddrIPv6WithZone = peg.Seq(ipaddrIPv6, peg.T("%"), DecInteger)

	// IP addresses.
	IP   = ipaddrIP
	IPv4 = ipaddrIPv4
	IPv6 = ipaddrIPv6

	// IP addresses with subnetwork mask.
	CIDR   = ipaddrCIDR
	CIDRv4 = ipaddrCIDRv4
	CIDRv6 = ipaddrCIDRv6

	// IPv6 address with subnetwork mask.
	IPv6WithZone = ipaddrIPv6WithZone
)

// URI definitions described in RFC 3986.
var (
	// The rune sets described in RFC 3986.
	uriEncodedByte       = peg.Seq(peg.T("%"), HexDigit, HexDigit)
	uriSchemeRune        = peg.Alt(ASCIILetterDigit, peg.S("+-."))
	uriUserInfoRune      = peg.Alt(ASCIILetterDigit, peg.S(":-_.~!$&'()*+,;="), uriEncodedByte)
	uriIPvFutureRune     = peg.Alt(ASCIILetterDigit, peg.S(":-_.~!$&'()*+,;="))
	uriRegNameRune       = peg.Alt(ASCIILetterDigit, peg.S("-_.~!$&'()*+,;="), uriEncodedByte)
	uriPathRune          = peg.Alt(ASCIILetterDigit, peg.S(":@-_.~!$&'()*+,;="), uriEncodedByte)
	uriPathNoSchemeRune  = peg.Alt(ASCIILetterDigit, peg.S("@-_.~!$&'()*+,;="), uriEncodedByte)
	uriQueryFragmentRune = peg.Alt(ASCIILetterDigit, peg.S("?/:@-_.~!$&'()*+,;="), uriEncodedByte)

	// Basic parts:
	//   host := ipv4 | regname | '[' ( ipv6 | 'v' version '.' data ) ']' .
	uriHost = peg.Alt(
		peg.Seq(peg.TI("[v"), peg.Q1(HexDigit), peg.T("."), peg.Q1(uriIPvFutureRune), peg.T("]")),
		peg.Seq(peg.T("["), ipaddrIPv6, peg.T("]")),
		ipaddrIPv4,
		peg.Q1(uriRegNameRune))

	// Basic parts:
	//   authority := '//' [ userinfo '@' ] host [ ':' port ] .
	uriAuthority = peg.Seq(
		peg.T("//"),
		peg.Q01(peg.Seq(
			peg.NG("userinfo", peg.Q1(uriUserInfoRune)),
			peg.T("@"))),
		peg.NG("host", uriHost),
		peg.Q01(peg.Seq(
			peg.T(":"),
			peg.NG("port", DecUint16))))

	// URI without fragment.
	uriAbsolute = peg.Seq(
		peg.NG("scheme", peg.Seq(ASCIILetter, peg.Q0(uriSchemeRune))),
		peg.Alt(
			// authority and path.
			peg.Seq(
				uriAuthority,
				peg.NG("path",
					peg.Q0(peg.Seq(
						peg.T("/"),
						peg.Q0(uriPathRune))))),
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

	// Standard URI.
	uriStandard = peg.Seq(
		// uri without fragment.
		uriAbsolute,
		// fragment part.
		peg.Q01(peg.Seq(peg.T("#"), peg.NG("fragment", peg.Q0(uriQueryFragmentRune)))))

	// URI as a reference.
	uriReference = peg.Alt(
		// has scheme.
		uriStandard,
		// no scheme.
		peg.Seq(
			peg.Alt(
				// authority and path.
				peg.Seq(
					uriAuthority,
					peg.NG("path",
						peg.Q0(peg.Seq(
							peg.T("/"),
							peg.Q0(uriPathRune))))),
				// or use bare path, avoided conflicting with scheme:
				// rune ':' is forbidden in the first segment of relative path.
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
			// query part.
			peg.Q01(peg.Seq(peg.T("?"), peg.NG("query", peg.Q0(uriQueryFragmentRune)))),
			// fragment part.
			peg.Q01(peg.Seq(peg.T("#"), peg.NG("fragment", peg.Q0(uriQueryFragmentRune))))))

	URI         = uriStandard
	AbsoluteURI = uriAbsolute
	HRef        = uriReference
	Host        = uriHost
)

// Useful web things.
var (
	webLetterHypen = peg.R('a', 'z', 'A', 'Z', '-', '-')
	webSlug        = peg.Q1(webLetterHypen)
	webDomain      = peg.Trunc(253,
		peg.Seq(
			peg.Jmn(1, 127, peg.Qmn(1, 63, webLetterHypen), peg.Seq(peg.T("."))),
			peg.Q01(peg.T("."))))

	// URL slug.
	Slug = webSlug

	// DNS domain name.
	Domain = webDomain
)

// EMail address described in RFC 5322.
//
// This is a rewrite of the simplified regexp taking from http://emailregex.com/.
var (
	emailLocalRune   = peg.Alt(ASCIILetterDigit, peg.S("!#$%&'*+/=?^_`{|}~-"))
	emailLocalQuoted = peg.Seq(
		peg.T(`"`),
		peg.Q0(
			peg.Seq(peg.T(`\`),
				peg.R(
					'\x01', '\x09',
					'\x0b', '\x0c',
					'\x0e', '\x7f')),
			peg.R(
				'\x01', '\x08',
				'\x0b', '\x0c',
				'\x0e', '\x1f',
				'\x20', '\x21',
				'\x23', '\x5b',
				'\x5d', '\x7f')),
		peg.T(`"`))
	emailDomainData = peg.Seq(
		peg.T(`[`),
		peg.Alt(
			peg.Seq(
				peg.Q0(peg.Seq(peg.Alt(ASCIILetterDigit, peg.T("-")), peg.Not(peg.T(":")))),
				ASCIILetterDigit,
				peg.T(":"),
				peg.Q1(
					peg.R(
						'\x01', '\x08',
						'\x0b', '\x0c',
						'\x0e', '\x1f',
						'\x21', '\x5a',
						'\x5e', '\x7f'),
					peg.Seq(peg.T(`\`),
						peg.R(
							'\x01', '\x09',
							'\x0b', '\x0c',
							'\x0e', '\x7f')))),
			ipaddrIPv4),
		peg.T(`]`))

	email = peg.Seq(
		peg.NG("local", peg.Trunc(64, peg.Alt(
			emailLocalQuoted,
			peg.J1(peg.Q1(emailLocalRune), peg.T("."))))),
		peg.T("@"),
		peg.NG("domain", peg.Trunc(255, peg.Alt(
			emailDomainData,
			peg.Seq(
				peg.J1(peg.Q1(ASCIILetterDigit, peg.T("-")), peg.T(".")),
				peg.Q01(peg.T(".")))))))

	// E-mail address.
	EMail = email
)

// Helpers for IPv6 address validation.

func ipv6EllipsisIPv4(s string) (n int, ok bool) {
	split := strings.Split(s, "::")
	left := strings.Split(split[0], ":")
	nleft := len(left)
	if split[0] == "" {
		nleft = 0
	}
	noipv4 := strings.TrimSuffix(strings.TrimRight(split[1], ".0123456789"), ":")
	right := strings.Split(noipv4, ":")
	nright := len(right)
	if noipv4 == "" {
		nright = 0
	}

	if nleft+nright <= 5 {
		return len(s), true
	}
	if nleft+nright <= 7 {
		return len(split[0]) + 2 + len(noipv4), true
	}
	right = right[:7-len(left)]
	rightlen := len(strings.Join(right, ":"))
	return len(split[0]) + 2 + rightlen, true
}

func ipv6EllipsisNoIPv4(s string) (n int, ok bool) {
	split := strings.Split(s, "::")
	left := strings.Split(split[0], ":")
	nleft := len(left)
	if split[0] == "" {
		nleft = 0
	}
	right := strings.Split(split[1], ":")
	nright := len(right)
	if split[1] == "" {
		nright = 0
	}

	if nleft+nright <= 7 {
		return len(s), true
	}
	right = right[:7-len(left)]
	rightlen := len(strings.Join(right, ":"))
	return len(split[0]) + 2 + rightlen, true
}
