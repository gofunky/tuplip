package tupliplib

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"math"
	"os/exec"
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

// isFromInstruction marks if a line is a Docker FROM instruction.
func fromInstruction(input string) bool {
	return strings.HasPrefix(input, DockerFromInstruction)
}

// withoutFrom removes the FROM instruction prefix.
func withoutFrom(input string) string {
	return strings.TrimSpace(strings.TrimPrefix(input, DockerFromInstruction))
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

// toRootInstruction matches the repository name and patches it to a root wildcard prefix instruction.
// Returns an empty instruction if the instruction was unversioned.
func toRootInstruction(inst string) (rootInst string, repository string) {
	repository, version, _ := splitFromInstruction(inst)
	if version != "" {
		var buffer bytes.Buffer
		buffer.WriteString(DockerFromInstruction)
		buffer.WriteString(" ")
		buffer.WriteString(WildcardDependency)
		buffer.WriteString(VersionSeparator)
		buffer.WriteString(version)
		rootInst = buffer.String()
	}
	return
}

// toTagVector converts the given FROM instruction to a tag vector.
func toTagVector(inst string) (vector string) {
	repository, version, alias := splitFromInstruction(inst)
	if alias == "" {
		repoParts := strings.SplitN(repository, RepositorySeparator, 2)
		if len(repoParts) > 1 {
			vector = repoParts[1]
		} else {
			vector = repoParts[0]
		}
	} else {
		vector = alias
	}
	if version != "" {
		vector = strings.Join([]string{vector, version}, VersionSeparator)
	}
	return
}

// markRootInstruction matches the root vector repository in the given Dockerfile, converts it to a root instruction
// (i.e., instruction without alias and underscore as repository), and returns the updated Dockerfile content.
// If the designated instruction is unversioned, it is removed from the result set.
func markRootInstruction(lines []string) (marked []string, repository string, err error) {
	var lastFromInst = -1
	for i, l := range lines {
		if fromInstruction(l) {
			lastFromInst = i
		}
	}
	if lastFromInst < 0 {
		err = errors.New("the given Dockerfile does not contain any FROM instructions")
		return
	}
	var rootInst string
	rootInst, repository = toRootInstruction(lines[lastFromInst])
	marked = lines
	marked[lastFromInst] = rootInst
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

// requireDocker ensures that docker is available in the PATH.
func requireDocker() error {
	cmd := exec.Command("docker")
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}
