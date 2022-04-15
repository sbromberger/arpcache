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
	m.Stop()
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
	rand.Seed(2)
	rand.Shuffle(len(accesses), func(i, j int) { accesses[i], accesses[j] = accesses[j], accesses[i] })
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, i := range accesses {
			m.Get(i)
		}
	}
	m.Stop()
}

func ip2uint(i net.IP) uint32 {
	return binary.BigEndian.Uint32(i)
}

func uint2ip(n uint32) net.IP {
	ip := make([]byte, 4)
	ip[3] = byte(n)
	n = n >> 8
	ip[2] = byte(n)
	n = n >> 8
	ip[1] = byte(n)
	n = n >> 8
	ip[0] = byte(n)

	return net.IP(ip)
}

func benchmarkMapGet(size int, b *testing.B) {
	m := map[uint32]int{}
	accesses := make([]net.IP, size*2)
	for i := uint32(0); i < uint32(size); i++ {
		m[i] = int(i)
		accesses[i] = uint2ip(i)
		accesses[size+int(i)] = uint2ip(uint32(size) + i)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(accesses), func(i, j int) { accesses[i], accesses[j] = accesses[j], accesses[i] })
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, i := range accesses {
			_ = m[ip2uint(i)]
		}
	}
}

// func BenchmarkArpCacheSet1(b *testing.B) { benchmarkArpCacheSet(10, b) }
//
// func BenchmarkArpCacheSet2(b *testing.B) { benchmarkArpCacheSet(100, b) }
// func BenchmarkArpCacheSet3(b *testing.B) { benchmarkArpCacheSet(1000, b) }
// func BenchmarkArpCacheSet4(b *testing.B) { benchmarkArpCacheSet(10000, b) }
// func BenchmarkArpCacheSet5(b *testing.B) { benchmarkArpCacheSet(100000, b) }
// func BenchmarkArpCacheSet6(b *testing.B) { benchmarkArpCacheSet(1000000, b) }
//
// func BenchmarkMapSet1(b *testing.B) { benchmarkMapSet(10, b) }
// func BenchmarkMapSet2(b *testing.B) { benchmarkMapSet(100, b) }
// func BenchmarkMapSet3(b *testing.B) { benchmarkMapSet(1000, b) }
// func BenchmarkMapSet4(b *testing.B) { benchmarkMapSet(10000, b) }
// func BenchmarkMapSet5(b *testing.B) { benchmarkMapSet(100000, b) }
// func BenchmarkMapSet6(b *testing.B) { benchmarkMapSet(1000000, b) }
//
// func BenchmarkArpCacheGet1(b *testing.B) { benchmarkArpCacheGet(10, b) }
// func BenchmarkArpCacheGet2(b *testing.B) { benchmarkArpCacheGet(100, b) }
// func BenchmarkArpCacheGet3(b *testing.B) { benchmarkArpCacheGet(1000, b) }
// func BenchmarkArpCacheGet4(b *testing.B) { benchmarkArpCacheGet(10000, b) }
// func BenchmarkArpCacheGet5(b *testing.B) { benchmarkArpCacheGet(100000, b) }
func BenchmarkArpCacheGet6(b *testing.B) { benchmarkArpCacheGet(5000000, b) }

// func BenchmarkMapGet1(b *testing.B) { benchmarkMapGet(10, b) }
// func BenchmarkMapGet2(b *testing.B) { benchmarkMapGet(100, b) }
//
// func BenchmarkMapGet3(b *testing.B) { benchmarkMapGet(1000, b) }
// func BenchmarkMapGet4(b *testing.B) { benchmarkMapGet(10000, b) }
// func BenchmarkMapGet5(b *testing.B) { benchmarkMapGet(100000, b) }
func BenchmarkMapGet6(b *testing.B) { benchmarkMapGet(5000000, b) }

// func BenchmarkArpCacheGet7(b *testing.B) { benchmarkArpCacheGet(10_000_000, b) }
