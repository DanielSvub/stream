package stream

type transformer[T, U any] struct {
	DefaultConsumer[T]
	DefaultProducer[U]
	DefaultClosable
	transform func(T) U
}

func NewTransformer[T, U any](fn func(T) U) Transformer[T, U] {
	ego := &transformer[T, U]{transform: fn}
	ego.DefaultProducer = *NewDefaultProducer[U](ego)
	return ego
}

func (ego *transformer[T, U]) Get() (value U, valid bool, err error) {
	inValue, valid, err := ego.Consume()
	if !valid || err != nil {
		ego.Close()
	} else {
		value = ego.transform(inValue)
	}
	return
}
