package arpcache

import (
	"encoding/binary"
	"math/rand"
	"net"
	"testing"
	"time"
)

func benchmarkArpCacheSet(size int, b *testing.B) {
	m := New(30)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := uint32(0); i < uint32(size); i++ {
			// bs := make([]byte, 4)
			// binary.BigEndian.PutUint32(bs, i)
			// hw := make([]byte, 6)
			// binary.BigEndian.PutUint32(bs, i)
			//
			m.Set(net.IP{}, net.HardwareAddr{})
		}
	}
}

func benchmarkMapSet(size int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		// m := map[int]int{}
		m := make(map[int]int, size*2)
		for i := 0; i < size; i++ {
			m[i] = i
		}
	}
}

func benchmarkArpCacheGet(size int, b *testing.B) {
	m := New(30)
	accesses := make([]net.IP, size*2)
	for i := uint32(0); i < uint32(size); i++ {
		bs := make([]byte, 4)
		binary.BigEndian.PutUint32(bs, i)
		hw := make([]byte, 6)
		binary.BigEndian.PutUint32(bs, i)
		m.Set(bs, hw)
		accesses[int(i)] = bs
		bs[0] = 0xff
		accesses[size+int(i)] = bs
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(accesses), func(i, j int) { accesses[i], accesses[j] = accesses[j], accesses[i] })
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, i := range accesses {
			_, _ = m.Get(i)
		}
	}
}

func benchmarkMapGet(size int, b *testing.B) {
	m := map[int]int{}
	accesses := make([]int, size*2)
	for i := 0; i < size; i++ {
		m[i] = i
		accesses[i] = i
		accesses[size+i] = size + i
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(accesses), func(i, j int) { accesses[i], accesses[j] = accesses[j], accesses[i] })
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := range accesses {
			_ = m[i]
		}
	}
}

func BenchmarkArpCacheSet1(b *testing.B) { benchmarkArpCacheSet(10, b) }

func BenchmarkArpCacheSet2(b *testing.B) { benchmarkArpCacheSet(100, b) }
func BenchmarkArpCacheSet3(b *testing.B) { benchmarkArpCacheSet(1000, b) }
func BenchmarkArpCacheSet4(b *testing.B) { benchmarkArpCacheSet(10000, b) }
func BenchmarkArpCacheSet5(b *testing.B) { benchmarkArpCacheSet(100000, b) }
func BenchmarkArpCacheSet6(b *testing.B) { benchmarkArpCacheSet(1000000, b) }

func BenchmarkMapSet1(b *testing.B) { benchmarkMapSet(10, b) }
func BenchmarkMapSet2(b *testing.B) { benchmarkMapSet(100, b) }
func BenchmarkMapSet3(b *testing.B) { benchmarkMapSet(1000, b) }
func BenchmarkMapSet4(b *testing.B) { benchmarkMapSet(10000, b) }
func BenchmarkMapSet5(b *testing.B) { benchmarkMapSet(100000, b) }
func BenchmarkMapSet6(b *testing.B) { benchmarkMapSet(1000000, b) }

func BenchmarkArpCacheGet1(b *testing.B) { benchmarkArpCacheGet(10, b) }
func BenchmarkArpCacheGet2(b *testing.B) { benchmarkArpCacheGet(100, b) }
func BenchmarkArpCacheGet3(b *testing.B) { benchmarkArpCacheGet(1000, b) }
func BenchmarkArpCacheGet4(b *testing.B) { benchmarkArpCacheGet(10000, b) }
func BenchmarkArpCacheGet5(b *testing.B) { benchmarkArpCacheGet(100000, b) }
func BenchmarkArpCacheGet6(b *testing.B) { benchmarkArpCacheGet(1000000, b) }

func BenchmarkMapGet1(b *testing.B) { benchmarkMapGet(10, b) }
func BenchmarkMapGet2(b *testing.B) { benchmarkMapGet(100, b) }

func BenchmarkMapGet3(b *testing.B) { benchmarkMapGet(1000, b) }
func BenchmarkMapGet4(b *testing.B) { benchmarkMapGet(10000, b) }
func BenchmarkMapGet5(b *testing.B) { benchmarkMapGet(100000, b) }
func BenchmarkMapGet6(b *testing.B) { benchmarkMapGet(1000000, b) }
