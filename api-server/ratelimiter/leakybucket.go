package ratelimiter

import (
	"math"
	"sync/atomic"
	"time"
)

//var F *os.File

type paddedUint64 struct {
	value uint64
	_     [56]byte
}

type LeakyBucket struct {
	burstCapacity float64
	leakRate      float64
	currentFill   paddedUint64
	lastLeakTime  paddedUint64
}

func NewLeakyBucket(leakRate, burstCapacity float64) *LeakyBucket {
	b := &LeakyBucket{
		leakRate:      leakRate,
		burstCapacity: burstCapacity,
	}

	now := time.Now().UnixNano()
	b.lastLeakTime.value = uint64(now)
	b.currentFill.value = 0
	return b
}

func (b *LeakyBucket) Allow() bool {
	const maxRetries = 100 // Prevent infinite loop under extreme contention

	for retries := 0; retries < maxRetries; retries++ {
		// Single time snapshot to avoid time skew
		now := time.Now()
		nowNano := uint64(now.UnixNano())

		lastFillBits := atomic.LoadUint64(&b.currentFill.value)
		lastFill := math.Float64frombits(lastFillBits)

		lastTimeNano := atomic.LoadUint64(&b.lastLeakTime.value)
		lastTime := time.Unix(0, int64(lastTimeNano))

		// Calculate leak using the single time snapshot
		elapsed := now.Sub(lastTime).Seconds()
		leaked := elapsed * b.leakRate

		// Use local variable instead of global
		newFill := math.Max(0, lastFill-leaked)

		if newFill < b.burstCapacity {
			newFill++
			newFillBits := math.Float64bits(newFill)

			if atomic.CompareAndSwapUint64(&b.currentFill.value, lastFillBits, newFillBits) {
				atomic.StoreUint64(&b.lastLeakTime.value, nowNano)
				AllowedTotal.Inc()
				return true
			}
			// CAS failed, retry
		} else {
			RejectedTotal.Inc()
			return false
		}
	}

	// Max retries exceeded - fail safe to reject
	RejectedTotal.Inc()
	return false
}

func (b *LeakyBucket) PeekFill() float64 {
	lastFillBits := atomic.LoadUint64(&b.currentFill.value)
	lastFill := math.Float64frombits(lastFillBits)
	lastTime := time.Unix(0, int64(atomic.LoadUint64(&b.lastLeakTime.value)))
	now := time.Now()
	elapsed := now.Sub(lastTime).Seconds()
	leaked := elapsed * b.leakRate
	return math.Max(0, lastFill-leaked)
}
