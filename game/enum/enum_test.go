package enum

import (
	"reflect"
	"testing"
)

type testOption int

func (o testOption) String() string {
	return []string{"first", "second"}[o]
}

func TestGenericEnumGetOptions(t *testing.T) {
	tests := []struct {
		name string
		enum any
		want []string
	}{
		{
			name: "stringer values",
			enum: GenericEnum[testOption]{List: []testOption{0, 1}},
			want: []string{"first", "second"},
		},
		{
			name: "primitive values",
			enum: GenericEnum[int]{List: []int{10, 20}},
			want: []string{"10", "20"},
		},
		{
			name: "empty list",
			enum: GenericEnum[string]{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			switch enum := tt.enum.(type) {
			case GenericEnum[testOption]:
				got = enum.GetOptions()
			case GenericEnum[int]:
				got = enum.GetOptions()
			case GenericEnum[string]:
				got = enum.GetOptions()
			default:
				t.Fatalf("unsupported enum type %T", tt.enum)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenericEnumGetOptionsReturnsIndependentSlice(t *testing.T) {
	enum := GenericEnum[string]{List: []string{"first", "second"}}

	options := enum.GetOptions()
	options[0] = "changed"

	if enum.List[0] != "first" {
		t.Fatalf("GetOptions() result aliases List: List[0] = %q", enum.List[0])
	}
}
