package fjson

import (
	"errors"
	"io"
)

var (
	ErrScan      error = errors.New("FJSON Request scanning failed")
	ErrMarshal   error = errors.New("FJSON marshal error")
	ErrUnmarshal error = errors.New("FJSON unmarshal error")
	ErrRead      error = errors.New("FJSON read error")
	ErrWrite     error = errors.New("FJSON write error")
	ErrHandler   error = errors.New("FJSON handler error")
	ErrTimeout   error = errors.New("FJSON timeout")
	ErrDial      error = errors.New("FJSON dial error")
)

func scanPack(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Scan until 0, marking end of pack.
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return i, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, io.ErrUnexpectedEOF
	}

	// Request more data.
	return 0, nil, nil
}
