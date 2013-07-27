# lfucache
--
    import "github.com/calmh/lfucache"

Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing the
LFU cache eviction scheme" by K. Shah, A. Mitra and D. Matani.

It is based on two levels of doubly linked lists and gives O(1) insert, access
and delete operations. The cache supports sending evicted items to interested
listeners via channels, and manually evicting all cache items matching a
criteria. This is useful for example when using the cache as a write cache for a
database, where evicted items must be written to the database.

## Usage

#### type Cache

```go
type Cache struct {
}
```

Cache is a LFU cache structure.

#### func  Create

```go
func Create(maxItems int) *Cache
```
Create a new LFU Cache structure.

#### func (*Cache) Access

```go
func (c *Cache) Access(key string) (interface{}, bool)
```
Access an item in the cache. Returns "value, ok" similar to map indexing.
Increases the item's use count.

#### func (*Cache) Delete

```go
func (c *Cache) Delete(key string)
```
Delete an item from the cache. Does nothing if the item is not present in the
cache.

#### func (*Cache) EvictIf

```go
func (c *Cache) EvictIf(test func(interface{}) bool) int
```
Applies test to each item in the cache and evicts it if the test returns true.
Returns the number of items that was evicted.

#### func (*Cache) Evictions

```go
func (c *Cache) Evictions() <-chan interface{}
```
Return a new channel used to report items that get evicted from the cache. Only
items evicted due to LFU or EvictIf() will be sent on the channel, not items
removed by calling Delete().

#### func (*Cache) Insert

```go
func (c *Cache) Insert(key string, value interface{})
```
Insert a new item into the cache. Does nothing if the key already exists in the
cache.

#### func (*Cache) Print

```go
func (c *Cache) Print()
```

#### func (*Cache) Statistics

```go
func (c *Cache) Statistics() Statistics
```
Returns the cache operation statistics.

#### func (*Cache) UnregisterEvictions

```go
func (c *Cache) UnregisterEvictions(exp <-chan interface{})
```
Removes the channel from the list of channels to be notified on item eviction.
Must be called when there is no longer a reader for the channel in question.

#### type Statistics

```go
type Statistics struct {
	Items       int // Number of items currently in the cache
	ItemsFreq0  int // Number of items at frequency zero, i.e Inserted but not Accessed
	Inserts     int // Number of Insert()s
	Hits        int // Number of hits (Access() to item)
	Misses      int // Number of misses (Access() to non-existant key)
	Evictions   int // Number of evictions (due to size constraints on Insert(), or EvictIf() calls)
	Deletes     int // Number of Delete()s.
	FreqListLen int // Current length of frequency list, i.e. the number of distinct usage levels
}
```

Current item counts and operation counters.
