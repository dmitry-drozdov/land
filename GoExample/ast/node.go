package main

import "sort"

type Offset struct {
	Start int
	End   int
}

func (o Offset) Contains(other Offset) bool {
	return o.Start <= other.Start && o.End >= other.End
}

func (o Offset) Overlaps(other Offset) bool {
	return o.Start < other.End && o.End > other.Start
}

type Node struct {
	Offs    Offset
	Chldren []*Node
}

func MergeTrees(node1, node2 *Node) *Node {
	if node1 == nil {
		return node2
	}
	if node2 == nil {
		return node1
	}

	if node1.Offs == node2.Offs {
		return &Node{
			Offs:    node1.Offs,
			Chldren: MergeChildLists(node1.Chldren, node2.Chldren),
		}
	}

	if node1.Offs.Contains(node2.Offs) {
		node1.Chldren = MergeChildIntoChildren(node1.Chldren, node2)
		return node1
	}

	if node2.Offs.Contains(node1.Offs) {
		node2.Chldren = MergeChildIntoChildren(node2.Chldren, node1)
		return node2
	}

	// Обработка перекрывающихся узлов
	if node1.Offs.Overlaps(node2.Offs) {
		mergedNode := &Node{
			Offs: Offset{
				Start: min(node1.Offs.Start, node2.Offs.Start),
				End:   max(node1.Offs.End, node2.Offs.End),
			},
			Chldren: []*Node{node1, node2},
		}
		//mergedNode.Children.Sort((a, b) => a.Offset.Start.CompareTo(b.Offset.Start));
		sort.Slice(mergedNode.Chldren, func(i, j int) bool {
			return mergedNode.Chldren[i].Offs.Start < mergedNode.Chldren[j].Offs.Start
		})
		return mergedNode
	}

	// Узлы не пересекаются — создаем родительский узел
	parentNode := &Node{
		Offs: Offset{
			Start: min(node1.Offs.Start, node2.Offs.Start),
			End:   max(node1.Offs.End, node2.Offs.End),
		},
		Chldren: []*Node{node1, node2},
	}
	//parentNode.Children.Sort((a, b) => a.Offset.Start.CompareTo(b.Offset.Start));
	sort.Slice(parentNode.Chldren, func(i, j int) bool {
		return parentNode.Chldren[i].Offs.Start < parentNode.Chldren[j].Offs.Start
	})
	return parentNode
}

func MergeChildLists(list1, list2 []*Node) []*Node {
	mergedList := make([]*Node, 0, len(list1)+len(list2))
	i, j := 0, 0

	for i < len(list1) && j < len(list2) {
		child1 := list1[i]
		child2 := list2[j]

		if child1.Offs == child2.Offs {
			mergedList = append(mergedList, MergeTrees(child1, child2))
			i++
			j++
		} else if child1.Offs.Start < child2.Offs.Start {
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
		mergedList = append(mergedList, list2[i:]...)
	}

	if len(mergedList) == 0 {
		return nil
	}

	return mergedList
}

func MergeChildIntoChildren(children []*Node, childToMerge *Node) []*Node {
	merged := false
	for i, child := range children {
		if child.Offs.Contains(childToMerge.Offs) || child.Offs == childToMerge.Offs {
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
		return children[i].Offs.Start < children[j].Offs.Start
	})
	return children
}
