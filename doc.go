/*
Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing
the LFU cache eviction scheme" by  K. Shah, A. Mitra and D.  Matani.

It is based on two levels of doubly linked lists and gives O(1) insert, access
and delete operations. The cache supports sending evicted items to interested
listeners via channels, and manually evicting all cache items matching a
criteria. This is useful for example when using the package as a write cache
for a database, where items must be written to the backing store on eviction.

It is safe to make calls on the cache concurrently from multiple goroutines.

Example

	c := lfucache.Create(1024)
	c.Insert("mykey", 2345) // => true
	c.Access("foo")         // => nil, false
	c.Access("mykey")       // => interface{}{2345}, true
	c.Delete("mykey")       // => true

License

The MIT license.

*/
package lfucache
