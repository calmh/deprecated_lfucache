package lfucache

import (
	"fmt"
	"testing"
)

func TestMinimalFrequencyNodesDuringAccess(t *testing.T) {
	c := New(10)
	c.EnableChecking()

	c.Insert("test1", 42) // usage=1
	c.Insert("test2", 43) // usage=1
	c.Insert("test3", 44) // usage=1
	c.Insert("test4", 45) // usage=1

	if n := c.numFrequencyNodes(); n != 1 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=2
	c.Access("test2") // usage=2
	c.Access("test3") // usage=2
	c.Access("test4") // usage=2

	if n := c.numFrequencyNodes(); n != 2 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=3
	c.Access("test2") // usage=3

	if n := c.numFrequencyNodes(); n != 3 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test3") // usage=3
	c.Access("test4") // usage=3

	if n := c.numFrequencyNodes(); n != 2 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}
}

func TestMinimalFrequencyNodesDuringDelete1(t *testing.T) {
	c := New(10)
	c.EnableChecking()

	c.Insert("test1", 42) // usage=1
	c.Insert("test2", 43) // usage=1
	c.Insert("test3", 44) // usage=1
	c.Insert("test4", 45) // usage=1

	if n := c.numFrequencyNodes(); n != 1 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=2
	c.Access("test2") // usage=2
	c.Access("test3") // usage=2
	c.Access("test4") // usage=2

	if n := c.numFrequencyNodes(); n != 2 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=3
	c.Access("test2") // usage=3

	if n := c.numFrequencyNodes(); n != 3 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Delete("test1")
	c.Delete("test2")

	if n := c.numFrequencyNodes(); n != 2 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}
}

func TestMinimalFrequencyNodesDuringDelete2(t *testing.T) {
	c := New(10)
	c.EnableChecking()

	c.Insert("test1", 42) // usage=1
	c.Insert("test2", 43) // usage=1
	c.Insert("test3", 44) // usage=1
	c.Insert("test4", 45) // usage=1

	if n := c.numFrequencyNodes(); n != 1 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=2
	c.Access("test2") // usage=2
	c.Access("test3") // usage=2
	c.Access("test4") // usage=2

	if n := c.numFrequencyNodes(); n != 2 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Access("test1") // usage=3
	c.Access("test2") // usage=3

	if n := c.numFrequencyNodes(); n != 3 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}

	c.Delete("test3")
	c.Delete("test4")

	if n := c.numFrequencyNodes(); n != 2 {
		t.Errorf("Non-minimal number of frequency nodes %d\n", n)
	}
}

func BenchmarkInsertUnsafe(b *testing.B) {
	c := New(b.N)
	c.threadUnsafe = true

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("k%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Insert(keys[i], i)
	}
}

func BenchmarkAccessUnsafe(b *testing.B) {
	c := New(b.N)
	c.threadUnsafe = true

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
