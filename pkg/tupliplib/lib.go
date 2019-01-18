package tupliplib

import (
	"github.com/deckarep/golang-set"
	"github.com/gofunky/automi/emitters"
	"github.com/gofunky/automi/stream"
	"io"
	"io/ioutil"
	"strings"
)

// Tuplip contains the parameters for the Docker tag generation.
type Tuplip struct {
	// ExcludeMajor excludes the major versions from the considered version variants.
	ExcludeMajor bool `short:"m" help:"excludes the major versions from the considered version variants"`
	// ExcludeMinor excludes the minor versions from the considered version variants.
	ExcludeMinor bool `short:"i" help:"excludes the minor versions from the considered version variants"`
	// ExcludeBase excludes the base alias without version suffix from the considered version variants.
	ExcludeBase bool `short:"b" help:"excludes the base alias without version suffix from the considered version variants"`
	// AddLatest adds an additional 'latest' tag to the result set.
	AddLatest bool `short:"l" help:"adds an additional 'latest' tag to the result set"`
	// Simulate prevents the execution of any Docker commands.
	Simulate bool `hidden:""`
}

// TuplipSource is the intermediary-built Tuplip stream containing only the source parsing steps.
type TuplipSource struct {
	tuplip *Tuplip
	stream *stream.Stream
	// Repository is the Docker Hub repository of the root tag vector in the format `organization/repository`.
	Repository string
}

// FromReader builds a tuplip source from a io.Reader as scanner.
// The separator is used to split the tag vectors from the same row. It defaults to an empty space.
func (t *Tuplip) FromReader(src io.Reader, sep string) *TuplipSource {
	logger.InfoWith("queueing read from reader").
		String("separator", sep).
		Write()
	stm := stream.New(emitters.Scanner(src, nil))
	stm.FlatMap(t.splitBySeparator(sep))
	stm.Filter(nonEmpty)
	return &TuplipSource{tuplip: t, stream: stm}
}

// FromSlice builds a tuplip source from a slice.
func (t *Tuplip) FromSlice(src []string) *TuplipSource {
	logger.Info("queueing read from slice")
	stm := stream.New(emitters.Slice(src))
	stm.Filter(nonEmpty)
	return &TuplipSource{tuplip: t, stream: stm}
}

// FromFile builds a tuplip source from a Dockerfile.
func (t *Tuplip) FromFile(src string) (source *TuplipSource, err error) {
	logger.InfoWith("queueing read from Dockerfile").
		String("file", src).
		Write()
	if content, readErr := ioutil.ReadFile(src); readErr != nil {
		return nil, readErr
	} else {
		lines := strings.Split(string(content), "\n")
		if markedLines, repository, err := markRootInstruction(lines); err != nil {
			return nil, err
		} else {
			stm := stream.New(emitters.Slice(markedLines))
			stm.Filter(nonEmpty)
			stm.Filter(fromInstruction)
			stm.Map(toTagVector)
			source = &TuplipSource{tuplip: t, stream: stm, Repository: repository}
			return source, nil
		}
	}
}

// Build defines a tuplip stream that builds a complete set of Docker tags. The returned stream has no configured sink.
// requireSemver enables semantic version checks. Short versions are not allowed then.
func (s *TuplipSource) Build(requireSemver bool) (stream *stream.Stream) {
	logger.InfoWith("queueing build").
		Bool("require semantic version", requireSemver).
		Write()
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

// Tag extends the *TuplipSource.Build stream by a `docker tag` execution for all generated tags.
// requireSemver enables semantic version checks. Short versions are not allowed then.
func (s *TuplipSource) Tag(requireSemver bool, sourceTag string) (stream *stream.Stream, err error) {
	logger.InfoWith("queueing tagging").
		Bool("require semantic version", requireSemver).
		String("source tag", sourceTag).
		Write()
	if err = s.requireDocker(); err != nil {
		return
	}
	stream = s.Build(requireSemver)
	stream.Map(s.dockerTag(sourceTag))
	return
}

// Push extends the *TuplipSource.Tag stream by a `docker push` execution for all generated tags.
// requireSemver enables semantic version checks. Short versions are not allowed then.
func (s *TuplipSource) Push(requireSemver bool, sourceTag string) (stream *stream.Stream, err error) {
	logger.InfoWith("queueing push").
		Bool("require semantic version", requireSemver).
		String("source tag", sourceTag).
		Write()
	if sourceTag == "" {
		if err = s.requireDocker(); err != nil {
			return
		}
		stream = s.Build(requireSemver)
	} else {
		if stream, err = s.Tag(requireSemver, sourceTag); err != nil {
			return
		}
	}
	stream.Map(s.dockerPush)
	return
}

// DirectPush performs a tagging and `docker push` execution for all given tags on the given source tag.
// TODO Check if this can be done in Push()
func (s *TuplipSource) DirectPush(sourceTag string) (stream *stream.Stream, err error) {
	logger.InfoWith("queueing direct push").
		String("source tag", sourceTag).
		Write()
	if err = s.requireDocker(); err != nil {
		return
	}
	stream = s.stream
	if sourceTag == "" {
		stream.Map(s.dockerTag(sourceTag))
	}
	stream.Map(s.dockerPush)
	return
}

// Find defines a tuplip stream that finds an appropriate matching Docker tag in the given Docker Hub repository.
// The returned stream has no configured sink.
func (s *TuplipSource) Find() (stream *stream.Stream, err error) {
	logger.Info("queueing find")
	stream = s.stream
	if tagMap, err := s.getTags(); err != nil {
		return nil, err
	} else {
		stream.Map(s.tuplip.splitVersion(false))
		stream.Reduce(tagMap, removeCommon)
		stream.Map(keyForSmallest)
	}
	return
}
