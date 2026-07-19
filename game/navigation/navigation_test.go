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
	if !reflect.DeepEqual(got.Stairs, want) {
		t.Fatalf("ParseMap() stairs = %#v, want %#v", got.Stairs, want)
	}
	if got.Width != 3 || got.Height != 2 {
		t.Fatalf("ParseMap() dimensions = %dx%d, want 3x2", got.Width, got.Height)
	}
	for index, walkable := range got.Walkable {
		if !walkable {
			t.Fatalf("ParseMap() cell %d is blocked, want walkable", index)
		}
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
	root := t.TempDir()
	path := filepath.Join(root, "map.dat")
	writeTestGraphicInfo(t, root, map[uint16]mapTileProperty{
		12000: {East: 1, South: 1, Passable: true},
		12002: {East: 1, South: 1, Passable: true},
	})
	if err := os.WriteFile(path, testMapData(1, 1, map[int]uint16{0: 12000}), 0o600); err != nil {
		t.Fatal(err)
	}

	cache := new(FileCache)
	first, err := cache.Load(root, path)
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}
	if len(first.Stairs) != 1 || first.Stairs[0].Type != StairUp {
		t.Fatalf("first Load() = %#v, want one up route", first)
	}

	if err := os.WriteFile(path, testMapData(2, 1, map[int]uint16{1: 12002}), 0o600); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}

	second, err := cache.Load(root, path)
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}
	want := []Stair{{East: 1, South: 0, Type: StairDown}}
	if !reflect.DeepEqual(second.Stairs, want) {
		t.Fatalf("second Load() stairs = %#v, want %#v", second.Stairs, want)
	}
}

