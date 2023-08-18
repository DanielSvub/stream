# Stream

This package provides lazy generic data streams. Data are not loaded until they are needed, which allows to process a large amount of data with constant space complexity. To operate, the streams have to be connected together by so-called *pipes*, creating a data flow, *pipeline*. The processing is initiated then by reading from the end of the pipeline.

*Stream* is an entity capable of streaming generic data. There are two main kinds of streams:

- **Producer -** a stream producing data of a certain type,
- **Consumer -** a stream capable of attaching to some producer and consuming data from it.

Some streams are both producer and consumer. Those are called *two-sided* and are used to alter the flow or the data themselves. There are five of them:

- **Transformer -** transforms a value of each item,
- **Filter -** discards items not satisfying a certain condition,
- **Multiplexer -** creates multiple streams identical to the source stream,
- **Splitter -** splits one stream into multiple based on a certain condition,
- **Merger -** merges multiple streams into one.

They can be connected together arbitrarily, creating a pipeline. To move the data to/from the pipeline, *inputs* and *outputs* are used. The input loads data from the outside (Golang variable, file, remote API, etc.) and serves as a starting producer. The output serves as a final consumer, exporting the data to an external resource (not required for storing to a variable though, as every producer is readable - will be explained in Usage section).

The flow of the data has to be terminated at some point. Thus each producer can be in two states - open or closed. The stream is closed, when there is no more data to read. The closed state propagates from the start to the end of the pipeline, until the last stream is closed, what makes the whole process to end.
