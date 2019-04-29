package vedux

import (
	"sync"
)

// topic maintains a collection of subscribers
type topic struct {
	mutex *sync.RWMutex
	first *Subscription
	last  *Subscription
}

func newTopic() *topic {
	return &topic{mutex: new(sync.RWMutex)}
}

// subscribe returns a new Subscription
func (t *topic) subscribe() *Subscription {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	sub := newSubscription(t)
	if t.first == nil && t.last == nil {
		t.first = sub
		t.last = sub
	} else {
		sub.prev = t.last
		t.last.next = sub
		t.last = sub
	}

	go sub.forward()

	return sub
}

func (t *topic) unsubscribe(sub *Subscription) {
	if sub.topic != t {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	sub.shutdown()

	if sub.next == nil {
		t.last = sub.prev
	} else {
		sub.next.prev = sub.prev
	}
	if sub.prev == nil {
		t.first = sub.next
	} else {
		sub.prev.next = sub.next
	}
}

func (t *topic) notify(v interface{}) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for sub := t.first; sub != nil; sub = sub.next {
		sub.notify(v)
	}
}

func (t *topic) shutdown() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for sub := t.first; sub != nil; sub = sub.next {
		sub.shutdown()
	}

	t.first = nil
	t.last = nil
}

type Subscription struct {
	pub     chan interface{}
	c       chan interface{}
	topic   *topic
	closeFn *sync.Once
	next    *Subscription
	prev    *Subscription
}

func newSubscription(topic *topic) *Subscription {
	return &Subscription{
		pub:     make(chan interface{}, 16),
		c:       make(chan interface{}),
		closeFn: new(sync.Once),
		topic:   topic,
	}
}

func (s *Subscription) forward() {
	defer close(s.c)

	for v := range s.pub {
		s.c <- v
	}
}

func (s *Subscription) Chan() <-chan interface{} {
	return s.c
}

func (s *Subscription) notify(data interface{}) {
	s.pub <- data
}

func (s *Subscription) Close() {
	s.topic.unsubscribe(s)
}

func (s *Subscription) shutdown() {
	s.closeFn.Do(func() { close(s.pub) })
}
