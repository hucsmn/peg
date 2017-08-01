package peg

import (
	"unicode"
	"unicode/utf8"
)

var (
	// performance workaround: avoid ASCII runes being treated as special rune.
	foldCaseWorkAround = map[rune]rune{
		'ſ': 'ſ', // => ASCII 'S'
		'K': 'K', // => ASCII 'K'
	}

	// runes that utf8len(runeFoldCase(r)) != utf8len(r).
	lengthChangedAfterFoldCase = map[rune]Pattern{
		'ᲀ': S("ᲀВв"),
		'В': S("ᲀВв"),
		'в': S("ᲀВв"),
		'ᲁ': S("ᲁДд"),
		'Д': S("ᲁДд"),
		'д': S("ᲁДд"),
		'ᲂ': S("ᲂОо"),
		'О': S("ᲂОо"),
		'о': S("ᲂОо"),
		'ᲃ': S("ᲃСс"),
		'С': S("ᲃСс"),
		'с': S("ᲃСс"),
		'ᲄ': S("ᲄᲅТт"),
		'ᲅ': S("ᲄᲅТт"),
		'Т': S("ᲄᲅТт"),
		'т': S("ᲄᲅТт"),
		'ᲆ': S("ᲆЪъ"),
		'Ъ': S("ᲆЪъ"),
		'ъ': S("ᲆЪъ"),
		'ᲇ': S("ᲇѢѣ"),
		'Ѣ': S("ᲇѢѣ"),
		'ѣ': S("ᲇѢѣ"),
		'ẞ': S("ẞß"),
		'ß': S("ẞß"),
		'ι': S("ιͅΙι"),
		'ͅ': S("ιͅΙι"),
		'Ι': S("ιͅΙι"),
		'ι': S("ιͅΙι"),
		'Ω': S("ΩΩω"),
		'Ω': S("ΩΩω"),
		'ω': S("ΩΩω"),
		'Å': S("ÅÅå"),
		'Å': S("ÅÅå"),
		'å': S("ÅÅå"),
		'Ɫ': S("Ɫɫ"),
		'ɫ': S("Ɫɫ"),
		'Ɽ': S("Ɽɽ"),
		'ɽ': S("Ɽɽ"),
		'ⱥ': S("ⱥȺ"),
		'Ⱥ': S("ⱥȺ"),
		'ⱦ': S("ⱦȾ"),
		'Ⱦ': S("ⱦȾ"),
		'Ɑ': S("Ɑɑ"),
		'ɑ': S("Ɑɑ"),
		'Ɱ': S("Ɱɱ"),
		'ɱ': S("Ɱɱ"),
		'Ɐ': S("Ɐɐ"),
		'ɐ': S("Ɐɐ"),
		'Ɒ': S("Ɒɒ"),
		'ɒ': S("Ɒɒ"),
		'Ȿ': S("Ȿȿ"),
		'ȿ': S("Ȿȿ"),
		'Ɀ': S("Ɀɀ"),
		'ɀ': S("Ɀɀ"),
		'Ɥ': S("Ɥɥ"),
		'ɥ': S("Ɥɥ"),
		'Ɦ': S("Ɦɦ"),
		'ɦ': S("Ɦɦ"),
		'Ɜ': S("Ɜɜ"),
		'ɜ': S("Ɜɜ"),
		'Ɡ': S("Ɡɡ"),
		'ɡ': S("Ɡɡ"),
		'Ɬ': S("Ɬɬ"),
		'ɬ': S("Ɬɬ"),
		'Ɪ': S("Ɪɪ"),
		'ɪ': S("Ɪɪ"),
		'Ʞ': S("Ʞʞ"),
		'ʞ': S("Ʞʞ"),
		'Ʇ': S("Ʇʇ"),
		'ʇ': S("Ʇʇ"),
		'Ʝ': S("Ʝʝ"),
		'ʝ': S("Ʝʝ"),
	}
)

// Check if string would keep its length unchanged after case folding.
func couldSafelyFoldCase(text string) bool {
	for _, r := range text {
		if _, ok := lengthChangedAfterFoldCase[r]; ok {
			return false
		}
	}
	return true
}

// Unicode case folding for strings (with several workarounds).
func foldCase(s string) string {
	encoded := make([]byte, 0, 16)
	buf := make([]byte, 4)
	for i, r := range s {
		if r == unicode.ReplacementChar {
			encoded = append(encoded, s[i])
		} else {
			n := utf8.EncodeRune(buf, runeFoldCase(r))
			encoded = append(encoded, buf[:n]...)
		}
	}
	return string(encoded)
}

// Unicode case folding for runes (with several workarounds).
func runeFoldCase(r rune) rune {
	if w, ok := foldCaseWorkAround[r]; ok {
		return w
	}

	r0 := unicode.SimpleFold(r)
	if r0 == r {
		return r
	}
	for r0 > r {
		r0 = unicode.SimpleFold(r0)
	}
	return r0
}
