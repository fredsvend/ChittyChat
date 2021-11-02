package service

import "sync/atomic"

type Clock struct {
	Counter uint64
}

type Time uint64

func NewClock() *Clock {
	return &Clock{
		Counter: 1,
	}
}

func (l *Clock) Time() Time {
	return Time(atomic.LoadUint64(&l.Counter))
}

func (l *Clock) Increment() Time {
	return Time(atomic.AddUint64(&l.Counter, 1))
}
