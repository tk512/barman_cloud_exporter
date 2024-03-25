package collector

import (
	"bytes"
	"os"
)

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
	_, err = file.ReadAt(buf, start)

	// Ensure we start on a new line
	newLineIndex := bytes.IndexByte(buf, []byte("\r")[0])
	if newLineIndex != -1 && len(buf) > newLineIndex {
		return buf[newLineIndex+1:], nil
	}

	if err != nil {
		return []byte{}, err
	}

	return buf, nil
}
