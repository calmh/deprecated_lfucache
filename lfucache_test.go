package lfucache_test

import (
	"github.com/calmh/lfucache"
	"testing"
	"testing/quick"
	"fmt"
)

func TestInstantiateCache(t *testing.T) {
	_ = lfucache.New(42)
}

func TestInsertAccess(t *testing.T) {
	c := lfucache.New(10)
	c.Insert("test", 42)
	v, _ := c.Access("test")
	if v.(int) != 42 {
		t.Error("Didn't get the right value back from the cache")
	}
}

func TestExpiry(t *testing.T) {
	c := lfucache.New(3)

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
	c := lfucache.New(3)

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

func TestDoubleInsert(t *testing.T) {
	c := lfucache.New(3)

	c.Insert("test1", 42)
	c.Insert("test1", 43)
	c.Insert("test1", 44)

	if c.Statistics().Items != 1 {
		t.Error("Unexpected size")
	}

	if v, ok := c.Access("test1"); !ok || v.(int) != 44 {
		t.Error("Incorrect entry")
	}

	c.Delete("test1")

	if c.Statistics().Items != 0 {
		t.Error("Unexpected size")
	}
}

func TestEvictionsChannel(t *testing.T) {
	c := lfucache.New(3)
	exp := c.Evictions()

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
					return
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
	// Will evict test2
	c.Insert("test4", 45) // usage=1
	<-done

	c.UnregisterEvictions(exp)
	// Will evict test3, there is noone listening on the expired channel
	c.Insert("test5", 45) // usage=1
}

func TestStats(t *testing.T) {
	c := lfucache.New(3)

	c.Access("test1")
	c.Access("test2")

	c.Insert("test1", 42) // usage=1
	c.Access("test1")
	c.Access("test1")

	c.Insert("test2", 43) // usage=1

	c.Insert("test3", 44) // usage=1
	c.Access("test3")

	c.Access("test1")
	c.Access("test2")
	c.Access("test3")

	// Will evict test2
	c.Insert("test4", 45) // usage=1

	c.Access("test2")

	// Will evict test3
	c.Insert("test5", 45) // usage=1

	c.Delete("test1")
	c.Delete("test2")

	stats := c.Statistics()

	if stats.Items != 2 {
		t.Error("Stats items incorrect", stats.Items)
	}
	if stats.ItemsFreq0 != 1 {
		t.Error("Stats itemsfreq0 incorrect")
	}
	if stats.Inserts != 5 {
		t.Error("Stats inserts incorrect")
	}
	if stats.Hits != 6 {
		t.Error("Stats hits incorrect")
	}
	if stats.Misses != 3 {
		t.Error("Stats misses incorrect")
	}
	if stats.Evictions != 2 {
		t.Error("Stats evictions incorrect")
	}
	if stats.Deletes != 1 {
		t.Error("Stats deletes incorrect")
	}
	if stats.FreqListLen != 2 {
		t.Error("Stats freqlistlen incorrect")
	}
}

func TestEvictIf(t *testing.T) {
	c := lfucache.New(10)

	c.Insert("test1", 42)
	c.Insert("test2", 43)
	c.Insert("test3", 44)
	c.Insert("test4", 45)
	c.Insert("test5", 46)

	ev := c.EvictIf(func(v interface{}) bool {
		return v.(int)%2 == 0
	})

	if ev != 3 {
		t.Error("Incorrect number of items evicted", ev)
	}

	if _, ok := c.Access("test1"); ok {
		t.Error("test1 not expected to exist")
	}
	if _, ok := c.Access("test2"); !ok {
		t.Error("test2 expected to exist")
	}
	if _, ok := c.Access("test3"); ok {
		t.Error("test3 not expected to exist")
	}
	if _, ok := c.Access("test4"); !ok {
		t.Error("test4 expected to exist")
	}
	if _, ok := c.Access("test5"); ok {
		t.Error("test5 not expected to exist")
	}
}

func TestRandomAccess(t *testing.T) {
	c := lfucache.New(1024)
	quick.Check(func(key string, val int) bool {
		c.Insert(key, val)
		c.Statistics()
		v, ok := c.Access(key)
		return ok && v.(int) == val
	}, &quick.Config{MaxCount: 100000})
}

func BenchmarkInsert(b *testing.B) {
	c := lfucache.New(b.N)
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Insert(keys[i], i)
	}
}

func BenchmarkAccess(b *testing.B) {
	c := lfucache.New(b.N)

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}
	for i := 0; i < b.N; i++ {
		c.Insert(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Access(keys[i])
	}
}

