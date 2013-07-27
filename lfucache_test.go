package lfucache_test

import (
	"github.com/calmh/lfucache"
	"testing"
)

func TestInstantiateCache(t *testing.T) {
	_ = lfucache.Create(42)
}

func TestInsertAccess(t *testing.T) {
	c := lfucache.Create(10)
	c.Insert("test", 42)
	v, _ := c.Access("test")
	if v.(int) != 42 {
		t.Error("Didn't get the right value back from the cache")
	}
}

func TestExpiry(t *testing.T) {
	c := lfucache.Create(3)

	c.Insert("test1", 42) // usage=1
	c.Access("test1")     // usage=2
	c.Access("test1")     // usage=3

	c.Insert("test2", 43) // usage=1

	c.Insert("test3", 44) // usage=1
	c.Access("test3")     // usage=2

	if v, _ := c.Access("test1"); v.(int) != 42 {
		t.Error("Didn't get the right value back from the cache (test1)")
	}

	if v, _ := c.Access("test2"); v.(int) != 43 {
		t.Error("Didn't get the right value back from the cache (test2)")
	}

	if v, _ := c.Access("test3"); v.(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}

	c.Insert("test4", 45) // usage=1, should remove test2 which is lfu

	if v, _ := c.Access("test1"); v.(int) != 42 {
		t.Error("Didn't get the right value back from the cache (test1)")
	}

	if _, ok := c.Access("test2"); ok {
		t.Error("Node test2 was not removed")
	}

	if v, _ := c.Access("test3"); v.(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}

	if v, _ := c.Access("test4"); v.(int) != 45 {
		t.Error("Didn't get the right value back from the cache (test4)")
	}
}

func TestDelete(t *testing.T) {
	c := lfucache.Create(3)

	c.Insert("test1", 42) // usage=1
	c.Access("test1")     // usage=2
	c.Access("test1")     // usage=3

	c.Insert("test2", 43) // usage=1

	c.Insert("test3", 44) // usage=1
	c.Access("test3")     // usage=2

	c.Delete("test1")

	if _, ok := c.Access("test1"); ok {
		t.Error("test1 was not deleted")
	}

	if v, _ := c.Access("test2"); v.(int) != 43 {
		t.Error("Didn't get the right value back from the cache (test2)")
	}

	if v, _ := c.Access("test3"); v.(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}
}

func TestSize(t *testing.T) {
	c := lfucache.Create(3)

	c.Insert("test1", 42)
	c.Insert("test2", 43)
	c.Insert("test3", 44)
	c.Insert("test4", 45) // test3 is deleted
	c.Delete("test1")

	if c.Size() != 2 {
		t.Error("Unexpected size")
	}
}

func TestDoubleInsert(t *testing.T) {
	c := lfucache.Create(3)

	c.Insert("test1", 42)
	c.Insert("test1", 43)
	c.Insert("test1", 44)

	if c.Size() != 1 {
		t.Error("Unexpected size")
	}

	c.Delete("test1")

	if c.Size() != 0 {
		t.Error("Unexpected size")
	}
}

func TestExpiredChannel(t *testing.T) {
	c := lfucache.Create(3)
	exp := c.Expired()

	start := make(chan bool)
	done := make(chan bool)
	go func() {
		ready := false
		for {
			select {
			case e := <-exp:
				if !ready {
					t.Errorf("Unexpected expire %#v", e)
				} else if e.(int) != 43 {
					t.Errorf("Incorrect expire %#v", e)
				} else {
					done <- true
				}
			case <-start:
				ready = true
			}
		}
	}()


	c.Insert("test1", 42) // usage=1
	c.Access("test1")     // usage=2
	c.Access("test1")     // usage=3

	c.Insert("test2", 43) // usage=1

	c.Insert("test3", 44) // usage=1
	c.Access("test3")     // usage=2

	c.Access("test1")
	c.Access("test2")
	c.Access("test3")

	start <- true
	c.Insert("test4", 45) // usage=1
	<-done
}

func TestLFUPanic(t *testing.T) {
	c := lfucache.Create(0)
	defer func() {
		_ = recover()
	}()
	c.Insert("test1", 42)
	t.Error("Should not continue past error condition")
}
