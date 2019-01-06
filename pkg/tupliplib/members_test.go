package tupliplib

import (
	"github.com/blang/semver"
	"github.com/deckarep/golang-set"
	"reflect"
	"testing"
)

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
		{
			name:       "Split Tuple With Different Separator",
			t:          Tuplip{Separator: ","},
			args:       args{"foo, boo,hoo"},
			wantResult: []string{"foo", "boo", "hoo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.check()
			if gotResult := tt.t.splitBySeparator(tt.args.input); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Tuplip.splitBySeparator() = %v, want %v", gotResult, tt.wantResult)
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
