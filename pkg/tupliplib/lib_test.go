package tupliplib

import (
	"github.com/deckarep/golang-set"
	"strings"
	"testing"
	"time"

	"github.com/gofunky/automi/collectors"
)

func TestTuplipStream_BuildFromReader(t *testing.T) {
	type args struct {
		input []string
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
			args:    args{[]string{}},
			wantErr: true,
		},
		{
			name:    "Empty Element",
			args:    args{[]string{""}},
			wantErr: true,
		},
		{
			name: "Simple Unary Tag",
			args: args{[]string{"alias"}},
			want: []string{"alias"},
		},
		{
			name: "Simple Binary Tag",
			args: args{[]string{"alias", "foo"}},
			want: []string{"alias", "foo", "alias-foo"},
		},
		{
			name: "Simple Tertiary Tag",
			args: args{[]string{"alias", "foo", "boo"}},
			want: []string{"alias", "foo", "alias-foo", "boo", "alias-boo", "boo-foo", "alias-boo-foo"},
		},
		{
			name: "Simple Binary Tag With Short Version",
			args: args{[]string{"alias", "foo:2.0"}},
			want: []string{"alias", "foo", "foo2", "foo2.0", "alias-foo", "alias-foo2", "alias-foo2.0"},
		},
		{
			name: "Simple Unary Tag With Long Version",
			args: args{[]string{"foo:2.0.0"}},
			want: []string{"foo", "foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name: "Wildcard Binary Tag With Long Version",
			args: args{[]string{"_:2.0.0", "alias"}},
			want: []string{"alias", "2", "2.0", "2.0.0", "2-alias", "2.0-alias", "2.0.0-alias"},
		},
		{
			name: "Wildcard Binary Tag With Long Version And Major Exclusion",
			t:    Tuplip{ExcludeMajor: true},
			args: args{[]string{"_:2.0.0", "alias"}},
			want: []string{"alias", "2.0", "2.0.0", "2.0-alias", "2.0.0-alias"},
		},
		{
			name: "Wildcard Binary Tag With Long Version And Minor Exclusion",
			t:    Tuplip{ExcludeMinor: true},
			args: args{[]string{"_:2.0.0", "alias"}},
			want: []string{"alias", "2", "2.0.0", "2-alias", "2.0.0-alias"},
		},
		{
			name: "Simple Unary Tag With Long Version And Base Exclusion",
			t:    Tuplip{ExcludeBase: true},
			args: args{[]string{"foo:2.0.0"}},
			want: []string{"foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name: "Wildcard Unary Tag With Long Version And Base Exclusion",
			t:    Tuplip{ExcludeBase: true},
			args: args{[]string{"_:2.0.0"}},
			want: []string{"2", "2.0", "2.0.0"},
		},
		{
			name: "Wildcard Unary Tag With Long Version And Latest Addition",
			t:    Tuplip{AddLatest: true},
			args: args{[]string{"_:2.0.0", "foo"}},
			want: []string{"latest", "foo", "2", "2.0", "2.0.0", "2-foo", "2.0-foo", "2.0.0-foo"},
		},
		{
			name: "Wildcard Unary Tag With Long Version And a Different Separator",
			t:    Tuplip{Separator: ";"},
			args: args{[]string{" _:2.0.0; foo "}},
			want: []string{"foo", "2", "2.0", "2.0.0", "2-foo", "2.0-foo", "2.0.0-foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := strings.NewReader(strings.Join(tt.args.input, " "))
			tuplipSrc := tt.t.FromReader(src)
			tStream := tuplipSrc.Build()
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
		t       Tuplip
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
			t:       Tuplip{Repository: "gofunky/git"},
			args:    args{[]string{"unknown"}},
			wantErr: true,
		},
		{
			name: "Simple Unary Tag",
			t:    Tuplip{Repository: "gofunky/git"},
			args: args{[]string{"envload"}},
			want: "envload",
		},
		{
			name: "Simple Binary Tag",
			t:    Tuplip{Repository: "gofunky/git"},
			args: args{[]string{"envload", "alpine:3.8"}},
			want: "alpine3.8-envload",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := strings.NewReader(strings.Join(tt.args.input, " "))
			tuplipSrc := tt.t.FromReader(src)
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

func TestTuplip_getTags(t *testing.T) {
	type fields struct {
		ExcludeMajor bool
		ExcludeMinor bool
		ExcludeBase  bool
		AddLatest    bool
		Separator    string
		Repository   string
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
			tu := &Tuplip{
				ExcludeMajor: tt.fields.ExcludeMajor,
				ExcludeMinor: tt.fields.ExcludeMinor,
				ExcludeBase:  tt.fields.ExcludeBase,
				AddLatest:    tt.fields.AddLatest,
				Separator:    tt.fields.Separator,
				Repository:   tt.fields.Repository,
			}
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
