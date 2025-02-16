package ring_buffer

import (
	"sync"
	"sync/atomic"
)

type buffer[T any] struct {
	items []T
	// head is the type's zero value. items start at the next index.
	// tail is the last inserted value.
	head, tail, size int64
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
			size:  size,
		},
	}
}

func (rb *RingBuffer[T]) Len() int64 {
	return rb.len.Load()
}

func (rb *RingBuffer[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	// tail[1] head[2]
	//     v  v
	// [1, 3, 0, 2, 5]
	// 2 = (1 + 1) % 5
	rb.buf.tail = (rb.buf.tail + 1) % rb.buf.size
	if rb.buf.tail == rb.buf.head {
		newSize := rb.buf.size * 2
		newItems := make([]T, newSize)
		for i := int64(0); i < rb.buf.size; i++ {
			idx := (rb.buf.tail + i) % rb.buf.size
			newItems[i] = rb.buf.items[idx]
		}
		buf := &buffer[T]{
			items: newItems,
			tail:  rb.buf.size,
			size:  newSize,
		}
		rb.buf = buf
	}
	rb.len.Add(1)
	rb.buf.items[rb.buf.tail] = item
}

func (rb *RingBuffer[T]) Pop() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	var t T
	if rb.Len() == 0 {
		return t, false
	}
	//  head[1]  tail[4]
	//     v        v
	// [0, 0, 6, 2, 5]
	// 2 = (1 + 1) % 5
	rb.buf.head = (rb.buf.head + 1) % rb.buf.size
	item := rb.buf.items[rb.buf.head]
	rb.buf.items[rb.buf.head] = t
	rb.len.Add(-1)
	return item, true
}

func (rb *RingBuffer[T]) PopN(n int64) ([]T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.Len() == 0 {
		return nil, false
	}
	if n > rb.Len() {
		n = rb.Len()
	}
	items := make([]T, n)
	// head[0] tail[3]
	//  v        v
	// [0, 4, 6, 2, 0]
	// 2 = (1 + 1) % 5
	for i := int64(0); i < n; i++ {
		pos := (rb.buf.head + 1 + i) % rb.buf.size
		items[i] = rb.buf.items[pos]
		var t T
		rb.buf.items[pos] = t
	}
	rb.buf.head = (rb.buf.head + n) % rb.buf.size
	rb.len.Add(-n)
	return items, true
}
