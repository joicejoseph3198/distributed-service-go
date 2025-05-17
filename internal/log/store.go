package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// Go variable declaration that defines a package-level variable
var (
	// used for encoding and decoding numbers into byte slices.
	// BigEndian - predefined variable in binary package that implements ByteOrder interface.
	// ByteOrder defines how multibyte values (like uint32, int64, etc.) are encoded into byte slices.
	// Big Endian means the most significant byte is stored first.
	enc = binary.BigEndian
)

// package level constant
const ( 
	lenWidth = 8
)

type store struct {
	*os.File // pointer to the file
	mu sync.Mutex
	buf *bufio.Writer // reduces the number of actual disk writes by accumulating data in memory first.
	size uint64 // current size of the file, to calculate the offset
}

func newStore(f *os.File) (*store, error){
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf: bufio.NewWriter(f),
	}, nil
}

// persists the given bytes to the store.
// we return the number of bytes written, and the position where the store holds the record.
// The segment will use this position when it creates an associated index entry for this record.
func (s *store) Append(p []byte) (n uint64, pos uint64, err error){
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size
	// We write to the buffered writer instead of directly to the file.
	// We write the length of the record so that, when we read the record, we know how many bytes to read.
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil{
		return 0,0, err
	}

	w, err := s.buf.Write(p)
	if err != nil {
		return 0,0,err
	}
	// number of bytes written + the bytes used to indicate the length of the record
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
	
}

// To read a full record (a length-prefixed byte slice) starting from a known pos in the store.
// You store each record with a fixed-size prefix (8 bytes) that tells you how long the data is.
// You first read those 8 bytes to know how many bytes to read next.
func (s* store) Read(pos uint64) ([]byte, error){
	s.mu.Lock()
	defer s.mu.Unlock()
	// flush any record thats in buffer first
	// in case we’re about to try to read a record that the buffer hasn’t flushed to disk yet.
	if err := s.buf.Flush(); err != nil{
		return nil, err
	}

	size := make([]byte, lenWidth)
	// Each record in the store file is written like this: 
	// [8-byte length][record data...]
	// You need to know how much data to read. Storing the length as a prefix solves this.
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil{
		return nil, err
	}

	b := make([] byte, enc.Uint64(size))
	// read 
	if _,err := s.File.ReadAt(b, int64(pos + lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}

// A lower-level method that reads raw bytes into a buffer p starting at a given offset off.
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}