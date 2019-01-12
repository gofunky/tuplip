package tupliplib

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/gofunky/semver"
	"sort"
	"strings"
)

// buildTag parses a semantic version with the given version digits. Optionally, prefix an alias tag.
func (t Tuplip) buildTag(withBase bool, alias string, versionDigits ...uint64) (string, error) {
	var builder strings.Builder
	if withBase {
		_, err := builder.WriteString(alias)
		if err != nil {
			return "", err
		}
	}
	for n, digit := range versionDigits {
		if n > 0 {
			_, err := builder.WriteString(".")
			if err != nil {
				return "", err
			}
		}
		_, err := builder.WriteString(fmt.Sprint(digit))
		if err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}

// buildVersionSet parses all possible shortened version representations from a semantic version object.
func (t Tuplip) buildVersionSet(withBase bool, alias string, versionArity int, version semver.Version) (result mapset.Set,
	err error) {

	result = mapset.NewSet()
	if withBase && !t.ExcludeBase {
		result.Add(alias)
	}
	if !t.ExcludeMajor {
		if newTag, err := t.buildTag(withBase, alias, version.Major); err != nil {
			return nil, err
		} else {
			result.Add(newTag)
		}
	}
	if versionArity >= 2 {
		if !t.ExcludeMinor {
			if newTag, err := t.buildTag(withBase, alias, version.Major, version.Minor); err != nil {
				return nil, err
			} else {
				result.Add(newTag)
			}
		}
	}
	if versionArity >= 3 {
		if newTag, err := t.buildTag(withBase, alias, version.Major, version.Minor, version.Patch); err != nil {
			return nil, err
		} else {
			result.Add(newTag)
		}
	}
	return result, nil
}

// splitVersion takes a parsed semantic version string, builds a semantic version object and generates all possible
// shortened version strings from it.
// requireSemver enables semantic version checks. Short versions are not allowed then.
func (t Tuplip) splitVersion(requireSemver bool) func(inputTag string) (result mapset.Set, err error) {
	return func(inputTag string) (result mapset.Set, err error) {
		if strings.Contains(inputTag, VersionSeparator) {
			dependency := strings.SplitN(inputTag, VersionSeparator, 2)
			dependencyAlias := dependency[0]
			var dependencyVersionText = dependency[1]
			versionArity := strings.Count(dependencyVersionText, VersionDot) + 1
			var dependencyVersion semver.Version
			if requireSemver {
				dependencyVersion, err = semver.Parse(dependencyVersionText)
			} else {
				dependencyVersion, err = semver.ParseTolerant(dependencyVersionText)
			}
			if err != nil {
				return
			}
			withBase := dependencyAlias != WildcardDependency
			return t.buildVersionSet(withBase, dependencyAlias, versionArity, dependencyVersion)
		} else {
			return mapset.NewSetWith(inputTag), nil
		}
	}
}

// splitBySeparator generates a function separates the input string by the given character and trims superfluous spaces.
func (t Tuplip) splitBySeparator(sep string) func(input string) []string {
	if sep == "" {
		sep = VectorSeparator
	}
	return func(input string) (result []string) {
		result = strings.Split(input, sep)
		for i, el := range result {
			result[i] = strings.TrimSpace(el)
		}
		return
	}
}

// join joins all subtags (i.e., elements of the given set) to all possible representations by building a cartesian
// product of them. The subtags are separated by the given Docker separator. The subtags are ordered alphabetically
// to ensure that a root tag vector (i.e., a tag without an alias) is mentioned before alias tags.
func (t Tuplip) join(inputSet mapset.Set) (result mapset.Set) {
	result = mapset.NewSet()
	inputSlice := inputSet.ToSlice()
	subTagSlice := make(SortedSet, len(inputSlice))
	for i, subTag := range inputSlice {
		subTagSlice[i] = subTag.(mapset.Set)
	}
	sort.Sort(subTagSlice)
	for _, subTag := range subTagSlice {
		subTagSet := subTag.(mapset.Set)
		if result.Cardinality() == 0 {
			result = subTagSet
		} else {
			productSet := subTagSet.CartesianProduct(result)
			result = mapset.NewSet()
			for item := range productSet.Iter() {
				pair := item.(mapset.OrderedPair)
				concatPair := fmt.Sprintf("%s%s%s", pair.First, DockerTagSeparator, pair.Second)
				result.Add(concatPair)
			}
		}
	}
	return result
}

// addLatestTag adds an additional latest tag if requested in *Tuplip.
func (t Tuplip) addLatestTag(inputSet mapset.Set) mapset.Set {
	if t.AddLatest {
		inputSet.Add(mapset.NewSet(mapset.NewSet("latest")))
	}
	return inputSet
}
