
/*
Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing
the LFU cache eviction scheme" by  K. Shah, A. Mitra and D.  Matani.

It is based on two levels of doubly linked lists and gives O(1) insert, access
and delete operations. The cache supports sending evicted items to interested
listeners via channels, and manually evicting all cache items matching a
criteria. This is useful for example when using the cache as a write cache for
a database, where evicted items must be written to the database.
*/
package lfucache
