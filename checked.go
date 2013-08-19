// +build checked

package lfucache

func (c *Cache) check() {
	if c.numItems != len(c.index) {
		c.print()
		panic("bug: index/numItems mismatch")
	}

	count := 0
	var prevFn *frequencyNode
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		if fn.nodeList == nil && fn.usage != 0 {
			c.print()
			panic("bug: empty non-head frequency node")
		}
		if fn.prev != prevFn {
			c.print()
			panic("bug: incorrect prev frequencyNode pointer")
		}

		var prev *node
		for n := fn.nodeList; n != nil; n = n.next {
			if n.parent != fn {
				c.print()
				panic("bug: incorrect parent pointer")
			}
			if n.prev != prev {
				c.print()
				panic("bug: incorrect prev node pointer")
			}
			prev = n
			count++
		}

		prevFn = fn
	}

	if count != len(c.index) {
		c.print()
		panic("bug: index/item count mismatch")
	}
}
