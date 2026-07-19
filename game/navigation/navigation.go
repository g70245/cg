package navigation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	mapHeaderSize       = 20
	mapCellSize         = 2
	mapSectionCount     = 3
	stairTransitionCode = uint16(49155)
)

var (
	errInvalidMapHeader     = errors.New("invalid map header")
	errInvalidMapDimensions = errors.New("invalid map dimensions")
	errIncompleteMapData    = errors.New("incomplete map data")
)

type StairType string

const (
	StairUp      StairType = "Up"
	StairDown    StairType = "Down"
	StairPassage StairType = "Passage"
	StairUnknown StairType = "Unknown"
)

type Stair struct {
	East  int
	South int
	Type  StairType
}

type Point struct {
	East  int
	South int
}

type MapData struct {
	Width    int
	Height   int
	Known    []bool
	Walkable []bool
	Monsters []bool
	Objects  []uint16
	Stairs   []Stair
}

func (data MapData) IsKnown(point Point) bool {
	if point.East < 0 || point.South < 0 || point.East >= data.Width || point.South >= data.Height {
		return false
	}
	index := point.South*data.Width + point.East
	return index >= 0 && index < len(data.Known) && data.Known[index]
}

func (data MapData) IsWalkable(point Point) bool {
	if point.East < 0 || point.South < 0 || point.East >= data.Width || point.South >= data.Height {
		return false
	}
	index := point.South*data.Width + point.East
	return index >= 0 && index < len(data.Walkable) && data.Walkable[index]
}

type Route struct {
	Stair
	Direction       string
	DistanceSquared int64
}

func ParseMap(reader io.Reader) (MapData, error) {
	return parseMap(reader, nil)
}

func parseMap(reader io.Reader, graphics *graphicInfoIndex) (MapData, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return MapData{}, fmt.Errorf("read map data: %w", err)
	}
	if len(data) < mapHeaderSize || !bytes.Equal(data[:3], []byte("MAP")) {
		return MapData{}, errInvalidMapHeader
	}

	width := int64(int32(binary.LittleEndian.Uint32(data[12:16])))
	height := int64(int32(binary.LittleEndian.Uint32(data[16:20])))
	if width <= 0 || height <= 0 || width > math.MaxInt32/height {
		return MapData{}, errInvalidMapDimensions
	}

	cellCount := width * height
	if cellCount > (math.MaxInt64-mapHeaderSize)/(mapCellSize*mapSectionCount) {
		return MapData{}, errInvalidMapDimensions
	}
	sectionSize := cellCount * mapCellSize
	requiredSize := int64(mapHeaderSize) + sectionSize*mapSectionCount
	if requiredSize > int64(len(data)) {
		return MapData{}, errIncompleteMapData
	}

	objectOffset := int64(mapHeaderSize) + sectionSize
	transitionOffset := objectOffset + sectionSize
	result := MapData{
		Width:    int(width),
		Height:   int(height),
		Known:    make([]bool, int(cellCount)),
		Walkable: make([]bool, int(cellCount)),
		Monsters: make([]bool, int(cellCount)),
		Objects:  make([]uint16, int(cellCount)),
	}
	stairs := make([]Stair, 0)
	for index := int64(0); index < cellCount; index++ {
		transitionIndex := transitionOffset + index*mapCellSize
		transition := binary.LittleEndian.Uint16(data[transitionIndex : transitionIndex+mapCellSize])
		groundIndex := int64(mapHeaderSize) + index*mapCellSize
		objectIndex := objectOffset + index*mapCellSize
		groundID := binary.LittleEndian.Uint16(data[groundIndex : groundIndex+mapCellSize])
		objectID := binary.LittleEndian.Uint16(data[objectIndex : objectIndex+mapCellSize])
		result.Known[index] = groundID != 0 || objectID != 0 || transition != 0
		result.Walkable[index] = transition&0xff00 == 0xc000
		result.Monsters[index] = transition == 0xc002
		if graphics != nil {
			if property, ok := graphics.lookup(groundID); groundID != 0 && ok && !property.Passable {
				result.Walkable[index] = false
			}
			if property, ok := graphics.lookup(objectID); objectID != 0 && ok && !property.Passable {
				result.Walkable[index] = false
			}
		}
		result.Objects[index] = objectID
		if transition != stairTransitionCode {
			continue
		}

		stairs = append(stairs, Stair{
			East:  int(index % width),
			South: int(index / width),
			Type:  classifyStair(objectID),
		})
	}
	result.Stairs = stairs
	return result, nil
}

func classifyStair(objectID uint16) StairType {
	switch objectID {
	case 12000, 12001,
		13268, 13270, 13272, 13274,
		13996, 13998,
		15561,
		15887, 15889, 15891,
		16610, 16611, 16626, 16627, 16628, 16629,
		17952, 17954, 17956, 17958, 17960, 17962, 17964, 17966, 17968, 17970,
		17972, 17974, 17976, 17978, 17980, 17982, 17984, 17986, 17988, 17990,
		17992, 17994, 17996, 17998:
		return StairUp
	case 12002, 12003,
		13269, 13271, 13273, 13275,
		13997, 13999,
		15562,
		15888, 15890, 15892,
		16612, 16613, 16614, 16615,
		17953, 17955, 17957, 17959, 17961, 17963, 17965, 17967, 17969, 17971,
		17973, 17975, 17977, 17979, 17981, 17983, 17985, 17987, 17989, 17991,
		17993, 17995, 17997, 17999:
		return StairDown
	case 0, 14676:
		return StairPassage
	default:
		return StairUnknown
	}
}

