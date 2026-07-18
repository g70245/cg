package items

import (
	"reflect"
	"testing"
)

func TestBombOptionsUseItemNames(t *testing.T) {
	want := []string{"7B", "8B", "9A"}
	if got := Bombs.GetOptions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Bombs.GetOptions() = %v, want %v", got, want)
	}
}

func TestFindBomb(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		want      Item
		wantFound bool
	}{
		{name: "finds current name", param: "7B", want: I_B_7B, wantFound: true},
		{name: "finds legacy struct parameter", param: "{8B 14614527}", want: I_B_8B, wantFound: true},
		{name: "rejects unknown parameter", param: "unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, found := FindBomb(test.param)
			if got != test.want || found != test.wantFound {
				t.Fatalf("FindBomb(%q) = (%v, %t), want (%v, %t)", test.param, got, found, test.want, test.wantFound)
			}
		})
	}
}
