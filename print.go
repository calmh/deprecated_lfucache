package lfucache

import "fmt"

func (c *Cache) print() {
	fmt.Printf("C %+v\n", c)

	for fn := c.frequencyList; fn != nil; fn = fn.next {
		c.printFreqNode(fn)
	}
}

func (c *Cache) printFreqNode(fn *frequencyNode) {
	fmt.Printf("- FN %+v\n", fn)
	for n := fn.head; n != nil; n = n.next {
		c.printNode(n)
	}
}

func (c *Cache) printNode(n *node) {
	fmt.Printf("-- N %+v\n", n)
}
