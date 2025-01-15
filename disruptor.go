package matcher

import (
	"runtime"
	"sync/atomic"
)

const (
	// 必须是2的幂
	defaultBufferSize = 1024
	// 用于CPU缓存行对齐，避免伪共享
	cacheLineSize = 64
)

// 对齐到缓存行的序号
type Sequence struct {
	value atomic.Int64
	pad   [cacheLineSize - 8]byte
}

type Disruptor struct {
	ringBuffer []*Event
	mask       int64
	size       int64

	// 生产者序号
	cursor Sequence
	// 消费者序号
	gating Sequence
}

func NewDisruptor(size int64) *Disruptor {
	if size < 1 {
		size = defaultBufferSize
	}
	// 确保 size 是2的幂
	size = roundUpToPowerOf2(size)

	return &Disruptor{
		ringBuffer: make([]*Event, size),
		mask:       size - 1,
		size:       size,
	}
}

// 发布事件
func (d *Disruptor) TryPublish(event *Event) bool {
	current := d.cursor.value.Load()
	next := current + 1
	wrap := next - d.size

	if wrap > d.gating.value.Load() {
		return false // buffer is full
	}

	d.ringBuffer[next&d.mask] = event
	d.cursor.value.Store(next)
	return true
}

// 消费事件
func (d *Disruptor) Process(handler func(*Event)) {
	var cursor int64
	for {
		current := d.cursor.value.Load()

		for i := cursor + 1; i <= current; i++ {
			event := d.ringBuffer[i&d.mask]
			handler(event)
		}

		cursor = current
		d.gating.value.Store(cursor)

		if cursor == current {
			runtime.Gosched() // 让出CPU
		}
	}
}

// 工具函数：将数字向上取整到最近的2的幂
func roundUpToPowerOf2(v int64) int64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}
