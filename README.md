# lfucache
--
    import "github.com/calmh/lfucache"

Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing the
LFU cache eviction scheme" by K. Shah, A. Mitra and D. Matani.

It is based on two levels of doubly linked lists and gives O(1) insert, access
and delete operations. The cache supports sending evicted items to interested
listeners via channels, and manually evicting all cache items matching a
criteria. This is useful for example when using the package as a write cache for
a database, where items must be written to the backing store on eviction.

It is safe to make calls on the cache concurrently from multiple goroutines.

### Example

    c := lfucache.Create(1024)
    c.Insert("mykey", 2345) // => true
    c.Access("foo")         // => nil, false
    c.Access("mykey")       // => interface{}{2345}, true
    c.Delete("mykey")       // => true


### License

The MIT license.

## Usage

#### type Cache

```go
type Cache struct {
	sync.Mutex
}
```

Cache is a LFU cache structure.

#### func  New

```go
func New(maxItems int) *Cache
```
New initializes a new LFU Cache structure.

#### func (*Cache) Access

```go
func (c *Cache) Access(key string) (interface{}, bool)
```
Access an item in the cache. Returns "value, ok" similar to map indexing.
Increases the item's use count.

#### func (*Cache) Delete

```go
func (c *Cache) Delete(key string) bool
```
Delete deletes an item from the cache and returns true. Does nothing and returns
false if the key was not present in the cache.

#### func (*Cache) EvictIf

```go
func (c *Cache) EvictIf(test func(interface{}) bool) int
```
EvictIf applies test to each item in the cache and evicts it if the test returns
true. Returns the number of items that was evicted.

#### func (*Cache) Evictions

```go
func (c *Cache) Evictions() <-chan interface{}
```
Evictions returns a new channel used to report items that get evicted from the
cache. Only items evicted due to LFU or EvictIf() will be sent on the channel,
not items removed by calling Delete(). The channel must be unregistered using
UnregisterEvictions() prior to ceasing reads in order to avoid deadlocking
evictions.

#### func (*Cache) Insert

```go
func (c *Cache) Insert(key string, value interface{})
```
Insert inserts an item into the cache. If the key already exists, the existing
item is evicted and the new one inserted.

#### func (*Cache) Statistics

```go
func (c *Cache) Statistics() Statistics
```
Statistics returns the cache statistics.

#### func (*Cache) UnregisterEvictions

```go
func (c *Cache) UnregisterEvictions(exp <-chan interface{})
```
UnregisterEvictions removes the channel from the list of channels to be notified
on item eviction. Must be called when there is no longer a reader for the
channel in question.

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
