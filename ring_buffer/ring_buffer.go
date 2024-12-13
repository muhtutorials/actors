package ring_buffer

import (
	"sync"
	"sync/atomic"
)

type buffer[T any] struct {
	items           []T
	head, tail, cap int64
}

type RingBuffer[T any] struct {
	buf *buffer[T]
	len atomic.Int64
	mu  sync.Mutex
}

func New[T any](size int64) *RingBuffer[T] {
	return &RingBuffer[T]{
		buf: &buffer[T]{
			items: make([]T, size),
			cap:   size,
		},
	}
}

func (rb *RingBuffer[T]) Len() int64 {
	return rb.len.Load()
}

func (rb *RingBuffer[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buf.tail = (rb.buf.tail + 1) % rb.buf.cap
	if rb.buf.tail == rb.buf.head {
		size := rb.buf.cap * 2
		newItems := make([]T, size)
		for i := int64(0); i < rb.buf.cap; i++ {
			idx := (rb.buf.tail + i) % rb.buf.cap
			newItems[i] = rb.buf.items[idx]
		}
		buf := &buffer[T]{
			items: newItems,
			tail:  rb.buf.cap,
			cap:   size,
		}
		rb.buf = buf
	}
	rb.len.Add(1)
	rb.buf.items[rb.buf.tail] = item
}

func (rb *RingBuffer[T]) Pop() (T, bool) {
	if rb.Len() == 0 {
		var t T
		return t, false
	}
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buf.head = (rb.buf.head + 1) % rb.buf.cap
	item := rb.buf.items[rb.buf.head]
	var t T
	rb.buf.items[rb.buf.head] = t
	rb.len.Add(-1)
	return item, true
}

func (rb *RingBuffer[T]) PopN(n int64) ([]T, bool) {
	if rb.Len() == 0 {
		return nil, false
	}
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if n >= rb.len.Load() {
		n = rb.len.Load()
	}
	items := make([]T, n)
	for i := int64(0); i < n; i++ {
		pos := (rb.buf.head + 1 + i) % rb.buf.cap
		items[i] = rb.buf.items[pos]
		var t T
		rb.buf.items[pos] = t
	}
	rb.buf.head = (rb.buf.head + n) % rb.buf.cap
	rb.len.Add(-n)
	return items, true
}
