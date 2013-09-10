package lfucache

func (c *Cache) check() {
	if c.length != len(c.index) {
		c.bug("index/numItems mismatch")
	}

	count := 0
	var prevFn *frequencyNode
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		if fn.head == nil && fn.usage != 0 {
			c.bug("empty non-head frequency node")
		}
		if fn.prev != prevFn {
			c.bug("incorrect prev frequencyNode pointer")
		}

		var prev *node
		for n := fn.head; n != nil; n = n.next {
			if n.parent != fn {
				c.bug("incorrect parent pointer")
			}
			if n.prev != prev {
				c.bug("incorrect prev node pointer")
			}
			prev = n
			count++

			if n.next == nil {
				if fn.tail != n {
					c.bug("tail pointer not pointing to last node")
				}
			}
		}

		prevFn = fn
	}

	if count != len(c.index) {
		c.bug("index/item count mismatch")
	}
}

func (c *Cache) bug(msg string) {
	c.print()
	panic("bug: " + msg)
}
