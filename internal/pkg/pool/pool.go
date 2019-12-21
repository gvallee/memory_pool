/*
 * Copyright(c) 2019 Geoffroy Vallee. All rights reserved.
 * This software is licensed under a 3-clause BSD license. Please consult the
 * LICENSE.md file distributed with the sources of this project regarding your
 * rights to use or distribute this software.
 */

package pool

import (
	"sync"
)

// Pool is the data structure representing a memory pool
type Pool struct {
	ObjSize    int64
	NObj       int64
	GrowFactor int8
	Erase      bool
	lock       sync.Mutex
	data       chan byte
}

// Obj represents an object from a memory pool
type Obj []byte

func (p *Pool) growPool(newSize int64) error {
	if p == nil {
		return nil
	}

	sizeToAdd := newSize - (p.ObjSize * p.NObj)
	// Increase the capacity
	p.data = make(chan byte, newSize)
	var i int64
	// The channel has enough space, we create allocate the new (and only the new) memory
	for i = 0; i < sizeToAdd; i++ {
		var b byte
		p.data <- b
	}

	return nil
}

// New initializes a new memory pool
func (p *Pool) New() error {
	if p == nil {
		return nil
	}

	p.data = make(chan byte, p.NObj*p.ObjSize)
	for i := 0; i < cap(p.data); i++ {
		var b byte
		p.data <- b
	}

	return nil
}

// Get returns an object from a memory pool
func (p *Pool) Get() Obj {
	if p == nil {
		return nil
	}

	// Check if we have an object available
	if len(p.data) == 0 {
		// If not, can we grow the pool? If not return an error
		if p.GrowFactor <= 0 {
			return nil
		}

		// Grow the pool
		totalSize := (p.NObj * p.ObjSize) * int64(p.GrowFactor)
		p.growPool(totalSize)
	}

	// Lock the pool
	p.lock.Lock()
	defer p.lock.Unlock()

	// Get an object from the passive queue
	var i int64
	var o []byte // empty slice we add bytes from the pool
	for i = 0; i < p.ObjSize; i++ {
		o = append(o, <-p.data)
	}

	return o
}

// Return puts an object into the memory pool for later reuse
func (p *Pool) Return(o Obj) error {
	if p == nil {
		return nil
	}

	var i int64
	if p.Erase {
		for i = 0; i < p.ObjSize; i++ {
			o[i] = 0
		}
	}

	// Lock the active queue
	p.lock.Lock()
	defer p.lock.Unlock()

	for i = 0; i < p.ObjSize; i++ {
		p.data <- o[i]
	}

	return nil
}
