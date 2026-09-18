package game

import "testing"

func TestPetBaseAddress(t *testing.T) {
	for index := 0; index < PET_SLOT_COUNT; index++ {
		got, err := petBaseAddress(index)
		if err != nil {
			t.Fatalf("petBaseAddress(%d) error = %v", index, err)
		}
		want := uint32(MEMORY_PET_BASE + index*MEMORY_PET_STRIDE)
		if got != want {
			t.Fatalf("petBaseAddress(%d) = %#x, want %#x", index, got, want)
		}
	}
}

func TestPetBaseAddressRejectsInvalidIndex(t *testing.T) {
	for _, index := range []int{-1, PET_SLOT_COUNT} {
		if _, err := petBaseAddress(index); err == nil {
			t.Fatalf("petBaseAddress(%d) error = nil, want an error", index)
		}
	}
}
