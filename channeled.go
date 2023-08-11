package stream

import (
	"errors"
	"fmt"
)

type channeledInput[T any] struct {
	DefaultClosable
	DefaultProducer[T]
	channel  chan T
	capacity int
}

func NewChanneledInput[T any](capacity int) *channeledInput[T] {
	ego := &channeledInput[T]{DefaultClosable: DefaultClosable{false}, channel: make(chan T, capacity), capacity: capacity}
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
	ego.DefaultClosable.Close()
}

func (ego *channeledInput[T]) Write(value ...T) (n int, err error) {
	if value == nil {
		return 0, errors.New("input slice is not initialized")
	}

	//defer to catch writing to closed channel or simmilar panicing problems (are there any simmilar?)
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

	for i, v := range value {
		ego.Channel() <- v
		n = i + 1
	}
	return len(value), err
}
