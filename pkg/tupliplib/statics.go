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

// hasPrefix marks if a line has the given prefix.
func hasPrefix(prefix string) func(input string) bool {
	return func(input string) bool {
		return strings.HasPrefix(input, prefix)
	}
}

// withoutFrom removes the FROM instruction prefix.
func withoutFrom(input string) string {
	return strings.TrimSpace(strings.TrimPrefix(input, DockerFromInstruction))
}

// withoutWildcard removes the wildcard prefixes.
func withoutWildcard(input string) string {
	return strings.TrimSpace(strings.TrimPrefix(input, WildcardDependency+VersionSeparator))
}

// splitFromInstruction splits up a FROM instruction to the repository, the tag, and the alias.
func splitFromInstruction(inst string) (repo string, version string, alias string) {
	withoutFrom := withoutFrom(inst)
	aliasTuple := strings.SplitN(withoutFrom, DockerAs, 2)
	if len(aliasTuple) > 1 {
		alias = strings.TrimSpace(aliasTuple[1])
	}
	repoTuple := strings.SplitN(aliasTuple[0], VersionSeparator, 2)
	repo = strings.TrimSpace(repoTuple[0])
	if len(repoTuple) > 1 {
		version = strings.TrimSpace(repoTuple[1])
	}
	return
}

// transformRootArgument matches the version of the root tag vector from the Dockerfile's VERSION ARG
// and transforms it to a FROM instruction.
// If remove is true, the instruction will be removed instead of transformed.
func transformRootVersion(remove bool) func(inst string) string {
	return func(inst string) string {
		if hasPrefix(VersionInstruction)(inst) {
			if remove {
				return ""
			}
			version := strings.TrimSpace(
				strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(inst, VersionInstruction)), ArgEquation),
			)
			return WildcardInstruction + version
		}
		return inst
	}
}

// toTagVector converts the given FROM instruction to a tag vector.
func toTagVector(inst string) (vector []string) {
	vector = make([]string, 0)
	if inst == ScratchInstruction {
		return
	}
	var firstVector string
	repository, version, alias := splitFromInstruction(inst)
	if repository == DockerScratch && alias != "" {
		firstVector = alias
	} else {
		repoParts := strings.SplitN(repository, RepositorySeparator, 2)
		if len(repoParts) > 1 {
			firstVector = repoParts[1]
		} else {
			firstVector = repoParts[0]
		}
	}
	parts := strings.Split(version, DockerTagSeparator)
	firstVersion := parts[0]
	if firstVersion != "" {
		firstVector = strings.Join([]string{firstVector, parts[0]}, VersionSeparator)
	}
	vector = append(vector, firstVector)
	if len(parts) > 1 {

		for _, p := range parts[1:] {
			if strings.ContainsAny(p, VersionChars) {
				versionIndex := strings.IndexAny(p, VersionChars)
				partSlice := []rune(p)
				partVector := string(partSlice[:versionIndex]) + VersionSeparator + string(partSlice[versionIndex:])
				vector = append(vector, partVector)
			} else {
				vector = append(vector, p)
			}
		}
	}
	return
}

// findRepository checks if the given Dockerfile lines are from a valid Dockerfile and returns the REPOSITORY ARG
// if given.
func findRepository(lines []string) (repository string, err error) {
	var hasVectors bool
	for _, l := range lines {
		if hasPrefix(DockerArgInstruction + Space + VersionArg)(l) {
			hasVectors = true
		}
		if hasPrefix(RepositoryInstruction)(l) {
			repository = strings.TrimSpace(
				strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(l, RepositoryInstruction)), ArgEquation),
			)
		}
		if hasPrefix(DockerFromInstruction)(l) {
			hasVectors = true
		}
	}
	if !hasVectors {
		err = errors.New("the given Dockerfile does not contain any FROM instructions or VERSION ARG")
		return
	}
	return
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
		return nil, fmt.Errorf("the given tag vector '%v' was not found in any remote tags", next.ToSlice())
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
