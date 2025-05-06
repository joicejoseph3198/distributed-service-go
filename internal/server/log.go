package server

import (
	"fmt"
	"sync"
)


type Log struct {
	mu sync.Mutex;
	records []Record
}

type Record struct{
	Value string `json:"value"`
	Offset uint64 `json:"offset"`
}

var ErrOffsetNotFound = fmt.Errorf("offset not found")

// Returns a new instance of log
func NewLog() *Log{
	return &Log{}
}

// func (c *Log) associates a method with the concrete type *Log.

/*
defer c.mu.Unlock();
the unlock happens no matter what — even if:
    An error is returned
    A panic happens (in Go)
    A return is called early

Similar to finally in Java
*/

func (c *Log) Append(record Record) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	record.Offset = uint64(len(c.records));
	c.records = append(c.records, record)
	return record.Offset, nil
}

func (c *Log) Read(offset uint64) (Record, error){
	c.mu.Lock()
    defer c.mu.Unlock()
	if(offset >= uint64(len(c.records))){
		return Record{}, ErrOffsetNotFound
	}
	return c.records[offset], nil
}