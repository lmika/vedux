package vedux

import "sync"

type sharedMap struct {
	mutex *sync.RWMutex
	data  map[string]interface{}
}

func newSharedMap() *sharedMap {
	return &sharedMap{new(sync.RWMutex), make(map[string]interface{})}
}

func (sm *sharedMap) get(key string) (interface{}, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	v, hasV := sm.data[key]
	return v, hasV
}

// getOrCalc checks if the value exists.  if it does returns the value + true.  Otherwise, it calls
// func, stores it as the value for the key and returns the value + false.
func (sm *sharedMap) getOrCalc(key string, calcFn func() interface{}) (interface{}, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	v, hasV := sm.data[key]
	if hasV {
		return v, true
	}

	v = calcFn()
	sm.data[key] = v

	return v, false
}

func (sm *sharedMap) put(key string, val interface{}) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.data[key] = val
}

func (sm *sharedMap) delete(key string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.data, key)
}
