package tupliplib

import (
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/gofunky/semver"
	"github.com/nokia/docker-registry-client/registry"
	"os/exec"
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

// withFilter excludes all tags without the given set of tag vectors from the output set.
func (t Tuplip) withFilter(inputSet mapset.Set) bool {
	for _, filterVector := range t.Filter {
		var contains bool
		inputSet.Each(func(i interface{}) bool {
			if i.(mapset.Set).Contains(filterVector) {
				contains = true
			}
			return contains
		})
		if !contains {
			logger.InfoWith("filtering tag since the required filter vector is missing").
				String("tag", inputSet.String()).
				String("filter vector", filterVector).
				Write()
			return false
		}
	}
	return true
}

// getTags fetches the set of tags for the given Docker repository.
// The returned TagMap is always initialized.
func (s *TuplipSource) getTags() (tagMap map[string]mapset.Set, err error) {
	tagMap = make(map[string]mapset.Set)
	if err = validation.Validate(s.Repository,
		validation.Required,
	); err != nil {
		return nil, err
	}
	logger.InfoWith("fetching tags for remote repository").String("repository", s.Repository).Write()
	if s.tuplip.Simulate {
		return make(map[string]mapset.Set), nil
	}
	hub, err := registry.NewCustom(DockerRegistry, registry.Options{
		Logf: registry.Quiet,
	})
	if err != nil {
		return nil, err
	}
	tags, err := hub.Tags(s.Repository)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, errors.New("no Docker tags could be found on the given remote")
	}
	for _, tag := range tags {
		tagVectors := strings.Split(tag, DockerTagSeparator)
		vectorSet := mapset.NewSet()
		for _, v := range tagVectors {
			vectorSet.Add(v)
		}
		tagMap[tag] = vectorSet
	}
	return
}

// dockerTag tags all inputTags given the sourceTag.
func (s *TuplipSource) dockerTag(sourceTag string) func(inputTag string) (string, error) {
	return func(inputTag string) (o string, err error) {
		var targetTag = inputTag
		if repo := s.Repository; repo != "" {
			targetTag = strings.Join([]string{repo, targetTag}, VersionSeparator)
		}
		cmd := exec.Command("docker", "tag", sourceTag, targetTag)
		logger.InfoWith("execute").
			String("args", strings.Join(cmd.Args, " ")).
			Write()
		if !s.tuplip.Simulate {
			if _, err := cmd.CombinedOutput(); err != nil {
				return "", err
			}
		}
		logger.InfoWith("tagged").String("tag", targetTag).Write()
		return inputTag, nil
	}
}

// dockerPush pushes all inputTags to the Docker Hub and prepends a success or fail message to the respective tags.
func (s *TuplipSource) dockerPush() func(inputTag string) (tagMsg string, err error) {
	tagMap, _ := s.getTags()
	return func(inputTag string) (tagMsg string, err error) {
		var targetTag = inputTag
		if repo := s.Repository; repo != "" {
			targetTag = strings.Join([]string{repo, targetTag}, VersionSeparator)
		}
		cmd := exec.Command("docker", "push", targetTag)
		logger.InfoWith("execute").
			String("args", strings.Join(cmd.Args, " ")).
			Write()
		if !s.tuplip.Simulate {
			if _, err := cmd.CombinedOutput(); err != nil {
				return "", err
			}
		}
		if _, exist := tagMap[inputTag]; exist {
			logger.InfoWith("repushed").String("tag", targetTag).Write()
		} else {
			logger.InfoWith("pushed").String("tag", targetTag).Write()
		}
		return targetTag, nil
	}
}

// requireDocker ensures that docker is available in the PATH.
func (s *TuplipSource) requireDocker() error {
	cmd := exec.Command("docker")
	logger.InfoWith("execute").
		String("args", strings.Join(cmd.Args, " ")).
		Write()
	if !s.tuplip.Simulate {
		if _, err := cmd.CombinedOutput(); err != nil {
			return err
		}
	}
	return nil
}
