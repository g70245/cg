package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

var ErrNoLogFiles = errors.New("no log files found")

func GetLastLine(fileDir string) (string, error) {
	lines, err := GetLastLines(fileDir, 1)
	if err != nil || len(lines) == 0 {
		return "", err
	}
	return lines[0], nil
}

func GetLastLines(logDir string, lineCount int) ([]string, error) {
	if lineCount <= 0 {
		return []string{}, nil
	}

	path, _, err := findLastModifiedFileBefore(logDir, time.Now().Add(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("find latest log file in %q: %w", logDir, err)
	}
	if path == "" {
		return nil, fmt.Errorf("find latest log file in %q: %w", logDir, ErrNoLogFiles)
	}

	return getLastLinesWithSeek(path, lineCount)
}

func findLastModifiedFileBefore(dir string, t time.Time) (path string, info os.FileInfo, err error) {
	isFirst := true
	min := 0 * time.Second
	err = filepath.Walk(dir, func(p string, i os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if !i.IsDir() && i.ModTime().Before(t) {
			if isFirst {
				isFirst = false
				path = p
				info = i
				min = t.Sub(i.ModTime())
			}
			if diff := t.Sub(i.ModTime()); diff < min {
				path = p
				min = diff
				info = i
			}
		}
		return nil
	})
	return
}

func getLastLinesWithSeek(path string, lineCount int) ([]string, error) {
	fileHandle, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", path, err)
	}
	defer fileHandle.Close()

	stat, err := fileHandle.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat log file %q: %w", path, err)
	}
	if stat.Size() == 0 || lineCount <= 0 {
		return []string{}, nil
	}

	const chunkSize int64 = 4096
	position := stat.Size()
	var tail []byte
	for position > 0 {
		readSize := min(chunkSize, position)
		position -= readSize

		chunk := make([]byte, readSize)
		if _, err := fileHandle.ReadAt(chunk, position); err != nil && err != io.EOF {
			return nil, fmt.Errorf("read log file %q: %w", path, err)
		}
		tail = append(chunk, tail...)

		normalized := normalizeLineEndings(tail)
		if bytes.Count(normalized, []byte{'\n'}) > lineCount {
			break
		}
	}

	normalized := normalizeLineEndings(tail)
	rawLines := bytes.Split(normalized, []byte{'\n'})
	for len(rawLines) > 0 && len(rawLines[len(rawLines)-1]) == 0 {
		rawLines = rawLines[:len(rawLines)-1]
	}
	if len(rawLines) > lineCount {
		rawLines = rawLines[len(rawLines)-lineCount:]
	}

	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		lines = append(lines, byteToString(line))
	}
	return lines, nil
}

func normalizeLineEndings(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
	return bytes.ReplaceAll(data, []byte{'\r'}, []byte{'\n'})
}

func byteToString(buffer []byte) string {
	transformReader := transform.NewReader(bytes.NewReader(buffer), traditionalchinese.Big5.NewDecoder())
	decBytes, _ := io.ReadAll(transformReader)
	return string(decBytes)
}
