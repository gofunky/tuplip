package tupliplib

import (
	"errors"
	"fmt"
	"github.com/blang/semver"
	"github.com/deckarep/golang-set"
	"github.com/gofunky/automi/emitters"
	"github.com/gofunky/automi/stream"
	"io"
	"sort"
	"strings"
)

type Tuplip struct {
	ExcludeMajor bool
	ExcludeMinor bool
	ExcludeBase  bool
}

const VersionSeparator = ":"
const WildcardDependency = "_"
const VersionDot = "."
const DockerTagSeparator = "-"

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

func (t Tuplip) parseVersions(withBase bool, alias string, isShort bool, version semver.Version) (result mapset.Set,
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
		return t.parseVersions(withBase, dependencyAlias, versionIsShort, dependencyVersion)
	} else {
		return mapset.NewSetWith(inputTag), nil
	}
}

func (t Tuplip) nonEmpty(input string) bool {
	return input != ""
}

func (t Tuplip) splitBySeparator(input string) (result []string) {
	return strings.Split(input, " ")
}

func (t Tuplip) packInSet(subSet mapset.Set) (result mapset.Set) {
	return mapset.NewSetWith(subSet)
}

func (t Tuplip) mergeSets(left mapset.Set, right mapset.Set) (result mapset.Set) {
	return left.Union(right)
}

func (t Tuplip) power(inputSet mapset.Set) mapset.Set {
	return inputSet.PowerSet()
}

func (t Tuplip) failOnEmpty(inputSet mapset.Set) (mapset.Set, error) {
	if inputSet.Cardinality() <= 1 {
		return nil, errors.New("no input tags could be detected")
	}
	return inputSet, nil
}

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

func (t Tuplip) FromScanner(src io.Reader) *stream.Stream {
	iStream := stream.New(emitters.Scanner(src, nil))
	iStream.FlatMap(t.splitBySeparator)
	iStream.Filter(t.nonEmpty)
	iStream.Map(t.splitVersion)
	iStream.Map(t.packInSet)
	iStream.Reduce(mapset.NewSet(), t.mergeSets)
	iStream.Map(t.power)
	iStream.FlatMap(t.failOnEmpty)
	iStream.FlatMap(t.join)
	iStream.Filter(t.nonEmpty)
	return iStream
}
