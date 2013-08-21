package lfucache

import (
	"container/list"

	"fmt"
)

func (c *Cache) print() {
	fmt.Printf("C %+v\n", c)

	for fne := c.frequencyList.Front(); fne != nil; fne = fne.Next() {
		c.printFreqNode(fne)
	}
}

func (c *Cache) printFreqNode(fne *list.Element) {
	fmt.Printf("- FN %+v\n", fne)
	fmt.Printf("     %+v\n", fn(fne))
	for ne := fn(fne).nodeList.Front(); ne != nil; ne = ne.Next() {
		c.printNode(ne)
	}
}

func (c *Cache) printNode(ne *list.Element) {
	fmt.Printf("-- N %+v\n", ne)
	fmt.Printf("     %+v\n", n(ne))
}
