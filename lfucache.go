package lfucache

import (
	"container/list"
)

// Cache is a LFU cache structure.
type Cache struct {
	maxItems      int
	numItems      int
	frequencyList list.List
	index         map[string]*list.Element
	evictedChans  list.List
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
	nodeList list.List
}

type node struct {
	key    string
	value  interface{}
	parent *list.Element
}

// New initializes a new LFU Cache structure.
func New(maxItems int) *Cache {
	if maxItems == 0 {
		panic("cannot create zero-sized cache")
	}

	c := Cache{}
	c.maxItems = maxItems
	c.index = make(map[string]*list.Element)
	c.frequencyList.PushFront(&frequencyNode{})
	return &c
}

// Insert inserts an item into the cache.
// If the key already exists, the existing item is evicted and the new one inserted.
func (c *Cache) Insert(key string, value interface{}) {
	if e, ok := c.index[key]; ok {
		c.evict(e)
	}

	if c.numItems == c.maxItems {
		c.evict(c.lfu())
	}

	fne := c.frequencyList.Front()
	fnv := fn(fne)
	nv := &node{key: key, value: value, parent: fne}
	ne := fnv.nodeList.PushBack(nv)
	c.index[key] = ne
	c.numItems++
	c.stats.Inserts++

}

// Delete deletes an item from the cache and returns true. Does nothing and
// returns false if the key was not present in the cache.
func (c *Cache) Delete(key string) bool {
	e, ok := c.index[key]
	if ok {
		c.deleteNode(e)
		c.stats.Deletes++
	}

	return ok
}

// Access an item in the cache. Returns "value, ok" similar to map indexing.
// Increases the item's use count.
func (c *Cache) Access(key string) (interface{}, bool) {
	ne, ok := c.index[key]
	if !ok {
		c.stats.Misses++
		return nil, false
	}

	nv := n(ne)
	nextUsage := fn(nv.parent).usage + 1
	var nextFn *list.Element

	if pe := nv.parent.Next(); pe == nil || fn(pe).usage != nextUsage {
		newfn := &frequencyNode{usage: nextUsage}
		nextFn = c.frequencyList.InsertAfter(newfn, nv.parent)
	} else {
		nextFn = pe
	}

	c.moveNodeToFn(ne, nextFn)
	c.stats.Hits++

	return nv.value, true
}

// Statistics returns the cache statistics.
func (c *Cache) Statistics() Statistics {
	c.stats.Items = c.numItems
	c.stats.ItemsFreq0 = c.items0()
	c.stats.FreqListLen = c.numFrequencyNodes()
	return c.stats
}

// Evictions registers a channel used to report items that get evicted from
// the cache.  Only items evicted due to LFU or EvictIf() will be sent on the
// channel, not items removed by calling Delete(). The channel must be
// unregistered using UnregisterEvictions() prior to ceasing reads in order to
// avoid deadlocking evictions.
func (c *Cache) Evictions(e chan<- interface{}) {
	c.evictedChans.PushBack(e)
}

// UnregisterEvictions removes the channel from the list of channels to be
// notified on item eviction.  Must be called when there is no longer a reader
// for the channel in question.
func (c *Cache) UnregisterEvictions(e chan<- interface{}) {
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
	cnt := 0
	for _, ne := range c.index {
		if test(n(ne).value) {
			c.evict(ne)
			cnt++
		}
	}

	return cnt
}

func (c *Cache) items0() int {
	return fn(c.frequencyList.Front()).nodeList.Len()
}

func (c *Cache) evict(ne *list.Element) {
	nv := n(ne)
	for c := c.evictedChans.Front(); c != nil; c = c.Next() {
		c.Value.(chan<- interface{}) <- nv.value
	}
	c.deleteNode(ne)
	c.stats.Evictions++
}

func (c *Cache) deleteNode(ne *list.Element) {
	nv := n(ne)
	pe := nv.parent
	pv := fn(pe)
	pv.nodeList.Remove(ne)

	// Delete empty non-head frequency node
	if pv.usage != 0 && pv.nodeList.Len() == 0 {
		c.frequencyList.Remove(pe)
	}

	delete(c.index, nv.key)
	c.numItems--
}

func (c *Cache) lfu() *list.Element {
	for fne := c.frequencyList.Front(); fne != nil; fne = fne.Next() {
		if ne := fn(fne).nodeList.Front(); ne != nil {
			return ne
		}
	}
	panic("no item to evict")
}

func (c *Cache) moveNodeToFn(ne *list.Element, fne *list.Element) {
	nv := n(ne)
	fnv := fn(fne)

	if curPar := nv.parent; curPar != nil {
		// Remove from existing parent
		curParV := fn(curPar)
		curParV.nodeList.Remove(ne)
		if curParV.usage != 0 && curParV.nodeList.Len() == 0 {
			// Delete empty parent
			c.frequencyList.Remove(curPar)
		}
	}

	nv.parent = fne
	ne = fnv.nodeList.PushBack(nv)
	c.index[nv.key] = ne
}

func (c *Cache) numFrequencyNodes() int {
	return c.frequencyList.Len()
}

func n(e *list.Element) *node {
	return e.Value.(*node)
}

func fn(e *list.Element) *frequencyNode {
	return e.Value.(*frequencyNode)
}
