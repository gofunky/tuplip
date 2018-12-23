package tupliplib

import (
	"github.com/deckarep/golang-set"
)

type SortedSet []mapset.Set

func (a SortedSet) Len() int           { return len(a) }
func (a SortedSet) Less(i, j int) bool { return a[i].String() > a[j].String() }
func (a SortedSet) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
