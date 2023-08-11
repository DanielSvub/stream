package stream

type multiplexer[T any] struct {
	DefaultConsumer[T]
	outputs []ChanneledInput[T]
}

func NewMultiplexer[T any](branches int, capacity int) Multiplexer[T] {
	ego := &multiplexer[T]{}
	ego.outputs = make([]ChanneledInput[T], branches)
	for i := 0; i < branches; i++ {
		ego.outputs[i] = NewChanneledInput[T](capacity)
	}
	return ego
}

func (ego *multiplexer[T]) pipeData() {

	for _, output := range ego.outputs {
		defer output.Close()
	}

	for {
		value, valid, err := ego.Consume()
		if err != nil || !valid {
			return
		}
		for _, output := range ego.outputs {
			output.Write(value)
		}
	}

}

func (ego *multiplexer[T]) Out(index int) Producer[T] {
	return ego.outputs[index]
}

func (ego *multiplexer[T]) SetSource(s Producer[T]) error {
	if err := ego.DefaultConsumer.SetSource(s); err != nil {
		return err
	}
	go ego.pipeData()
	return nil
}
