package tupliplib

import (
	"errors"
	"github.com/deckarep/golang-set"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/gofunky/automi/emitters"
	"github.com/gofunky/automi/stream"
	"github.com/nokia/docker-registry-client/registry"
	"io"
	"strings"
)

// DockerRegistry is the Docker Hub registry URL.
const DockerRegistry = "https://registry-1.docker.io/"

// VersionSeparator is the separator that separates the alias form the semantic version.
const VersionSeparator = ":"

// WildcardDependency is the alias for a wildcard dependency to build a root tag vector
// (i.e., semantic version without a prefix).
const WildcardDependency = "_"

// VersionDot is the separator that separates the digits of a semantic version.
const VersionDot = "."

// DockerTagSeparator is the separator that separates the sub tags in a Docker tag.
const DockerTagSeparator = "-"

// VectorSeparator is the default tag vector separator.
const VectorSeparator = " "

// Tuplip contains the parameters for the Docker tag generation.
type Tuplip struct {
	// ExcludeMajor excludes the major versions from the result set.
	ExcludeMajor bool
	// ExcludeMinor excludes the minor versions from the result set.
	ExcludeMinor bool
	// ExcludeBase excludes the base alias without version suffix from the result set.
	ExcludeBase bool
	// AddLatest adds an additional 'latest' tag to the result set.
	AddLatest bool
	// Separator to split the separate tag vector aliases. The default separator is single space.
	Separator string
	// Docker repository of the root tag vector in the format `organization/repository`.
	Repository string
}

// tuplipSource is the intermediarily-built Tuplip stream containing only the source parsing steps.
type tuplipSource struct {
	tuplip *Tuplip
	stream *stream.Stream
}

// check performs post-construction checks.
func (t *Tuplip) check() {
	if t.Separator == "" {
		t.Separator = VectorSeparator
	}
}

// FromReader builds a tuplip source from a io.Reader as scanner.
func (t *Tuplip) FromReader(src io.Reader) *tuplipSource {
	t.check()
	stm := stream.New(emitters.Scanner(src, nil))
	stm.FlatMap(t.splitBySeparator)
	stm.Filter(nonEmpty)
	return &tuplipSource{t, stm}
}

// Build defines a tuplip stream that builds a complete set of Docker tags. The returned stream has no configured sink.
func (s *tuplipSource) Build() (stream *stream.Stream) {
	stream = s.stream
	stream.Map(s.tuplip.splitVersion)
	stream.Map(packInSet)
	stream.Reduce(mapset.NewSet(), mergeSets)
	stream.Map(power)
	stream.Map(s.tuplip.addLatestTag)
	stream.FlatMap(failOnEmpty)
	stream.FlatMap(s.tuplip.join)
	stream.Filter(nonEmpty)
	return
}

// getTags fetches the set of tags for the given Docker repository.
func (t *Tuplip) getTags() (tagMap map[string]mapset.Set, err error) {
	if err = validation.Validate(t.Repository,
		validation.Required,
	); err != nil {
		return nil, err
	}
	if hub, err := registry.New(DockerRegistry, "", ""); err != nil {
		return nil, err
	} else {
		if tags, err := hub.Tags(t.Repository); err != nil {
			return nil, err
		} else {
			if len(tags) == 0 {
				return nil, errors.New("no Docker tags could be found on the given remote")
			}
			tagMap := make(map[string]mapset.Set)
			for _, tag := range tags {
				tagVectors := strings.Split(tag, DockerTagSeparator)
				vectorSet := mapset.NewSet()
				for _, v := range tagVectors {
					vectorSet.Add(v)
				}
				tagMap[tag] = vectorSet
			}
			return tagMap, nil
		}
	}
}

// Find defines a tuplip stream that finds the matching Docker tag in the Docker Hub.
// The returned stream has no configured sink.
func (s *tuplipSource) Find() (stream *stream.Stream, err error) {
	stream = s.stream
	if tagMap, err := s.tuplip.getTags(); err != nil {
		return nil, err
	} else {
		stream.Map(s.tuplip.splitVersion)
		stream.Reduce(tagMap, removeCommon)
		stream.Map(keyForSmallest)
	}
	return
}
