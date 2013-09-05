/*

Package lfucache implements an O(1) LFU (Least Frequenty Used) cache.

This structure is described in the paper "An O(1) algorithm for implementing
the LFU cache eviction scheme" by  K. Shah, A. Mitra and D.  Matani. It is
based on two levels of doubly linked lists and gives O(1) insert, access and
delete operations. It is implemented here with a few practical optimizations
and extra features.

The cache supports sending evicted items to interested listeners via channels,
and manually evicting cache items matching a certain criteria. This is useful
for example when using the package as a write cache for a database, where
items must be written to the backing store on eviction.

Example:

	c := lfucache.Create(1024) // The cache will hold up to 1024 items.
	c.Access("mykey")          // => nil, false
	c.Insert("mykey", 2345)    // => true
	v, ok := c.Access("mykey") // => v = interface{}{2345}, ok = true
	c.Delete("mykey")          // => true

---

Copyright (c) 2013 Jakob Borg. Licensed under the MIT license.

*/
package lfucache
