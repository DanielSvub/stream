package stream

type Closable interface {
	Closed() bool
	Close()
}

type Writer[T any] interface {
	Write(value ...T) (int, error)
}

type Reader[T any] interface {
	Read(dst []T) (int, error)
}

type Collector[T any] interface {
	Collect() ([]T, error)
}

type Iterator[T any] interface {
	ForEach(fn func(T) error) error
}

type Producer[T any] interface {
	Closable
	Reader[T]
	Collector[T]
	Iterator[T]
	Get() (T, bool, error)
	Pipe(c Consumer[T]) Consumer[T]
}

type ChanneledProducer[T any] interface {
	Producer[T]
	Channel() chan T
}

type Consumer[T any] interface {
	SetSource(s Producer[T]) error
	CanSetSource() bool
	Consume() (T, bool, error)
}

type Transformer[T, U any] interface {
	Consumer[T]
	Producer[U]
}

type Filter[T any] interface {
	Consumer[T]
	Producer[T]
}

type Multiplexer[T any] interface {
	Consumer[T]
	Out(index int) Producer[T]
}

type Splitter[T any] interface {
	Consumer[T]
	Cond(index int) Producer[T]
	Default() Producer[T]
}

type Merger[T any] interface {
	Consumer[T]
	Producer[T]
}

type ChanneledInput[T any] interface {
	ChanneledProducer[T]
	Writer[T]
}
