package system

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

func FindLastModifiedFileBefore(dir string, t time.Time) (path string, info os.FileInfo, err error) {
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

func GetLinesWithSeek(filepath string, lineNumber int) []string {
	fileHandle, err := os.Open(filepath)

	if err != nil {
		panic("Cannot open file")
	}
	defer fileHandle.Close()

	lines := make([]string, 0)
	var buffer []byte

	var cursor int64 = -2
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()
	for {
		cursor--
		fileHandle.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		fileHandle.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			lines = append(lines, byteToString(buffer))
			if len(lines) == lineNumber {
				break
			} else {
				cursor--
			}
		}

		buffer = append([]byte{char[0]}, buffer...)

		if cursor == -filesize { // stop if we are at the begining
			break
		}

	}

	return lines
}

func byteToString(buffer []byte) string {
	transformReader := transform.NewReader(bytes.NewReader(buffer), traditionalchinese.Big5.NewDecoder())
	decBytes, _ := io.ReadAll(transformReader)
	return string(decBytes)
}

func GetLastLineOfLog(logDir string) string {
	path, _, _ := FindLastModifiedFileBefore(logDir, time.Now().Add(10*time.Second))
	return GetLinesWithSeek(path, 1)[0]
}

func GetLinesOfLog(logDir string, lineNumber int) []string {
	path, _, _ := FindLastModifiedFileBefore(logDir, time.Now().Add(10*time.Second))
	return GetLinesWithSeek(path, lineNumber)
}
