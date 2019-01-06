package tupliplib

import (
	"github.com/deckarep/golang-set"
	"reflect"
	"testing"
)

func TestTuplip_nonEmpty(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
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
			if got := nonEmpty(tt.args.input); got != tt.want {
				t.Errorf("Tuplip.nonEmpty() = %v, want %v", got, tt.want)
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
			if gotResult := packInSet(tt.args.subSet); !reflect.DeepEqual(gotResult.ToSlice(),
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
			if gotResult := mergeSets(tt.args.left, tt.args.right); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Tuplip.mergeSets() = %v, want %v", gotResult, tt.wantResult)
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
			got := power(tt.args.inputSet).ToSlice()
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
			got, err := failOnEmpty(tt.args.inputSet)
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

func Test_removeCommon(t *testing.T) {
	type args struct {
		seed map[string]mapset.Set
		next mapset.Set
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]mapset.Set
		wantErr bool
	}{
		{
			name: "Unary seed",
			args: args{
				seed: map[string]mapset.Set{
					"tag": mapset.NewSet("tag"),
				},
				next: mapset.NewSet("tag"),
			},
			want: map[string]mapset.Set{
				"tag": mapset.NewSet(),
			},
		},
		{
			name: "No matching set elements",
			args: args{
				seed: map[string]mapset.Set{
					"tag": mapset.NewSet("tag"),
				},
				next: mapset.NewSet("unknown"),
			},
			wantErr: true,
		},
		{
			name: "Binary seed",
			args: args{
				seed: map[string]mapset.Set{
					"tag":        mapset.NewSet("tag"),
					"second-tag": mapset.NewSet("tag", "second"),
				},
				next: mapset.NewSet("tag"),
			},
			want: map[string]mapset.Set{
				"tag":        mapset.NewSet(),
				"second-tag": mapset.NewSet("second"),
			},
		},
		{
			name: "Complex seed",
			args: args{
				seed: map[string]mapset.Set{
					"empty":              mapset.NewSet(),
					"tag":                mapset.NewSet("tag"),
					"second-tag":         mapset.NewSet("tag", "second"),
					"1.2.3-alias-tag1.2": mapset.NewSet("tag1.2", "alias", "1.2.3"),
				},
				next: mapset.NewSet("tag", "tag1", "tag1.2", "tag1.2.3"),
			},
			want: map[string]mapset.Set{
				"tag":                mapset.NewSet(),
				"second-tag":         mapset.NewSet("second"),
				"1.2.3-alias-tag1.2": mapset.NewSet("alias", "1.2.3"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := removeCommon(tt.args.seed, tt.args.next)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tuplip.failOnEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeCommon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyForSmallest(t *testing.T) {
	type args struct {
		seed map[string]mapset.Set
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "Empty Input",
			args: args{seed: map[string]mapset.Set{}},
		},
		{
			name: "Unary Input",
			args: args{seed: map[string]mapset.Set{
				"tag": mapset.NewSet(),
			}},
			wantResult: "tag",
		},
		{
			name: "Binary Input",
			args: args{seed: map[string]mapset.Set{
				"tag":        mapset.NewSet(),
				"second-tag": mapset.NewSet("second"),
			}},
			wantResult: "tag",
		},
		{
			name: "Complex Input",
			args: args{seed: map[string]mapset.Set{
				"tag":                mapset.NewSet(),
				"second-tag":         mapset.NewSet("second"),
				"1.2.3-alias-tag1.2": mapset.NewSet("alias", "1.2.3"),
			}},
			wantResult: "tag",
		},
		{
			name: "Duplicates Input",
			args: args{seed: map[string]mapset.Set{
				"first-tag":          mapset.NewSet(),
				"second-tag":         mapset.NewSet("second"),
				"1.2.3-alias-tag1.2": mapset.NewSet(),
			}},
			wantResult: "1.2.3-alias-tag1.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := keyForSmallest(tt.args.seed); gotResult != tt.wantResult {
				t.Errorf("keyForSmallest() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_minVal(t *testing.T) {
	type args struct {
		numbers map[string]mapset.Set
	}
	tests := []struct {
		name          string
		args          args
		wantMinNumber int
	}{
		{
			name: "Sample Input With Duplicates",
			args: args{numbers: map[string]mapset.Set{
				"foo":   mapset.NewSet("a", "b"),
				"boo":   mapset.NewSet("c", "d"),
				"other": mapset.NewSet("a", "b", "c"),
			}},
			wantMinNumber: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMinNumber := minVal(tt.args.numbers); gotMinNumber != tt.wantMinNumber {
				t.Errorf("minVal() = %v, want %v", gotMinNumber, tt.wantMinNumber)
			}
		})
	}
}

func Test_mostSeparators(t *testing.T) {
	type args struct {
		values []string
		sep    string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "Sample Input With Duplicates",
			args: args{
				values: []string{
					"a;b",
					"d;b;c",
					"a;b;c",
					"x",
				},
				sep: ";",
			},
			wantResult: "a;b;c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := mostSeparators(tt.args.values, tt.args.sep); gotResult != tt.wantResult {
				t.Errorf("mostSeparators() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
