/*
Package lfucache implements an O(1) LFU cache structure.

This structure is described in the paper "An O(1) algorithm for implementing
the LFU cache eviction scheme" by  K. Shah, A. Mitra and D.  Matani, August 16,
2010. It is based on two levels of doubly linked lists and gives O(1) insert,
access and delete operations.
*/
package lfucache

import "fmt"

// Cache is a LFU cache structure.
type Cache struct {
	size          int
	nItems        int
	frequenceNode *frequencyNode
	index         map[string]*node
}

type frequencyNode struct {
	usage     int
	prev      *frequencyNode
	next      *frequencyNode
	firstNode *node
}

type node struct {
	key           string
	value         interface{}
	frequencyNode *frequencyNode
	next          *node
	prev          *node
}

// Create a new LFU Cache structure.
// size is the maximum number of items contained in the cache.
func Create(size int) *Cache {
	c := Cache{}
	c.size = size
	c.index = make(map[string]*node)
	c.frequenceNode = &frequencyNode{1, nil, nil, nil}
	return &c
}

// Insert a new item into the cache.
func (c *Cache) Insert(key string, value interface{}) {
	if c.nItems == c.size {
		n := c.lfu()
		c.deleteNode(n)
	}

	// Create node
	n := &node{}
	n.key = key
	n.value = value

	// Insert into map
	c.index[key] = n

	// Insert into LFU Cache
	c.moveNodeToFn(n, c.frequenceNode)

	c.nItems++
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
	node, ok := c.index[key]
	if !ok {
		return nil
	}

	nextUsage := node.frequencyNode.usage + 1
	var nextFn *frequencyNode
	if node.frequencyNode.next == nil || node.frequencyNode.next.usage != nextUsage {
		nextFn = c.newFrequencyNode(nextUsage, node.frequencyNode, node.frequencyNode.next)
	} else {
		nextFn = node.frequencyNode.next
	}

	c.moveNodeToFn(node, nextFn)

	return node.value
}

func (c *Cache) deleteNode(n *node) {
	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}

	if n.frequencyNode.firstNode == n {
		n.frequencyNode.firstNode = n.next
	}

	delete(c.index, n.key)
	c.nItems--

	if n.frequencyNode.firstNode == nil {
		c.deleteFrequencyNode(n.frequencyNode)
	}
}

func (c *Cache) lfu() *node {
	for fn := c.frequenceNode; fn != nil; fn = fn.next {
		if fn.firstNode != nil {
			return fn.firstNode
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

	if n.frequencyNode != nil && n.frequencyNode.firstNode == n {
		n.frequencyNode.firstNode = n.next
	}

	n.frequencyNode = fn
	n.next = fn.firstNode
	if n.next != nil {
		n.next.prev = n
	}
	fn.firstNode = n
}

// Debug

func (c *Cache) Print() {
	fmt.Printf("C %#v\n", c)

	for fn := c.frequenceNode; fn != nil; fn = fn.next {
		c.printFreqNode(fn)
	}
}

func (c *Cache) printFreqNode(fn *frequencyNode) {
	fmt.Printf(" FN %#v\n", fn)
	for n := fn.firstNode; n != nil; n = n.next {
		c.printNode(n)
	}
}

func (c *Cache) printNode(n *node) {
	fmt.Printf("  N %#v\n", n)
}
