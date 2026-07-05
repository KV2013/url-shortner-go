package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	p sync.Pool
}

func New[T Resetter](fn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{New: func() any { return fn() }},
	}
}

func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.p.Put(v)
}
