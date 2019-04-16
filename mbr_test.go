package rtree

import (
	"testing"
)

func Test_MbrInt32(t *testing.T) {
	var mbr Mbr = NewMbrInt32([]int32{1, 2}, []int32{0, 0})

	if mbr == nil {
		t.Errorf("NewMbrInt32() failed")
	}

	if mbr.Type() != MbrTypeInt32 {
		t.Errorf("Wrong type")
	}

	if mbr.Dim() != 2 {
		t.Errorf("Wrong dimension")
	}
}

func Test_MbrInt32_Equals(t *testing.T) {
	mbr1 := NewMbrInt32([]int32{0, 0, 0, 0}, []int32{4, 4, 4, 4})
	mbr2 := NewMbrInt32([]int32{0, 0, 0, 0}, []int32{4, 4, 4, 4})

	if !mbr1.Equals(mbr2) {
		t.Errorf("MbrInt32.Equals() got wrong result")
	}
}

func Test_MbrInt32_Contains(t *testing.T) {
	mbr1 := NewMbrInt32([]int32{0, 0, 0, 0}, []int32{4, 4, 4, 4})
	mbr2 := NewMbrInt32([]int32{1, 1, 1, 0}, []int32{3, 3, 3, 4})
	mbr3 := NewMbrInt32([]int32{2, 2, 2, 2}, []int32{4, 4, 4, 4, 1})

	if !mbr1.Contains(mbr2) {
		t.Errorf("mbr1.Contains(mbr2) got wrong result, mbr1=%s, mbr2=%s", mbr1, mbr2)
	}

	if mbr1.Contains(mbr3) {
		t.Errorf("mbr1.Contains(mbr3) got wrong result, mbr1=%s, mbr3=%s", mbr1, mbr3)
	}
}

func Test_MbrInt32_Intersects(t *testing.T) {
	mbr1 := NewMbrInt32([]int32{0, 0, 0, 0}, []int32{4, 4, 4, 4})
	mbr2 := NewMbrInt32([]int32{1, 1, 1, 1}, []int32{2, 2, 2, 2})
	mbr3 := NewMbrInt32([]int32{10, 10, 10, 10}, []int32{1, 1, 1, 1, 1})

	if !mbr1.Intersects(mbr2) {
		t.Errorf("mbr1.Intersects(mbr2) got wrong result, mbr1=%s, mbr2=%s", mbr1, mbr2)
	}

	if mbr1.Intersects(mbr3) {
		t.Errorf("mbr1.Intersects(mbr3) got wrong result, mbr1=%s, mbr3=%s", mbr1, mbr3)
	}
}

func Test_MbrInt32_Clone(t *testing.T) {
	mbr1 := NewMbrInt32([]int32{0, 0}, []int32{4, 4})
	mbr2 := mbr1.Clone()
	(*mbr1)[0] += 1

	if mbr1.Equals(mbr2) {
		t.Errorf("MbrInt32.Clone() got wrong result")
	}
}

func Test_MergeMbrs(t *testing.T) {
	mbr1 := NewMbrInt32([]int32{0, 0, 0}, []int32{2, 2, 2})
	mbr2 := NewMbrInt32([]int32{1, 1, 1}, []int32{2, 2, 2})
	mbr3 := NewMbrInt32([]int32{3, 3, 3}, []int32{2, 2, 2})

	mbr := MergeMbrs(mbr1, mbr2, mbr3)
	if !mbr.Equals(NewMbrInt32([]int32{0, 0, 0}, []int32{5, 5, 5})) {
		t.Errorf("MergeMbrs() got wrong result")
	}
}

func Test_MergeMbrsTypeDismatch(t *testing.T) {
	mbr1 := NewMbrInt32([]int32{0, 0, 0}, []int32{2, 2, 2})
	mbr2 := NewMbrFloat64([]float64{1.0, 2.0}, []float64{2, 2})

	mbr := MergeMbrs(mbr1, mbr2)
	if !mbr.Equals(mbr1) {
		t.Errorf("MergeMbrs() got wrong result")
	}
}
