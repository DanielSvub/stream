package stream

type filter[T any] struct {
	DefaultConsumer[T]
	DefaultProducer[T]
	DefaultClosable
	filter func(T) bool
}

func NewFilter[T any](fn func(T) bool) Filter[T] {
	ego := &filter[T]{filter: fn}
	ego.DefaultProducer = *NewDefaultProducer[T](ego)
	return ego
}

func (ego *filter[T]) Get() (value T, valid bool, err error) {
	value, valid, err = ego.Consume()
	for err == nil && valid && !ego.filter(value) {
		value, valid, err = ego.Consume()
	}
	if !valid || err != nil {
		ego.Close()
	}
	return
}
