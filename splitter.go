package stream

/*
Implements:
  - Splitter
*/
type splitter[T any] struct {
	DefaultConsumer[T]
	predicates    []func(T) bool
	outputs       []ChanneledInput[T]
	defaultOutput ChanneledInput[T]
}

/*
NewSplitter is a constructor of the splitter.

Type parameters:
  - T - type of the consumed and produced values.

Parameters:
  - capacity - size of the channel buffer,
  - fn - any amount of conditional functions.

Returns:
  - pointer to the new splitter.
*/
func NewSplitter[T any](capacity int, fn ...func(T) bool) Splitter[T] {
	branches := len(fn)
	ego := &splitter[T]{}
	ego.predicates = fn
	ego.outputs = make([]ChanneledInput[T], branches)
	for i := 0; i < branches; i++ {
		ego.outputs[i] = NewChanneledInput[T](capacity)
	}
	ego.defaultOutput = NewChanneledInput[T](capacity)
	return ego
}

/*
Consumes the data from the source Producer and pushes them to the result streams.
Each value is pushed to exactly one of the streams depending on the conditional functions.
If the value does not satisfy any of the conditions, it is pushed to the default stream.
It runs asynchronously.
*/
func (ego *splitter[T]) pipeData() {

	defer ego.defaultOutput.Close()
	for _, output := range ego.outputs {
		defer output.Close()
	}

pipe:
	for {
		value, valid, err := ego.Consume()
		if err != nil || !valid {
			return
		}
		for i, output := range ego.outputs {
			if ego.predicates[i](value) {
				output.Write(value)
				continue pipe
			}
		}
		ego.defaultOutput.Write(value)
	}

}

func (ego *splitter[T]) Cond(index int) Producer[T] {
	if len(ego.outputs) <= index {
		return nil
	}
	return ego.outputs[index]
}

func (ego *splitter[T]) Default() Producer[T] {
	return ego.defaultOutput
}

func (ego *splitter[T]) SetSource(s Producer[T]) error {
	if err := ego.DefaultConsumer.SetSource(s); err != nil {
		return err
	}
	go ego.pipeData()
	return nil
}
