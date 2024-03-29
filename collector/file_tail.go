package collector

import (
	"bytes"
	"os"
)

const (
	tailBufSize = 1000
	newLine     = '\n'
)

// ReadTailOfFile Read only the tail of the file, given the desired buffer size
// and ensure we start after a newline
func ReadTailOfFile(fname string, tailBufSize int64) ([]byte, error) {
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := make([]byte, tailBufSize)
	stat, statErr := file.Stat()
	if statErr != nil {
		panic(statErr)
	}
	start := stat.Size() - tailBufSize
	if start < 0 {
		start = 0
	}
	_, err = file.ReadAt(buf, start)

	// Ensure we start on a new line
	newLineIndex := bytes.IndexByte(buf, newLine)

	// If only a single line, return entire buffer immediately
	if newLineIndex+1 == int(stat.Size()) {
		return buf, nil
	}

	if newLineIndex != -1 && len(buf) > newLineIndex {
		return buf[newLineIndex+1:], nil
	}

	return []byte{}, err
}
