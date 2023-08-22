# Stream

This package provides lazy generic data streams. Data are not loaded until they are needed, which allows to process a large amount of data with constant space complexity. The data flow is modelled by a so-called *pipeline*. The processing of the data is initiated by reading from the sink, the end of the pipeline (there may be more of them, as the pipeline can be branched).

Stream is an entity in the pipeline. There are two main kinds of streams:

- **Producer -** a stream producing data of a certain type,
- **Consumer -** a stream capable of attaching to some producer and consuming data from it.

Some streams are both producer and consumer. Those are called *two-sided* and are used to alter the flow or the data themselves. There are five of them:

- **Transformer -** transforms a value of each item,
- **Filter -** discards items not satisfying a certain condition,
- **Multiplexer -** creates multiple streams identical to the source stream,
- **Splitter -** splits one stream into multiple based on certain conditions,
- **Merger -** merges multiple streams into one.

They can be connected together arbitrarily, which creates the pipeline. To move the data to/from the pipeline, *inputs* and *outputs* are used. The input loads data from the outside (Golang variable, file, remote API, etc.) and serves as a starting producer. The output serves as a final consumer, exporting the data to an external resource (not required for storing to a variable though, as every producer is readable - will be explained in [Usage section](#reading-data)).

The flow of the data has to be terminated at some point. Thus each producer can be in two states - open or closed. The stream is closed, when there is no more data to read. The closed state propagates from the start to the end of the pipeline, until the sink is closed, what makes the whole process to end.

## Usage

### Inputs

Inputs serve as a source of the data (first producer in the pipeline). They can be created by implementing the ``Producer`` interface. This library contains one pre-implemented input, ``BufferInput``. TIn this case, the source of the data is a buffered channel of a defined capacity. The data are passed to the stream by *Write* method. The method can be called multiple times. If the buffer is full, the program waits for some space to be freed. When the writing is done, the stream has to be manually closed.

```go
s := stream.NewChanneledInput[int](3)
s.Write(1, 2, 3)
s.Close()
```

### Reading data

Data can be read from any producer. There are 4 methods usable to do this:

- ``Get`` - acquires a single value together with information whether the value is valid (direct approach),
```go
var value int
var err error
valid := true
for valid {
	if value, valid, err = s.Get(); err != nil {
		panic(err)
	}
	fmt.Println(value)
}
```

- ``Read`` - reads a concrete amount of values to a given slice (provided by ``Reader`` interface),
```go
out := make([]int, 3)
if n, err := s.Read(out); err != nil {
	panic(err)
} else if n < 3 {
	panic("not enough values")
} else {
	for _, value := range out {
		fmt.Println(value)
	}
}
```

- ``Collect`` - collects all data in the pipeline and returns them as a slice (provided by ``Collector`` interface),
```go
if out, err := s.Collect(); err != nil {
	panic(err)
} else {
	for _, value := range out {
		fmt.Println(value)
	}
}
```

- ``ForEach`` - iterates over all data and executes a given function on each value (provided by ``Iterable`` interface).
```go
if err := s.ForEach(func(value int) error {
	_, err := fmt.Println(value)
	return err
}); err != nil {
	panic(err)
}
```

**Note:** Read, Collect and ForEach are available only if the producer is not attached to any other consumer (therefore it is a sink).

### Transformer

The transformer works as a *map* method in many programming languages. Each item of the stream is modified by the given transformation function. The output can be of a different type than the input.

```go
t := stream.NewTransformer(func(x int) int {
    return x * x
})
s.Pipe(t)
```

### Filter

The filter simply filters the data by dropping all items for which the given function returns false.

```go
f := stream.NewFilter(func(x int) bool {
	return x <= 2
})
s.Pipe(f)
```

### Multiplexer

The multiplexer clones its source stream. It contains multiple channeled inputs that are filled automatically when the stream is attached to a source. The outputs are accessed by calling the ``Out`` method.

```go
capacity := 3
branches := 2
m := stream.NewMultiplexer[int](capacity, branches)
s.Pipe(m)
b0 := m.Out(0)
b1 := m.Out(1)
```

### Splitter

The splitter is similar to multiplexer, but each item is written to only one of the nested channeled inputs, depending on which of the given conditions is the first satisfied by the item's value. If the value does not satisfy any of the conditions, it goes to the default branch. The outputs are accessed by calling the ``Cond`` and ``Default`` methods.

```go
capacity := 3
ss := stream.NewSplitter[int](capacity, func(x int) bool {
	return x <= 2
})
s.Pipe(ss)
b0 := ss.Cond(0)
b1 := ss.Default()
```

### Merger

The merger merges multiple streams into a single one. It can be configured to close automatically (after closing of all currently attached sources) or manually.

```go
capacity := 3
autoclose := true
m := stream.NewChanneledMerger[int](capacity, autoclose)
s1.Pipe(m)
s2.Pipe(m)
```

### Outputs

Output is a sink which exports the data to some destination. Read, Collect and ForEach methods then cannot be used anywhere in the corresponding branch of the pipeline. It has to be implemented by user, this library does not provide any.

## Custom streams

The package provides tools to conveniently create your own implementations of all interfaces mentioned above (for more information, check the code documentation). Inputs and outputs are likely to be most common.

### Inputs

Custom inputs can be created by implementing the ``Producer`` interface. For this purpose, the ``DefaultProducer`` struct is available to use. It defines all methods of the producer interface excluding ``Get`` thich defines the concrete way of acquiring the data. To insert a simple closing mechanism, the ``DefaultClosable`` struct can be used. Example:

```go
type RandomIntInput struct {
	stream.DefaultClosable
	stream.DefaultProducer[int]
}

func NewRandomIntInput() *RandomIntInput {
	ego := new(RandomIntInput)
	ego.DefaultProducer = *stream.NewDefaultProducer[int](ego)
	return ego
}

func (ego *RandomIntInput) Get() (value int, valid bool, err error) {
	if ego.Closed() {
		return
	}
	value = rand.Intn(100)
	valid = true
	return
}
```

### Outputs
Custom outputs must implement the ``Consumer`` interface. The ``DefaultConsumer`` struct is ready to use for this. Example:

```go
type StdoutOutput struct {
	stream.DefaultConsumer[int]
}

func NewStdoutOutput() *StdoutOutput {
	return new(StdoutOutput)
}

func (ego *StdoutOutput) Run() (err error) {
	value, valid, err := ego.Consume()
	for valid && err == nil {
		fmt.Println(value)
		value, valid, err = ego.Consume()
	}
	return
}
```

## Examples

1. Computes a square of three numbers. The result will be [1, 4, 9].

```go
input := stream.NewChanneledInput[int](3)
transformer := stream.NewTransformer(func(x int) int {
    return x * x
})

input.Write(1, 2, 3)
input.Close()
input.Pipe(transformer)

result, err := transformer.Collect()
```

2. Parallelly creates a million of numbers and prints them increased by 1. Only one integer is stored in memory at the time.

```go
input := stream.NewChanneledInput[int](0)
transformer := stream.NewTransformer(func(x int) int {
	return x + 1
})

var wg sync.WaitGroup
wg.Add(2)

write := func() {
	defer wg.Done()
	defer input.Close()
	for i := 0; i < 1000000; i++ {
		input.Write(i)
	}
}

input.Pipe(transformer)

read := func() {
	defer wg.Done()
	transformer.ForEach(func(x int) error {
		_, err := fmt.Println(x)
		return err
	})
}

go write()
go read()
wg.Wait()
```
