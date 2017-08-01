package peg

import "testing"

// Test if position calculator works correctly.
func TestPositionCalculator(t *testing.T) {
	data := []struct {
		text    string
		inputs  []int
		outputs []Position
	}{
		{"", []int{0}, []Position{{0, 0, 0}}},
		{"A\n", []int{0, 1, 2}, []Position{
			{0, 0, 0},
			{1, 0, 1},
			{2, 1, 0},
		}},
		{"\nAA\r\r\nA\n\n", []int{1, 3, 4, 5, 6, 9}, []Position{
			{1, 1, 0},
			{3, 1, 2},
			{4, 2, 0},
			{5, 2, 1},
			{6, 3, 0},
			{9, 5, 0},
		}},
		{"\nAA\r\r\nA\n\n", []int{1, 5, 3, 4, 6, 9}, []Position{
			{1, 1, 0},
			{5, 2, 1},
			{3, 1, 2},
			{4, 2, 0},
			{6, 3, 0},
			{9, 5, 0},
		}},
	}

	for _, d := range data {
		pcalc := &positionCalculator{text: d.text}
		for i := range d.inputs {
			pos := pcalc.calculate(d.inputs[i])
			if d.outputs[i] != pos {
				t.Errorf("%q.position(%d) => %v != %v (lnends=%v)",
					d.text, d.inputs[i], pos, d.outputs[i], pcalc.lnends)
			}
		}
	}
}
