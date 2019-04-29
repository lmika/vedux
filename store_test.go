package vedux

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
)

func TestStore_GetPut(t *testing.T) {
	store := New()
	store.On("test", func(ctx ActionContext) error {
		ctx.Put("value", 123)
		return nil
	})

	then.AssertThat(t, store.Has("value"), is.False())
	then.AssertThat(t, store.Get("value"), is.Nil())

	store.Dispatch("test")

	then.AssertThat(t, store.Has("value"), is.True())
	then.AssertThat(t, store.Get("value"), is.EqualTo(123))
}

func TestStore_Observe(t *testing.T) {
	store := New()
	store.On("test", func(ctx ActionContext) error {
		ctx.Put("value", 123)
		return nil
	})

	seenValues, _ := observeAndForward(store, "value")

	store.Dispatch("test")

	then.AssertThat(t, valuesOfChan(t, seenValues, 1), is.EqualTo([]interface{}{123}))
}

func TestStore_DispatchArgs(t *testing.T) {
	store := New()
	store.On("test", func(ctx ActionContext) error {
		ctx.Put("value", ctx.Arg(0))
		return nil
	})

	seenValues, _ := observeAndForward(store, "value")

	store.Dispatch("test", 111)
	store.Dispatch("test", 222)
	store.Dispatch("test", 333)

	then.AssertThat(t, valuesOfChan(t, seenValues, 3), is.EqualTo([]interface{}{111, 222, 333}))
}

func TestStore_MultiObservers(t *testing.T) {
	store := New()
	store.On("test", func(ctx ActionContext) error {
		ctx.Put("value", ctx.Arg(0))
		ctx.Put("value2", ctx.Arg(0).(int)*2)
		return nil
	})

	seenValues1a, _ := observeAndForward(store, "value")
	seenValues1b, _ := observeAndForward(store, "value")
	seenValues2, _ := observeAndForward(store, "value2")

	store.Dispatch("test", 111)
	store.Dispatch("test", 222)
	store.Dispatch("test", 333)

	then.AssertThat(t, valuesOfChan(t, seenValues1a, 3), is.EqualTo([]interface{}{111, 222, 333}))
	then.AssertThat(t, valuesOfChan(t, seenValues1b, 3), is.EqualTo([]interface{}{111, 222, 333}))
	then.AssertThat(t, valuesOfChan(t, seenValues2, 3), is.EqualTo([]interface{}{222, 444, 666}))
}

func TestStore_Unsubscribe(t *testing.T) {
	store := New()
	store.On("test", func(ctx ActionContext) error {
		ctx.Put("value", ctx.Arg(0))
		return nil
	})

	seenValues1a, sub1a := observeAndForward(store, "value")
	seenValues1b, _ := observeAndForward(store, "value")

	store.Dispatch("test", 111)
	store.Dispatch("test", 222)

	sub1a.Close()

	store.Dispatch("test", 333)
	store.Dispatch("test", 444)

	then.AssertThat(t, valuesOfChan(t, seenValues1a, 2), is.EqualTo([]interface{}{111, 222}))
	then.AssertThat(t, valuesOfChan(t, seenValues1b, 4), is.EqualTo([]interface{}{111, 222, 333, 444}))
}

func TestStore_Delete(t *testing.T) {
	store := New()
	store.On("test", func(ctx ActionContext) error {
		ctx.Put("value", ctx.Arg(0))
		return nil
	})
	store.On("deleteValue", func(ctx ActionContext) error {
		ctx.Delete("value")
		return nil
	})

	seenValues1a, sub1a := observeAndForward(store, "value")
	seenValues1b, _ := observeAndForward(store, "value")

	store.Dispatch("test", 111)
	store.Dispatch("test", 222)
	store.Dispatch("deleteValue")

	sub1a.Close()

	then.AssertThat(t, valuesOfChan(t, seenValues1a, 2), is.EqualTo([]interface{}{111, 222}))
	then.AssertThat(t, valuesOfChan(t, seenValues1b, 4), is.EqualTo([]interface{}{111, 222}))
}

func observeAndForward(store *Store, key string) (chan interface{}, *Subscription) {
	sub := store.Observe(key)
	seenValues := make(chan interface{})

	go func() {
		defer close(seenValues)

		for v := range sub.Chan() {
			seenValues <- v
		}
	}()

	return seenValues, sub
}

func valuesOfChan(t *testing.T, c chan interface{}, waitFor int) []interface{} {
	res := make([]interface{}, 0)
	for i := 0; i < waitFor; i++ {
		select {
		case v, ok := <-c:
			if !ok {
				t.Logf("channel has been closed")
				return res
			}
			res = append(res, v)
		case <-time.After(5 * time.Second):
			t.Errorf("timeout waiting for value")
			return res
		}
	}

	return res
}
