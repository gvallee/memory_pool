# memory_pool
Go package enabling the management of memory pools

# Overview

I love Go, I appreciate its slice concept. But the fact that I have no control over the
management of the actual data in the slice is a problem when doing system software management.
A typical example is the development of a new communication package where it is needed to
handle RX and TX buffers to send/receive message in order to achieve some level of decent
performance in the context of many endpoints and large datasets.

A typical way to handle such a problem in C is to create a pool of memory and rely on `void*` pointer to get RX and TX before, handle the data, and return these buffers to the pool once
done. This prevents the allocation of memory in the critical path.

Fortunately, it is possible with Go to allocate memory and make sure the runtime will not
touch it: channels.
So this package is based on a simple idea:
- Allocate a bunch of memory at once (create a pool) based on the concept of objects (an object
is simply a given amount of bytes; so a memory pool has a size of the number of objects times
the size of an object).
- When you need an object, get it from the pool.
- When you are done with the object, return it to the pool.

This package assumes that an object is simply a slice of bytes. When creating the memory
pool, we basically get a big buffer:
```
_____________________________________
|slice 0|slice 1|slice 2|...|slice n|
|_______|_______|_______|___|_______|
```
When getting an object, we declare an empty slice of bytes and gets bytes from the pool.

Of course, this is not as optimal in term of memory allocations than what we would do in C since when we get an object, we declare a new slice of bytes, which will allocate some memory for the slice object itself. But it is nevertheless a useful building block to build other service and
capabilities and still ensure decent performance when handling large memory pools.

# Drawbacks

Channels are by definition slower than the allocation of slices (at least for a lot of cases), simply because
channels are designed for concurrent executions and therefore all perform a runtime level lock. This means
that the cost of a lock is overwelmingly more expensive than memory allocation and garbage collection is more
cases.

Does it mean memory pools are not interesting? No, it means that we need to be clever. For instance, since the
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
