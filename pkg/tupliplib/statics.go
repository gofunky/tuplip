package tupliplib

import (
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"math"
	"sort"
	"strings"
)

// packInSet packs a set as subset into a new set.
func packInSet(subSet mapset.Set) (result mapset.Set) {
	return mapset.NewSetWith(subSet)
}

// mergeSets merges the second given set into the first one.
func mergeSets(left mapset.Set, right mapset.Set) (result mapset.Set) {
	return left.Union(right)
}

// power build a power of the given set.
func power(inputSet mapset.Set) mapset.Set {
	return inputSet.PowerSet()
}

// failOnEmpty returns an error if the given power set is empty (i.e, has cardinality < 2).
func failOnEmpty(inputSet mapset.Set) (mapset.Set, error) {
	if inputSet.Cardinality() <= 1 {
		return nil, errors.New("no input tags could be detected")
	}
	return inputSet, nil
}

// nonEmpty marks if a string is not empty.
func nonEmpty(input string) bool {
	return input != ""
}

// removeCommon finds common denominators in the seed and removes them.
// Empty sets are removed on the second run to be able to find equal sets in the subsequent run.
// Fails if none of next's elements is found in any of seed's sets.
func removeCommon(seed map[string]mapset.Set, next mapset.Set) (result map[string]mapset.Set, err error) {
	var found = false
	for k, v := range seed {
		if v.Cardinality() == 0 {
			delete(seed, k)
		} else {
			if v.Intersect(next).Cardinality() > 0 {
				found = true
			}
			seed[k] = v.Difference(next)
		}
	}
	if !found {
		return nil, fmt.Errorf("the given tag vector '%v' was not found in any remote tags", next)
	}
	return seed, nil
}

// keyForSmallest finds the smallest set in the map and returns its key.
func keyForSmallest(seed map[string]mapset.Set) (result string) {
	smallestSets := make([]string, 0)
	minVal := minVal(seed)
	for k, v := range seed {
		if v.Cardinality() == minVal {
			smallestSets = append(smallestSets, k)
		}
	}
	return mostSeparators(smallestSets, DockerTagSeparator)
}

// minVal finds the smallest set in the given map.
func minVal(numbers map[string]mapset.Set) (minNumber int) {
	minNumber = math.MaxInt8
	for _, v := range numbers {
		if c := v.Cardinality(); c < minNumber {
			minNumber = c
		}
	}
	return minNumber
}

// mostSeparators finds the element in the given slice that contains the most separators.
func mostSeparators(values []string, sep string) (result string) {
	sort.Strings(values)
	maxNumber := math.MinInt8
	for _, v := range values {
		if c := strings.Count(v, sep); c > maxNumber {
			maxNumber = c
			result = v
		}
	}
	return result
}
