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
	// ExcludeMajor excludes the major versions from the considered version variants.
	ExcludeMajor bool `help:"excludes the major versions from the considered version variants"`
	// ExcludeMinor excludes the minor versions from the considered version variants.
	ExcludeMinor bool `help:"excludes the minor versions from the considered version variants"`
	// ExcludeBase excludes the base alias without version suffix from the considered version variants.
	ExcludeBase bool `help:"excludes the base alias without version suffix from the considered version variants"`
	// AddLatest adds an additional 'latest' tag to the result set.
	AddLatest bool `short:"l" help:"adds an additional 'latest' tag to the result set"`
}

// TuplipSource is the intermediarily-built Tuplip stream containing only the source parsing steps.
type TuplipSource struct {
	tuplip *Tuplip
	stream *stream.Stream
}

// FromReader builds a tuplip source from a io.Reader as scanner.
// The separator is used to split the tag vectors from the same row. It defaults to an empty space.
func (t *Tuplip) FromReader(src io.Reader, sep string) *TuplipSource {
	stm := stream.New(emitters.Scanner(src, nil))
	stm.FlatMap(t.splitBySeparator(sep))
	stm.Filter(nonEmpty)
	return &TuplipSource{t, stm}
}

// FromSlice builds a tuplip source from a slice.
func (t *Tuplip) FromSlice(src []string) *TuplipSource {
	stm := stream.New(emitters.Slice(src))
	stm.Filter(nonEmpty)
	return &TuplipSource{t, stm}
}

// Build defines a tuplip stream that builds a complete set of Docker tags. The returned stream has no configured sink.
// requireSemver enables semantic version checks. Short versions are not allowed then.
func (s *TuplipSource) Build(requireSemver bool) (stream *stream.Stream) {
	stream = s.stream
	stream.Map(s.tuplip.splitVersion(requireSemver))
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
func (t *Tuplip) getTags(repository string) (tagMap map[string]mapset.Set, err error) {
	if err = validation.Validate(repository,
		validation.Required,
	); err != nil {
		return nil, err
	}
	if hub, err := registry.NewCustom(DockerRegistry, registry.Options{
		Logf: registry.Quiet,
	}); err != nil {
		return nil, err
	} else {
		hub.Logf = registry.Quiet
		if tags, err := hub.Tags(repository); err != nil {
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

// Find defines a tuplip stream that finds an appropriate matching Docker tag in the given Docker Hub repository.
// The returned stream has no configured sink.
func (s *TuplipSource) Find(repository string) (stream *stream.Stream, err error) {
	stream = s.stream
	if tagMap, err := s.tuplip.getTags(repository); err != nil {
		return nil, err
	} else {
		stream.Map(s.tuplip.splitVersion(false))
		stream.Reduce(tagMap, removeCommon)
		stream.Map(keyForSmallest)
	}
	return
}
