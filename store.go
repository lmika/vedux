package vedux

import "log"

type Store struct {
	data    *sharedMap
	topics  *sharedMap
	actions *sharedMap
}

func New() *Store {
	return &Store{
		data:    newSharedMap(),
		topics:  newSharedMap(),
		actions: newSharedMap(),
	}
}

func (s *Store) Has(key string) bool {
	_, hasVal := s.data.get(key)
	return hasVal
}

func (s *Store) Get(key string) interface{} {
	val, _ := s.data.get(key)
	return val
}

func (s *Store) put(key string, val interface{}) {
	s.data.put(key, val)

	t, hasTopic := s.topics.get(key)
	if !hasTopic {
		return
	}

	t.(*topic).notify(val)
}

func (s *Store) Observe(key string) *Subscription {
	t, _ := s.topics.getOrCalc(key, func() interface{} { return newTopic() })
	return t.(*topic).subscribe()
}

func (s *Store) On(action string, handler Handler) {
	s.actions.put(action, handler)
}

func (s *Store) Dispatch(action string, args ...interface{}) {
	h, hasAction := s.actions.get(action)
	if !hasAction {
		return
	}

	handler := h.(Handler)
	err := handler(&actionContext{args, s})
	if err != nil {
		// TODO: Handle errors
		log.Printf("dispatcher error %v: %v", action, err)
	}
}

func (s *Store) delete(key string) {
	s.data.delete(key)

	t, hasTopic := s.topics.get(key)
	if !hasTopic {
		return
	}

	t.(*topic).shutdown()
}
