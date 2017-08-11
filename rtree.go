package rtree

import (
	"math"
	"sort"
)

type Rtree struct {
	dim     int
	fan     int
	halfFan int

	root   *node
	size   int32
	height int8
}

func NewRtree(dim int, fan int, features ...Feature) *Rtree {
	t := &Rtree{
		dim:     dim,
		fan:     fan,
		halfFan: (fan / 2),

		root: &node{
			objs:  []*object{},
			leaf:  true,
			level: 1,
		},

		size:   0,
		height: 1,
	}

	if len(features) <= fan {
		for _, feature := range features {
			t.Insert(feature)
		}
	} else {
		t.bulkLoad(features)
	}

	return t
}

func (t *Rtree) Dim() int {
	return t.dim
}

func (t *Rtree) Size() int32 {
	return t.size
}

func (t *Rtree) Height() int8 {
	return t.height
}

func (t *Rtree) Insert(feature Feature) {
	obj := &object{
		mbr:     feature.Mbr(),
		feature: feature,
	}

	t.insertObj(obj, 1)

	t.size++
}

func (t *Rtree) insertObj(e *object, level int8) {
	leaf := t.chooseNode(t.root, e, level)
	leaf.objs = append(leaf.objs, e)

	if e.node != nil {
		e.node.parent = leaf
	}

	var split *node
	if len(leaf.objs) > t.fan {
		leaf, split = leaf.split(t.halfFan)
	}

	root, splitRoot := t.adjustTree(leaf, split)
	if splitRoot != nil {
		oldRoot := root
		t.height++
		t.root = &node{
			parent: nil,
			level:  t.height,
			objs: []*object{
				&object{
					mbr:  oldRoot.computeMbr(),
					node: oldRoot,
				},
				&object{
					mbr:  splitRoot.computeMbr(),
					node: splitRoot,
				},
			},
		}

		oldRoot.parent = t.root
		splitRoot.parent = t.root
	}
}

func (t *Rtree) bulkLoad(features []Feature) {
	n := len(features)

	objs := make([]*object, n)
	for i, feature := range features {
		objs[i] = &object{
			mbr:     feature.Mbr(),
			feature: feature,
		}
	}

	t.root.leaf = false
	t.size = int32(n)
	t.height = int8(math.Ceil(math.Log(float64(n)) / float64(math.Log(float64(t.fan)))))
	t.root.level = t.height
	nsub := int(math.Pow(float64(t.fan), float64(t.height-1)))
	s := int(math.Floor(math.Sqrt(math.Ceil(float64(n) / float64(nsub)))))

	sortByDim(0, objs)

	t.root.objs = make([]*object, s)

	for i, part := range splitInS(s, objs) {
		node := t.omt(t.root.level-1, part, t.fan)
		node.parent = t.root

		t.root.objs[i] = &object{
			mbr:  node.computeMbr(),
			node: node,
		}
	}
}

func (t *Rtree) omt(level int8, objs []*object, m int) *node {
	if len(objs) <= m {
		return &node{
			leaf:  true,
			objs:  objs,
			level: level,
		}
	}

	sortByDim(int(t.height-level)%t.dim, objs)

	n := &node{
		level: level,
		objs:  make([]*object, 0, m),
	}

	for _, part := range splitByM(m, objs) {
		node := t.omt(level-1, part, m)
		node.parent = n

		n.objs = append(n.objs, &object{
			mbr:  node.computeMbr(),
			node: node,
		})
	}

	return n
}

func (t *Rtree) chooseNode(n *node, obj *object, level int8) *node {
	if n.leaf || n.level == level {
		return n
	}

	diff := math.MaxFloat64
	var chosen *object
	for _, en := range n.objs {
		mbr := MergeMbrs(en.mbr, obj.mbr)
		d := mbr.size() - en.mbr.size()
		if d < diff || (d == diff && en.mbr.size() < chosen.mbr.size()) {
			diff = d
			chosen = en
		}
	}

	return t.chooseNode(chosen.node, obj, level)
}

