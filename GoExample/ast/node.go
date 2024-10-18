package main

import "sort"

type Shift struct {
	Start int
	End   int
}

func (o Shift) Contains(other Shift) bool {
	return o.Start <= other.Start && o.End >= other.End
}

func (o Shift) Overlaps(other Shift) bool {
	return o.Start < other.End && o.End > other.Start
}

type Node struct {
	Shft    Shift
	Chldren []*Node
}

func Shft(start, end int) Shift {
	return Shift{start, end}
}

func MergeTrees(node1, node2 *Node) *Node {
	if node1 == nil {
		return node2
	}
	if node2 == nil {
		return node1
	}

	if node1.Shft == node2.Shft {
		return &Node{
			Shft:    node1.Shft,
			Chldren: MergeChildLists(node1.Chldren, node2.Chldren),
		}
	}

	if node1.Shft.Contains(node2.Shft) {
		node1.Chldren = MergeChildIntoChildren(node1.Chldren, node2)
		return node1
	}

	if node2.Shft.Contains(node1.Shft) {
		node2.Chldren = MergeChildIntoChildren(node2.Chldren, node1)
		return node2
	}

	// Обработка перекрывающихся узлов
	if node1.Shft.Overlaps(node2.Shft) {
		mergedNode := &Node{
			Shft: Shift{
				Start: min(node1.Shft.Start, node2.Shft.Start),
				End:   max(node1.Shft.End, node2.Shft.End),
			},
			Chldren: []*Node{node1, node2},
		}
		//mergedNode.Children.Sort((a, b) => a.Offset.Start.CompareTo(b.Offset.Start));
		sort.Slice(mergedNode.Chldren, func(i, j int) bool {
			return mergedNode.Chldren[i].Shft.Start < mergedNode.Chldren[j].Shft.Start
		})
		return mergedNode
	}

	// Узлы не пересекаются — создаем родительский узел
	parentNode := &Node{
		Shft: Shift{
			Start: min(node1.Shft.Start, node2.Shft.Start),
			End:   max(node1.Shft.End, node2.Shft.End),
		},
		Chldren: []*Node{node1, node2},
	}
	//parentNode.Children.Sort((a, b) => a.Offset.Start.CompareTo(b.Offset.Start));
	sort.Slice(parentNode.Chldren, func(i, j int) bool {
		return parentNode.Chldren[i].Shft.Start < parentNode.Chldren[j].Shft.Start
	})
	return parentNode
}

func MergeChildLists(list1, list2 []*Node) []*Node {
	mergedList := make([]*Node, 0, len(list1)+len(list2))
	i, j := 0, 0

	for i < len(list1) && j < len(list2) {
		child1 := list1[i]
		child2 := list2[j]

		if child1.Shft == child2.Shft {
			mergedList = append(mergedList, MergeTrees(child1, child2))
			i++
			j++
		} else if child1.Shft.Start < child2.Shft.Start {
			mergedList = append(mergedList, child1)
			i++
		} else {
			mergedList = append(mergedList, child2)
			j++
		}
	}

	if i < len(list1) {
		mergedList = append(mergedList, list1[i:]...)
	}

	if j < len(list2) {
		mergedList = append(mergedList, list2[j:]...)
	}

	if len(mergedList) == 0 {
		return nil
	}

	return mergedList
}

func MergeChildIntoChildren(children []*Node, childToMerge *Node) []*Node {
	merged := false
	for i, child := range children {
		if child.Shft.Contains(childToMerge.Shft) || child.Shft == childToMerge.Shft {
			children[i] = MergeTrees(child, childToMerge)
			merged = true
			break
		}
	}
	if !merged {
		children = append(children, childToMerge)
	}

	//children.Sort((a, b) => a.Offset.Start.CompareTo(b.Offset.Start));
	sort.Slice(children, func(i, j int) bool {
		return children[i].Shft.Start < children[j].Shft.Start
	})
	return children
}
