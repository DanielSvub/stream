package stream

import (
	"errors"
	"fmt"

	"golang.org/x/exp/slices"
)

// NewBinarySplitter returns Splitter with only two branches - positive and negative.
func NewBinarySplitter[T any](buffersCapacity int, predicate func(T) bool) Splitter[T] {
	predMap := map[string]func(T) bool{}
	predMap["positive"] = predicate
	predMap["negative"] = func(a T) bool { return !predicate(a) }
	res := splitter[T]{predicates: predMap, outputs: map[string]*channeledInput[T]{}, evalOrder: []string{"positive", "negative"}}
	res.outputs["negative"] = NewChanneledInput[T](buffersCapacity)
	res.outputs["positive"] = NewChanneledInput[T](buffersCapacity)
	return &res
}

type splitter[T any] struct {
	DefaultConsumer[T]
	predicates map[string]func(T) bool
	outputs    map[string]*channeledInput[T]
	evalOrder  []string
}

// NewSplitter creates new Splitter with output stream for each specified predicate
// Predicates are evaluated in the specified evalOrder for each passing value. Value is always sent to the output associated with the first predicate to return true.
// There is "_default" output which is used if value does not pass any of supplied predicates.
// Do not use "_default" as predicate name - it is always a fallback stream and the outptus would behave unexpectedly-
func NewSplitter[T any](predicates map[string]func(T) bool, evalOrder []string, capacity int) (Splitter[T], error) {
	for _, n := range evalOrder {
		if predicates[n] == nil {
			return nil, errors.New(fmt.Sprintf("Undefined predicate in eval order: %s", n))
		}
	}

	for n := range predicates {
		if !slices.Contains(evalOrder, n) {
			return nil, errors.New(fmt.Sprintf("Predicate name not in eval order: %s", n))
		}
	}
	if predicates["_default"] != nil {
		return nil, errors.New("Predicate named \"_default\" is not allowed.")
	}

	res := &splitter[T]{predicates: predicates, evalOrder: evalOrder, outputs: map[string]*channeledInput[T]{}}
	for n := range predicates {
		res.outputs[n] = NewChanneledInput[T](capacity)
	}
	res.outputs["_default"] = NewChanneledInput[T](capacity)
	return res, nil
}

func (ego *splitter[T]) pipeData() {

	var value T
	var err error
	valid := true

	for _, o := range ego.outputs {
		defer o.Close()
	}

	for valid {
		if value, valid, err = ego.Consume(); err != nil || !valid {
			return
		}
		sent := false
		for _, n := range ego.evalOrder {
			if ego.predicates[n](value) {
				ego.outputs[n].Write(value)
				sent = true
				break
			}
		}
		if !sent {
			ego.outputs["_default"].Write(value)
			sent = true
		}
	}

}

func (ego *splitter[T]) Out(name string) Producer[T] {
	return ego.outputs[name]
}

func (ego *splitter[T]) SetSource(s Producer[T]) error {
	if err := ego.DefaultConsumer.SetSource(s); err != nil {
		return err
	}
	go ego.pipeData()
	return nil
}