func BuildRoutes(stairs []Stair, currentEast, currentSouth int) []Route {
	routes := make([]Route, 0, len(stairs))
	for _, stair := range stairs {
		deltaEast := int64(stair.East - currentEast)
		deltaSouth := int64(stair.South - currentSouth)
		routes = append(routes, Route{
			Stair:           stair,
			Direction:       direction(deltaEast, deltaSouth),
			DistanceSquared: deltaEast*deltaEast + deltaSouth*deltaSouth,
		})
	}

	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].DistanceSquared != routes[j].DistanceSquared {
			return routes[i].DistanceSquared < routes[j].DistanceSquared
		}
		if routes[i].East != routes[j].East {
			return routes[i].East < routes[j].East
		}
		if routes[i].South != routes[j].South {
			return routes[i].South < routes[j].South
		}
		return routes[i].Type < routes[j].Type
	})

	return routes
}

func direction(deltaEast, deltaSouth int64) string {
	if deltaEast == 0 && deltaSouth == 0 {
		return "Here"
	}

	angle := math.Atan2(float64(deltaSouth), float64(deltaEast)) * 180 / math.Pi
	sector := int(math.Floor((angle+22.5+360)/45)) % 8
	return [...]string{"➡", "↘", "⬇", "↙", "⬅", "↖", "⬆", "↗"}[sector]
}

type PathResolver struct {
	mu    sync.Mutex
	root  string
	paths map[uint32]string
}

func (resolver *PathResolver) Resolve(gameDir string, code uint32) (string, error) {
	if strings.TrimSpace(gameDir) == "" || code == 0 {
		return "", errors.New("map location is unavailable")
	}
	root, err := filepath.Abs(gameDir)
	if err != nil {
		return "", fmt.Errorf("resolve game folder: %w", err)
	}
	mapRoot := filepath.Join(root, "map")

	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	if resolver.root != root {
		resolver.root = root
		resolver.paths = make(map[uint32]string)
	}
	if path := resolver.paths[code]; isRegularFile(path) {
		return path, nil
	}
	path, err := findMapFileAtCommonDepth(mapRoot, code)
	if err != nil {
		return "", err
	}
	if path != "" {
		if resolver.paths == nil {
			resolver.paths = make(map[uint32]string)
		}
		resolver.paths[code] = path
		return path, nil
	}
	return "", errors.New("map data is unavailable")
}

func findMapFileAtCommonDepth(mapRoot string, code uint32) (string, error) {
	name := strconv.FormatUint(uint64(code), 10) + ".dat"
	patterns := []string{
		filepath.Join(mapRoot, name),
		filepath.Join(mapRoot, "*", name),
		filepath.Join(mapRoot, "*", "*", name),
		filepath.Join(mapRoot, "*", "*", "*", name),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return "", fmt.Errorf("match map data: %w", err)
		}
		for _, match := range matches {
			if isRegularFile(match) {
				return match, nil
			}
		}
	}
	return "", nil
}

func isRegularFile(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func IsMazePath(gameDir, path string) bool {
	mapRoot, err := filepath.Abs(filepath.Join(gameDir, "map"))
	if err != nil {
		return false
	}
	resolvedPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(mapRoot, resolvedPath)
	if err != nil || relative == "." || filepath.IsAbs(relative) {
		return false
	}
	parts := strings.Split(filepath.Clean(relative), string(filepath.Separator))
	return len(parts) > 1 && parts[0] == "1"
}

type FileCache struct {
	mu             sync.Mutex
	path           string
	size           int64
	modified       time.Time
	data           MapData
	graphics       graphicInfoCache
	graphicVersion string
}

func (cache *FileCache) Invalidate() {
	cache.mu.Lock()
	cache.path = ""
	cache.size = 0
	cache.modified = time.Time{}
	cache.data = MapData{}
	cache.graphicVersion = ""
	cache.mu.Unlock()
}

func (cache *FileCache) Load(gameDir, path string) (MapData, error) {
	graphics, graphicVersion, err := cache.graphics.Load(gameDir)
	if err != nil {
		return MapData{}, err
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()

	info, err := os.Stat(path)
	if err != nil {
		return MapData{}, fmt.Errorf("inspect map data: %w", err)
	}
	if !info.Mode().IsRegular() {
		return MapData{}, errors.New("map data is not a regular file")
	}
	if cache.path == path && cache.size == info.Size() && cache.modified.Equal(info.ModTime()) && cache.graphicVersion == graphicVersion {
		return cloneMapData(cache.data), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return MapData{}, fmt.Errorf("open map data: %w", err)
	}
	data, parseErr := parseMap(file, graphics)
	closeErr := file.Close()
	if parseErr != nil {
		return MapData{}, parseErr
	}
	if closeErr != nil {
		return MapData{}, fmt.Errorf("close map data: %w", closeErr)
	}

	cache.path = path
	cache.size = info.Size()
	cache.modified = info.ModTime()
	cache.graphicVersion = graphicVersion
	cache.data = cloneMapData(data)
	return cloneMapData(data), nil
}

func (cache *FileCache) Reset() {
	cache.mu.Lock()
	cache.path = ""
	cache.size = 0
	cache.modified = time.Time{}
	cache.data = MapData{}
	cache.graphicVersion = ""
	cache.mu.Unlock()
}

func cloneMapData(data MapData) MapData {
	data.Known = append([]bool(nil), data.Known...)
	data.Walkable = append([]bool(nil), data.Walkable...)
	data.Monsters = append([]bool(nil), data.Monsters...)
	data.Objects = append([]uint16(nil), data.Objects...)
	data.Stairs = append([]Stair(nil), data.Stairs...)
	return data
}
