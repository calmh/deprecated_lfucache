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
	v := c.Access("test")
	if v.(int) != 42 {
		t.Error("Didn't get the right value back from the cache")
	}
}

func TestExpiry(t *testing.T) {
	c := lfucache.Create(3)

	c.Insert("test1", 42) // usage=1
	c.Access("test1") // usage=2
	c.Access("test1") // usage=3

	c.Insert("test2", 43) // usage=1

	c.Insert("test3", 44) // usage=1
	c.Access("test3") // usage=2

	if c.Access("test1").(int) != 42 {
		t.Error("Didn't get the right value back from the cache (test1)")
	}

	if c.Access("test2").(int) != 43 {
		t.Error("Didn't get the right value back from the cache (test2)")
	}

	if c.Access("test3").(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}

	c.Insert("test4", 45) // usage=1, should remove test2 which is lfu

	if c.Access("test1").(int) != 42 {
		t.Error("Didn't get the right value back from the cache (test1)")
	}

	if c.Access("test2") != nil {
		t.Error("Node test2 was not removed")
	}

	if c.Access("test3").(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}

	if c.Access("test4").(int) != 45 {
		t.Error("Didn't get the right value back from the cache (test4)")
	}
}

func TestDelete(t *testing.T) {
	c := lfucache.Create(3)

	c.Insert("test1", 42) // usage=1
	c.Access("test1") // usage=2
	c.Access("test1") // usage=3

	c.Insert("test2", 43) // usage=1

	c.Insert("test3", 44) // usage=1
	c.Access("test3") // usage=2

	c.Delete("test1")

	if c.Access("test1") != nil{
		t.Error("test1 was not deleted")
	}

	if c.Access("test2").(int) != 43 {
		t.Error("Didn't get the right value back from the cache (test2)")
	}

	if c.Access("test3").(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}

	c.Insert("test4", 45) // usage=1

	if c.Access("test2").(int) != 43 {
		t.Error("Didn't get the right value back from the cache (test2)")
	}

	if c.Access("test3").(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
	}

	if c.Access("test4").(int) != 45 {
		t.Error("Didn't get the right value back from the cache (test4)")
	}
}
