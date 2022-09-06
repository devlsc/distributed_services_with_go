package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var enc = binary.BigEndian

// bytes of uint 64 in bytes
const lenWidth = 8

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())

	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

/*
* Appends given data to the store.
 */
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// safe current size, as position for the index
	pos = s.size
	// store size of the byte-slice p before the actual data, that the reading part knows how much bytes needs to be read for the record.
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// Write data to buffer
	bytesWritten, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	// written data + the length information
	bytesWritten += lenWidth

	// increase size of the store after successful appending the log
	s.size += uint64(bytesWritten)

	return uint64(bytesWritten), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write buffer to disk
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

    // Read the size from the file
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

    // Read data
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64)(int, error){
    s.mu.Lock()
    defer s.mu.Unlock()

    if err := s.buf.Flush();  err != nil {
        return 0, err
    }

    return s.File.ReadAt(p, off)
}


func (s *store) Close() error{
    s.mu.Lock()
    defer s.mu.Unlock()

    err := s.buf.Flush()
    if err != nil {
        return err
    }
    return s.File.Close()
}
