package stream

type duplexer[T any] struct {
	DefaultConsumer[T]
	first  ChanneledInput[T]
	second ChanneledInput[T]
}

func NewDuplexer[T any](capacity int) *duplexer[T] {
	return &duplexer[T]{
		first:  NewChanneledInput[T](capacity),
		second: NewChanneledInput[T](capacity),
	}
}

func (ego *duplexer[T]) pipeData() {
	var value T
	var valid bool = true
	var err error
	defer ego.first.Close()
	defer ego.second.Close()
	for valid {
		if value, valid, err = ego.Consume(); err != nil || !valid {
			return
		}
		ego.first.Write(value)
		ego.second.Write(value)
	}
}

func (ego *duplexer[T]) First() Producer[T] {
	return ego.first
}

func (ego *duplexer[T]) Second() Producer[T] {
	return ego.second
}

func (ego *duplexer[T]) SetSource(s Producer[T]) error {
	if err := ego.DefaultConsumer.SetSource(s); err != nil {
		return err
	}
	go ego.pipeData()
	return nil
}
