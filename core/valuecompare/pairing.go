package valuecompare

import "sort"

// BestAssignment is the largest total over one-to-one pairings of the rows of
// score with its columns, and the pairing that reaches it: rowTo[i] is the
// column row i pairs with, or -1. A cell pairs only when positive, so a word
// with nothing in common stays unpaired. The search is exact over the shorter
// side, so it does not depend on which side is the row; sides longer than
// maxExact pair greedily by descending score. rowTo is in score's own
// orientation.
func BestAssignment(score [][]float64, maxExact int) (total float64, rowTo []int) {
	rows := len(score)
	rowTo = make([]int, rows)
	for i := range rowTo {
		rowTo[i] = -1
	}
	if rows == 0 || len(score[0]) == 0 {
		return 0, rowTo
	}
	cols := len(score[0])
	transposed := false
	at := func(i, j int) float64 { return score[i][j] }
	if rows < cols {
		transposed = true
		rows, cols = cols, rows
		at = func(i, j int) float64 { return score[j][i] }
	}
	assign := make([]int, rows)
	for i := range assign {
		assign[i] = -1
	}
	if cols > maxExact {
		total = greedyAssign(rows, cols, at, assign)
	} else {
		total = exactAssign(rows, cols, at, assign)
	}
	if !transposed {
		return total, assign
	}
	// assign is indexed by the transposed rows: the original columns.
	for col, row := range assign {
		if row >= 0 {
			rowTo[row] = col
		}
	}
	return total, rowTo
}

func exactAssign(rows, cols int, at func(i, j int) float64, assign []int) float64 {
	type step struct {
		col  int
		mask int
	}
	best := make([]float64, 1<<cols)
	steps := make([][]step, rows)
	for i := 0; i < rows; i++ {
		next := append([]float64(nil), best...)
		ch := make([]step, 1<<cols)
		for mask := range ch {
			ch[mask] = step{col: -1, mask: mask}
		}
		for mask, v := range best {
			for j := 0; j < cols; j++ {
				s := at(i, j)
				if mask&(1<<j) != 0 || s <= 0 {
					continue
				}
				nm := mask | (1 << j)
				if t := v + s; t > next[nm] {
					next[nm] = t
					ch[nm] = step{col: j, mask: mask}
				}
			}
		}
		best = next
		steps[i] = ch
	}
	topMask := 0
	var top float64
	for mask, v := range best {
		if v > top {
			top = v
			topMask = mask
		}
	}
	mask := topMask
	for i := rows - 1; i >= 0; i-- {
		ch := steps[i][mask]
		if ch.col >= 0 {
			assign[i] = ch.col
		}
		mask = ch.mask
	}
	return top
}

func greedyAssign(rows, cols int, at func(i, j int) float64, assign []int) float64 {
	type pair struct {
		i, j int
		s    float64
	}
	var pairs []pair
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if s := at(i, j); s > 0 {
				pairs = append(pairs, pair{i, j, s})
			}
		}
	}
	sort.SliceStable(pairs, func(a, b int) bool { return pairs[a].s > pairs[b].s })
	colUsed := make([]bool, cols)
	var total float64
	for _, p := range pairs {
		if assign[p.i] < 0 && !colUsed[p.j] {
			assign[p.i] = p.j
			colUsed[p.j] = true
			total += p.s
		}
	}
	return total
}
