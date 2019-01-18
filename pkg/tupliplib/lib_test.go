package tupliplib

import (
	"github.com/deckarep/golang-set"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/gofunky/automi/collectors"
)

func TestTuplipStream_BuildFromReader(t *testing.T) {
	type args struct {
		input         []string
		sep           string
		requireSemver bool
	}
	tests := []struct {
		name    string
		t       Tuplip
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "Empty Slice",
			args:    args{input: []string{}},
			wantErr: true,
		},
		{
			name:    "Empty Element",
			args:    args{input: []string{""}},
			wantErr: true,
		},
		{
			name: "Simple Unary Tag",
			args: args{input: []string{"alias"}},
			want: []string{"alias"},
		},
		{
			name: "Simple Binary Tag",
			args: args{input: []string{"alias", "foo"}},
			want: []string{"alias", "foo", "alias-foo"},
		},
		{
			name: "Simple Tertiary Tag",
			args: args{input: []string{"alias", "foo", "boo"}},
			want: []string{"alias", "foo", "alias-foo", "boo", "alias-boo", "boo-foo", "alias-boo-foo"},
		},
		{
			name: "Simple Binary Tag With Short Version",
			args: args{input: []string{"alias", "foo:2.0"}},
			want: []string{"alias", "foo", "foo2", "foo2.0", "alias-foo", "alias-foo2", "alias-foo2.0"},
		},
		{
			name: "Simple Unary Tag With Long Version",
			args: args{input: []string{"foo:2.0.0"}},
			want: []string{"foo", "foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name: "Wildcard Binary Tag With Long Version",
			args: args{input: []string{"_:2.0.0", "alias"}},
			want: []string{"alias", "2", "2.0", "2.0.0", "2-alias", "2.0-alias", "2.0.0-alias"},
		},
		{
			name: "Wildcard Binary Tag With Long Version And Major Exclusion",
			t:    Tuplip{ExcludeMajor: true},
			args: args{input: []string{"_:2.0.0", "alias"}},
			want: []string{"alias", "2.0", "2.0.0", "2.0-alias", "2.0.0-alias"},
		},
		{
			name: "Wildcard Binary Tag With Long Version And Minor Exclusion",
			t:    Tuplip{ExcludeMinor: true},
			args: args{input: []string{"_:2.0.0", "alias"}},
			want: []string{"alias", "2", "2.0.0", "2-alias", "2.0.0-alias"},
		},
		{
			name: "Simple Unary Tag With Long Version And Base Exclusion",
			t:    Tuplip{ExcludeBase: true},
			args: args{input: []string{"foo:2.0.0"}},
			want: []string{"foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name: "Wildcard Unary Tag With Long Version And Base Exclusion",
			t:    Tuplip{ExcludeBase: true},
			args: args{input: []string{"_:2.0.0"}},
			want: []string{"2", "2.0", "2.0.0"},
		},
		{
			name: "Wildcard Unary Tag With Long Version And Latest Addition",
			t:    Tuplip{AddLatest: true},
			args: args{input: []string{"_:2.0.0", "foo"}},
			want: []string{"latest", "foo", "2", "2.0", "2.0.0", "2-foo", "2.0-foo", "2.0.0-foo"},
		},
		{
			name: "Wildcard Unary Tag With Long Version And a Different Separator",
			args: args{input: []string{" _:2.0.0; foo "}, sep: ";"},
			want: []string{"foo", "2", "2.0", "2.0.0", "2-foo", "2.0-foo", "2.0.0-foo"},
		},
		{
			name: "Unary Patch Version Tag Without Base",
			args: args{input: []string{" _:2"}},
			want: []string{"2"},
		},
		{
			name:    "Invalid Semantic Version With Semver Required",
			args:    args{input: []string{" _:2.0 foo"}, requireSemver: true},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := strings.NewReader(strings.Join(tt.args.input, " "))
			tuplipSrc := tt.t.FromReader(src, tt.args.sep)
			tStream := tuplipSrc.Build(tt.args.requireSemver)
			collector := collectors.Slice()
			tStream.Into(collector)
			select {
			case gotErr := <-tStream.Open():
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("Tuplip.Build() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
			case <-time.After(500 * time.Millisecond):
				t.Fatal("Waited too long ...")
			}
			gotOutput := mapset.NewSetFromSlice(collector.Get())
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
			gotOutput := mapset.NewSetFromSlice(collector.Get())
			wantSet := mapset.NewSet(tt.want)
			if !tt.wantErr && !gotOutput.Equal(wantSet) {
				t.Errorf("Tuplip.Open() = %v, want %v, difference %v",
					gotOutput, wantSet, gotOutput.Difference(wantSet))
			}
		})
	}
}

func TestTuplipStream_BuildFromFile(t *testing.T) {
	t.Run("Test Dockerfile", func(t *testing.T) {
		expectedFile, err := ioutil.ReadFile("../../test/tags.txt")
		if err != nil {
			t.Errorf("test error = %v", err)
			return
		}
		expectedLines := strings.Split(string(expectedFile), "\n")
		expectedSet := mapset.NewSet()
		for _, l := range expectedLines {
			if l != "" {
				expectedSet.Add(l)
			}
		}
		tuplipSrc, err := new(Tuplip).FromFile("../../test/Dockerfile")
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
		gotOutput := mapset.NewSetFromSlice(collector.Get())
		if !gotOutput.Equal(expectedSet) {
			t.Errorf("Tuplip.Build() = %v, want %v, difference %v",
				gotOutput, expectedSet, gotOutput.Difference(expectedSet))
		}
	})
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
			tu := NewTuplipSource(&Tuplip{}, nil)
			tu.Repository = tt.fields.Repository
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
