package stream

/*
Implements:
  - Filter
*/
type filter[T any] struct {
	DefaultConsumer[T]
	DefaultProducer[T]
	DefaultClosable
	filter func(T) bool
}

/*
NewFilter is a constructor of the filter.

Type parameters:
  - T - type of the consumed and produced values.

Parameters:
  - fn - filter function.

Returns:
  - pointer to the new filter.
*/
func NewFilter[T any](fn func(T) bool) Filter[T] {
	ego := &filter[T]{filter: fn}
	ego.DefaultProducer = *NewDefaultProducer[T](ego)
	return ego
}

func (ego *filter[T]) Get() (value T, valid bool, err error) {
	value, valid, err = ego.Consume()
	for valid && err == nil && !ego.filter(value) {
		value, valid, err = ego.Consume()
	}
	if !valid || err != nil {
		ego.Close()
	}
	return
}
