package battle

import (
	"reflect"
	"testing"
)

func TestMPRatioOptions(t *testing.T) {
	want := []string{"0.05", "0.1", "0.15", "0.2", "0.25", "0.3", "0.35", "0.4", "0.45", "0.5", "0.55"}
	if got := MPRatios.GetOptions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("MPRatios.GetOptions() = %v, want %v", got, want)
	}
}
