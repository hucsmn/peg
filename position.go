package peg

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Position is offset and line-column numbers counting from zero.
type Position struct {
	Offest int
	Line   int
	Column int
}

func (pos *Position) String() string {
	return fmt.Sprintf("%d:%d+%d", pos.Line+1, pos.Column+1, pos.Offest)
}

// Simple utf-8 string line column calculator.
type positionCalculator struct {
	text   string
	cached int   // cached to where
	lnends []int // found "\r"|"\n"|"\r\n" line endings
}

func (calc *positionCalculator) calculate(offset int) Position {
	ln, lnstart := calc.search(offset)
	col := utf8.RuneCountInString(calc.text[lnstart:offset])
	return Position{
		Offest: offset,
		Line:   ln,
		Column: col,
	}
}

func (calc *positionCalculator) search(offset int) (ln, lnstart int) {
	calc.caching(offset)
	if len(calc.lnends) == 0 {
		return 0, 0
	}

	i, j := 0, len(calc.lnends)
	for i < j {
		m := i + (j-i)/2
		if offset > calc.lnends[m] {
			i = m + 1
		} else if offset < calc.lnends[m] {
			j = m
		} else {
			return m + 1, offset
		}
	}
	return i, calc.lnends[i-1]
}

func (calc *positionCalculator) caching(to int) {
	for ; calc.cached < to; calc.cached++ {
		switch calc.text[calc.cached] {
		case '\n':
			calc.lnends = append(calc.lnends, calc.cached+1)
		case '\r':
			if !strings.HasPrefix(calc.text[calc.cached+1:], "\n") {
				calc.lnends = append(calc.lnends, calc.cached+1)
			}
		}
	}
}
