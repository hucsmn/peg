package peg

// Internal search structure for TS/TSI.
type prefixTree struct {
	term  bool     // if searching could be terminated here
	width int      // length of each key
	keys  []string // sorted keys
	subs  []prefixTree
}

// Build prefix tree from a collection of texts.
func buildPrefixTree(sorted []string) prefixTree {
	tree := prefixTree{}
	var i int
	for ; i < len(sorted) && sorted[i] == ""; i++ {
		tree.term = true
	}
	sorted = sorted[i:]
	if len(sorted) == 0 {
		return tree
	}

	tree.width = len(sorted[0])
	for _, s := range sorted {
		if len(s) < tree.width {
			tree.width = len(s)
		}
	}

	var lastprefix = sorted[0][:tree.width]
	var lasttail = sorted[0][tree.width:]
	var tails = []string{lasttail}
	for _, s := range sorted[1:] {
		prefix, tail := s[:tree.width], s[tree.width:]
		if prefix == lastprefix {
			if tail != lasttail {
				tails = append(tails, tail)
				lasttail = tail
			}
		} else {
			tree.keys = append(tree.keys, lastprefix)
			tree.subs = append(tree.subs, buildPrefixTree(tails))
			lastprefix = prefix
			lasttail = tail
			tails = []string{lasttail}
		}
	}
	tree.keys = append(tree.keys, lastprefix)
	tree.subs = append(tree.subs, buildPrefixTree(tails))
	return tree
}

// Search the given prefix, where len(prefix) == tree.width.
func (tree prefixTree) search(s string) (int, bool) {
	if len(s) != tree.width {
		return 0, false
	}

	i, j := 0, len(tree.keys)
	for i < j {
		m := i + (j-i)/2
		if s == tree.keys[m] {
			return m, true
		} else if s > tree.keys[m] {
			i = m + 1
		} else {
			j = m
		}
	}
	return 0, false
}
