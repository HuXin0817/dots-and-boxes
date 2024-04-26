package pusher

import "time"

type Option[T any] func(*Pusher[T])

func WithPushLogic[T any](PushLogic func(...T) error) Option[T] {
	return func(p *Pusher[T]) {
		p.PushLogic = PushLogic
	}
}

func WithPushInterval[T any](PushInterval time.Duration) Option[T] {
	return func(p *Pusher[T]) {
		p.PushInterval = PushInterval
	}
}

func WithElements[T any](element ...T) Option[T] {
	return func(p *Pusher[T]) {
		p.MessagesBuffer = append(p.MessagesBuffer, element...)
	}
}
