package lfucache

import "container/list"

// Cache is a LFU cache structure.
type Cache struct {
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

// Create a new LFU Cache structure.
func Create(maxItems int) *Cache {
	c := Cache{}
	c.maxItems = maxItems
	c.index = make(map[string]*node)
	c.frequencyList = &frequencyNode{0, nil, nil, nil}
	c.evictedChans = list.New()
	return &c
}

// Insert a new item into the cache. Does nothing if the key already exists in
// the cache.
func (c *Cache) Insert(key string, value interface{}) {
	if _, ok := c.index[key]; ok {
		return
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
}

// Delete an item from the cache. Does nothing if the item is not present in
// the cache.
func (c *Cache) Delete(key string) {
	n, ok := c.index[key]
	if ok {
		c.deleteNode(n)
		c.stats.Deletes++
	}
}

// Access an item in the cache. Returns "value, ok" similar to map indexing.
// Increases the item's use count.
func (c *Cache) Access(key string) (interface{}, bool) {
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
	return n.value, true
}

// Returns the cache operation statistics.
func (c *Cache) Statistics() Statistics {
	c.stats.Items = c.numItems
	c.stats.ItemsFreq0 = c.items0()
	c.stats.FreqListLen = c.numFrequencyNodes()
	return c.stats
}

// Return a new channel used to report items that get evicted from the cache.
// Only items evicted due to LFU or EvictIf() will be sent on the channel, not
// items removed by calling Delete().
func (c *Cache) Evictions() <-chan interface{} {
	exp := make(chan interface{})
	c.evictedChans.PushBack(exp)
	return exp
}

// Removes the channel from the list of channels to be notified on item eviction.
// Must be called when there is no longer a reader for the channel in question.
func (c *Cache) UnregisterEvictions(exp <-chan interface{}) {
	for el := c.evictedChans.Front(); el != nil; el = el.Next() {
		if el.Value.(chan interface{}) == exp {
			c.evictedChans.Remove(el)
			return
		}
	}
}

// Applies test to each item in the cache and evicts it if the test returns true.
// Returns the number of items that was evicted.
func (c *Cache) EvictIf(test func(interface{}) bool) int {
	cnt := 0
	for _, n := range c.index {
		if test(n.value) {
			c.evict(n)
			cnt++
		}
	}
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

	if n.parent.nodeList == n {
		n.parent.nodeList = n.next
	}

	delete(c.index, n.key)
	c.numItems--

	if n.parent.usage != 0 && n.parent.nodeList == nil {
		c.deleteFrequencyNode(n.parent)
	}
}

func (c *Cache) lfu() *node {
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		if fn.nodeList != nil {
			return fn.nodeList
		}
	}

	panic("call to lfu() on empty cache")
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
		n.prev = nil
	}

	if n.next != nil {
		n.next.prev = n.prev
		n.prev = nil
	}

	if n.parent != nil && n.parent.nodeList == n {
		n.parent.nodeList = n.next
	}

	if n.parent != nil && n.parent.usage != 0 && n.parent.nodeList == nil {
		c.deleteFrequencyNode(n.parent)
	}

	n.parent = fn
	n.next = fn.nodeList
	if n.next != nil {
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
