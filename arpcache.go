package arpcache

import (
	"net"
	"sync/atomic"
	"time"
	"unsafe"
)

const cacheSize = 256 * 256

// timer updates every timeoutSeconds /  timeoutDivisor seconds
const timeoutDivisor = 10

// ipToIndex takes an IP address and returns the offset into the row.
func ipToIndex(ip net.IP) int {
	if len(ip) < 2 {
		return 0
	}
	return int(ip[len(ip)-1])<<8 + int(ip[len(ip)-2])
}

// hwToBytes takes a HardwareAddr and returns the last 6 bytes as the
// MAC address.
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

type ArpCache struct {
	cache          [cacheSize]*arpEntry
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
	ac := &ArpCache{cache: [cacheSize]*arpEntry{}, defaultTimeout: timeoutSeconds, tickerStop: done}
	for i := range ac.cache {
		ac.cache[i] = &arpEntry{}
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

func (a *ArpCache) Now() int64 {
	return atomic.LoadInt64(&a.now)
}

// SetDefaultTimeout sets the default timeout in seconds for an ArpCache.
func (a *ArpCache) SetDefaultTimeout(timeoutSeconds int64) {
	a.defaultTimeout = timeoutSeconds
}

// Get returns a hardware address (and a boolean indicating whether found) given an IP address.
// Get will return false if the entry is expired.
func (a *ArpCache) Get(ip net.IP) (net.HardwareAddr, bool) {
	i := ipToIndex(ip)

	target := (*unsafe.Pointer)(unsafe.Pointer(&a.cache[i]))
	entry := (*arpEntry)(atomic.LoadPointer(target))

	if entry.expires < atomic.LoadInt64(&a.now) {
		return net.HardwareAddr{}, false
	}
	return entry.hw[:], true
}

// Set assigns a hardware address to a given IP address and sets the expiration time.
func (a *ArpCache) Set(ip net.IP, hw net.HardwareAddr) {
	i := ipToIndex(ip)

	target := (*unsafe.Pointer)(unsafe.Pointer(&a.cache[i]))
	value := arpEntry{
		hw:      hwToBytes(hw),
		expires: a.Now() + a.defaultTimeout,
	}

	atomic.StorePointer(target, unsafe.Pointer(&value))
}
