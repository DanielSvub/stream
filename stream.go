package stream

type Closable interface {
	Closed() bool
	Close()
}

type Writer[T any] interface {
	Write(p ...T) (int, error)
}

type Reader[T any] interface {
	Read(dest []T) (int, error)
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
	Get() (value T, valid bool, err error)
	Pipe(Consumer[T]) Consumer[T]
}

type ChanneledProducer[T any] interface {
	Producer[T]
	Channel() chan T
}

type Consumer[T any] interface {
	SetSource(s Producer[T]) error
	CanSetSource() bool
	Consume() (value T, valid bool, err error)
}

type Transformer[T, U any] interface {
	Consumer[T]
	Producer[U]
}

type Filter[T any] interface {
	Consumer[T]
	Producer[T]
}

type Duplexer[T any] interface {
	Consumer[T]
	First() Producer[T]
	Second() Producer[T]
}

type Splitter[T any] interface {
	Consumer[T]
	Out(string) Producer[T]
}

type Merger[T any] interface {
	Consumer[T]
	Producer[T]
}

type ChanneledInput[T any] interface {
	ChanneledProducer[T]
	Writer[T]
}
