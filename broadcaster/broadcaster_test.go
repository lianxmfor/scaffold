package broadcaster

import (
	"sync"
	"testing"
)

func TestBroadcaster_Notify(t *testing.T) {
	b := NewBroadcaster(1)
	ch := make(chan interface{})

	b.Register(ch)
	go func() {
		b.Notify(1)
	}()
	<-ch
}

func TestBroadcaster(t *testing.T) {
	wg := sync.WaitGroup{}
	b := NewBroadcaster(10)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		ch := make(chan interface{})
		b.Register(ch)
		go func() {
			defer wg.Done()
			defer b.UnRegister(ch)
			<-ch
		}()
	}
	b.Notify(1)
	wg.Wait()
}
