package stream

import (
	"errors"
	"fmt"
)

type channeledInput[T any] struct {
	closed bool
	DefaultProducer[T]
	channel chan T
}

func NewChanneledInput[T any](capacity int) ChanneledInput[T] {
	ego := &channeledInput[T]{channel: make(chan T, capacity)}
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
	close(ego.channel)
	ego.closed = true
}

func (ego *channeledInput[T]) Closed() bool {
	return ego.closed && len(ego.channel) == 0
}

func (ego *channeledInput[T]) Write(value ...T) (n int, err error) {

	if value == nil {
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

	for _, v := range value {
		ego.Channel() <- v
	}

	n = len(value)
	return

}
