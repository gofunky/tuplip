package tupliplib

import (
	"bufio"
	"github.com/gofunky/pyraset/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofunky/automi/collectors"
)

func TestTuplipStream(t *testing.T) {
	type args struct {
		input         []string
		sep           string
		requireSemver bool
	}
	tests := []struct {
		name      string
		t         Tuplip
		buildArgs *args
		pushArgs  *args
		want      []string
		wantErr   bool
	}{
		{
			name:      "Empty Slice",
			buildArgs: &args{input: []string{}},
			wantErr:   true,
		},
		{
			name:      "Empty Element",
			buildArgs: &args{input: []string{""}},
			wantErr:   true,
		},
		{
			name:      "Simple Unary Tag",
			buildArgs: &args{input: []string{"alias"}},
			want:      []string{"alias"},
		},
		{
			name:      "Simple Binary Tag",
			buildArgs: &args{input: []string{"alias", "foo"}},
			want:      []string{"alias", "foo", "alias-foo"},
		},
		{
			name:      "Simple Tertiary Tag",
			buildArgs: &args{input: []string{"alias", "foo", "boo"}},
			want:      []string{"alias", "foo", "alias-foo", "boo", "alias-boo", "boo-foo", "alias-boo-foo"},
		},
		{
			name:      "Simple Binary Tag With Short Version",
			buildArgs: &args{input: []string{"alias", "foo:2.0"}},
			want:      []string{"alias", "foo", "foo2", "foo2.0", "alias-foo", "alias-foo2", "alias-foo2.0"},
		},
		{
			name:      "Simple Unary Tag With Long Version",
			buildArgs: &args{input: []string{"foo:2.0.0"}},
			want:      []string{"foo", "foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name:      "Wildcard Binary Tag With Long Version",
			buildArgs: &args{input: []string{"_:2.0.0", "alias"}},
			want:      []string{"alias", "2", "2.0", "2.0.0", "2-alias", "2.0-alias", "2.0.0-alias"},
		},
		{
			name:      "Wildcard Binary Tag With Long Version And Major Exclusion",
			t:         Tuplip{ExcludeMajor: true},
			buildArgs: &args{input: []string{"_:2.0.0", "alias"}},
			want:      []string{"alias", "2.0", "2.0.0", "2.0-alias", "2.0.0-alias"},
		},
		{
			name:      "Wildcard Binary Tag With Long Version And Minor Exclusion",
			t:         Tuplip{ExcludeMinor: true},
			buildArgs: &args{input: []string{"_:2.0.0", "alias"}},
			want:      []string{"alias", "2", "2.0.0", "2-alias", "2.0.0-alias"},
		},
		{
			name:      "Simple Unary Tag With Long Version And Base Exclusion",
			t:         Tuplip{ExcludeBase: true},
			buildArgs: &args{input: []string{"foo:2.0.0"}},
			want:      []string{"foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name:      "Wildcard Unary Tag With Long Version And Base Exclusion",
			t:         Tuplip{ExcludeBase: true},
			buildArgs: &args{input: []string{"_:2.0.0"}},
			want:      []string{"2", "2.0", "2.0.0"},
		},
		{
			name:      "Wildcard Unary Tag With Long Version And Latest Addition",
			t:         Tuplip{AddLatest: true},
			buildArgs: &args{input: []string{"_:2.0.0", "foo"}},
			want:      []string{"latest", "foo", "2", "2.0", "2.0.0", "2-foo", "2.0-foo", "2.0.0-foo"},
		},
		{
			name:      "Wildcard Unary Tag With Long Version And a Different Separator",
			buildArgs: &args{input: []string{" _:2.0.0; foo "}, sep: ";"},
			pushArgs:  &args{input: []string{"_:2.0.0", "foo"}},
			want:      []string{"foo", "2", "2.0", "2.0.0", "2-foo", "2.0-foo", "2.0.0-foo"},
		},
		{
			name:      "Unary Patch Version Tag Without Base",
			buildArgs: &args{input: []string{" _:2"}},
			pushArgs:  &args{input: []string{"_:2"}},
			want:      []string{"2"},
		},
		{
			name:      "Invalid Semantic Version With Semver Required",
			buildArgs: &args{input: []string{" _:2.0 foo"}, requireSemver: true},
			wantErr:   true,
		},
		{
			name:      "Simple Binary Tag With Exclusive Latest Enabled",
			buildArgs: &args{input: []string{"alias", "foo"}},
			want:      []string{"alias", "foo", "alias-foo"},
			t:         Tuplip{ExclusiveLatest: true},
		},
		{
			name:      "Invalid Latest Version With Semver Required",
			buildArgs: &args{input: []string{"_:latest"}, requireSemver: true},
			t:         Tuplip{ExclusiveLatest: true},
			wantErr:   true,
		},
		{
			name:      "Exclusive Latest Version",
			buildArgs: &args{input: []string{"_:latest", "foo"}},
			t:         Tuplip{ExclusiveLatest: true},
			want:      []string{"latest"},
		},
		{
			name:      "Filter Unary Unversioned",
			buildArgs: &args{input: []string{"_:1.0", "foo", "goo"}},
			t:         Tuplip{Filter: []string{"foo"}},
			want:      []string{"foo", "foo-goo", "1-foo", "1.0-foo", "1-foo-goo", "1.0-foo-goo"},
		},
		{
			name:      "Filter Unary Versioned",
			buildArgs: &args{input: []string{"_:1", "docker", "alpine:3.8"}},
			t:         Tuplip{Filter: []string{"alpine"}},
			want: []string{
				"1-alpine3.8", "1-alpine-docker", "alpine-docker", "alpine", "alpine3", "alpine3.8",
				"alpine3-docker", "alpine3.8-docker", "1-alpine", "1-alpine3-docker", "1-alpine3.8-docker", "1-alpine3",
			},
		},
		{
			name:      "Filter Binary Versioned",
			buildArgs: &args{input: []string{"_:1", "docker:2", "alpine:3.8"}},
			t:         Tuplip{Filter: []string{"alpine", "docker"}},
			want: []string{
				"1-alpine-docker", "1-alpine-docker2", "alpine-docker", "alpine-docker2",
				"1-alpine3-docker", "1-alpine3-docker2", "alpine3-docker", "alpine3-docker2",
				"1-alpine3.8-docker", "1-alpine3.8-docker2", "alpine3.8-docker", "alpine3.8-docker2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name+"_BuildFromReader", func(t *testing.T) {
			tt.t.Simulate = true
			src := strings.NewReader(strings.Join(tt.buildArgs.input, " "))
			buildSource := tt.t.FromReader(src, tt.buildArgs.sep)
			buildStream := buildSource.Build(tt.buildArgs.requireSemver)
			buildCollector := collectors.Slice()
			buildStream.Into(buildCollector)
			select {
			case gotErr := <-buildStream.Open():
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("Tuplip.Build() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
			case <-time.After(500 * time.Millisecond):
				t.Fatal("Waited too long ...")
			}
			gotOutput := mapset.NewSet(buildCollector.Get()...)
			wantSet := mapset.NewSet()
			for _, w := range tt.want {
				wantSet.Add(w)
			}
			if !gotOutput.Equal(wantSet) {
				t.Errorf("Tuplip.Build() = %v, want %v, difference %v",
					gotOutput, wantSet, gotOutput.Difference(wantSet))
			}
		})
	}
	for _, tt := range tests {
		t.Run(tt.name+"_PushFromSlice", func(t *testing.T) {
			var args = tt.buildArgs
			if tt.pushArgs != nil {
				args = tt.pushArgs
			}
			tt.t.Simulate = true
			tagSource := tt.t.FromSlice(args.input)
			var tagStream = tagSource.Build(args.requireSemver)
			tagStream, tagErr := tagSource.Tag("source")
			tagStream, pushErr := tagSource.Push()
			tagCollector := collectors.Slice()
			tagStream.Into(tagCollector)
			select {
			case gotErr := <-tagStream.Open():
				if (gotErr != nil || tagErr != nil || pushErr != nil) != tt.wantErr {
					t.Errorf("Tuplip.Push() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
			case <-time.After(1000 * time.Millisecond):
				t.Fatal("Waited too long ...")
			}
			gotTagOutput := mapset.NewSet(tagCollector.Get()...)
			wantTagOutput := mapset.NewSet()
			for _, w := range tt.want {
				wantTagOutput.Add(w)
			}
			if !gotTagOutput.Equal(wantTagOutput) {
				t.Errorf("Tuplip.Push() = %v, want %v, difference %v",
					gotTagOutput, wantTagOutput, gotTagOutput.Difference(wantTagOutput))
			}
		})
	}
}

func TestTuplipStream_FindFromReader(t *testing.T) {
	type args struct {
		input []string
	}
	tests := []struct {
		name    string
		s       TuplipSource
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Empty Repository",
			args:    args{[]string{""}},
			wantErr: true,
		},
		{
			name:    "Input No Match",
			s:       TuplipSource{Repository: "gofunky/git"},
			args:    args{[]string{"unknown"}},
			wantErr: true,
		},
		{
			name: "Simple Unary Tag",
			s:    TuplipSource{Repository: "gofunky/git"},
			args: args{[]string{"envload"}},
			want: "envload",
		},
		{
			name: "Simple Binary Tag",
			s:    TuplipSource{Repository: "gofunky/git"},
			args: args{[]string{"envload", "alpine:3.8"}},
			want: "alpine3.8-envload",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := strings.NewReader(strings.Join(tt.args.input, " "))
			tuplipSrc := new(Tuplip).FromReader(src, " ")
			tuplipSrc.Repository = tt.s.Repository
			tStream, srcErr := tuplipSrc.Find()
			collector := collectors.Slice()
			var gotErr error
			if tStream != nil {
				tStream.Into(collector)
				select {
				case gotErr = <-tStream.Open():

				case <-time.After(500 * time.Millisecond):
					t.Fatal("Waited too long ...")
					return
				}
			}
			if (gotErr != nil || srcErr != nil) != tt.wantErr {
				t.Errorf("Tuplip.Open() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			gotOutput := mapset.NewSet(collector.Get()...)
			wantSet := mapset.NewSet(tt.want)
			if !tt.wantErr && !gotOutput.Equal(wantSet) {
				t.Errorf("Tuplip.Open() = %v, want %v, difference %v",
					gotOutput, wantSet, gotOutput.Difference(wantSet))
			}
		})
	}
}

func TestTuplipStream_BuildFromFile_WithoutRepository(t *testing.T) {
	t.Run("Test Dockerfile", func(t *testing.T) {
		filePath, err := filepath.Abs("./../../test/tags_without_repository.txt")
		if err != nil {
			panic(err)
		}
		expectedSet := mapset.NewSet()
		file, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				expectedSet.Add(line)
			}
		}
		tuplipSrc, err := new(Tuplip).FromFile("../../test/WithoutRepository.Dockerfile", "")
		if err != nil {
			t.Errorf("Tuplip.Build() error = %v", err)
			return
		}
		tStream := tuplipSrc.Build(false)
		collector := collectors.Slice()
		tStream.Into(collector)
		select {
		case gotErr := <-tStream.Open():
			if gotErr != nil {
				t.Errorf("Tuplip.Build() error = %v", gotErr)
				return
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Waited too long ...")
		}
		gotOutput := mapset.NewSet(collector.Get()...)
		if !gotOutput.Equal(expectedSet) {
			t.Errorf("Tuplip.Build() = %v,\nwant %v,\ndifference %v",
				gotOutput, expectedSet, gotOutput.Difference(expectedSet))
		}
	})
}

func TestTuplipStream_BuildFromFile_WithRepository(t *testing.T) {
	t.Run("Test Dockerfile With Repository", func(t *testing.T) {
		filePath, err := filepath.Abs("./../../test/tags_with_repository.txt")
		if err != nil {
			panic(err)
		}
		expectedSet := mapset.NewSet()
		file, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				expectedSet.Add(line)
			}
		}
		tuplipSrc, err := new(Tuplip).FromFile("../../test/WithRepository.Dockerfile", "")
		if err != nil {
			t.Errorf("Tuplip.Build() error = %v", err)
			return
		}
		tStream := tuplipSrc.Build(false)
		collector := collectors.Slice()
		tStream.Into(collector)
		select {
		case gotErr := <-tStream.Open():
			if gotErr != nil {
				t.Errorf("Tuplip.Build() error = %v", gotErr)
				return
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Waited too long ...")
		}
		gotOutput := mapset.NewSet(collector.Get()...)
		if !gotOutput.Equal(expectedSet) {
			t.Errorf("Tuplip.Build() = %v,\nwant %v,\ndifference %v",
				gotOutput, expectedSet, gotOutput.Difference(expectedSet))
		}
	})
}

func TestTuplipStream_PushStraightFromSlice(t *testing.T) {
	type args struct {
		input         []string
		sep           string
		requireSemver bool
	}
	tests := []struct {
		name     string
		t        Tuplip
		pushArgs *args
		want     []string
		wantErr  bool
	}{
		{
			name:     "Empty Slice",
			pushArgs: &args{input: []string{}},
			want:     []string{},
		},
		{
			name:     "Empty Element",
			pushArgs: &args{input: []string{""}},
			want:     []string{},
		},
		{
			name:     "Simple Unary Tag",
			pushArgs: &args{input: []string{"alias"}},
			want:     []string{"alias"},
		},
		{
			name:     "Simple Binary Tag",
			pushArgs: &args{input: []string{"alias", "foo"}},
			want:     []string{"alias", "foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args = tt.pushArgs
			tt.t.Simulate = true
			tagSource := tt.t.FromSlice(args.input)
			var tagStream = tagSource.Straight()
			tagStream, tagErr := tagSource.Tag("source")
			tagStream, pushErr := tagSource.Push()
			tagCollector := collectors.Slice()
			tagStream.Into(tagCollector)
			select {
			case gotErr := <-tagStream.Open():
				if (gotErr != nil || tagErr != nil || pushErr != nil) != tt.wantErr {
					t.Errorf("Tuplip.Push() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
			case <-time.After(500 * time.Millisecond):
				t.Fatal("Waited too long ...")
			}
			gotTagOutput := mapset.NewSet(tagCollector.Get()...)
			wantTagOutput := mapset.NewSet()
			for _, w := range tt.want {
				wantTagOutput.Add(w)
			}
			if !gotTagOutput.Equal(wantTagOutput) {
				t.Errorf("Tuplip.Push() = %v, want %v, difference %v",
					gotTagOutput, wantTagOutput, gotTagOutput.Difference(wantTagOutput))
			}
		})
	}
}

func TestTuplip_getTags(t *testing.T) {
	type fields struct {
		Repository string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Unknown Repository",
			fields:  fields{Repository: "gofunky/unknown"},
			wantErr: true,
		},
		{
			name:    "Invalid Repository Name",
			fields:  fields{Repository: "$%"},
			wantErr: true,
		},
		{
			name:    "Empty Repository Name",
			fields:  fields{Repository: ""},
			wantErr: true,
		},
		{
			name:    "Unary Repository",
			fields:  fields{Repository: "alpine"},
			wantErr: true, // TODO: find a way to check official images
		},
		{
			name:   "Binary Repository",
			fields: fields{Repository: "gofunky/git"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tu := &TuplipSource{Repository: tt.fields.Repository, tuplip: &Tuplip{}}
			tagSet, err := tu.getTags()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tuplip.getTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(tagSet) == 0 {
				t.Errorf("Tuplip.getTags() returned no tags")
				return
			}
		})
	}
}
