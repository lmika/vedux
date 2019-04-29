package vedux

type Handler func(actCtx ActionContext) error

type ActionContext interface {
	NArgs() int
	Arg(i int) interface{}

	Has(key string) bool
	Get(key string) interface{}
	Put(key string, value interface{})
	Delete(key string)
}

type actionContext struct {
	args  []interface{}
	store *Store
}

func (ac *actionContext) NArgs() int {
	return len(ac.args)
}

func (ac *actionContext) Arg(i int) interface{} {
	return ac.args[i]
}

func (ac *actionContext) Has(key string) bool {
	return ac.store.Has(key)
}

func (ac *actionContext) Get(key string) interface{} {
	return ac.store.Get(key)
}

func (ac *actionContext) Put(key string, value interface{}) {
	ac.store.put(key, value)
}

func (ac *actionContext) Delete(key string) {
	ac.store.delete(key)
}
