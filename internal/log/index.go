package log

import (
	"os"
	"github.com/tysonmote/gommap"
)

var (
	// We store offsets as uint32s and positions as uint64s, so they take 
	// up 4 and 8 bytes of space, respectively.
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth = offWidth + posWidth
)

type index struct {
	file *os.File
	// maps a file's contents directly into memory. It allows your program to
	// read and write file contents as if it were just accessing an in-memory 
	// byte slice, which is faster and more efficient for many use cases.
	mmap gommap.MMap
	size uint64
}