package lfucache_test

import (
	"fmt"
	"github.com/calmh/lfucache"
	"math/rand"
	// "runtime"
	// "sync"
	"testing"
	"testing/quick"
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

func TestExpireOldest(t *testing.T) {
	c := lfucache.New(3)

	c.Insert("test1", 42)
	c.Insert("test2", 43)
	c.Insert("test3", 44)
	c.Insert("test4", 45) // should remove test1 which is oldest

	if _, ok := c.Access("test1"); ok {
		t.Error("test1 was not removed")
	}
}

func TestResize(t *testing.T) {
	c := lfucache.New(10)

	c.Insert("test1", 42) // usage=0
	c.Access("test1")     // usage=1
	c.Access("test1")     // usage=2

	c.Insert("test2", 43) // usage=0

	c.Insert("test3", 44) // usage=0
	c.Access("test3")     // usage=1

	c.Insert("test4", 45) // usage=0

	if cp := c.Cap(); cp != 10 {
		t.Errorf("incorrect cap, %d", cp)
	}

	if ln := c.Len(); ln != 4 {
		t.Errorf("incorrect length, %d", ln)
	}

	if s := c.Statistics(); s.Evictions != 0 {
		t.Errorf("premature evictions, %d", s.Evictions)
	}

	c.Resize(2)

	if cp := c.Cap(); cp != 2 {
		t.Errorf("incorrect cap, %d", cp)
	}

	if ln := c.Len(); ln != 2 {
		t.Errorf("incorrect length, %d", ln)
	}

	if s := c.Statistics(); s.Evictions != 2 {
		t.Errorf("missed evictions, %d", s.Evictions)
	}

	if _, ok := c.Access("test2"); ok {
		t.Error("Node test2 was not removed")
	}

	if _, ok := c.Access("test4"); ok {
		t.Error("Node test4 was not removed")
	}

	if v, _ := c.Access("test1"); v.(int) != 42 {
		t.Error("Didn't get the right value back from the cache (test1)")
	}

	if v, _ := c.Access("test3"); v.(int) != 44 {
		t.Error("Didn't get the right value back from the cache (test3)")
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

	if c.Len() != 1 {
		t.Error("Unexpected size")
	}

	if v, ok := c.Access("test1"); !ok || v.(int) != 44 {
		t.Error("Incorrect entry")
	}

	c.Delete("test1")

	if c.Len() != 0 {
		t.Error("Unexpected size")
	}
}

func TestEvictionsChannel(t *testing.T) {
	c := lfucache.New(3)

	exp := make(chan interface{})
	c.Evictions(exp)

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

	c.Access("test1") // miss
	c.Access("test2") // miss

	c.Insert("test1", 42) // usage=0
	c.Access("test1")     // usage=1
	c.Access("test1")     // usage=2

	c.Insert("test2", 43) // usage=0

	c.Insert("test3", 44) // usage=0
	c.Access("test3")     // usage=1

	c.Access("test1") // usage=3
	c.Access("test2") // usage=1
	c.Access("test3") // usage=2

	// Will evict test2
	c.Insert("test4", 45) // usage=0

	c.Access("test2") // miss

	// Will evict test4
	c.Insert("test5", 45) // usage=0

	c.Delete("test1")
	c.Delete("test2")

	stats := c.Statistics()

	if stats.LenFreq0 != 1 {
		t.Errorf("Stats itemsfreq0 incorrect, %d", stats.LenFreq0)
	}
	if stats.Inserts != 5 {
		t.Errorf("Stats inserts incorrect, %d", stats.Inserts)
	}
	if stats.Hits != 6 {
		t.Errorf("Stats hits incorrect, %d", stats.Hits)
	}
	if stats.Misses != 3 {
		t.Errorf("Stats misses incorrect, %d", stats.Misses)
	}
	if stats.Evictions != 2 {
		t.Errorf("Stats evictions incorrect, %d", stats.Evictions)
	}
	if stats.Deletes != 1 {
		t.Errorf("Stats deletes incorrect, %d", stats.Deletes)
	}
	if stats.FreqListLen != 2 {
		t.Errorf("Stats freqlistlen incorrect, %d", stats.FreqListLen)
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

	err := quick.Check(func(key string, val int) bool {
		c.Insert(key, val)
		v, ok := c.Access(key)
		return ok && v.(int) == val
	}, &quick.Config{MaxCount: 100000})

	if err != nil {
		t.Error(err)
	}
}

// func TestParallellAccess(t *testing.T) {
// 	n := 5000
// 	k := 16
// 	m := 50

// 	c := lfucache.New(n)

// 	keys := make([]string, n)
// 	for i := 0; i < n; i++ {
// 		keys[i] = fmt.Sprintf("k%d", i)
// 	}

// 	runtime.GOMAXPROCS(k)
// 	var wg sync.WaitGroup
// 	wg.Add(k)

// 	for i := 0; i < k; i++ {
// 		go func() {
// 			for j := 0; j < n*m; j++ {
// 				idx := rand.Int31n(int32(n))
// 				v, ok := c.Access(keys[idx])
// 				if !ok {
// 					c.Insert(keys[idx], idx)
// 				} else if v != idx {
// 					t.Errorf("key mismatch %d != %d", v, idx)
// 				}
// 			}
// 			wg.Done()
// 		}()
// 	}
// 	wg.Wait()
// }

const cacheSize = 1e6

func BenchmarkInsertStr(b *testing.B) {
	c := lfucache.New(cacheSize)

	keys := make([]string, cacheSize)
	for i := 0; i < cacheSize; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Insert(keys[i%cacheSize], i)
	}
}

func BenchmarkAccessHitBestCaseStr(b *testing.B) {
	c := lfucache.New(cacheSize)

	keys := make([]string, cacheSize)
	for i := 0; i < cacheSize; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}
	for i := 0; i < cacheSize; i++ {
		c.Insert(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Access(keys[i%cacheSize])
	}
}

func BenchmarkAccessHitRandomStr(b *testing.B) {
	c := lfucache.New(cacheSize)

	keys := make([]string, cacheSize)
	for i := 0; i < cacheSize; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}
	for i := 0; i < cacheSize; i++ {
		c.Insert(keys[i], i)
	}

	indexes := make([]string, cacheSize)
	for i := 0; i < cacheSize; i++ {
		indexes[i] = keys[int(rand.Int31n(cacheSize))]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Access(indexes[i%cacheSize])
	}
}

func BenchmarkAccessHitRandomInt(b *testing.B) {
	c := lfucache.New(cacheSize)

	for i := 0; i < cacheSize; i++ {
		c.Insert(i, i)
	}

	indexes := make([]int, cacheSize)
	for i := 0; i < cacheSize; i++ {
		indexes[i] = int(rand.Int31n(cacheSize))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Access(indexes[i%cacheSize])
	}
}

func BenchmarkAccessHitWorstCaseStr(b *testing.B) {
	c := lfucache.New(cacheSize)

	keys := make([]string, cacheSize)
	for i := 0; i < cacheSize; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}
	for i := 0; i < cacheSize; i++ {
		c.Insert(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Access(keys[0])
	}
}

func BenchmarkAccessMissStr(b *testing.B) {
	c := lfucache.New(cacheSize)

	keys := make([]string, cacheSize)
	for i := 0; i < cacheSize; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Access(keys[i%cacheSize])
	}
}