func TestFileCacheInvalidateReloadsSameSizeAndTimestamp(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "map.dat")
	writeTestGraphicInfo(t, root, map[uint16]mapTileProperty{
		12000: {East: 1, South: 1, Passable: true},
		12002: {East: 1, South: 1, Passable: true},
	})
	if err := os.WriteFile(path, testMapData(1, 1, map[int]uint16{0: 12000}), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	cache := new(FileCache)
	if _, err := cache.Load(root, path); err != nil {
		t.Fatalf("first Load() error = %v", err)
	}
	if err := os.WriteFile(path, testMapData(1, 1, map[int]uint16{0: 12002}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}

	stale, err := cache.Load(root, path)
	if err != nil {
		t.Fatalf("stale Load() error = %v", err)
	}
	if len(stale.Stairs) != 1 || stale.Stairs[0].Type != StairUp {
		t.Fatalf("stale Load() = %#v, want cached Up", stale.Stairs)
	}

	cache.Invalidate()
	fresh, err := cache.Load(root, path)
	if err != nil {
		t.Fatalf("fresh Load() error = %v", err)
	}
	if len(fresh.Stairs) != 1 || fresh.Stairs[0].Type != StairDown {
		t.Fatalf("fresh Load() = %#v, want reloaded Down", fresh.Stairs)
	}
}

func TestParseMapUsesTransitionHighByteForWalkability(t *testing.T) {
	data := testMapData(5, 1, nil)
	transitionOffset := mapHeaderSize + 5*mapCellSize*2
	values := []uint16{0, 0xc000, 0xc002, 0xc00a, 0xc100}
	for index, value := range values {
		binary.LittleEndian.PutUint16(data[transitionOffset+index*mapCellSize:], value)
	}

	parsed, err := ParseMap(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	want := []bool{false, true, true, true, false}
	if !reflect.DeepEqual(parsed.Walkable, want) {
		t.Fatalf("walkable = %v, want %v", parsed.Walkable, want)
	}
	wantKnown := []bool{false, true, true, true, true}
	if !reflect.DeepEqual(parsed.Known, wantKnown) {
		t.Fatalf("known = %v, want %v", parsed.Known, wantKnown)
	}
	wantMonsters := []bool{false, false, true, false, false}
	if !reflect.DeepEqual(parsed.Monsters, wantMonsters) {
		t.Fatalf("monsters = %v, want %v", parsed.Monsters, wantMonsters)
	}
}

func TestParseMapDistinguishesKnownBlockedCellsFromFog(t *testing.T) {
	data := testMapData(2, 1, nil)
	groundOffset := mapHeaderSize
	transitionOffset := mapHeaderSize + 2*mapCellSize*2
	binary.LittleEndian.PutUint16(data[groundOffset:], 123)
	binary.LittleEndian.PutUint16(data[transitionOffset:], 0)
	binary.LittleEndian.PutUint16(data[transitionOffset+mapCellSize:], 0)

	parsed, err := ParseMap(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Known[0] || parsed.Walkable[0] {
		t.Fatalf("known blocked cell = known %v, walkable %v", parsed.Known[0], parsed.Walkable[0])
	}
	if parsed.Known[1] || parsed.Walkable[1] {
		t.Fatalf("fog cell = known %v, walkable %v", parsed.Known[1], parsed.Walkable[1])
	}
}

func TestParseMapCombinesTransitionAndKnownWallObjectsForCollision(t *testing.T) {
	data := testMapData(4, 1, nil)
	objectOffset := mapHeaderSize + 4*mapCellSize
	transitionOffset := objectOffset + 4*mapCellSize
	binary.LittleEndian.PutUint16(data[objectOffset:], 13627)
	binary.LittleEndian.PutUint16(data[objectOffset+mapCellSize:], 15070)
	binary.LittleEndian.PutUint16(data[objectOffset+2*mapCellSize:], 12000)
	binary.LittleEndian.PutUint16(data[transitionOffset+mapCellSize:], 0xc100)
	binary.LittleEndian.PutUint16(data[transitionOffset+2*mapCellSize:], stairTransitionCode)

	graphics := testGraphicInfoIndex(map[uint16]mapTileProperty{
		13627: {East: 1, South: 1, Passable: false},
		15070: {East: 1, South: 1, Passable: false},
		12000: {East: 1, South: 1, Passable: true},
	})
	parsed, err := parseMap(bytes.NewReader(data), graphics)
	if err != nil {
		t.Fatal(err)
	}
	want := []bool{false, false, true, true}
	if !reflect.DeepEqual(parsed.Walkable, want) {
		t.Fatalf("walkable = %v, want combined collision flags %v", parsed.Walkable, want)
	}
}

func TestParsedObjectWallIsNotCrossedByPathfinding(t *testing.T) {
	const (
		width  = 5
		height = 3
	)
	data := testMapData(width, height, map[int]uint16{4: 12000})
	objectOffset := mapHeaderSize + width*height*mapCellSize
	for east := 1; east < width; east++ {
		binary.LittleEndian.PutUint16(data[objectOffset+(width+east)*mapCellSize:], 13637)
	}

	graphics := testGraphicInfoIndex(map[uint16]mapTileProperty{
		12000: {East: 1, South: 1, Passable: true},
		13637: {East: 1, South: 1, Passable: false},
	})
	parsed, err := parseMap(bytes.NewReader(data), graphics)
	if err != nil {
		t.Fatal(err)
	}
	path, _, ok := FindPath(parsed, Point{East: 4, South: 2}, StairUp, nil)
	if !ok {
		t.Fatal("FindPath() found no route around the object wall")
	}
	usedOpening := false
	for _, point := range path {
		if point.South == 1 && point.East > 0 {
			t.Fatalf("FindPath() crossed object wall: %v", path)
		}
		if point == (Point{East: 0, South: 1}) {
			usedOpening = true
		}
	}
	if !usedOpening {
		t.Fatalf("FindPath() did not route through the wall opening: %v", path)
	}
}

func TestParsedTransitionWallIsNotCrossedByPathfinding(t *testing.T) {
	data := testMapData(3, 2, map[int]uint16{2: 12000})
	transitionOffset := mapHeaderSize + 6*mapCellSize*2
	binary.LittleEndian.PutUint16(data[transitionOffset+mapCellSize:], 0xc100)

	parsed, err := ParseMap(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	path, _, ok := FindPath(parsed, Point{}, StairUp, nil)
	if !ok {
		t.Fatal("FindPath() found no route around the transition wall")
	}
	for _, point := range path {
		if point == (Point{East: 1}) {
			t.Fatalf("FindPath() crossed transition wall: %v", path)
		}
	}
}

func TestIsMazePath(t *testing.T) {
	root := t.TempDir()
	if !IsMazePath(root, filepath.Join(root, "map", "1", "1", "699.dat")) {
		t.Fatal("map\\1 path was not recognized as a maze")
	}
	if IsMazePath(root, filepath.Join(root, "map", "0", "1401.dat")) {
		t.Fatal("map\\0 path was recognized as a maze")
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
	for index := 0; index < cellCount; index++ {
		binary.LittleEndian.PutUint16(data[transitionOffset+index*mapCellSize:], 0xc000)
	}
	for index, objectID := range stairs {
		binary.LittleEndian.PutUint16(data[objectOffset+index*mapCellSize:], objectID)
		binary.LittleEndian.PutUint16(data[transitionOffset+index*mapCellSize:], stairTransitionCode)
	}
	return data
}
