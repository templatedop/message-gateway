package ratelimiter

import (
	"math"
	"sync/atomic"
	"time"
)

//var F *os.File

var newFill float64

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
	now := time.Now()
	nowNano := uint64(now.UnixNano())

	for {
		lastFillBits := atomic.LoadUint64(&b.currentFill.value)
		lastFill := math.Float64frombits(lastFillBits)

		lastTime := time.Unix(0, int64(atomic.LoadUint64(&b.lastLeakTime.value)))
		now := time.Now()
		elapsed := now.Sub(lastTime).Seconds()
		leaked := elapsed * b.leakRate

		newFill = math.Max(0, lastFill-leaked)

		if newFill < b.burstCapacity {
			newFill++
			if atomic.CompareAndSwapUint64(&b.currentFill.value, lastFillBits, math.Float64bits(newFill)) {
				atomic.StoreUint64(&b.lastLeakTime.value, nowNano)
				AllowedTotal.Inc()

				return true
			}
		} else {
			RejectedTotal.Inc()
			return false
		}
	}
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
