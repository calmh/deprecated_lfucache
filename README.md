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
Create a new LFU Cache structure. maxItems is the maximum number of items that
can be contained in the cache.

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

#### func (*Cache) Evicted

```go
func (c *Cache) Evicted() <-chan interface{}
```
Return a new channel used to report items that get evicted from the cache. Only
items evicted due to LFU will be sent on the channel, not items removed by
calling Delete().

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

#### func (*Cache) Size

```go
func (c *Cache) Size() int
```
Returns the number of items in the cache.

#### func (*Cache) Size0

```go
func (c *Cache) Size0() int
```
Returns the number of items at the first level (never Accessed) of the cache.

#### func (*Cache) Statistics

```go
func (c *Cache) Statistics() Statistics
```
Returns the cache operation statistics.

#### func (*Cache) UnregisterEvicted

```go
func (c *Cache) UnregisterEvicted(exp <-chan interface{})
```
Removes the channel from the list of channels to be notified on item eviction.
Must be called when there is no longer a reader for the channel in question.

#### type Statistics

```go
type Statistics struct {
	Inserts     int
	Hits        int
	Misses      int
	Evictions   int
	Deletes     int
	FreqListLen int
}
```

Statistics as monotonically increasing counters, apart from FreqListLen which is
a snapshot value.
