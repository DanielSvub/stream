/*
Package stream is a library providing lazy generic data streams for Golang.
*/
package stream

/*
Closable is an entity which can be closed.
*/
type Closable interface {

	/*
		Closed checks whether the entity is closed or not.

		Returns:
		  - true if the entity is closed, false otherwise.
	*/
	Closed() bool

	/*
		Close closes the entity.
	*/
	Close()
}

/*
Writer is an entity which can be written to.

Type parameters:
  - T - type of the written values.
*/
type Writer[T any] interface {
	/*
		Write writes the given values to the entity.

		Parameters:
		  - values - any amount of data items to write.

		Returns:
		  - count of data items actually written,
		  - error if any occurred.
	*/
	Write(values ...T) (int, error)
}

/*
Reader is an entity which can be read from.

Type parameters:
  - T - type of the read values.
*/
type Reader[T any] interface {
	/*
		Read reads values from the entity into the given slice.

		Parameters:
		  - dst - destination slice.

		Returns:
		  - count of data items actually read (maximum is the length of dst),
		  - error if any occurred.
	*/
	Read(dst []T) (int, error)
}

/*
Collector is an entity from which all its values can be acquired as a slice.

Type parameters:
  - T - type of the collected values.
*/
type Collector[T any] interface {
	/*
		Collect reads all values from the entity and returns them as a slice.

		Returns:
		  - slice of all data items (its length is not known in advance),
		  - error if any occurred.
	*/
	Collect() ([]T, error)
}

/*
Iterator is an entity which can be iterated over.

Type parameters:
  - T - type of the iterated values.
*/
type Iterator[T any] interface {
	/*
		ForEach executes a given function on an every data item.
		The function has one parameter, the current element, and returns error if any occurred.
		If an error occurrs, the rest of iterations is skipped.

		Parameters:
		  - fn - anonymous function to be executed.

		Returns:
		  - error if any of the iterations returned error.
	*/
	ForEach(fn func(T) error) error
}

/*
Producer is a data stream producing data.
Can be connected to a Consumer via pipe.

Embeds:
  - Closable
  - Reader
  - Collector
  - Iterable

Type parameters:
  - T - type of the produced values.
*/
type Producer[T any] interface {
	Closable
	Reader[T]
	Collector[T]
	Iterator[T]

	/*
		Get acquires a next item from the stream.

		Returns:
			- value - the value of the item,
			- valid - true if the value is present, false otherwise,
			- err - error, if any occurred.
	*/
	Get() (T, bool, error)

	/*
		Pipe attaches the producer to the given consumer.

		Parameters:
		  - c - the destination consumer.

		Returns:
		  - the destination consumer.
	*/
	Pipe(c Consumer[T]) Consumer[T]
}

/*
ChanneledProducer is a Producer whose data are drawn from a channel.
We may use Get, or work directly with the Channel.

Embeds:
  - Producer

Type parameters:
  - T - type of the produced values.
*/
type ChanneledProducer[T any] interface {
	Producer[T]

	/*
			Channel acquires the underlying chan T.

			Returns:
		      - underlying channel.
	*/
	Channel() chan T
}

/*
Producer is a data stream consuming data.
A Producer can be attached to it via pipe.

Type parameters:
  - T - type of the consumed values.
*/
type Consumer[T any] interface {

	/*
		SetSource sets some producer as a source of the values for the Consumer.

		Parameters:
		  - s - Producer to be set as a source.

		Returns:
		  - error if any occurred.
	*/
	SetSource(s Producer[T]) error

	/*
		CanAddSource signals if the Consumer is able to accept more Producers.

		Returns:
		  - true if the source can be set at the moment, false otherwise.
	*/
	CanSetSource() bool

	/*
		Consume consumes the net value from the Producer.

		Returns:
		  - value - the value of the item,
		  - valid - true if the value is present, false otherwise,
		  - err - error, if any occurred.
	*/
	Consume() (T, bool, error)
}

/*
Transformer is a stream which can consume values of T from some Producer and produce values of U.

Embeds:
  - Consumer
  - Producer

Type parameters:
  - T - type of the consumed values,
  - U - type of the produced values.
*/
type Transformer[T, U any] interface {
	Consumer[T]
	Producer[U]
}

/*
Filter is a stream which can consume values of T from some Producer and produce values of the same type.
The produced data is a subset of the consumed data.

Embeds:
  - Consumer
  - Producer

Type parameters:
  - T - type of the consumed and produced values.
*/
type Filter[T any] interface {
	Consumer[T]
	Producer[T]
}

/*
Multiplexer multiplies the attached producer.
The cloned streams can be consumed separately.

Embeds:
  - Consumer

Type parameters:
  - T - type of the consumed and produced values.
*/
type Multiplexer[T any] interface {
	Consumer[T]

	/*
		Out acquires one of the cloned Producers.

		Parameters:
		  - index - index of the desired Producer.

		Returns:
		  - the corresponding producer, nil if the index does not exist.
	*/
	Out(index int) Producer[T]
}

/*
Splitter splits the attached producer into multiple branches due to some conditions.
If the data do not satisfy any of the conditions, it is written to the default branch.
Each input value from the input Producer goes to exactly one output.
The output streams can be consumed separately.

Embeds:
  - Consumer

Type parameters:
  - T - type of the consumed and produced values.
*/
type Splitter[T any] interface {
	Consumer[T]

	/*
		Cond acquires one of the Producers corresponding the the concrete condition.

		Parameters:
		  - index - index of the desired Producer.

		Returns:
		  - the corresponding producer, nil if the index does not exist.
	*/
	Cond(index int) Producer[T]

	/*
		Default acquires the Producer containing data not satisfying any of the conditions.

		Returns:
		  - default Producer.
	*/
	Default() Producer[T]
}

/*
Merger accepts multiple Producers as a source and combines data from all of them into one stream.

Embeds:
  - Consumer
  - Producer

Type parameters:
  - T - type of the consumed and produced values.
*/
type Merger[T any] interface {
	Consumer[T]
	Producer[T]
}

/*
Channeled input is a Producer getting its data from the outside world.
The source of the data is a channel which is written to by the programmer.
It is closed manually, what causes all attached streams to close also.

Embeds:
  - ChanneledProducer
  - Writer

Type parameters:
  - T - type of the produced values.
*/
type ChanneledInput[T any] interface {
	ChanneledProducer[T]
	Writer[T]
}
