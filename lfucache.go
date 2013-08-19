package lfucache

import (
	"container/list"
	"sync"
)

// Cache is a LFU cache structure.
type Cache struct {
	sync.Mutex
	maxItems      int
	numItems      int
	frequencyList *frequencyNode
	index         map[string]*node
	evictedChans  *list.List
	stats         Statistics
}

// Current item counts and operation counters.
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

type frequencyNode struct {
	usage    int
	prev     *frequencyNode
	next     *frequencyNode
	nodeList *node
}

type node struct {
	key    string
	value  interface{}
	parent *frequencyNode
	next   *node
	prev   *node
}

// New initializes a new LFU Cache structure.
func New(maxItems int) *Cache {
	if maxItems == 0 {
		panic("cannot create zero-sized cache")
	}

	c := Cache{}
	c.maxItems = maxItems
	c.index = make(map[string]*node)
	c.frequencyList = &frequencyNode{}
	c.evictedChans = list.New()
	return &c
}

// Insert inserts an item into the cache.
// If the key already exists, the existing item is evicted and the new one inserted.
func (c *Cache) Insert(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	c.check()

	if n, ok := c.index[key]; ok {
		c.evict(n)
	}

	if c.numItems == c.maxItems {
		c.evict(c.lfu())
	}

	n := &node{}
	n.key = key
	n.value = value
	c.index[key] = n
	c.moveNodeToFn(n, c.frequencyList)
	c.numItems++
	c.stats.Inserts++

	c.check()
}

// Delete deletes an item from the cache and returns true. Does nothing and
// returns false if the key was not present in the cache.
func (c *Cache) Delete(key string) bool {
	c.Lock()
	defer c.Unlock()

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
func (c *Cache) Access(key string) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()

	c.check()

	n, ok := c.index[key]
	if !ok {
		c.stats.Misses++
		return nil, false
	}

	nextUsage := n.parent.usage + 1
	var nextFn *frequencyNode
	if n.parent.next == nil || n.parent.next.usage != nextUsage {
		nextFn = c.newFrequencyNode(nextUsage, n.parent, n.parent.next)
	} else {
		nextFn = n.parent.next
	}

	c.moveNodeToFn(n, nextFn)
	c.stats.Hits++

	c.check()

	return n.value, true
}

// Statistics returns the cache statistics.
func (c *Cache) Statistics() Statistics {
	c.Lock()
	defer c.Unlock()

	c.check()

	c.stats.Items = c.numItems
	c.stats.ItemsFreq0 = c.items0()
	c.stats.FreqListLen = c.numFrequencyNodes()
	return c.stats
}

// Evictions returns a new channel used to report items that get evicted from
// the cache.  Only items evicted due to LFU or EvictIf() will be sent on the
// channel, not items removed by calling Delete(). The channel must be
// unregistered using UnregisterEvictions() prior to ceasing reads in order to
// avoid deadlocking evictions.
func (c *Cache) Evictions() <-chan interface{} {
	c.Lock()
	defer c.Unlock()

	c.check()

	exp := make(chan interface{})
	c.evictedChans.PushBack(exp)
	return exp
}

// UnregisterEvictions removes the channel from the list of channels to be
// notified on item eviction.  Must be called when there is no longer a reader
// for the channel in question.
func (c *Cache) UnregisterEvictions(exp <-chan interface{}) {
	c.Lock()
	defer c.Unlock()

	c.check()

	for el := c.evictedChans.Front(); el != nil; el = el.Next() {
		if el.Value.(chan interface{}) == exp {
			c.evictedChans.Remove(el)
			return
		}
	}
}

// EvictIf applies test to each item in the cache and evicts it if the test
// returns true.  Returns the number of items that was evicted.
func (c *Cache) EvictIf(test func(interface{}) bool) int {
	c.Lock()
	defer c.Unlock()

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
		c.Value.(chan interface{}) <- n.value
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

	// Delete empty non-head frequency node
	if fn.usage != 0 && fn.nodeList == nil {
		c.deleteFrequencyNode(fn)
	}

	delete(c.index, n.key)
	c.numItems--
}

func (c *Cache) lfu() *node {
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		if fn.nodeList != nil {
			return fn.nodeList
		}
	}

	panic("bug: call to lfu() on empty cache")
}

func (c *Cache) newFrequencyNode(usage int, prev, next *frequencyNode) *frequencyNode {
	fn := &frequencyNode{usage, prev, next, nil}
	if fn.prev != nil {
		fn.prev.next = fn
	}
	if fn.next != nil {
		fn.next.prev = fn
	}
	return fn
}

func (c *Cache) deleteFrequencyNode(fn *frequencyNode) {
	if fn.next != nil {
		fn.next.prev = fn.prev
	}

	if fn.prev != nil {
		fn.prev.next = fn.next
	}
}

func (c *Cache) moveNodeToFn(n *node, fn *frequencyNode) {
	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}

	if n.parent != nil && n.parent.nodeList == n {
		n.parent.nodeList = n.next
	}

	if n.parent != nil && n.parent.usage != 0 && n.parent.nodeList == nil {
		c.deleteFrequencyNode(n.parent)
	}

	n.prev = nil
	n.next = nil

	n.parent = fn
	if fn.nodeList != nil {
		n.next = fn.nodeList
		n.next.prev = n
	}
	fn.nodeList = n
}

func (c *Cache) numFrequencyNodes() int {
	count := 0
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		count++
	}
	return count
}
