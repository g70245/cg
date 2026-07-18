package navigation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestParseMap(t *testing.T) {
	data := testMapData(3, 2, map[int]uint16{
		0: 12000,
		2: 12002,
		4: 14676,
		5: 999,
	})

	got, err := ParseMap(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseMap() error = %v", err)
	}
	want := []Stair{
		{East: 0, South: 0, Type: StairUp},
		{East: 2, South: 0, Type: StairDown},
		{East: 1, South: 1, Type: StairPassage},
		{East: 2, South: 1, Type: StairUnknown},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseMap() = %#v, want %#v", got, want)
	}
}

func TestParseMapRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "short header", data: []byte("MAP"), want: errInvalidMapHeader},
		{name: "wrong signature", data: make([]byte, mapHeaderSize), want: errInvalidMapHeader},
		{name: "zero dimensions", data: testMapData(0, 0, nil), want: errInvalidMapDimensions},
		{name: "truncated sections", data: testMapData(2, 2, nil)[:mapHeaderSize+1], want: errIncompleteMapData},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseMap(bytes.NewReader(test.data))
			if !errors.Is(err, test.want) {
				t.Fatalf("ParseMap() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestBuildRoutesSortsByDistanceAndCoordinates(t *testing.T) {
	stairs := []Stair{
		{East: 5, South: 5, Type: StairDown},
		{East: 1, South: 0, Type: StairUp},
		{East: 0, South: 1, Type: StairPassage},
		{East: 0, South: 0, Type: StairUnknown},
	}

	got := BuildRoutes(stairs, 0, 0)
	want := []Route{
		{Stair: stairs[3], Direction: "Here", DistanceSquared: 0},
		{Stair: stairs[2], Direction: "⬇", DistanceSquared: 1},
		{Stair: stairs[1], Direction: "➡", DistanceSquared: 1},
		{Stair: stairs[0], Direction: "↘", DistanceSquared: 50},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildRoutes() = %#v, want %#v", got, want)
	}
}

func TestDirectionCoversAllSectors(t *testing.T) {
	tests := []struct {
		east, south int64
		want        string
	}{
		{east: 0, south: 1, want: "⬇"},
		{east: 1, south: 1, want: "↘"},
		{east: 1, south: 0, want: "➡"},
		{east: 1, south: -1, want: "↗"},
		{east: 0, south: -1, want: "⬆"},
		{east: -1, south: -1, want: "↖"},
		{east: -1, south: 0, want: "⬅"},
		{east: -1, south: 1, want: "↙"},
		{want: "Here"},
	}

	for _, test := range tests {
		if got := direction(test.east, test.south); got != test.want {
			t.Fatalf("direction(%d, %d) = %q, want %q", test.east, test.south, got, test.want)
		}
	}
}

func TestPathResolverFindsMapCodeAtSupportedDepth(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "map", "1", "1", "699.dat")
	if err := os.MkdirAll(filepath.Dir(want), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, testMapData(1, 1, nil), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(want), "ignore.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := new(PathResolver).Resolve(root, 699)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}

func TestPathResolverDoesNotRecursivelyScanDeepFolders(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "map", "a", "b", "c", "d", "699.dat")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, testMapData(1, 1, nil), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := new(PathResolver).Resolve(root, 699); err == nil {
		t.Fatal("Resolve() error = nil, want unsupported-depth error")
	}
}

func TestPathResolverRejectsUnavailableLocations(t *testing.T) {
	tests := []struct {
		name string
		root string
		code uint32
	}{
		{name: "empty root", code: 699},
		{name: "zero code", root: t.TempDir()},
		{name: "missing map folder", root: t.TempDir(), code: 699},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := new(PathResolver).Resolve(test.root, test.code); err == nil {
				t.Fatal("Resolve() error = nil, want error")
			}
		})
	}
}

func TestFileCacheReloadsChangedMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "map.dat")
	if err := os.WriteFile(path, testMapData(1, 1, map[int]uint16{0: 12000}), 0o600); err != nil {
		t.Fatal(err)
	}

	cache := new(FileCache)
	first, err := cache.Load(path)
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}
	if len(first) != 1 || first[0].Type != StairUp {
		t.Fatalf("first Load() = %#v, want one up route", first)
	}

	if err := os.WriteFile(path, testMapData(2, 1, map[int]uint16{1: 12002}), 0o600); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}

	second, err := cache.Load(path)
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}
	want := []Stair{{East: 1, South: 0, Type: StairDown}}
	if !reflect.DeepEqual(second, want) {
		t.Fatalf("second Load() = %#v, want %#v", second, want)
	}
}

func testMapData(width, height int32, stairs map[int]uint16) []byte {
	cellCount := int(width * height)
	if cellCount < 0 {
		cellCount = 0
	}
	data := make([]byte, mapHeaderSize+cellCount*mapCellSize*mapSectionCount)
	copy(data, "MAP")
	binary.LittleEndian.PutUint32(data[12:16], uint32(width))
	binary.LittleEndian.PutUint32(data[16:20], uint32(height))
	objectOffset := mapHeaderSize + cellCount*mapCellSize
	transitionOffset := objectOffset + cellCount*mapCellSize
	for index, objectID := range stairs {
		binary.LittleEndian.PutUint16(data[objectOffset+index*mapCellSize:], objectID)
		binary.LittleEndian.PutUint16(data[transitionOffset+index*mapCellSize:], stairTransitionCode)
	}
	return data
}
