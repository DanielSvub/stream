package stream

/*
Implements:
  - Transformer
*/
type transformer[T, U any] struct {
	DefaultConsumer[T]
	DefaultProducer[U]
	DefaultClosable
	transform func(T) U
}

/*
NewTransformer is a constructor of the transformer.

Type parameters:
  - T - type of the consumed values,
  - U - type of the produced values.

Parameters:
  - fn - transform function.

Returns:
  - pointer to the new transformer.
*/
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
