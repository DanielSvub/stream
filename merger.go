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
	DefaultClosable
	DefaultProducer[T]
	sources        []ChanneledProducer[T]
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
	ego := &channeledMerger[T]{
		autoclose:   autoclose,
		sources:     make([]ChanneledProducer[T], 0),
		sourcesLock: sync.Mutex{},
	}
	ego.DefaultProducer = *NewDefaultProducer[T](ego)
	return ego
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

	chp, ischanneled := s.(ChanneledProducer[T])
	if ischanneled {

		ego.sources = append(ego.sources, chp)
	} else {
		chp := NewChanneledInput[T](1)
		go func() {
			for {

				if ego.Closed() {
					break
				}

				value, valid, err := s.Get()

				//source is exhausted
				if !valid {
					chp.Close()
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
				chp.Channel() <- value
			}

		}()
		ego.sources = append(ego.sources, chp)
	}

	//go ego.merge(s)

	return nil

}

func (ego *channeledMerger[T]) CanSetSource() bool {
	return true
}

func (ego *channeledMerger[T]) Close() {
	for _, s := range ego.sources {
		ego.unsetSource(s)
	}
	ego.DefaultClosable.Close()
}

var nextToReadFrom int = 0

func (ego *channeledMerger[T]) Consume() (value T, valid bool, err error) {
	//TODO beware: this is active waiting
	for {
		//has to be locked inside for loop so sources can be added between iterations
		ego.sourcesLock.Lock()
		if len(ego.sources) == 0 {
			ego.sourcesLock.Unlock()
			break
		}
		nextToReadFrom = (nextToReadFrom + 1) % len(ego.sources)
		select {
		case value, valid := <-ego.sources[nextToReadFrom].Channel():
			ego.sourcesLock.Unlock()
			if !valid {
				ego.unsetSource(ego.sources[nextToReadFrom])
			}
			return value, valid, nil
		default:
			ego.sourcesLock.Unlock()
			continue
		}
	}
	if !valid && len(ego.overflowBuffer) > 0 {
		value = ego.overflowBuffer[0]
		ego.overflowBuffer = ego.overflowBuffer[1:]
		valid = true
		return value, valid, nil
	}
	if ego.Closed() {
		return *new(T), false, nil
	}
	return *new(T), true, errors.New("no sources attached yet or all sources were unset and autoclose is not active")
}

func (ego *channeledMerger[T]) Get() (value T, valid bool, err error) {
	value, valid, err = ego.Consume()
	return
}
