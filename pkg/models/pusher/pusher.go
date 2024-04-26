package pusher

import (
	"log"
	"sync"
	"time"
)

type Pusher[T any] struct {
	MessagesBuffer []T
	PushLogic      func(...T) error
	PushInterval   time.Duration
	ErrorHandler   func(error)
	lock           sync.Mutex
	running        bool
}

func NewPusher[T any](options ...Option[T]) (newPusher *Pusher[T]) {
	newPusher = &Pusher[T]{
		running:      true,
		PushLogic:    func(...T) error { return nil },
		ErrorHandler: func(err error) { log.Println(err) },
		PushInterval: time.Second,
	}

	for _, option := range options {
		option(newPusher)
	}

	return
}

func (p *Pusher[T]) PushAll() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := p.PushLogic(p.MessagesBuffer...); err != nil {
		return err
	}

	p.MessagesBuffer = []T{}
	return nil
}

func (p *Pusher[T]) AddMessages(messages ...T) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.MessagesBuffer = append(p.MessagesBuffer, messages...)
}

func (p *Pusher[T]) Start() {
	go func() {
		for p.running {
			timer := time.NewTimer(p.PushInterval)
			if err := p.PushAll(); err != nil {
				p.ErrorHandler(err)
			}
			<-timer.C
		}
	}()
}

func (p *Pusher[T]) Stop() {
	p.running = false
}
