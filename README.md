# memory_pool
Go package enabling the management of memory pools

# Overview

I love Go, I appreciate its slice concept. But the fact that I have no control over the
management of the actual data in the slice is a problem when doing system software development.
A typical example is the development of a new communication package where it is needed to
handle RX and TX buffers to send/receive messages in order to achieve some level of decent
performance in the context of many endpoints and large datasets.

A typical way to handle such a problem in C is to create a pool of memory and rely on `void*` 
pointers to get RX and TX buffers, handle the data, and return these buffers to the pool once
done. This prevents the allocation of memory in the critical path.

Fortunately, it is possible with Go to allocate memory and make sure the runtime will not
touch it: channels.
So this package is based on a simple idea:
- Allocate a bunch of memory at once (create a pool of byte slices) based on the concept of
*objects* (an object is simply ultimately a byte slice from a pool); so a memory pool has a
size of the number of objects times the size of an object.
- When you need an object, get it from the pool.
- When you are done with the object, return it to the pool.

This package assumes that an object is simply a slice of bytes. When creating the memory
pool, we basically get a bug chunk of memory that has the following conceptual layout:
```
_____________________________________
|slice 0|slice 1|slice 2|...|slice n|
|_______|_______|_______|___|_______|
```

# Drawbacks

Channels are by definition slower than the allocation of slices for some sizes (basically small sizes),
simply because channels are designed for concurrent executions and therefore all perform a runtime level lock.
This means that the cost of a lock can be overwelmingly more expensive than memory allocation and garbage collection
for a given slice size.

Does it mean memory pools are not interesting? No, it means that we need to be clever and understand the
problem you are trying to solve before opting for memory pools. For instance, since the
runtime locks a channel every time we get data from it, it is not efficient to get a channel that deals with
1 byte requests at a time. Instead, the granularity of the channel should be the byte slice of the requested
size. Of course, it also means that for small slices, the benefit of the memory pool will be overcome by
the cost of memory allocation itself: since for all slices, we now have two operations per slice (get and
return), plus the cost of internal locking of the channel, it ends up being more expensive. By for larger
slices, the benefit is clear: on my system for a slice of 1M bytes, when requesting a slice, using it, 
throwing it away, using memory pools is about 33% faster. Where is the soft spot where memory pools are 
faster? It depends on your system and the problem you trying to solve.

Figures illustrating these results are available in the `doc/data` directory.

# Growing a memory pool

Using a channel is a powerful choice here since a channel already handle its content as a queue.
In other words, when we safely get an object from a memory pool, we know it cannot be given to
any other process, thread, or Go routine. 

It also means it is easy to grow a memory pool: when the memory pool is empty, make a new one with the new capacity and allocate the requested new objects. Then we can return the objects already in
use using the same memory pool.

Note that this package does not protect the pool with
a mutex so if you use memory pools with concurrent accesses, you are responsible for providing
the adequate protection. This problem arises only when the pool can grow: we do not protect the
pool structure with any lock, which means that in theory, growing the pool and getting or returning
a slice could happen at the same time and lead to some operations failing.

# Usage

For details, please refer to the tests and benchmarks.

## Pool initialization

```
	pool := Pool{
		ObjSize:    <object size>,
		NObj:       <number of objects added during initialization, must be at least 1>,
		GrowFactor: <0 means the pool cannot grow; any positive number means the size will be increase by the fact when empty>,
		Erase:      <false: the objects are simply returned; true: the objects' content is set to zero before being returned >,
	}
	pool.New()
```

## Get a slice of bytes from the pool

```
s := pool.Get()
if s == nil {
    fmt.Println("unable to get object from pool")
}
```

## Return an object to a pool

```
err := p.Return(s)
if err != nil {
	t.Fatal("failed to return object")
}
```

