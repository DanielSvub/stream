package stream

type splitter[T any] struct {
	DefaultConsumer[T]
	predicate func(T) bool
	positive  ChanneledInput[T]
	negative  ChanneledInput[T]
}

func NewSplitter[T any](buffersCapacity int, predicate func(T) bool) *splitter[T] {
	return &splitter[T]{
		positive:  NewChanneledInput[T](buffersCapacity),
		negative:  NewChanneledInput[T](buffersCapacity),
		predicate: predicate,
	}
}

func (ego *splitter[T]) pipeData() {

	var value T
	var err error
	valid := true

	defer ego.positive.Close()
	defer ego.negative.Close()

	for valid {
		if value, valid, err = ego.Consume(); err != nil || !valid {
			return
		}
		if ego.predicate(value) {
			ego.positive.Write(value)
		} else {
			ego.negative.Write(value)
		}
	}

}

func (ego *splitter[T]) Positive() Producer[T] {
	return ego.positive
}

func (ego *splitter[T]) Negative() Producer[T] {
	return ego.negative
}

func (ego *splitter[T]) SetSource(s Producer[T]) error {
	if err := ego.DefaultConsumer.SetSource(s); err != nil {
		return err
	}
	go ego.pipeData()
	return nil
}
