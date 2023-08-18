package stream

import "errors"

/*
DefaultClosable is a default implementation of the closing mechanism.
It is a simple closed flag which can be set to true or checked.
To use it, just include the stuct as promoted field.

Implements:
  - Closable
*/
type DefaultClosable struct {
	closed bool
}

func (ego *DefaultClosable) Closed() bool {
	return ego.closed
}

func (ego *DefaultClosable) Close() {
	ego.closed = true
}

/*
DefaultConsumer is a default partial implementation of the Consumer.

Implements:
  - Consumer
*/
type DefaultConsumer[T any] struct {
	source Producer[T]
}

func (ego *DefaultConsumer[T]) Consume() (value T, valid bool, err error) {
	if ego.source == nil {
		return *new(T), false, errors.New("no source to consume from")
	}
	return ego.source.Get()
}

func (ego *DefaultConsumer[T]) SetSource(s Producer[T]) error {
	if !ego.CanSetSource() {
		return errors.New("the source has already been set")
	}
	ego.source = s
	return nil
}

func (ego *DefaultConsumer[T]) CanSetSource() bool {
	return ego.source == nil
}

/*
DefaultProducer is a default partial implementation of the Producer.
It does not itself implement Producer interface, the Get method has to be defined by the full implementation.
To use it, include the struct as promoted field and initialize it with the provided constructor.
All methods work with a pointer to the embedding struct.

Implements:
  - Producer (partially)
*/
type DefaultProducer[T any] struct {
	producer Producer[T]
	piped    bool
}

/*
NewDefaultProducer is a constructor of the DefaultProducer.

Type parameters:
  - T - type of the produced values.

Parameters:
  - p - full implementation of the Producer (embedding this struct).

Returns:
  - pointer to the new DefaultProducer.
*/
func NewDefaultProducer[T any](p Producer[T]) *DefaultProducer[T] {
	return &DefaultProducer[T]{producer: p}
}

func (ego *DefaultProducer[T]) Pipe(c Consumer[T]) Consumer[T] {

	if !c.CanSetSource() {
		panic("the consumer does not accept new sources")
	}

	if err := c.SetSource(ego.producer); err != nil {
		panic(err)
	}

	ego.piped = true
	return c

}

func (ego *DefaultProducer[T]) Read(dst []T) (int, error) {

	if ego.piped {
		return 0, errors.New("the stream is piped")
	}

	if dst == nil {
		return 0, errors.New("the input slice is not initialized")
	}

	n := len(dst)

	for i := 0; i < n; i++ {
		value, valid, err := ego.producer.Get()
		if err != nil || !valid {
			return i, err
		}
		dst[i] = value
	}

	return n, nil

}

func (ego *DefaultProducer[T]) Collect() ([]T, error) {

	if ego.piped {
		return nil, errors.New("the stream is piped")
	}

	output := make([]T, 0)

	for {
		value, valid, err := ego.producer.Get()
		if err != nil || !valid {
			return output, err
		}
		output = append(output, value)
	}

}

func (ego *DefaultProducer[T]) ForEach(fn func(T) error) error {

	if ego.piped {
		return errors.New("the stream is piped")
	}

	for {
		value, valid, err := ego.producer.Get()
		if err != nil || !valid {
			return err
		}
		if err := fn(value); err != nil {
			return err
		}
	}

}
