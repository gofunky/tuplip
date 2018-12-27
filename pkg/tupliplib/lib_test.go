package tupliplib

import (
	"github.com/deckarep/golang-set"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/gofunky/automi/collectors"
)

func TestTuplipStream_FromReader(t *testing.T) {
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
			args: args{[]string{"latest"}},
			want: []string{"latest"},
		},
		{
			name: "Simple Binary Tag",
			args: args{[]string{"latest", "foo"}},
			want: []string{"latest", "foo", "foo-latest"},
		},
		{
			name: "Simple Tertiary Tag",
			args: args{[]string{"latest", "foo", "boo"}},
			want: []string{"latest", "foo", "foo-latest", "boo", "boo-latest", "boo-foo", "boo-foo-latest"},
		},
		{
			name: "Simple Binary Tag With Short Version",
			args: args{[]string{"latest", "foo:2.0"}},
			want: []string{"latest", "foo", "foo2", "foo2.0", "foo-latest", "foo2-latest", "foo2.0-latest"},
		},
		{
			name: "Simple Unary Tag With Long Version",
			args: args{[]string{"foo:2.0.0"}},
			want: []string{"foo", "foo2", "foo2.0", "foo2.0.0"},
		},
		{
			name: "Wildcard Binary Tag With Long Version",
			args: args{[]string{"_:2.0.0", "latest"}},
			want: []string{"latest", "2", "2.0", "2.0.0", "2-latest", "2.0-latest", "2.0.0-latest"},
		},
		{
			name: "Wildcard Binary Tag With Long Version And Major Exclusion",
			t:    Tuplip{ExcludeMajor: true},
			args: args{[]string{"_:2.0.0", "latest"}},
			want: []string{"latest", "2.0", "2.0.0", "2.0-latest", "2.0.0-latest"},
		},
		{
			name: "Wildcard Binary Tag With Long Version And Minor Exclusion",
			t:    Tuplip{ExcludeMinor: true},
			args: args{[]string{"_:2.0.0", "latest"}},
			want: []string{"latest", "2", "2.0.0", "2-latest", "2.0.0-latest"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := strings.NewReader(strings.Join(tt.args.input, " "))
			tStream := tt.t.FromReader(src)
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

func TestTuplip_buildTag(t *testing.T) {
	type args struct {
		withBase      bool
		alias         string
		versionDigits []uint64
	}
	tests := []struct {
		name    string
		t       Tuplip
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "With Base And 3 Digits",
			args: args{
				withBase:      true,
				alias:         "alias",
				versionDigits: []uint64{1, 0, 0},
			},
			want: "alias1.0.0",
		},
		{
			name: "With Base And 2 Digits",
			args: args{
				withBase:      true,
				alias:         "alias",
				versionDigits: []uint64{1, 0},
			},
			want: "alias1.0",
		},
		{
			name: "With Base And 1 Digit",
			args: args{
				withBase:      true,
				alias:         "alias",
				versionDigits: []uint64{1},
			},
			want: "alias1",
		},
		{
			name: "Without Base And 3 Digits",
			t:    Tuplip{},
			args: args{
				versionDigits: []uint64{2, 0, 0},
			},
			want: "2.0.0",
		},
		{
			name: "Without Base And 2 Digits",
			t:    Tuplip{},
			args: args{
				versionDigits: []uint64{2, 0},
			},
			want: "2.0",
		},
		{
			name: "Without Base And 1 Digit",
			t:    Tuplip{},
			args: args{
				versionDigits: []uint64{2},
			},
			want: "2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.t.buildTag(tt.args.withBase, tt.args.alias, tt.args.versionDigits...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tuplip.buildTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Tuplip.buildTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTuplip_parseVersions(t *testing.T) {
	type args struct {
		withBase bool
		alias    string
		isShort  bool
		version  semver.Version
	}
	tests := []struct {
		name       string
		t          Tuplip
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "Short With Base",
			args: args{
				withBase: true,
				isShort:  true,
				alias:    "alias",
				version:  semver.Version{Minor: 1},
			},
			wantResult: []string{"alias", "alias1", "alias1.0"},
		},
		{
			name: "Long With Base",
			args: args{
				withBase: true,
				alias:    "alias",
				version:  semver.Version{Major: 1},
			},
			wantResult: []string{"alias", "alias1", "alias1.0", "alias1.0.0"},
		},
		{
			name: "Short Without Base",
			args: args{
				withBase: false,
				isShort:  true,
				version:  semver.Version{Minor: 1},
			},
			wantResult: []string{"1", "1.0"},
		},
		{
			name: "Long Without Base",
			args: args{
				withBase: false,
				version:  semver.Version{Major: 1},
			},
			wantResult: []string{"1", "1.0", "1.0.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.t.buildVersionSet(tt.args.withBase, tt.args.alias, tt.args.isShort, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tuplip.buildVersionSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantSet := mapset.NewSet()
			for _, e := range tt.wantResult {
				wantSet.Add(e)
			}
			if !reflect.DeepEqual(gotResult, wantSet) {
				t.Errorf("Tuplip.buildVersionSet() = %v, want %v", gotResult, wantSet)
			}
		})
	}
}

func TestTuplip_splitVersion(t *testing.T) {
	type args struct {
		inputTag string
	}
	tests := []struct {
		name       string
		t          Tuplip
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "Short With Base",
			args: args{
				inputTag: "alias:1.0",
			},
			wantResult: []string{"alias", "alias1", "alias1.0"},
		},
		{
			name: "Long With Base",
			args: args{
				inputTag: "alias:1.0.0",
			},
			wantResult: []string{"alias", "alias1", "alias1.0", "alias1.0.0"},
		},
		{
			name: "Short Without Base",
			args: args{
				inputTag: "_:1.0",
			},
			wantResult: []string{"1", "1.0"},
		},
		{
			name: "Long Without Base",
			args: args{
				inputTag: "_:1.0.0",
			},
			wantResult: []string{"1", "1.0", "1.0.0"},
		},
		{
			name: "Invalid Version",
			args: args{
				inputTag: "_:invalid.stuff",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.t.splitVersion(tt.args.inputTag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tuplip.splitVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				wantSet := mapset.NewSet()
				for _, e := range tt.wantResult {
					wantSet.Add(e)
				}
				if !reflect.DeepEqual(gotResult, wantSet) {
					t.Errorf("Tuplip.splitVersion() = %v, want %v", gotResult, wantSet)
				}
			}
		})
	}
}

func TestTuplip_nonEmpty(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		t    Tuplip
		args args
		want bool
	}{
		{
			name: "Empty String",
			args: args{""},
			want: false,
		},
		{
			name: "Nonempty String",
			args: args{"foo"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.nonEmpty(tt.args.input); got != tt.want {
				t.Errorf("Tuplip.nonEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTuplip_splitBySeparator(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name       string
		t          Tuplip
		args       args
		wantResult []string
	}{
		{
			name:       "Empty Split",
			args:       args{""},
			wantResult: []string{""},
		},
		{
			name:       "Unary Split",
			args:       args{"foo"},
			wantResult: []string{"foo"},
		},
		{
			name:       "Split Tuple",
			args:       args{"foo boo hoo"},
			wantResult: []string{"foo", "boo", "hoo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.t.splitBySeparator(tt.args.input); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Tuplip.splitBySeparator() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestTuplip_packInSet(t *testing.T) {
	type args struct {
		subSet mapset.Set
	}
	tests := []struct {
		name       string
		t          Tuplip
		args       args
		wantResult mapset.Set
	}{
		{
			name:       "Empty Set",
			args:       args{mapset.NewSet()},
			wantResult: mapset.NewSetWith(mapset.NewSet()),
		},
		{
			name:       "Unary Set",
			args:       args{mapset.NewSet("foo")},
			wantResult: mapset.NewSetWith(mapset.NewSet("foo")),
		},
		{
			name:       "Tuple Set",
			args:       args{mapset.NewSet("foo", "boo")},
			wantResult: mapset.NewSetWith(mapset.NewSet("foo", "boo")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.t.packInSet(tt.args.subSet); !reflect.DeepEqual(gotResult.ToSlice(),
				tt.wantResult.ToSlice()) {
				t.Errorf("Tuplip.packInSet() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestTuplip_mergeSets(t *testing.T) {
	type args struct {
		left  mapset.Set
		right mapset.Set
	}
	tests := []struct {
		name       string
		t          Tuplip
		args       args
		wantResult mapset.Set
	}{
		{
			name:       "Merge Empty Sets",
			args:       args{mapset.NewSet(), mapset.NewSet()},
			wantResult: mapset.NewSet(),
		},
		{
			name:       "Merge Empty Set With Nonempty Set",
			args:       args{mapset.NewSet(), mapset.NewSet("foo")},
			wantResult: mapset.NewSet("foo"),
		},
		{
			name:       "Merge Two Nonempty Sets",
			args:       args{mapset.NewSet("foo", "boo"), mapset.NewSet("hoo")},
			wantResult: mapset.NewSet("foo", "boo", "hoo"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.t.mergeSets(tt.args.left, tt.args.right); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Tuplip.mergeSets() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestTuplip_join(t *testing.T) {
	type args struct {
		inputSet mapset.Set
	}
	tests := []struct {
		name       string
		t          Tuplip
		args       args
		wantResult mapset.Set
	}{
		{
			name:       "Empty Set",
			args:       args{mapset.NewSet()},
			wantResult: mapset.NewSet(),
		},
		{
			name:       "Unary Set",
			args:       args{mapset.NewSet(mapset.NewSet("alias"))},
			wantResult: mapset.NewSet("alias"),
		},
		{
			name:       "Binary Set",
			args:       args{mapset.NewSet(mapset.NewSet("alias"), mapset.NewSet("foo"))},
			wantResult: mapset.NewSet("alias-foo"),
		},
		{
			name:       "Cartesian Product Check",
			args:       args{mapset.NewSet(mapset.NewSet("alias", "alias2"), mapset.NewSet("foo", "boo"))},
			wantResult: mapset.NewSet("alias-foo", "alias-boo", "alias2-foo", "alias2-boo"),
		},
		{
			name:       "Cartesian Product Check With Base Version",
			args:       args{mapset.NewSet(mapset.NewSet("1.0", "1.0.0"), mapset.NewSet("foo", "boo"))},
			wantResult: mapset.NewSet("1.0-foo", "1.0-boo", "1.0.0-foo", "1.0.0-boo"),
		},
		{
			name:       "Tertiary Set",
			args:       args{mapset.NewSet(mapset.NewSet("alias"), mapset.NewSet("foo"), mapset.NewSet("boo"))},
			wantResult: mapset.NewSet("alias-boo-foo"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.t.join(tt.args.inputSet); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Tuplip.join() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestTuplip_power(t *testing.T) {
	type args struct {
		inputSet mapset.Set
	}
	tests := []struct {
		name string
		t    Tuplip
		args args
		want []mapset.Set
	}{
		{
			name: "Empty Set",
			args: args{mapset.NewSet()},
			want: []mapset.Set{
				mapset.NewSet(),
			},
		},
		{
			name: "Unary Set",
			args: args{mapset.NewSet("alias")},
			want: []mapset.Set{
				mapset.NewSet(),
				mapset.NewSet("alias"),
			},
		},
		{
			name: "Binary Set",
			args: args{mapset.NewSet("alias", "foo")},
			want: []mapset.Set{
				mapset.NewSet(),
				mapset.NewSet("alias"),
				mapset.NewSet("foo"),
				mapset.NewSet("alias", "foo"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.power(tt.args.inputSet).ToSlice()
			for _, tagSet := range tt.want {
				var found bool
				for _, val := range got {
					if tagSet.Equal(val.(mapset.Set)) {
						found = true
					}
				}
				if !found {
					t.Errorf("Tuplip.power() = %v, want %v, missing %v", got, tt.want, tagSet)
				}
			}
		})
	}
}

func TestTuplip_failOnEmpty(t *testing.T) {
	nonemptySet := mapset.NewSet(mapset.NewSet(), mapset.NewSet("alias"))
	type args struct {
		inputSet mapset.Set
	}
	tests := []struct {
		name    string
		t       Tuplip
		args    args
		want    mapset.Set
		wantErr bool
	}{
		{
			name:    "Empty Set",
			args:    args{mapset.NewSet()},
			wantErr: true,
		},
		{
			name:    "Empty Power Set",
			args:    args{mapset.NewSet(mapset.NewSet())},
			wantErr: true,
		},
		{
			name: "Nonempty Power Set",
			args: args{nonemptySet},
			want: nonemptySet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.t.failOnEmpty(tt.args.inputSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tuplip.failOnEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tuplip.failOnEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
