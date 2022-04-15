package arpcache

import (
	"net"
	"sync"
	"time"
)

const rowSize = 256
const numRows = 256

// ipToInd takes an IP address and returns the cache row and the offset into the
// row. We use the last octet as the cache row so that within a /24, each IP
// address gets its own mutex lock.
func ipToInd(ip net.IP) (row, ind int) {
	return int(ip[len(ip)-1]), int(ip[len(ip)-2])
}

func hwToBytes(hw net.HardwareAddr) [6]byte {
	hb := [6]byte{}
	copy(hb[:], hw[len(hw)-6:])
	return hb
}

type arpEntry struct {
	expires int64
	hw      [6]byte
}

type arpRow [rowSize]arpEntry

type cacheRow struct {
	mu  *sync.RWMutex
	row arpRow
}
type ArpCache struct {
	cache          [numRows]cacheRow
	defaultTimeout int64
}

func New(timeoutSeconds int64) *ArpCache {
	var cache [numRows]cacheRow
	return &ArpCache{cache: cache, defaultTimeout: timeoutSeconds}
}

func (a *ArpCache) SetDefaultTimeout(timeoutSeconds int64) {
	a.defaultTimeout = timeoutSeconds
}

func (a *ArpCache) Get(ip net.IP) (net.HardwareAddr, bool) {
	i, j := ipToInd(ip)

	a.cache[i].mu.RLock()
	entry := a.cache[i].row[j]
	a.cache[i].mu.RUnlock()
	if entry.expires < time.Now().Unix() {
		var hw net.HardwareAddr
		return hw, false
	}
	return net.HardwareAddr(entry.hw[:]), true
}

func (a *ArpCache) Set(ip net.IP, hw net.HardwareAddr) {
	i, j := ipToInd(ip)
	a.cache[i].mu.Lock()
	defer a.cache[i].mu.Unlock()
	entry := a.cache[i].row[j]
	entry.hw = hwToBytes(hw)
	entry.expires = time.Now().Unix() + a.defaultTimeout
}

func (a *ArpCache) SetExpiry(ip net.IP, epoch int64) bool {
	i, j := ipToInd(ip)
	a.cache[i].mu.Lock()
	defer a.cache[i].mu.Unlock()
	entry := a.cache[i].row[j]
	entry.expires = epoch
	return true
}

func (a *ArpCache) Delete(ip net.IP) bool {
	return a.SetExpiry(ip, 0)
}
