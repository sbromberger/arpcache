package arpcache

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const rowSize = 256
const numRows = 256
const tickerDivisor = 10

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

type arpRow [rowSize]arpEntry

// cacheRow has its own mutex so we can hit multiple rows concurrently.
type cacheRow struct {
	mu  sync.RWMutex
	row arpRow
}
type ArpCache struct {
	cache          [numRows]cacheRow
	defaultTimeout int64     // timeout in seconds
	epoch          int64     // keep track of the epoch time here to reduce time.Now() calls
	tickerCtl      chan bool // send true to stop ticker at end
}

// New creates a new ArpCache.
func New(timeoutSeconds int64) *ArpCache {
	n := timeoutSeconds / tickerDivisor
	d, _ := time.ParseDuration(fmt.Sprintf("%ds", n))
	ticker := time.NewTicker(d)
	tCtl := make(chan bool)
	a := &ArpCache{cache: [numRows]cacheRow{}, defaultTimeout: timeoutSeconds, epoch: time.Now().Unix(), tickerCtl: tCtl}
	go func() {
		for {
			select {
			case <-tCtl:
				return
			case <-ticker.C:
				atomic.StoreInt64(&a.epoch, time.Now().Unix())
			}
		}
	}()
	return a
}

// Stop stops the timer subroutine.
func (a *ArpCache) Stop() {
	a.tickerCtl <- true
}

// SetDefaultTimeout sets the default timeout in seconds for an ArpCache.
func (a *ArpCache) SetDefaultTimeout(timeoutSeconds int64) {
	a.defaultTimeout = timeoutSeconds
}

// Get returns a hardware address (and a boolean indicating whether found) given an IP address.
// Get will return false if the entry is expired.
func (a *ArpCache) Get(ip net.IP) (net.HardwareAddr, bool) {
	i, j := ipToInd(ip)

	a.cache[i].mu.RLock()
	entry := a.cache[i].row[j]
	a.cache[i].mu.RUnlock()
	if entry.expires < a.epoch {
		var hw net.HardwareAddr
		return hw, false
	}
	return net.HardwareAddr(entry.hw[:]), true
}

// Set assigns a hardware address to a given IP address and sets the expiration time.]
func (a *ArpCache) Set(ip net.IP, hw net.HardwareAddr) {
	i, j := ipToInd(ip)
	a.cache[i].mu.Lock()
	defer a.cache[i].mu.Unlock()
	entry := a.cache[i].row[j]
	entry.hw = hwToBytes(hw)
	entry.expires = a.epoch + a.defaultTimeout
}

// SetExpiry updates / changes the expiration time of a cache entry given its IP address.
func (a *ArpCache) SetExpiry(ip net.IP, epoch int64) bool {
	i, j := ipToInd(ip)
	a.cache[i].mu.Lock()
	defer a.cache[i].mu.Unlock()
	entry := a.cache[i].row[j]
	entry.expires = epoch
	return true
}

// Delete invalidates a cache entry by setting its expiration time to 0.
func (a *ArpCache) Delete(ip net.IP) bool {
	return a.SetExpiry(ip, 0)
}
