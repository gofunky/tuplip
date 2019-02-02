package tupliplib

import "github.com/gofunky/pyraset/v2"

// SortedSet is a a slice of map sets to make them sortable.
// It will be sorted alphabetically by the string representations of the subsets.
type SortedSet []mapset.Set

// Len gives the size of the slice.
func (a SortedSet) Len() int { return len(a) }

// Less determines a order for the slice from the string representations of the subsets.
func (a SortedSet) Less(i, j int) bool { return a[i].String() > a[j].String() }

// Swap swaps the two slice elements at the given positions.
func (a SortedSet) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
