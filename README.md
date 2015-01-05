> **DEPRECATION WARNING**
> This package is unmaintained and probably not what you're looking for.

lfucache [![Build Status](https://drone.io/github.com/calmh/lfucache/status.png)](https://drone.io/github.com/calmh/lfucache/latest)
========

```go
import "github.com/calmh/lfucache"
```

Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing the
LFU cache eviction scheme" by K. Shah, A. Mitra and D. Matani.

It is based on two levels of doubly linked lists and gives O(1) insert, access
and delete operations. The cache supports sending evicted items to interested
listeners via channels, and manually evicting all cache items matching a
criteria. This is useful for example when using the package as a write cache for
a database, where items must be written to the backing store on eviction.

The cache structure is not thread safe.

Example
-------

```go
c := lfucache.Create(1024) // The cache will hold up to 1024 items.
c.Access("mykey")          // => nil, false
c.Insert("mykey", 2345)    // => true
v, ok := c.Access("mykey") // => v = interface{}{2345}, ok = true
c.Delete("mykey")          // => true
```

Documentation
-------------

http://godoc.org/github.com/calmh/lfucache

License
-------

MIT
