/*
Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing
the LFU cache eviction scheme" by  K. Shah, A. Mitra and D.  Matani, August 16,
2010. It is based on two levels of doubly linked lists and gives O(1) insert,
access and delete operations.
*/
package lfucache

// Cache is a LFU cache structure.
type Cache struct {
	maxItems      int
	numItems      int
	frequencyList *frequencyNode
	index         map[string]*node
}

// Expirer is the interface for an object that wants to know when it expires
// from the cache. If a value in the cache implements the Expirer interface the
// Expire() method will get called just prior to the object being removed from
// cache. Note: this will only happen when the value is being removed due to an
// LFU cache eviction, not on explicit calls to Delete.
type Expirer interface {
	Expire()
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
// maxItems is the maximum number of items contained in the cache.
func Create(maxItems int) *Cache {
	c := Cache{}
	c.maxItems = maxItems
	c.index = make(map[string]*node)
	c.frequencyList = &frequencyNode{1, nil, nil, nil}
	return &c
}

// Insert a new item into the cache.
func (c *Cache) Insert(key string, value interface{}) {
	if c.numItems == c.maxItems {
		n := c.lfu()
		if v, ok := n.value.(Expirer); ok {
			v.Expire()
		}
		c.deleteNode(n)
	}

	// Create node
	n := &node{}
	n.key = key
	n.value = value

	// Insert into map
	c.index[key] = n

	// Insert into LFU Cache
	c.moveNodeToFn(n, c.frequencyList)

	c.numItems++
}

// Delete an item from the cache.
func (c *Cache) Delete(key string) {
	n, ok := c.index[key]
	if ok {
		c.deleteNode(n)
	}
}

// Access an item in the cache.
// Increases the items use count.
func (c *Cache) Access(key string) interface{} {
	n, ok := c.index[key]
	if !ok {
		return nil
	}

	nextUsage := n.parent.usage + 1
	var nextFn *frequencyNode
	if n.parent.next == nil || n.parent.next.usage != nextUsage {
		nextFn = c.newFrequencyNode(nextUsage, n.parent, n.parent.next)
	} else {
		nextFn = n.parent.next
	}

	c.moveNodeToFn(n, nextFn)

	return n.value
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

	if n.parent.nodeList == nil {
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

	n.parent = fn
	n.next = fn.nodeList
	if n.next != nil {
		n.next.prev = n
	}
	fn.nodeList = n
}
