package lfucache // import "github.com/calmh/deprecated_lfucache"

import (
	"errors"
)

// Cache is an LFU cache structure.
type Cache struct {
	capacity      int
	length        int
	frequencyList *frequencyNode
	index         map[interface{}]*node
	evictedChans  []chan<- interface{}
	stats         Statistics
}

// Statistics contains current item counts and operation counters.
type Statistics struct {
	LenFreq0    int // Number of items at frequency zero, i.e Inserted but not Accessed
	Inserts     int // Number of Insert()s
	Hits        int // Number of hits (Access() to item)
	Misses      int // Number of misses (Access() to non-existant key)
	Evictions   int // Number of evictions (due to size constraints on Insert(), or EvictIf() calls)
	Deletes     int // Number of Delete()s.
	FreqListLen int // Current length of frequency list, i.e. the number of distinct usage levels
}

// The "frequencyNode" and "node" types make up the two levels of linked lists
// that we use to keep track of the usage per node. It could be argued that we
// should use the built in list type instead of implementing the linked list
// structure again directly in the nodes. I tried that and while it results in
// slightly less code, the extra layer of indirection and resulting type
// assertions make the code less readable. Also the Access() method becomes
// several times slower and requires a heap allocation per call. All in all,
// this was preferable.

type frequencyNode struct {
	usage int
	prev  *frequencyNode
	next  *frequencyNode
	head  *node
	tail  *node // most recently inserted
}

type node struct {
	key    interface{}
	value  interface{}
	parent *frequencyNode
	next   *node
	prev   *node
}

var (
	errZeroSizeCache = errors.New("create zero-sized cache")
	errEmptyLFU      = errors.New("lfu on empty cache")
)

// New initializes a new LFU Cache structure with the specified capacity.
func New(capacity int) *Cache {
	if capacity == 0 {
		panic(errZeroSizeCache)
	}

	return &Cache{
		capacity:      capacity,
		index:         make(map[interface{}]*node, capacity),
		frequencyList: &frequencyNode{},
	}
}

// Resize the cache to a new capacity. When shrinking, items may get evicted.
func (c *Cache) Resize(capacity int) {
	c.capacity = capacity
	for c.length > c.capacity {
		c.evict(c.lfu())
	}
}

// Insert inserts an item into the cache. If the key already exists, the
// existing item is evicted and the new one inserted. The key type is
// restricted to those acceptable as map keys
// (http://golang.org/ref/spec#Map_types).
func (c *Cache) Insert(key interface{}, value interface{}) {
	if debug {
		c.check()
	}

	if n, ok := c.index[key]; ok {
		c.evict(n)
	}

	if c.length == c.capacity {
		c.evict(c.lfu())
	}

	n := &node{key: key, value: value}
	c.index[key] = n
	c.moveNodeToFn(n, c.frequencyList)
	c.length++
	c.stats.Inserts++

	if debug {
		c.check()
	}
}

// Delete deletes an item from the cache and returns true. Does nothing and
// returns false if the key was not present in the cache.
func (c *Cache) Delete(key interface{}) bool {
	if debug {
		c.check()
	}

	n, ok := c.index[key]
	if ok {
		c.deleteNode(n)
		c.stats.Deletes++
	}

	if debug {
		c.check()
	}

	return ok
}

// Access an item in the cache. Returns "value, ok" similar to map indexing.
// Increases the item's use count.
func (c *Cache) Access(key interface{}) (interface{}, bool) {
	if debug {
		c.check()
	}

	n, ok := c.index[key]
	if !ok {
		c.stats.Misses++
		return nil, false
	}

	nextUsage := n.parent.usage + 1
	var nextFn *frequencyNode
	if n.parent.next == nil || n.parent.next.usage != nextUsage {
		nextFn = c.newFrequencyNode(nextUsage, n.parent)
	} else {
		nextFn = n.parent.next
	}

	c.moveNodeToFn(n, nextFn)
	c.stats.Hits++

	if debug {
		c.check()
	}

	return n.value, true
}

// Len returns the number of items currently stored in the cache.
func (c *Cache) Len() int {
	return c.length
}

// Cap returns the maximum number of items the cache will hold.
func (c *Cache) Cap() int {
	return c.capacity
}

