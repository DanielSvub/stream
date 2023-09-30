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
type activeMerger[T any] struct {
	DefaultClosable
	DefaultProducer[T]
	sources        []ChanneledProducer[T]
	autoclose      bool
	sourcesLock    sync.Mutex
	nextToReadFrom int
}

/*
NewActiveMerger creates new activeMerger.

	activeMerger is merger which actively in round-robin style polls attached sources (producers) in it's get.
	Beware that if attached source is not channeled then new goroutine is spawned to push data through channel the merger can select on.

Type parameters:
  - T - type of the consumed and produced values.

Parameters:
  - autoclose - if true, the stream closes automatically when all attached streams close.

Returns:
  - pointer to the new merger.
*/
func NewActiveMerger[T any](autoclose bool) Merger[T] {
	ego := &activeMerger[T]{
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
func (ego *activeMerger[T]) unsetSource(s Producer[T]) {

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

func (ego *activeMerger[T]) SetSource(s Producer[T]) error {

	if ego.Closed() {
		return errors.New("the stream is already closed")
	}

	chp, ischanneled := s.(ChanneledProducer[T])
	if ischanneled {
		ego.sourcesLock.Lock()
		ego.sources = append(ego.sources, chp)
		ego.sourcesLock.Unlock()
	} else {
		chp := NewChanneledInput[T](0)
		go func() {
			for {

				if ego.Closed() {
					break
				}

				value, valid, err := s.Get()

				//source is exhausted
				if !valid {
					chp.Close()
					return
				}

				if err != nil {
					log.Default().Println(err)
					return
				}
				chp.Channel() <- value
			}

		}()
		ego.sourcesLock.Lock()
		ego.sources = append(ego.sources, chp)
		ego.sourcesLock.Unlock()
	}

	return nil

}

func (ego *activeMerger[T]) CanSetSource() bool {
	return true
}

func (ego *activeMerger[T]) Close() {
	for _, s := range ego.sources {
		ego.unsetSource(s)
	}
	ego.DefaultClosable.Close()
}

func (ego *activeMerger[T]) Consume() (value T, valid bool, err error) {
	//TODO beware: this is active waiting
	//The merger implementation has been  changed from original lazy implementation to decrease number of goroutines -> now we create goroutine only for nonchanneled sources
	//In the original implementation each source had its own goroutine - it was waiting on sources get, until value appearead, then it pushed it into merger's output buffer (see older commits - cca 14.9.2023).
	//The price to pay is this active polling of sources in round robin style when Get is requested
	for {
		//has to be locked inside loop so sources can be added/removed between iterations
		ego.sourcesLock.Lock()
		if len(ego.sources) == 0 {
			ego.sourcesLock.Unlock()
			break
		}
		ego.nextToReadFrom = (ego.nextToReadFrom + 1) % len(ego.sources)
		select {
		case value, valid := <-ego.sources[ego.nextToReadFrom].Channel():
			if !valid {
				toUnset := ego.sources[ego.nextToReadFrom]
				ego.sourcesLock.Unlock()
				ego.unsetSource(toUnset)
				continue
			} else {
				ego.sourcesLock.Unlock()
				return value, valid, nil
			}

		default:
			ego.sourcesLock.Unlock()
			continue
		}
	}
	if ego.Closed() {
		return *new(T), false, nil
	}
	return *new(T), true, errors.New("no sources attached yet or all sources were unset and autoclose is not active")
}

func (ego *activeMerger[T]) Get() (value T, valid bool, err error) {
	value, valid, err = ego.Consume()
	return
}
