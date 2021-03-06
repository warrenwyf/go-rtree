package rtree

import (
	"fmt"
	"strings"
)

const (
	MbrTypeInt32 = iota
	MbrTypeFloat64
)

func NewMbrInt32(mins []int32, spans []int32) *MbrInt32 {
	minsLen, spansLen := len(mins), len(spans)

	if minsLen == 0 || spansLen == 0 {
		return nil
	}

	dim := minsLen
	if spansLen < dim {
		dim = spansLen
	}

	mbr := make(MbrInt32, dim*2)
	for i := 0; i < dim; i++ {
		mbr[i*2] = mins[i]
		mbr[i*2+1] = spans[i]
	}

	return &mbr
}

func NewMbrFloat64(mins []float64, spans []float64) *MbrFloat64 {
	minsLen, spansLen := len(mins), len(spans)

	if minsLen == 0 || spansLen == 0 {
		return nil
	}

	if minsLen == spansLen {
		return &MbrFloat64{
			mins:  mins,
			spans: spans,
		}
	} else if minsLen > spansLen {
		return &MbrFloat64{
			mins:  mins[:spansLen],
			spans: spans,
		}
	} else if minsLen < spansLen {
		return &MbrFloat64{
			mins:  mins,
			spans: spans[:minsLen],
		}
	}

	return nil
}

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

type MbrInt32 []int32

func (m *MbrInt32) Type() int {
	return MbrTypeInt32
}

func (m *MbrInt32) Dim() int {
	return len(*m) / 2
}

func (a *MbrInt32) Equals(mbr Mbr) bool {
	b, ok := mbr.(*MbrInt32)
	if !ok {
		return false
	}

	if len(*a) != len(*b) {
		return false
	}

	for i := 0; i < len(*a); i += 2 {
		aMin, aSpan, bMin, bSpan := (*a)[i], (*a)[i+1], (*b)[i], (*b)[i+1]

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

	if len(*a) != len(*b) {
		return false
	}

	for i := 0; i < len(*a); i += 2 {
		aMin, aSpan, bMin, bSpan := (*a)[i], (*a)[i+1], (*b)[i], (*b)[i+1]

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

	if len(*a) != len(*b) {
		return false
	}

	notIntersects := false

	for i := 0; i < len(*a); i += 2 {
		aMin, aSpan, bMin, bSpan := (*a)[i], (*a)[i+1], (*b)[i], (*b)[i+1]

		if aMin > (bMin+bSpan) || (aMin+aSpan) < bMin {
			notIntersects = true
			break
		}
	}

	return !notIntersects
}

func (mbr *MbrInt32) Clone() Mbr {
	newMbr := make(MbrInt32, len(*mbr))
	for i, v := range *mbr {
		newMbr[i] = v
	}

	return &newMbr
}

func (mbr *MbrInt32) String() string {
	dim := len(*mbr) / 2

	minStrs := make([]string, dim)
	spanStrs := make([]string, dim)

	for i := 0; i < dim; i++ {
		minStrs[i] = fmt.Sprintf("%d", (*mbr)[i*2])
		spanStrs[i] = fmt.Sprintf("%d", (*mbr)[i*2+1])
	}

	minStr := strings.Join(minStrs, ",")
	spanStr := strings.Join(spanStrs, ",")

	return fmt.Sprintf("[(%s),(%s)]", minStr, spanStr)
}

func (mbr *MbrInt32) size() float64 {
	var size float64 = 1

	for i := 1; i < len(*mbr); i += 2 {
		size *= float64((*mbr)[i])
	}

	return size
}

type MbrFloat64 struct {
	mins  []float64
	spans []float64
}

func (m *MbrFloat64) Type() int {
	return MbrTypeFloat64
}

func (m *MbrFloat64) Dim() int {
	return len(m.mins)
}

func (a *MbrFloat64) Equals(mbr Mbr) bool {
	b, ok := mbr.(*MbrFloat64)
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

func (a *MbrFloat64) Contains(mbr Mbr) bool {
	b, ok := mbr.(*MbrFloat64)
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

func (a *MbrFloat64) Intersects(mbr Mbr) bool {
	b, ok := mbr.(*MbrFloat64)
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

func (mbr *MbrFloat64) Clone() Mbr {
	dim := mbr.Dim()

	mins := make([]float64, dim)
	spans := make([]float64, dim)

	for i, min := range mbr.mins {
		mins[i] = min
		spans[i] = mbr.spans[i]
	}

	return NewMbrFloat64(mins, spans)
}

func (mbr *MbrFloat64) String() string {
	minStrs := make([]string, len(mbr.mins))
	spanStrs := make([]string, len(mbr.spans))

	for i, min := range mbr.mins {
		minStrs[i] = fmt.Sprintf("%f", min)
		spanStrs[i] = fmt.Sprintf("%f", mbr.spans[i])
	}

	minStr := strings.Join(minStrs, ",")
	spanStr := strings.Join(spanStrs, ",")

	return fmt.Sprintf("[(%s),(%s)]", minStr, spanStr)
}

func (m *MbrFloat64) size() float64 {
	var size float64 = 1

	for _, span := range m.spans {
		size *= span
	}

	return size
}

func MergeMbrs(mbrs ...Mbr) Mbr {
	mbrLen := len(mbrs)
	if mbrLen == 0 {
		return nil
	}

	mbrType := mbrs[0].Type()
	switch mbrType {
	case MbrTypeInt32:
		return mergeInt32Mbrs(mbrs...)
	case MbrTypeFloat64:
		return mergeFloat64Mbrs(mbrs...)
	}

	return nil
}

func mergeInt32Mbrs(mbrs ...Mbr) Mbr {
	mbrLen := len(mbrs)

	mbr := mbrs[0].Clone().(*MbrInt32)
	for i := 1; i < mbrLen; i++ {

		m, ok := mbrs[i].(*MbrInt32)
		if !ok {
			continue
		}

		for j := 0; j < len(*mbr); j += 2 {
			currentMin := (*mbr)[j]
			currentMax := currentMin + (*mbr)[j+1]
			min := (*m)[j]
			max := min + (*m)[j+1]

			if min < currentMin {
				currentMin = min
			}
			if max > currentMax {
				currentMax = max
			}

			(*mbr)[j] = currentMin
			(*mbr)[j+1] = currentMax - currentMin
		}

	}

	return mbr
}

func mergeFloat64Mbrs(mbrs ...Mbr) Mbr {
	mbrLen := len(mbrs)

	mbr := mbrs[0].Clone().(*MbrFloat64)
	dim := mbrs[0].Dim()
	for i := 1; i < mbrLen; i++ {

		m, ok := mbrs[i].(*MbrFloat64)
		if !ok {
			continue
		}

		for j := 0; j < dim; j++ {
			currentMin := mbr.mins[j]
			currentMax := currentMin + mbr.spans[j]
			min := m.mins[j]
			max := min + m.spans[j]

			if min < currentMin {
				currentMin = min
			}
			if max > currentMax {
				currentMax = max
			}

			mbr.mins[j] = currentMin
			mbr.spans[j] = currentMax - currentMin
		}

	}

	return mbr
}
