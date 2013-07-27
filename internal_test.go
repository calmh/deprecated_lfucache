package lfucache

import "testing"

func numFrequencyNodes(c *Cache) int {
	count := 0
	for fn := c.frequencyList; fn != nil; fn = fn.next {
		count++
	}
	return count
}

func TestMinimalFrequencyNodesDuringAccess(t *testing.T) {
	c := Create(10)
	c.Insert("test1", 42) // usage=1
	c.Insert("test2", 43) // usage=1
	c.Insert("test3", 44) // usage=1
	c.Insert("test4", 45) // usage=1

	if n := numFrequencyNodes(c); n != 1 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=2
	c.Access("test2") // usage=2
	c.Access("test3") // usage=2
	c.Access("test4") // usage=2

	if n := numFrequencyNodes(c); n != 2 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=3
	c.Access("test2") // usage=3

	if n := numFrequencyNodes(c); n != 3 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test3") // usage=3
	c.Access("test4") // usage=3

	if n := numFrequencyNodes(c); n != 2 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}
}

func TestMinimalFrequencyNodesDuringDelete1(t *testing.T) {
	c := Create(10)
	c.Insert("test1", 42) // usage=1
	c.Insert("test2", 43) // usage=1
	c.Insert("test3", 44) // usage=1
	c.Insert("test4", 45) // usage=1

	if n := numFrequencyNodes(c); n != 1 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=2
	c.Access("test2") // usage=2
	c.Access("test3") // usage=2
	c.Access("test4") // usage=2

	if n := numFrequencyNodes(c); n != 2 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=3
	c.Access("test2") // usage=3

	if n := numFrequencyNodes(c); n != 3 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Delete("test1")
	c.Delete("test2")

	if n := numFrequencyNodes(c); n != 2 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}
}

func TestMinimalFrequencyNodesDuringDelete2(t *testing.T) {
	c := Create(10)
	c.Insert("test1", 42) // usage=1
	c.Insert("test2", 43) // usage=1
	c.Insert("test3", 44) // usage=1
	c.Insert("test4", 45) // usage=1

	if n := numFrequencyNodes(c); n != 1 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=2
	c.Access("test2") // usage=2
	c.Access("test3") // usage=2
	c.Access("test4") // usage=2

	if n := numFrequencyNodes(c); n != 2 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=3
	c.Access("test2") // usage=3

	if n := numFrequencyNodes(c); n != 3 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}

	c.Delete("test3")
	c.Delete("test4")

	if n := numFrequencyNodes(c); n != 2 {
		t.Errorf("Non-minimal number fo frequency nodes %d\n", n)
	}
}

