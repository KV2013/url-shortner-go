package pool

import (
	"reflect"
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	p sync.Pool
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				var zero T
				rt := reflect.TypeOf(&zero).Elem()
				if rt.Kind() == reflect.Pointer {
					return reflect.New(rt.Elem()).Interface()
				}
				return reflect.New(rt).Elem().Interface()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.p.Put(v)
}
