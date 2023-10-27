package stream

import (
	"errors"
	"fmt"
	"sync"
)

/*
Implements:
  - ChanneledInput
*/
type channeledInput[T any] struct {
	closed bool
	DefaultProducer[T]
	channel     chan T
	closingLock sync.Mutex
}

/*
NewChanneledInput is a constructor of the channeled input.

Type parameters:
  - T - type of the produced values.

Parameters:
  - capacity - size of the channel buffer.

Returns:
  - pointer to the new channeled input.
*/
func NewChanneledInput[T any](capacity int) ChanneledInput[T] {
	ego := &channeledInput[T]{channel: make(chan T, capacity), closingLock: sync.Mutex{}}
	ego.DefaultProducer = *NewDefaultProducer[T](ego)
	return ego
}

func (ego *channeledInput[T]) Channel() chan T {
	return ego.channel
}

func (ego *channeledInput[T]) Get() (value T, valid bool, err error) {
	value, valid = <-ego.channel
	return
}

func (ego *channeledInput[T]) Close() {
	ego.closingLock.Lock()
	if !ego.closed {
		close(ego.channel)
		ego.closed = true
	}
	ego.closingLock.Unlock()
}

func (ego *channeledInput[T]) Closed() bool {
	return ego.closed && len(ego.channel) == 0
}

func (ego *channeledInput[T]) Write(values ...T) (n int, err error) {

	if values == nil {
		return 0, errors.New("input slice is not initialized")
	}

	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case error:
				err = e
			default:
				err = errors.New(fmt.Sprint(e))
			}
		}
	}()

	for _, v := range values {
		ego.Channel() <- v
	}

	n = len(values)
	return

}
