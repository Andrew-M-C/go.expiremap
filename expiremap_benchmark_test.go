package expiremap

import (
	"strconv"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	expire  = time.Second
	cleanup = 500 * time.Millisecond
	mask    = 0xFFF
)

func BenchmarkExpiremap(b *testing.B) {
	m := New(expire, cleanup)
	start := time.Now()

	var e bool
	b.ResetTimer()
	nonexistCount := 0

	for i := 0; i < b.N; i++ {
		k := strconv.FormatInt(int64(i&mask), 16)
		_, e = m.Load(k)
		if false == e {
			nonexistCount++
			m.Store(k, "Hello, expiremap")
		}
	}
	// b.Logf("Start %v", start)
	// b.Logf("Stop %v", time.Now())
	b.Logf("expiremap done, elapsed %v, b.N = %d, nonexist percentage %.4f%%", time.Now().Sub(start), b.N, float64(nonexistCount)/float64(b.N)*100)
}

func BenchmarkGoCache(b *testing.B) {
	c := cache.New(expire, cleanup)
	start := time.Now()

	var e bool
	b.ResetTimer()
	nonexistCount := 0

	for i := 0; i < b.N; i++ {
		k := strconv.FormatInt(int64(i&mask), 16)
		_, e = c.Get(k)
		if false == e {
			nonexistCount++
			c.Set(k, "Hello, go-cache!", 0)
		}
	}
	// b.Logf("Start %v", start)
	// b.Logf("Stop %v", time.Now())
	b.Logf("go-cache done, elapsed %v, b.N = %d, nonexist percentage %.4f%%", time.Now().Sub(start), b.N, float64(nonexistCount)/float64(b.N)*100)
}
