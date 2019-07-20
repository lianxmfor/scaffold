package broadcaster

type Broadcaster struct {
	stop       chan struct{}
	input      chan interface{}
	register   chan chan<- interface{}
	unregister chan chan<- interface{}
	notifys    map[chan<- interface{}]bool
}

func NewBroadcaster(buffer int) *Broadcaster {
	b := &Broadcaster{
		stop:       make(chan struct{}),
		input:      make(chan interface{}, buffer),
		register:   make(chan chan<- interface{}),
		unregister: make(chan chan<- interface{}),
		notifys:    make(map[chan<- interface{}]bool),
	}
	go b.Run()
	return b
}
func (b *Broadcaster) Register(ch chan<- interface{}) {
	b.register <- ch
}

func (b *Broadcaster) UnRegister(ch chan<- interface{}) {
	b.unregister <- ch
}

func (b *Broadcaster) Notify(msg interface{}) {
	b.input <- msg
}

func (b *Broadcaster) Stop() {
	b.stop <- struct{}{}
}

func (b *Broadcaster) Run() {
	for {
		select {
		case <-b.stop:
			close(b.register)
			return
		case ch, ok := <-b.register:
			if !ok {
				return
			}
			b.notifys[ch] = true
		case ch, ok := <-b.unregister:
			if !ok {
				return
			}
			delete(b.notifys, ch)
		case msg := <-b.input:
			for ch := range b.notifys {
				ch <- msg
			}
		}
	}
}
