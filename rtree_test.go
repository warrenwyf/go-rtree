package rtree

import (
	"fmt"
	"math/rand"
	"testing"
)

type Point struct {
	x  int
	y  int
	id string
}

func (p *Point) Mbr() Mbr {
	return NewMbrInt32([]int32{int32(p.x), int32(p.y)}, []int32{0, 0})
}

func (p *Point) Equals(f Feature) bool {
	p2, ok := f.(*Point)
	if !ok {
		return false
	}

	return p.id == p2.id
}

func (p *Point) String() string {
	return fmt.Sprintf("{x=%d, y=%d, id=%d}", p.x, p.y, p.id)
}

func Test_Insert_Search(t *testing.T) {
	tree := NewRtree(2, 16)

	nx := 100
	ny := 100

	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			feature := &Point{i, j, fmt.Sprintf("%d-%d", i, j)}
			tree.Insert(feature)
		}
	}

	for k := 0; k < nx*ny; k++ {
		x := rand.Intn(nx)
		y := rand.Intn(ny)

		result := tree.Search(NewMbrInt32([]int32{int32(x), int32(y)}, []int32{0, 0, 0}))

		if len(result) != 1 {
			t.Errorf("Search() got wrong result: %s", result)
			break
		}

		pt := result[0].(*Point)

		if pt.x != x || pt.y != y || pt.id != fmt.Sprintf("%d-%d", x, y) {
			t.Errorf("Search() got wrong result: %s", pt)
			break
		}
	}

	t.Logf("Tree height = %d", tree.Height())
}

func Test_Load_Search(t *testing.T) {
	nx := 100
	ny := 100

	features := []Feature{}
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			feature := &Point{i, j, fmt.Sprintf("%d-%d", i, j)}
			features = append(features, feature)
		}
	}

	tree := NewRtree(2, 16, features...)

	for k := 0; k < nx*ny; k++ {
		x := rand.Intn(nx)
		y := rand.Intn(ny)

		result := tree.Search(NewMbrInt32([]int32{int32(x), int32(y)}, []int32{0, 0, 0}))

		if len(result) != 1 {
			t.Errorf("Search() got wrong result: %s", result)
			break
		}

		pt := result[0].(*Point)

		if pt.x != x || pt.y != y || pt.id != fmt.Sprintf("%d-%d", x, y) {
			t.Errorf("Search() got wrong result: %s", pt)
			break
		}
	}

	t.Logf("Tree height = %d", tree.Height())
}

func Test_Search_1M(t *testing.T) {
	nx := 1000
	ny := 1000

	features := []Feature{}
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			feature := &Point{i, j, fmt.Sprintf("%d-%d", i, j)}
			features = append(features, feature)
		}
	}

	tree := NewRtree(2, 16, features...)
	t.Logf("Tree size = %d", tree.Size())

	for k := 0; k < 10000; k++ {
		x := rand.Intn(nx)
		y := rand.Intn(ny)

		result := tree.Search(NewMbrInt32([]int32{int32(x), int32(y)}, []int32{0, 0, 0}))

		if len(result) != 1 {
			t.Errorf("Search() got wrong result: %s", result)
		}

		pt := result[0].(*Point)

		if pt.x != x || pt.y != y || pt.id != fmt.Sprintf("%d-%d", x, y) {
			t.Errorf("Search() got wrong result: %s", pt)
		}
	}

	t.Logf("Tree height = %d", tree.Height())
}
