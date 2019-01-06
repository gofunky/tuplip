package tupliplib

import (
	"errors"
	"fmt"
	"github.com/blang/semver"
	"github.com/deckarep/golang-set"
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
func (t Tuplip) buildVersionSet(withBase bool, alias string, isShort bool, version semver.Version) (result mapset.Set,
	err error) {

	result = mapset.NewSet()
	if withBase && !t.ExcludeBase {
		result.Add(alias)
	}
	if isShort {
		if !t.ExcludeMajor {
			newTag, err := t.buildTag(withBase, alias, version.Minor)
			if err != nil {
				return nil, err
			}
			result.Add(newTag)
		}
		newTag, err := t.buildTag(withBase, alias, version.Minor, version.Patch)
		if err != nil {
			return nil, err
		}
		result.Add(newTag)
	} else {
		if !t.ExcludeMajor {
			newTag, err := t.buildTag(withBase, alias, version.Major)
			if err != nil {
				return nil, err
			}
			result.Add(newTag)
		}
		if !t.ExcludeMinor {
			newTag, err := t.buildTag(withBase, alias, version.Major, version.Minor)
			if err != nil {
				return nil, err
			}
			result.Add(newTag)
		}
		newTag, err := t.buildTag(withBase, alias, version.Major, version.Minor, version.Patch)
		if err != nil {
			return nil, err
		}
		result.Add(newTag)
	}
	return result, nil
}

// splitVersion takes a parsed semantic version string, builds a semantic version object and generates all possible
// shortened version strings from it.
func (t Tuplip) splitVersion(inputTag string) (result mapset.Set, err error) {
	if strings.Contains(inputTag, VersionSeparator) {
		dependency := strings.SplitN(inputTag, VersionSeparator, 2)
		dependencyAlias := dependency[0]
		var dependencyVersionText = dependency[1]
		versionIsShort := strings.Count(dependencyVersionText, VersionDot) == 1
		if versionIsShort {
			dependencyVersionText = "0." + dependencyVersionText
		}
		dependencyVersion, err := semver.Make(dependencyVersionText)
		if err != nil {
			return nil, err
		}
		withBase := dependencyAlias != WildcardDependency
		return t.buildVersionSet(withBase, dependencyAlias, versionIsShort, dependencyVersion)
	} else {
		return mapset.NewSetWith(inputTag), nil
	}
}

// nonEmpty marks if a string is not empty.
func (t Tuplip) nonEmpty(input string) bool {
	return input != ""
}

// splitBySeparator separates the input string by the chosen character and trims superfluous spaces.
func (t Tuplip) splitBySeparator(input string) (result []string) {
	result = strings.Split(input, t.Separator)
	for i, el := range result {
		result[i] = strings.TrimSpace(el)
	}
	return
}

// packInSet packs a set as subset into a new set.
func (t Tuplip) packInSet(subSet mapset.Set) (result mapset.Set) {
	return mapset.NewSetWith(subSet)
}

// mergeSets merges the second given set into the first one.
func (t Tuplip) mergeSets(left mapset.Set, right mapset.Set) (result mapset.Set) {
	return left.Union(right)
}

// power build a power of the given set.
func (t Tuplip) power(inputSet mapset.Set) mapset.Set {
	return inputSet.PowerSet()
}

// failOnEmpty returns an error if the given power set is empty (i.e, has cardinality < 2).
func (t Tuplip) failOnEmpty(inputSet mapset.Set) (mapset.Set, error) {
	if inputSet.Cardinality() <= 1 {
		return nil, errors.New("no input tags could be detected")
	}
	return inputSet, nil
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
