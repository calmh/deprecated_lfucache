package lfucache

import (
	"container/list"
	"errors"
)

// Cache is a LFU cache structure.
type Cache struct {
	capacity      int
	length        int
	frequencyList *frequencyNode
	index         map[interface{}]*node
	evictedChans  list.List
	stats         Statistics
}

// Current item counts and operation counters.
type Statistics struct {
	LenFreq0    int // Number of items at frequency zero, i.e Inserted but not Accessed
	Inserts     int // Number of Insert()s
	Hits        int // Number of hits (Access() to item)
	Misses      int // Number of misses (Access() to non-existant key)
	Evictions   int // Number of evictions (due to size constraints on Insert(), or EvictIf() calls)
	Deletes     int // Number of Delete()s.
	FreqListLen int // Current length of frequency list, i.e. the number of distinct usage levels
}

type frequencyNode struct {
	usage    int
	prev     *frequencyNode
	next     *frequencyNode
	nodeList *node
	lastNode *node
}

type node struct {
	key    interface{}
	value  interface{}
	parent *frequencyNode
	next   *node
	prev   *node
}

var (
	zeroSizeCache = errors.New("create zero-sized cache")
	emptyLfu      = errors.New("lfu on empty cache")
)

// New initializes a new LFU Cache structure.
func New(capacity int) *Cache {
	if capacity == 0 {
		panic(zeroSizeCache)
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
// existing item is evicted and the new one inserted.
func (c *Cache) Insert(key interface{}, value interface{}) {
	c.check()

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

	c.check()
}

// Delete deletes an item from the cache and returns true. Does nothing and
// returns false if the key was not present in the cache.
func (c *Cache) Delete(key interface{}) bool {
	c.check()

	n, ok := c.index[key]
	if ok {
		c.deleteNode(n)
		c.stats.Deletes++
	}

	c.check()

	return ok
}

// Access an item in the cache. Returns "value, ok" similar to map indexing.
// Increases the item's use count.
func (c *Cache) Access(key interface{}) (interface{}, bool) {
	c.check()

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

	c.check()

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
	c.check()

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
	c.check()

	c.evictedChans.PushBack(e)
}

// UnregisterEvictions removes the channel from the list of channels to be
// notified on item eviction.  Must be called when there is no longer a reader
// for the channel in question.
func (c *Cache) UnregisterEvictions(e chan<- interface{}) {
	c.check()

	for el := c.evictedChans.Front(); el != nil; el = el.Next() {
		if el.Value.(chan<- interface{}) == e {
			c.evictedChans.Remove(el)
			return
		}
	}
}

// EvictIf applies test to each item in the cache and evicts it if the test
// returns true.  Returns the number of items that was evicted.
func (c *Cache) EvictIf(test func(interface{}) bool) int {
	c.check()

	cnt := 0
	for _, n := range c.index {
		if test(n.value) {
			c.evict(n)
			cnt++
		}
	}

	c.check()

	return cnt
}

func (c *Cache) items0() int {
	cnt := 0
	for n := c.frequencyList.nodeList; n != nil; n = n.next {
		cnt++
	}
	return cnt
}

func (c *Cache) evict(n *node) {
	for c := c.evictedChans.Front(); c != nil; c = c.Next() {
		c.Value.(chan<- interface{}) <- n.value
	}
	c.deleteNode(n)
	c.stats.Evictions++
}

func (c *Cache) deleteNode(n *node) {
	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}

	fn := n.parent
	if fn.nodeList == n {
		fn.nodeList = n.next
	}
	if fn.lastNode == n {
		fn.lastNode = n.prev
	}

	if fn.usage != 0 && fn.nodeList == nil {
		c.deleteFrequencyNode(fn)
	}

	delete(c.index, n.key)
	c.length--
}

func (c *Cache) lfu() *node {
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		if fn.nodeList != nil {
			return fn.nodeList
		}
	}

	panic(emptyLfu)
}

func (c *Cache) newFrequencyNode(usage int, parent *frequencyNode) *frequencyNode {
	fn := &frequencyNode{
		usage: usage,
		prev:  parent,
		next:  parent.next,
	}

	if fn.next != nil {
		fn.next.prev = fn
	}

	parent.next = fn

	return fn
}

func (c *Cache) deleteFrequencyNode(fn *frequencyNode) {
	if fn.next != nil {
		fn.next.prev = fn.prev
	}

	fn.prev.next = fn.next
}

func (c *Cache) moveNodeToFn(n *node, fn *frequencyNode) {
	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}

	if n.parent != nil {
		if n.parent.nodeList == n {
			n.parent.nodeList = n.next
		}
		if n.parent.lastNode == n {
			n.parent.lastNode = n.prev
		}
		if n.parent.nodeList == nil && n.parent.usage != 0 {
			c.deleteFrequencyNode(n.parent)
		}
	}

	n.prev = nil
	n.next = nil

	n.parent = fn
	if fn.lastNode != nil {
		n.prev = fn.lastNode
		n.prev.next = n
	}

	if fn.nodeList == nil {
		fn.nodeList = n
	}

	fn.lastNode = n
}

func (c *Cache) numFrequencyNodes() int {
	count := 0
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		count++
	}
	return count
}
