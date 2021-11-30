package main

import (
	"fmt"
	"sync"
	"time"
)

type idDso struct {
	count  uint
	prefix string
	mutex  *sync.Mutex
}

func newId(prefix string) *idDso {
	s := &idDso{}
	s.mutex = new(sync.Mutex)
	s.prefix = prefix
	return s
}

func (id *idDso) next(key string) string {
	count := id.increment()
	//1678 - January 1, 1970 UTC - 2262
	//last 3 digits are always zero
	now := time.Now()
	when := now.Format("20060102T150405.000")
	return fmt.Sprintf("%s-%d-%s-%s", when, count, id.prefix, key)
}

func (id *idDso) increment() uint {
	defer id.mutex.Unlock()
	id.mutex.Lock()
	id.count++
	count := id.count
	return count
}
