package navigation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseGraphicInfoIndexesMapIDProperties(t *testing.T) {
	properties := map[uint16]mapTileProperty{
		12497: {East: 1, South: 1, Passable: true},
		13637: {East: 1, South: 1, Passable: false},
		15080: {East: 2, South: 2, Passable: false},
	}
	index, err := parseGraphicInfo(bytes.NewReader(testGraphicInfoData(properties)))
	if err != nil {
		t.Fatal(err)
	}
	for mapID, want := range properties {
		got, ok := index.lookup(mapID)
		if !ok {
			t.Fatalf("lookup(%d) was not found", mapID)
		}
		want.Found = true
		if got != want {
			t.Fatalf("lookup(%d) = %#v, want %#v", mapID, got, want)
		}
	}
	if _, ok := index.lookup(999); ok {
		t.Fatal("unknown map ID was found")
	}
}

func TestParseGraphicInfoRejectsInvalidLength(t *testing.T) {
	if _, err := parseGraphicInfo(bytes.NewReader(make([]byte, graphicInfoRecordSize-1))); err == nil {
		t.Fatal("parseGraphicInfo() error = nil")
	}
}

func TestGraphicInfoCollisionUsesOnlyEachCellsPassability(t *testing.T) {
	data := testMapData(4, 3, nil)
	objectOffset := mapHeaderSize + 4*3*mapCellSize
	binary.LittleEndian.PutUint16(data[objectOffset+5*mapCellSize:], 15080)
	binary.LittleEndian.PutUint16(data[objectOffset+3*mapCellSize:], 12497)
	graphics := testGraphicInfoIndex(map[uint16]mapTileProperty{
		15080: {East: 2, South: 2, Passable: false},
		12497: {East: 1, South: 1, Passable: true},
	})

	parsed, err := parseMap(bytes.NewReader(data), graphics)
	if err != nil {
		t.Fatal(err)
	}
	want := []bool{
		true, true, true, true,
		true, false, true, true,
		true, true, true, true,
	}
	if !reflect.DeepEqual(parsed.Walkable, want) {
		t.Fatalf("walkable = %v, want %v", parsed.Walkable, want)
	}
}

func TestFindGraphicInfoFileUsesNewestBaseFile(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"GraphicInfo_9.bin", "GraphicInfo_66.bin", "GraphicInfo_Joy_99.bin"} {
		if err := os.WriteFile(filepath.Join(binDir, name), make([]byte, graphicInfoRecordSize), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := findGraphicInfoFile(root)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(binDir, "GraphicInfo_66.bin")
	if got != want {
		t.Fatalf("findGraphicInfoFile() = %q, want %q", got, want)
	}
}

func TestFindGraphicInfoFileRejectsUnavailableFolder(t *testing.T) {
	if _, err := findGraphicInfoFile(t.TempDir()); !errors.Is(err, errGraphicInfoUnavailable) {
		t.Fatalf("findGraphicInfoFile() error = %v, want %v", err, errGraphicInfoUnavailable)
	}
}

func TestSandboxHostGameDir(t *testing.T) {
	got := sandboxHostGameDir(`C:\Sandbox\user\box\user\current\Documents\CG\CG`, `C:\Users\user`)
	want := filepath.Join(`C:\Users\user`, "Documents", "CG", "CG")
	if !stringsEqualFoldPath(got, want) {
		t.Fatalf("sandboxHostGameDir() = %q, want %q", got, want)
	}
	if got := sandboxHostGameDir(`C:\Users\user\Documents\CG\CG`, `C:\Users\user`); got != "" {
		t.Fatalf("non-Sandboxie path mapped to %q", got)
	}
}

func stringsEqualFoldPath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

func writeTestGraphicInfo(t *testing.T, root string, properties map[uint16]mapTileProperty) {
	t.Helper()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "GraphicInfo_1.bin"), testGraphicInfoData(properties), 0o600); err != nil {
		t.Fatal(err)
	}
}

func testGraphicInfoIndex(properties map[uint16]mapTileProperty) *graphicInfoIndex {
	index, err := parseGraphicInfo(bytes.NewReader(testGraphicInfoData(properties)))
	if err != nil {
		panic(err)
	}
	return index
}

func testGraphicInfoData(properties map[uint16]mapTileProperty) []byte {
	data := make([]byte, max(1, len(properties))*graphicInfoRecordSize)
	record := 0
	for mapID, property := range properties {
		offset := record * graphicInfoRecordSize
		binary.LittleEndian.PutUint32(data[offset:], uint32(record))
		data[offset+graphicInfoEastOffset] = property.East
		data[offset+graphicInfoSouthOffset] = property.South
		if property.Passable {
			data[offset+graphicInfoFlagOffset] = 1
		}
		binary.LittleEndian.PutUint32(data[offset+graphicInfoMapIDOffset:], uint32(mapID))
		record++
	}
	return data
}
