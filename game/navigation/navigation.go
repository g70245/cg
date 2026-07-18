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

type Route struct {
	Stair
	Direction       string
	DistanceSquared int64
}

func ParseMap(reader io.Reader) ([]Stair, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read map data: %w", err)
	}
	if len(data) < mapHeaderSize || !bytes.Equal(data[:3], []byte("MAP")) {
		return nil, errInvalidMapHeader
	}

	width := int64(int32(binary.LittleEndian.Uint32(data[12:16])))
	height := int64(int32(binary.LittleEndian.Uint32(data[16:20])))
	if width <= 0 || height <= 0 || width > math.MaxInt32/height {
		return nil, errInvalidMapDimensions
	}

	cellCount := width * height
	if cellCount > (math.MaxInt64-mapHeaderSize)/(mapCellSize*mapSectionCount) {
		return nil, errInvalidMapDimensions
	}
	sectionSize := cellCount * mapCellSize
	requiredSize := int64(mapHeaderSize) + sectionSize*mapSectionCount
	if requiredSize > int64(len(data)) {
		return nil, errIncompleteMapData
	}

	objectOffset := int64(mapHeaderSize) + sectionSize
	transitionOffset := objectOffset + sectionSize
	stairs := make([]Stair, 0)
	for index := int64(0); index < cellCount; index++ {
		transitionIndex := transitionOffset + index*mapCellSize
		transition := binary.LittleEndian.Uint16(data[transitionIndex : transitionIndex+mapCellSize])
		if transition != stairTransitionCode {
			continue
		}

		objectIndex := objectOffset + index*mapCellSize
		objectID := binary.LittleEndian.Uint16(data[objectIndex : objectIndex+mapCellSize])
		stairs = append(stairs, Stair{
			East:  int(index % width),
			South: int(index / width),
			Type:  classifyStair(objectID),
		})
	}

	return stairs, nil
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

type FileCache struct {
	mu       sync.Mutex
	path     string
	size     int64
	modified time.Time
	stairs   []Stair
}

func (cache *FileCache) Load(path string) ([]Stair, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect map data: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("map data is not a regular file")
	}
	if cache.path == path && cache.size == info.Size() && cache.modified.Equal(info.ModTime()) {
		return append([]Stair(nil), cache.stairs...), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open map data: %w", err)
	}
	stairs, parseErr := ParseMap(file)
	closeErr := file.Close()
	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close map data: %w", closeErr)
	}

	cache.path = path
	cache.size = info.Size()
	cache.modified = info.ModTime()
	cache.stairs = append([]Stair(nil), stairs...)
	return append([]Stair(nil), stairs...), nil
}

func (cache *FileCache) Reset() {
	cache.mu.Lock()
	cache.path = ""
	cache.size = 0
	cache.modified = time.Time{}
	cache.stairs = nil
	cache.mu.Unlock()
}
