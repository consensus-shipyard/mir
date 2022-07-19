// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

// actionQueue holds scheduled actions and implements the heap
// interface such that the earliest action is in the root of the tree,
// at index 0.
type actionQueue []*action

func (aq actionQueue) Len() int           { return len(aq) }
func (aq actionQueue) Less(i, j int) bool { return aq[i].deadline < aq[j].deadline }

func (aq actionQueue) Swap(i, j int) {
	aq[i], aq[j] = aq[j], aq[i]
	aq[i].index = i
	aq[j].index = j
}

func (aq *actionQueue) Push(v any) {
	e := v.(*action)
	e.index = len(*aq)
	*aq = append(*aq, e)
}

func (aq *actionQueue) Pop() any {
	n := len(*aq)
	e := (*aq)[n-1]
	*aq, (*aq)[n-1] = (*aq)[:n-1], nil
	e.index = -1
	return e
}

func (aq actionQueue) top() *action {
	return aq[0]
}
