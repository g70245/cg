package navigation

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	graphicInfoRecordSize  = 40
	graphicInfoEastOffset  = 28
	graphicInfoSouthOffset = 29
	graphicInfoFlagOffset  = 30
	graphicInfoMapIDOffset = 36
	maxMapTileID           = 1<<16 - 1
)

var errGraphicInfoUnavailable = errors.New("graphic map metadata is unavailable")

type mapTileProperty struct {
	East     uint8
	South    uint8
	Passable bool
	Found    bool
}

type graphicInfoIndex struct {
	tiles []mapTileProperty
}

func parseGraphicInfo(reader io.Reader) (*graphicInfoIndex, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read graphic map metadata: %w", err)
	}
	if len(data) == 0 || len(data)%graphicInfoRecordSize != 0 {
		return nil, errors.New("invalid graphic map metadata")
	}

	index := &graphicInfoIndex{tiles: make([]mapTileProperty, maxMapTileID+1)}
	for offset := 0; offset < len(data); offset += graphicInfoRecordSize {
		mapID := int64(int32(binary.LittleEndian.Uint32(data[offset+graphicInfoMapIDOffset : offset+graphicInfoMapIDOffset+4])))
		if mapID < 0 || mapID > maxMapTileID {
			continue
		}
		index.tiles[mapID] = mapTileProperty{
			East:     data[offset+graphicInfoEastOffset],
			South:    data[offset+graphicInfoSouthOffset],
			Passable: data[offset+graphicInfoFlagOffset] == 1,
			Found:    true,
		}
	}
	return index, nil
}

func (index *graphicInfoIndex) lookup(mapID uint16) (mapTileProperty, bool) {
	if index == nil || int(mapID) >= len(index.tiles) {
		return mapTileProperty{}, false
	}
	property := index.tiles[mapID]
	return property, property.Found
}

type graphicInfoCache struct {
	mu       sync.Mutex
	root     string
	path     string
	size     int64
	modified time.Time
	index    *graphicInfoIndex
}

func (cache *graphicInfoCache) Load(gameDir string) (*graphicInfoIndex, string, error) {
	root, err := filepath.Abs(gameDir)
	if err != nil {
		return nil, "", fmt.Errorf("resolve game folder for graphic metadata: %w", err)
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.root != root || !isRegularFile(cache.path) {
		cache.root = root
		cache.path = ""
		cache.index = nil
		cache.size = 0
		cache.modified = time.Time{}
		cache.path, err = findGraphicInfoFile(root)
		if err != nil {
			return nil, "", err
		}
	}

	info, err := os.Stat(cache.path)
	if err != nil {
		return nil, "", fmt.Errorf("inspect graphic map metadata: %w", err)
	}
	version := fmt.Sprintf("%s:%d:%d", cache.path, info.Size(), info.ModTime().UnixNano())
	if cache.index != nil && cache.size == info.Size() && cache.modified.Equal(info.ModTime()) {
		return cache.index, version, nil
	}

	file, err := os.Open(cache.path)
	if err != nil {
		return nil, "", fmt.Errorf("open graphic map metadata: %w", err)
	}
	index, parseErr := parseGraphicInfo(file)
	closeErr := file.Close()
	if parseErr != nil {
		return nil, "", parseErr
	}
	if closeErr != nil {
		return nil, "", fmt.Errorf("close graphic map metadata: %w", closeErr)
	}

	cache.size = info.Size()
	cache.modified = info.ModTime()
	cache.index = index
	return cache.index, version, nil
}

func findGraphicInfoFile(gameDir string) (string, error) {
	roots := []string{gameDir}
	if home, err := os.UserHomeDir(); err == nil {
		if hostDir := sandboxHostGameDir(gameDir, home); hostDir != "" && !strings.EqualFold(hostDir, gameDir) {
			roots = append(roots, hostDir)
		}
	}

	for _, root := range roots {
		matches, err := filepath.Glob(filepath.Join(root, "bin", "GraphicInfo_*.bin"))
		if err != nil {
			return "", fmt.Errorf("match graphic map metadata: %w", err)
		}
		type candidate struct {
			path    string
			version int
		}
		candidates := make([]candidate, 0, len(matches))
		for _, match := range matches {
			name := filepath.Base(match)
			middle := strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(name), "graphicinfo_"), ".bin")
			version, parseErr := strconv.Atoi(middle)
			if parseErr == nil && isRegularFile(match) {
				candidates = append(candidates, candidate{path: match, version: version})
			}
		}
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].version > candidates[j].version })
		if len(candidates) > 0 {
			return candidates[0].path, nil
		}
	}
	return "", errGraphicInfoUnavailable
}

func sandboxHostGameDir(gameDir, userHome string) string {
	clean := filepath.Clean(gameDir)
	volume := filepath.VolumeName(clean)
	rest := strings.TrimPrefix(clean, volume)
	parts := strings.FieldsFunc(rest, func(r rune) bool { return r == '\\' || r == '/' })
	for index := 0; index+1 < len(parts); index++ {
		if strings.EqualFold(parts[index], "user") && strings.EqualFold(parts[index+1], "current") {
			if index+2 >= len(parts) {
				return filepath.Clean(userHome)
			}
			return filepath.Join(append([]string{userHome}, parts[index+2:]...)...)
		}
	}
	return ""
}
