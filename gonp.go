// The algorithm implemented here is based on "An O(NP) Sequence Comparison Algorithm"
// by described by Sun Wu, Udi Manber and Gene Myers

package gonp

import (
	"container/list"
	"fmt"
	"unicode/utf8"
)

const (
	// SesDelete is manipulaton type of deleting element in SES
	SesDelete SesType = iota
	// SesCommon is manipulaton type of same element in SES
	SesCommon
	// SesAdd is manipulaton type of adding element in SES
	SesAdd
)

// SesType is manipulaton type
type SesType int

// Point is coordinate in edit graph
type Point struct {
	x, y, k int
}

// SesElem is element of SES
type SesElem struct {
	c rune
	t SesType
}

// Diff is context for calculating difference between a and b
type Diff struct {
	a    []rune
	b    []rune
	m, n int
	ed   int
	meta Meta
	lcs  *list.List
	ses  *list.List
}

// Meta is internal state for calculating difference
type Meta struct {
	reverse  bool
	path     []int
	onlyEd   bool
	pathposi map[int]Point
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// New is initializer of Diff
func New(a string, b string) *Diff {
	m, n := utf8.RuneCountInString(a), utf8.RuneCountInString(b)
	diff := new(Diff)
	diff.a, diff.b = []rune(a), []rune(b)
	diff.m, diff.n = m, n
	diff.meta.reverse = false
	if m >= n {
		diff.a, diff.b = diff.b, diff.a
		diff.m, diff.n = n, m
		diff.meta.reverse = true
	}
	diff.meta.onlyEd = false
	return diff
}

// OnlyEd enables to calculate only edit distance
func (diff *Diff) OnlyEd() {
	diff.meta.onlyEd = true
}

// Editdistance returns edit distance between a and b
func (diff *Diff) Editdistance() int {
	return diff.ed
}

// Lcs returns LCS (Longest Common Subsequence) between a and b
func (diff *Diff) Lcs() string {
	var b = make([]rune, diff.lcs.Len())
	for i, e := 0, diff.lcs.Front(); e != nil; i, e = i+1, e.Next() {
		b[i] = e.Value.(rune)
	}
	return string(b)
}

// Ses return SES (Shortest Edit Script) between a and b
func (diff *Diff) Ses() []SesElem {
	seq := make([]SesElem, diff.ses.Len())
	for i, e := 0, diff.ses.Front(); e != nil; i, e = i+1, e.Next() {
		seq[i].c = e.Value.(SesElem).c
		seq[i].t = e.Value.(SesElem).t
	}
	return seq
}

// PrintSes prints shortest edit script between a and b
func (diff *Diff) PrintSes() {
	for _, e := 0, diff.ses.Front(); e != nil; e = e.Next() {
		ee := e.Value.(SesElem)
		switch ee.t {
		case SesDelete:
			fmt.Printf("- %c\n", ee.c)
		case SesAdd:
			fmt.Printf("+ %c\n", ee.c)
		case SesCommon:
			fmt.Printf("  %c\n", ee.c)
		}
	}
}

// Compose composes diff between a and b
func (diff *Diff) Compose() {
	offset := diff.m + 1
	delta := diff.n - diff.m
	size := diff.m + diff.n + 3
	fp := make([]int, size)
	diff.meta.path = make([]int, size)
	diff.meta.pathposi = make(map[int]Point)
	diff.lcs = list.New()
	diff.ses = list.New()

	for i := range fp {
		fp[i] = -1
		diff.meta.path[i] = -1
	}

	for p := 0; ; p++ {

		for k := -p; k <= delta-1; k++ {
			fp[k+offset] = diff.snake(k, fp[k-1+offset]+1, fp[k+1+offset], offset)
		}

		for k := delta + p; k >= delta+1; k-- {
			fp[k+offset] = diff.snake(k, fp[k-1+offset]+1, fp[k+1+offset], offset)
		}

		fp[delta+offset] = diff.snake(delta, fp[delta-1+offset]+1, fp[delta+1+offset], offset)

		if fp[delta+offset] >= diff.n {
			diff.ed = delta + 2*p
			break
		}
	}

	if diff.meta.onlyEd {
		return
	}

	r := diff.meta.path[delta+offset]
	epc := make(map[int]Point)
	for r != -1 {
		epc[len(epc)] = Point{x: diff.meta.pathposi[r].x, y: diff.meta.pathposi[r].y, k: -1}
		r = diff.meta.pathposi[r].k
	}
	diff.recordSeq(epc)
}

func (diff *Diff) snake(k, p, pp, offset int) int {
	r := 0
	if p > pp {
		r = diff.meta.path[k-1+offset]
	} else {
		r = diff.meta.path[k+1+offset]
	}

	y := max(p, pp)
	x := y - k

	for x < diff.m && y < diff.n && diff.a[x] == diff.b[y] {
		x++
		y++
	}

	if !diff.meta.onlyEd {
		diff.meta.path[k+offset] = len(diff.meta.pathposi)
		diff.meta.pathposi[len(diff.meta.pathposi)] = Point{x: x, y: y, k: r}
	}

	return y
}

func (diff *Diff) recordSeq(epc map[int]Point) {
	xIdx, yIdx := 1, 1
	pxIdx, pyIdx := 0, 0
	for i := len(epc) - 1; i >= 0; i-- {
		for (pxIdx < epc[i].x) || (pyIdx < epc[i].y) {
			var t SesType
			if (epc[i].y - epc[i].x) > (pyIdx - pxIdx) {
				elem := diff.b[pyIdx]
				if diff.meta.reverse {
					t = SesDelete
				} else {
					t = SesAdd
				}
				diff.ses.PushBack(SesElem{c: elem, t: t})
				yIdx++
				pyIdx++
			} else if epc[i].y-epc[i].x < pyIdx-pxIdx {
				elem := diff.a[pxIdx]
				if diff.meta.reverse {
					t = SesAdd
				} else {
					t = SesDelete
				}
				diff.ses.PushBack(SesElem{c: elem, t: t})
				xIdx++
				pxIdx++
			} else {
				elem := diff.a[pxIdx]
				t = SesCommon
				diff.lcs.PushBack(elem)
				diff.ses.PushBack(SesElem{c: elem, t: t})
				xIdx++
				yIdx++
				pxIdx++
				pyIdx++
			}
		}
	}
}
