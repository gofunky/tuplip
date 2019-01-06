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
					t.Errorf("Tuplip.Open() error = %v, wantErr %v", gotErr, tt.wantErr)
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
				t.Errorf("Tuplip.Open() = %v, want %v, difference %v",
					gotOutput, wantSet, gotOutput.Difference(wantSet))
			}
		})
	}
}
