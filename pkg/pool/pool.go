/*
 * Copyright(c) 2019 Geoffroy Vallee. All rights reserved.
 * This software is licensed under a 3-clause BSD license. Please consult the
 * LICENSE.md file distributed with the sources of this project regarding your
 * rights to use or distribute this software.
 */

package pool

// Obj represents an object from a memory pool
//type Obj []byte

// Pool is the data structure representing a memory pool
type Pool struct {
	ObjSize    int64
	NObj       int64
	GrowFactor int8
	Erase      bool
	data       chan []byte // channels have a built-in locking mechanism, no need to protect it with a mutex
}

func (p *Pool) growPool(nNewObj int64) error {
	if p == nil {
		return nil
	}

	// Increase the size of the channel
	p.data = make(chan []byte, nNewObj)
	// The channel has enough space, we create allocate the new (and only the new) memory
	var i int64
	for i = 0; i < nNewObj-p.NObj; i++ {
		b := make([]byte, p.ObjSize)
		p.data <- b
	}

	return nil
}

// New initializes a new memory pool
func (p *Pool) New() error {
	if p == nil {
		return nil
	}

	p.data = make(chan []byte, p.NObj)
	var i int64
	for i = 0; i < p.NObj; i++ {
		b := make([]byte, p.ObjSize)
		p.data <- b
	}

	return nil
}

// Get returns an object from a memory pool
func (p *Pool) Get() []byte {
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
		p.growPool(p.NObj * int64(p.GrowFactor))
	}

	// Get an object from the passive queue
	return <-p.data
}

// Return puts an object into the memory pool for later reuse
func (p *Pool) Return(o []byte) error {
	if p == nil {
		return nil
	}

	var i int64
	if p.Erase {
		for i = 0; i < p.ObjSize; i++ {
			o[i] = 0
		}
	}

	p.data <- o
	/*
		for i = 0; i < p.ObjSize; i++ {
			p.data <- (*o)[i]
		}
	*/

	o = nil

	return nil
}