func (t *Rtree) adjustTree(n, nn *node) (*node, *node) {
	if n == t.root {
		return n, nn
	}

	en := n.getObject()
	en.mbr = n.computeMbr()

	if nn == nil {
		return t.adjustTree(n.parent, nil)
	}

	enn := &object{
		mbr:     nn.computeMbr(),
		node:    nn,
		feature: nil,
	}
	n.parent.objs = append(n.parent.objs, enn)

	if len(n.parent.objs) > t.fan {
		return t.adjustTree(n.parent.split(t.halfFan))
	}

	return t.adjustTree(n.parent, nil)
}

func (t *Rtree) Search(mbr Mbr) []Feature {
	return t.searchIntersect([]Feature{}, t.root, mbr)
}

func (t *Rtree) searchIntersect(results []Feature, n *node, mbr Mbr) []Feature {
	for _, e := range n.objs {
		if !mbr.Intersects(e.mbr) {
			continue
		}

		if !n.leaf {
			results = t.searchIntersect(results, e.node, mbr)
			continue
		}

		results = append(results, e.feature)
	}

	return results
}

func (t *Rtree) Remove(feature Feature) bool {
	n := t.findLeaf(t.root, feature)
	if n == nil {
		return false
	}

	ind := -1
	for i, e := range n.objs {
		if e.feature.Equals(feature) {
			ind = i
		}
	}
	if ind < 0 {
		return false
	}

	n.objs = append(n.objs[:ind], n.objs[ind+1:]...)

	t.condenseTree(n)

	t.size--

	if !t.root.leaf && len(t.root.objs) == 1 {
		t.root = t.root.objs[0].node
	}

	t.height = t.root.level

	return true
}

func (t *Rtree) findLeaf(n *node, feature Feature) *node {
	if n.leaf {
		return n
	}

	for _, e := range n.objs {
		if e.mbr.Contains(feature.Mbr()) {
			leaf := t.findLeaf(e.node, feature)
			if leaf == nil {
				continue
			}
			for _, leafEntry := range leaf.objs {
				if leafEntry.feature.Equals(feature) {
					return leaf
				}
			}
		}
	}

	return nil
}

func (t *Rtree) condenseTree(n *node) {
	deleted := []*node{}

	for n != t.root {
		if len(n.objs) < t.halfFan {
			objs := []*object{}
			for _, obj := range n.parent.objs {
				if obj.node != n {
					objs = append(objs, obj)
				}
			}

			n.parent.objs = objs

			if len(n.objs) > 0 {
				deleted = append(deleted, n)
			}
		} else {
			n.getObject().mbr = n.computeMbr()
		}

		n = n.parent
	}

	for _, node := range deleted {
		obj := &object{
			mbr:     node.computeMbr(),
			node:    node,
			feature: nil,
		}

		t.insertObj(obj, node.level+1)
	}
}

type node struct {
	parent *node
	leaf   bool
	objs   []*object
	level  int8
}

func (n *node) getObject() *object {
	var e *object

	for i := range n.parent.objs {
		if n.parent.objs[i].node == n {
			e = n.parent.objs[i]
			break
		}
	}

	return e
}

func (n *node) computeMbr() Mbr {
	mbrs := make([]Mbr, len(n.objs))
	for i, e := range n.objs {
		mbrs[i] = e.mbr
	}

	return MergeMbrs(mbrs...)
}

func (n *node) split(minGroupSize int) (left, right *node) {
	l, r := n.pickSeeds()
	leftSeed, rightSeed := n.objs[l], n.objs[r]

	remaining := append(n.objs[:l], n.objs[l+1:r]...)
	remaining = append(remaining, n.objs[r+1:]...)

	left = n
	left.objs = []*object{leftSeed}
	right = &node{
		parent: n.parent,
		leaf:   n.leaf,
		level:  n.level,
		objs:   []*object{rightSeed},
	}

	// TODO
	if rightSeed.node != nil {
		rightSeed.node.parent = right
	}
	if leftSeed.node != nil {
		leftSeed.node.parent = left
	}

	for len(remaining) > 0 {
		next := pickNext(left, right, remaining)
		e := remaining[next]

		if len(remaining)+len(left.objs) <= minGroupSize {
			assign(e, left)
		} else if len(remaining)+len(right.objs) <= minGroupSize {
			assign(e, right)
		} else {
			assignGroup(e, left, right)
		}

		remaining = append(remaining[:next], remaining[next+1:]...)
	}

	return
}

