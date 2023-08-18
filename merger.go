package stream

import (
	"errors"
	"log"
	"sync"
)

/*
Implements:
  - Merger
*/
type channeledMerger[T any] struct {
	ChanneledInput[T]
	sources        []Producer[T]
	autoclose      bool
	sourcesLock    sync.Mutex
	overflowBuffer []T
}

/*
NewChanneledMerge is a constructor of the channeled merger.
Channeled merger is a merger implementation based on ChanneledInput.

Type parameters:
  - T - type of the consumed and produced values.

Parameters:
  - capacity - size of the channel buffer,
  - autoclose - if true, the stream closes automatically when all attached streams close.

Returns:
  - pointer to the new merger.
*/
func NewChanneledMerger[T any](capacity int, autoclose bool) Merger[T] {
	return &channeledMerger[T]{
		autoclose:      autoclose,
		sources:        make([]Producer[T], 0),
		ChanneledInput: NewChanneledInput[T](capacity),
	}
}

/*
Consumes the data from the source Producer and pushes them to the result stream.
It runs asynchronously for each attached Producer.

Parameters:
  - s - producer to consume from.
*/
func (ego *channeledMerger[T]) merge(s Producer[T]) {

	for {

		if ego.Closed() {
			break
		}

		value, valid, err := s.Get()

		// The source is exhausted
		if !valid {
			ego.unsetSource(s)
			return
		}

		if err != nil {
			log.Default().Println(err)
			return
		}

		// It may happen that the routine was waiting on get while the merge stream got closed. Then we send the delayed data to overflowBuffer and then serve them through Get().
		defer func() {
			if r := recover(); r != nil {
				ego.overflowBuffer = append(ego.overflowBuffer, value)
				log.Default().Println("Channel closed externally, extra data sent to overflow buffer.", r, value)
			}
		}()

		ego.Channel() <- value

	}

}

/*
Unsets the source stream.

Parameters:
  - s - stream to unset.
*/
func (ego *channeledMerger[T]) unsetSource(s Producer[T]) {

	if ego.Closed() {
		return
	}

	defer ego.sourcesLock.Unlock()
	ego.sourcesLock.Lock()

	for i, source := range ego.sources {
		if source == s {
			ego.sources = append(ego.sources[:i], ego.sources[i+1:]...)
			break
		}
	}

	if ego.autoclose && len(ego.sources) == 0 {
		ego.Close()
	}

}

func (ego *channeledMerger[T]) SetSource(s Producer[T]) error {

	defer ego.sourcesLock.Unlock()
	ego.sourcesLock.Lock()

	if ego.Closed() {
		return errors.New("the stream is already closed")
	}

	ego.sources = append(ego.sources, s)

	go ego.merge(s)

	return nil

}

func (ego *channeledMerger[T]) CanSetSource() bool {
	return true
}

func (ego *channeledMerger[T]) Close() {
	for _, s := range ego.sources {
		ego.unsetSource(s)
	}
	ego.ChanneledInput.Close()
}

func (ego *channeledMerger[T]) Consume() (value T, valid bool, err error) {
	value, valid = <-ego.Channel()
	if !valid && len(ego.overflowBuffer) > 0 {
		value = ego.overflowBuffer[0]
		ego.overflowBuffer = ego.overflowBuffer[1:]
		valid = true
	}
	return

}

func (ego *channeledMerger[T]) Get() (value T, valid bool, err error) {
	value, valid, err = ego.Consume()
	return
}