// Statistics returns the cache statistics.
func (c *Cache) Statistics() Statistics {
	if debug {
		c.check()
	}

	c.stats.LenFreq0 = c.items0()
	c.stats.FreqListLen = c.numFrequencyNodes()
	return c.stats
}

// Evictions registers a channel used to report items that get evicted from
// the cache.  Only items evicted due to LFU or EvictIf() will be sent on the
// channel, not items removed by calling Delete(). The channel must be
// unregistered using UnregisterEvictions() prior to ceasing reads in order to
// avoid deadlocking evictions.
func (c *Cache) Evictions(e chan<- interface{}) {
	if debug {
		c.check()
	}

	c.evictedChans = append(c.evictedChans, e)
}

// UnregisterEvictions removes the channel from the list of channels to be
// notified on item eviction. Must be called when there is no longer a reader
// for the channel in question.
func (c *Cache) UnregisterEvictions(e chan<- interface{}) {
	if debug {
		c.check()
	}

	var i int
	var found bool

	for i = range c.evictedChans {
		if c.evictedChans[i] == e {
			found = true
			break
		}
	}

	if found {
		copy(c.evictedChans[i:], c.evictedChans[i+1:])
		c.evictedChans[len(c.evictedChans)-1] = nil
		c.evictedChans = c.evictedChans[:len(c.evictedChans)-1]
	}
}

// EvictIf applies test to each item in the cache and evicts it if the test
// returns true.  Returns the number of items that were evicted.
func (c *Cache) EvictIf(test func(interface{}) bool) int {
	if debug {
		c.check()
	}

	cnt := 0
	for _, n := range c.index {
		if test(n.value) {
			c.evict(n)
			cnt++
		}
	}

	if debug {
		c.check()
	}

	return cnt
}

// evict evicts a node from the cache by removing it from the structure and
// notifying any interested eviction listeners
func (c *Cache) evict(n *node) {
	for i := range c.evictedChans {
		c.evictedChans[i] <- n.value
	}
	c.deleteNode(n)
	c.stats.Evictions++
}

// deleteNode deletes a node from the cache, also deleting the frequency node
// if it became empty
func (c *Cache) deleteNode(n *node) {
	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}

	fn := n.parent
	if fn.head == n {
		fn.head = n.next
	}
	if fn.tail == n {
		fn.tail = n.prev
	}

	if fn.usage != 0 && fn.head == nil {
		c.deleteFrequencyNode(fn)
	}

	delete(c.index, n.key)
	c.length--
}

// lfu returns the least frequently used node in the cache, prefering the
// oldest if there are multiple nodes with the same lowest usage count
func (c *Cache) lfu() *node {
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		if fn.head != nil {
			return fn.head
		}
	}

	panic(errEmptyLFU)
}

// newFrequencyNode inserts a new frequency node after the specified prev node
func (c *Cache) newFrequencyNode(usage int, prev *frequencyNode) *frequencyNode {
	fn := &frequencyNode{
		usage: usage,
		prev:  prev,
		next:  prev.next,
	}

	if fn.next != nil {
		fn.next.prev = fn
	}

	prev.next = fn

	return fn
}

// deleteFrequencyNode removes a new frequency node from the list
func (c *Cache) deleteFrequencyNode(fn *frequencyNode) {
	if fn.next != nil {
		fn.next.prev = fn.prev
	}

	fn.prev.next = fn.next
}

// moveNodeToFn moves a node to become a child of a frequency node, while
// properly removing it from any current frequency node
func (c *Cache) moveNodeToFn(n *node, fn *frequencyNode) {
	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}

	if n.parent != nil {
		if n.parent.head == n {
			n.parent.head = n.next
		}
		if n.parent.tail == n {
			n.parent.tail = n.prev
		}
		if n.parent.head == nil && n.parent.usage != 0 {
			c.deleteFrequencyNode(n.parent)
		}
	}

	n.prev = nil
	n.next = nil

	n.parent = fn
	if fn.tail != nil {
		n.prev = fn.tail
		n.prev.next = n
	}

	if fn.head == nil {
		fn.head = n
	}

	fn.tail = n
}

// items0 returns the number of items at the head of the node list (usage
// count zero)
func (c *Cache) items0() (count int) {
	for n := c.frequencyList.head; n != nil; n = n.next {
		count++
	}
	return
}

// numFrequencyNodes returns the number of frequency nodes in the cache
func (c *Cache) numFrequencyNodes() (count int) {
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		count++
	}
	return
}
