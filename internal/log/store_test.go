package log

import (
	"os"
	"testing"
	"github.com/stretchr/testify/require"
)

var (
	write = []byte("Hello World")
	width = uint64(len(write) + lenWidth);
)


func TestStoreAppendRead(t *testing.T){
	// create a temp file
	f, err := os.CreateTemp("","temp_file_for_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s,err := newStore(f)
	require.NoError(t, err)

	testAppend(t,s)
	testRead(t,s)
	testReadAt(t,s)

	// re initiliaze the store
	// verify that our service will recover its state after a restart.
	s, err = newStore(f)
	require.NoError(t,err)
	testReadAt(t,s)
}

func testAppend(t *testing.T, s *store){
	t.Helper()
	for i := 1; i < 4; i++{
		n, pos, err := s.Append(write)
		require.NoError(t, err)
		// checks the size of 
		require.Equal(t, pos+n, width * uint64(i)) 
	}
}

func testRead(t *testing.T, s *store){
	t.Helper()
	pos := uint64(0)
	for i := 1; i < 4; i++{
		read, err := s.Read((pos))
		require.NoError(t, err)
		require.Equal(t, write, read)
		pos += width;
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()
	off := int64(0)
	for i := 1; i < 4; i++{
		b := make([]byte, lenWidth)
		// Read 8 bytes from the file at off. This is the length prefix of the next record.
		n, err := s.ReadAt(b, off)	
		require.NoError(t, err)
		require.Equal(t, lenWidth, n)
		
		off += int64(n)
		// decode the length of the record from the prefix length byte 
		size := enc.Uint64(b)
		// update slice to the size of the record you want to fetch
		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		// check if read worked, check if expected and actual content are equal, check size of the returned record
		require.NoError(t, err)
		require.Equal(t, write, b)
		require.Equal(t, int(size), n)
		off += int64(n)
		
	}
}

func TestStoreClose(t *testing.T) {
	f, err := os.CreateTemp("", "store_close_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	s, err := newStore(f)
	require.NoError(t, err)
	_, _, err = s.Append(write)
	require.NoError(t, err)
	f, beforeSize, err := openFile(f.Name())
	require.NoError(t, err)
	err = s.Close()
	require.NoError(t, err)
	_, afterSize, err := openFile(f.Name())
	require.NoError(t, err)
	require.True(t, afterSize > beforeSize)
}

func openFile(name string) (file *os.File, size int64, err error){
	f, err := os.OpenFile(
		name,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil
}




