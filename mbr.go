package rtree

import (
	"fmt"
	"strings"
)

const (
	MbrTypeInt32 = iota
	MbrTypeFloat64
)

type Mbr interface {
	Type() int
	Dim() int
	Equals(m Mbr) bool
	Contains(m Mbr) bool
	Intersects(m Mbr) bool
	Clone() Mbr
	String() string
	size() float64
}

type MbrInt32 struct {
	mins  []int32
	spans []int32
}

func (m *MbrInt32) Type() int {
	return MbrTypeInt32
}

func (m *MbrInt32) Dim() int {
	return len(m.mins)
}

func (a *MbrInt32) Equals(mbr Mbr) bool {
	b, ok := mbr.(*MbrInt32)
	if !ok {
		return false
	}

	if len(a.mins) != len(b.mins) {
		return false
	}

	for i, aMin := range a.mins {
		aSpan, bMin, bSpan := a.spans[i], b.mins[i], b.spans[i]

		if aMin != bMin || aSpan != bSpan {
			return false
		}
	}

	return true
}

func (a *MbrInt32) Contains(mbr Mbr) bool {
	b, ok := mbr.(*MbrInt32)
	if !ok {
		return false
	}

	if len(a.mins) != len(b.mins) {
		return false
	}

	for i, aMin := range a.mins {
		aSpan, bMin, bSpan := a.spans[i], b.mins[i], b.spans[i]

		if aMin > bMin || (aMin+aSpan) < (bMin+bSpan) {
			return false
		}
	}

	return true
}

func (a *MbrInt32) Intersects(mbr Mbr) bool {
	b, ok := mbr.(*MbrInt32)
	if !ok {
		return false
	}

	if len(a.mins) != len(b.mins) {
		return false
	}

	notIntersects := false

	for i, aMin := range a.mins {
		aSpan, bMin, bSpan := a.spans[i], b.mins[i], b.spans[i]

		if aMin > (bMin+bSpan) || (aMin+aSpan) < bMin {
			notIntersects = true
			break
		}
	}

	return !notIntersects
}

func (mbr *MbrInt32) Clone() Mbr {
	dim := mbr.Dim()

	mins := make([]int32, dim)
	spans := make([]int32, dim)

	for i, min := range mbr.mins {
		mins[i] = min
		spans[i] = mbr.spans[i]
	}

	return NewMbrInt32(mins, spans)
}

func (mbr *MbrInt32) String() string {
	minStrs := make([]string, len(mbr.mins))
	spanStrs := make([]string, len(mbr.spans))

	for i, min := range mbr.mins {
		minStrs[i] = fmt.Sprintf("%d", min)
		spanStrs[i] = fmt.Sprintf("%d", mbr.spans[i])
	}

	minStr := strings.Join(minStrs, ",")
	spanStr := strings.Join(spanStrs, ",")

	return fmt.Sprintf("[(%s),(%s)]", minStr, spanStr)
}

func (m *MbrInt32) size() float64 {
	var size float64 = 1

	for _, span := range m.spans {
		size *= float64(span)
	}

	return size
}
