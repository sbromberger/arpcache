package arpcache

import (
	"net"
	"sync/atomic"
	"time"
	"unsafe"
)

const rowSize = 256
const numRows = 256
const timeoutDivisor = 10

// ipToInd takes an IP address and returns the cache row and the offset into the
// row. We use the last octet as the cache row so that within a /24, each IP
// address gets its own mutex lock.
func ipToInd(ip net.IP) (row, ind int) {
	if len(ip) < 2 {
		return 0, 0
	}
	return int(ip[len(ip)-1]), int(ip[len(ip)-2])
}

func hwToBytes(hw net.HardwareAddr) [6]byte {
	hb := [6]byte{}
	if len(hw) >= 6 {
		copy(hb[:], hw[len(hw)-6:])
	}
	return hb
}

type arpEntry struct {
	expires int64   // epoch time in seconds
	hw      [6]byte // 48-bit mac
}

type arpRow [rowSize]*arpEntry

// cacheRow has its own mutex so we can hit multiple rows concurrently.
type cacheRow struct {
	row arpRow
}

type ArpCache struct {
	cache          [numRows]cacheRow
	defaultTimeout int64 // timeout in seconds
	now            int64
	tickerStop     chan bool
}

// New creates a new ArpCache.
func New(timeoutSeconds int64) *ArpCache {

	// pre allocate each entry
	// nyquist limit to achieve second resolution
	ticker := time.NewTicker(time.Duration(timeoutSeconds/timeoutDivisor) * time.Second)

	done := make(chan bool)
	ac := &ArpCache{cache: [numRows]cacheRow{}, defaultTimeout: timeoutSeconds, tickerStop: done}
	for i := range ac.cache {
		for j := range ac.cache[i].row {
			ac.cache[i].row[j] = &arpEntry{}
		}
	}

	go func() {
		select {
		case <-done:
			return
		case tick := <-ticker.C:
			atomic.StoreInt64(&ac.now, tick.Unix())
		}
	}()

	return ac
}

func (a *ArpCache) Stop() {
	a.tickerStop <- true
}

// SetDefaultTimeout sets the default timeout in seconds for an ArpCache.
func (a *ArpCache) SetDefaultTimeout(timeoutSeconds int64) {
	a.defaultTimeout = timeoutSeconds
}

// Get returns a hardware address (and a boolean indicating whether found) given an IP address.
// Get will return false if the entry is expired.
func (a *ArpCache) Get(ip net.IP) (net.HardwareAddr, bool) {
	i, j := ipToInd(ip)

	target := (*unsafe.Pointer)(unsafe.Pointer(&a.cache[i].row[j]))
	entry := (*arpEntry)(atomic.LoadPointer(target))

	if entry.expires < atomic.LoadInt64(&a.now) {
		return net.HardwareAddr{}, false
	}
	return entry.hw[:], true
}

// Set assigns a hardware address to a given IP address and sets the expiration time.]
func (a *ArpCache) Set(ip net.IP, hw net.HardwareAddr) {
	i, j := ipToInd(ip)

	target := (*unsafe.Pointer)(unsafe.Pointer(&a.cache[i].row[j]))
	value := arpEntry{
		hw:      hwToBytes(hw),
		expires: a.now + a.defaultTimeout,
	}

	atomic.StorePointer(target, unsafe.Pointer(&value))
}
