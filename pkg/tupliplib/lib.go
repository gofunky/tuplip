package tupliplib

import (
	"github.com/deckarep/golang-set"
	"github.com/gofunky/automi/emitters"
	"github.com/gofunky/automi/stream"
	"io"
)

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

// FromReader builds a tuplip source from a pipe reader.
func (t *Tuplip) FromReader(src io.Reader) *tuplipSource {
	t.check()
	stm := stream.New(emitters.Scanner(src, nil))
	stm.FlatMap(t.splitBySeparator)
	stm.Filter(t.nonEmpty)
	return &tuplipSource{t, stm}
}

// Build builds a tuplip stream from a io.Reader as scanner. The returned stream has no configured sink.
func (s *tuplipSource) Build() *stream.Stream {
	stm := s.stream
	stm.Map(s.tuplip.splitVersion)
	stm.Map(s.tuplip.packInSet)
	stm.Reduce(mapset.NewSet(), s.tuplip.mergeSets)
	stm.Map(s.tuplip.power)
	stm.Map(s.tuplip.addLatestTag)
	stm.FlatMap(s.tuplip.failOnEmpty)
	stm.FlatMap(s.tuplip.join)
	stm.Filter(s.tuplip.nonEmpty)
	return stm
}