func (n *node) pickSeeds() (int, int) {
	left, right := 0, 1
	maxWastedSpace := -1.0
	for i, e1 := range n.objs {
		for j, e2 := range n.objs[i+1:] {
			d := MergeMbrs(e1.mbr, e2.mbr).size() - e1.mbr.size() - e2.mbr.size()
			if d > maxWastedSpace {
				maxWastedSpace = d
				left, right = i, j+i+1
			}
		}
	}
	return left, right
}

func pickNext(left, right *node, objs []*object) (next int) {
	maxDiff := -1.0
	leftMbr := left.computeMbr()
	rightMbr := right.computeMbr()
	for i, obj := range objs {
		d1 := MergeMbrs(leftMbr, obj.mbr).size() - leftMbr.size()
		d2 := MergeMbrs(rightMbr, obj.mbr).size() - rightMbr.size()
		d := math.Abs(d1 - d2)
		if d > maxDiff {
			maxDiff = d
			next = i
		}
	}
	return
}

type object struct {
	mbr     Mbr
	node    *node
	feature Feature
}

type dimSorter struct {
	dim  int
	objs []*object
}

func (s *dimSorter) Len() int {
	return len(s.objs)
}

func (s *dimSorter) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}

func (s *dimSorter) Less(i, j int) bool {
	m1 := s.objs[i].mbr
	m2 := s.objs[j].mbr

	switch m1.Type() {
	case MbrTypeInt32:
		a, aok := m1.(*MbrInt32)
		b, bok := m2.(*MbrInt32)
		if aok && bok {
			return a.mins[s.dim] < b.mins[s.dim]
		}
	case MbrTypeFloat64:
		a, aok := m1.(*MbrFloat64)
		b, bok := m2.(*MbrFloat64)
		if aok && bok {
			return a.mins[s.dim] < b.mins[s.dim]
		}
	}

	return false
}

// splitByM splits objects into slices of maximum m objects.
// Split 10 in to 3 will yield 3 + 3 + 3 + 1
func splitByM(m int, objs []*object) [][]*object {
	perSlice := len(objs) / m

	numSlices := m
	if len(objs)%m != 0 {
		numSlices++
	}

	split := make([][]*object, numSlices)
	for i := 0; i < numSlices; i++ {
		if i == numSlices-1 {
			split[i] = objs[i*perSlice:]
			break
		}

		split[i] = objs[i*perSlice : i*perSlice+perSlice]
	}

	return split
}

func splitInS(s int, objs []*object) [][]*object {
	split := splitByM(s, objs)
	if len(split) < 2 {
		return split
	}

	last := split[len(split)-1]
	secondLast := split[len(split)-2]

	if len(last) < len(secondLast) {
		merged := append(secondLast, last...)
		split = split[:len(split)-1]
		split[len(split)-1] = merged
	}

	return split
}

func sortByDim(dim int, objs []*object) {
	sort.Sort(&dimSorter{dim, objs})
}

func assign(obj *object, group *node) {
	if obj.node != nil {
		obj.node.parent = group
	}

	group.objs = append(group.objs, obj)
}

func assignGroup(obj *object, left, right *node) {
	leftMbr := left.computeMbr()
	rightMbr := right.computeMbr()
	leftEnlarged := MergeMbrs(leftMbr, obj.mbr)
	rightEnlarged := MergeMbrs(rightMbr, obj.mbr)

	leftDiff := leftEnlarged.size() - leftMbr.size()
	rightDiff := rightEnlarged.size() - rightMbr.size()
	if diff := leftDiff - rightDiff; diff < 0 {
		assign(obj, left)
		return
	} else if diff > 0 {
		assign(obj, right)
		return
	}

	if diff := leftMbr.size() - rightMbr.size(); diff < 0 {
		assign(obj, left)
		return
	} else if diff > 0 {
		assign(obj, right)
		return
	}

	if diff := len(left.objs) - len(right.objs); diff <= 0 {
		assign(obj, left)
		return
	}
	assign(obj, right)
}
