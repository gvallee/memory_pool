/*
 * Copyright(c) 2019 Geoffroy Vallee All rights reserved.
 * This software is licensed under a 3-clause BSD license. Please consult the
 * LICENSE.md file distributed with the sources of this project regarding your
 * rights to use or distribute this software.
 */

package pool

import (
	"encoding/binary"
	"testing"
)

func TestNew(t *testing.T) {
	p := Pool{
		ObjSize:    16,    // Size of a single object for the pool
		NObj:       1,     // Number of objects in the pool
		Erase:      false, // We do not need to erase the data in the object when returning it to the pool
		GrowFactor: 0,     // The memory pool cannot grow
	}

	p.New()

	// Get some objects and write to it to check all is okay
	t.Log("Getting object 1...")
	obj1 := p.Get()
	if obj1 == nil {
		t.Fatal("failed to get object")
	}
	if len(obj1) != 16 {
		t.Fatalf("object is of the wrong size (%d vs. 16)", len(obj1))
	}

	t.Log("Getting object 2...")
	obj2 := p.Get()
	if obj2 == nil {
		t.Fatal("failed to get object")
	}
	if len(obj2) != 16 {
		t.Fatalf("object is of the wrong size (%d vs. 16)", len(obj2))
	}
	// Do something with the object
	s1 := binary.PutVarint(obj1, 42)
	s2 := binary.PutVarint(obj2, 11)

	// This one should fail
	t.Log("Getting object 13, which should fail...")
	obj3 := p.Get()
	if obj3 != nil {
		t.Fatal("we were able to get more objects than the capacity")
	}

	// Check obj1
	val, size := binary.Varint(obj1)
	if val != 42 && size != s1 {
		t.Fatal("data in first object is corrupted")
	}

	// Check obj2
	val, size = binary.Varint(obj2)
	if val != 11 && size != s2 {
		t.Fatal("data in first object is corrupted")
	}

	// Return obj1
	err := p.Return(obj1)
	if err != nil {
		t.Fatal("failed to return first object")
	}

	val, size = binary.Varint(obj2)
	if val != 11 && size != s2 {
		t.Fatal("data in first object is corrupted")
	}

	err = p.Return(obj2)
	if err != nil {
		t.Fatal("failed to return object")
	}
}

func TestGrow(t *testing.T) {
	p := Pool{
		ObjSize:    8,     // Size of a single object for the pool
		NObj:       1,     // Number of objects in the pool
		Erase:      false, // We do not need to erase the data in the object when returning it to the pool
		GrowFactor: 3,     // The memory pool grows by a factor of 3 everytime it needs to grow
	}

	p.New()
	t.Log("Getting object 1...")
	obj1 := p.Get()
	if obj1 == nil {
		t.Fatal("failed to get object")
	}
	if len(obj1) != 8 {
		t.Fatalf("object is of the wrong size (%d vs. 8)", len(obj1))
	}

	t.Log("Getting object 2, it should grow the memory pool...")
	obj2 := p.Get()
	if obj2 == nil {
		t.Fatal("failed to get object")
	}
	if len(obj2) != 8 {
		t.Fatalf("object is of the wrong size (%d vs. 8)", len(obj2))
	}

	err := p.Return(obj1)
	if err != nil {
		t.Fatal("failed to return object")
	}
	err = p.Return(obj2)
	if err != nil {
		t.Fatal("failed to return object")
	}

	// Note the length of the pool is the current size, not the capacity, so we return the objects first
	if int(len(p.data)) != int(3*p.ObjSize) {
		t.Fatalf("pool size of incorrect of growth (%d vs. %d)", len(p.data), 3*p.ObjSize)
	}
}
