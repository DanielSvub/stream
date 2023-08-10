package stream_test

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	. "stream"
	"sync"
	"testing"
)

const (
	testDataSize = 5000
)

func TestStreamBasics(t *testing.T) {
	t.Run("Closing", func(t *testing.T) {
		var closable Closable
		closable = NewChanneledInput[int](100)

		closable.Close()
		if !closable.Closed() {
			t.Error("Closing problems.")

		}
	})

	t.Run("Write-read", func(t *testing.T) {
		inS := NewChanneledInput[int](testDataSize)
		data := make([]int, testDataSize)
		for i := 0; i < testDataSize; i++ {
			data[i] = rand.Int()
			inS.Write(data[i])
		}
		inS.Close()

		for i := 0; i < testDataSize; i++ {
			value, valid, err := inS.Get()
			if !valid {
				t.Error("Unexpected invalid data red")
			}
			if err != nil {
				t.Error("Unexpected error:", err)
			}
			expected := data[i]
			if expected != value {
				t.Error("Unexpected value (wanted", expected, "got", value, ")")
			}
		}

		value, valid, err := inS.Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}
	})
}

func TestStream(t *testing.T) {

	t.Run("Transformer", func(t *testing.T) {

		transform := func(a int) int { return a * a }
		inS := NewChanneledInput[int](testDataSize)
		outS := NewTransformer(func(a int) int { return a * a })
		inS.Pipe(outS)

		data := make([]int, testDataSize)
		for i := 0; i < testDataSize; i++ {
			data[i] = rand.Int()
			inS.Write(data[i])
		}
		inS.Close()

		for i := 0; i < testDataSize; i++ {
			value, valid, err := outS.Get()
			if !valid {
				t.Error("Unexpected invalid data red")
			}
			if err != nil {
				t.Error("Unexpected error:", err)
			}
			expected := transform(data[i])
			if expected != value {
				t.Error("Unexpected value (wanted", expected, "got", value, ")")
			}
		}

		value, valid, err := outS.Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}
	})

	t.Run("Transformer (change type)", func(t *testing.T) {

		transform := func(a int) string { return fmt.Sprintf("%d", a) }
		inverse := func(a string) int {
			if v, err := strconv.Atoi(a); err != nil {
				return -1
			} else {
				return v
			}
		}
		inS := NewChanneledInput[int](testDataSize)
		traS := NewTransformer[int, string](transform)
		outS := NewTransformer[string, int](inverse)
		inS.Pipe(traS).(Producer[string]).Pipe(outS)

		data := make([]int, testDataSize)
		for i := 0; i < testDataSize; i++ {
			data[i] = rand.Int()
			inS.Write(data[i])
		}
		inS.Close()

		for i := 0; i < testDataSize; i++ {
			value, valid, err := outS.Get()
			if !valid {
				t.Error("Unexpected invalid data red")
			}
			if err != nil {
				t.Error("Unexpected error:", err)
			}
			expected := inverse(transform(data[i]))
			if expected != value {
				t.Error("Unexpected value (wanted", expected, "got", value, ")")
			}
		}

		value, valid, err := outS.Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}
	})

	t.Run("Filter", func(t *testing.T) {

		filter := func(a int) bool { return a%2 == 0 }
		inS := NewChanneledInput[int](testDataSize)
		outS := NewFilter(filter)
		inS.Pipe(outS)

		data := make([]int, testDataSize)
		for i := 0; i < testDataSize; i++ {
			data[i] = rand.Int()
			inS.Write(data[i])
		}
		inS.Close()
		i := 0
		for {
			value, valid, err := outS.Get()
			if !valid {
				break
			}
			if err != nil {
				t.Error("Unexpected error:", err)
			}
			var expected int
			for {
				if filter(data[i]) {
					expected = data[i]
					i++
					break
				} else {
					i++
				}

				if i >= len(data) {
					t.Error("Too many data items received.")
					return
				}
			}

			if expected != value {
				t.Error("Unexpected value (wanted", expected, "got", value, ")")
			}
		}

		value, valid, err := outS.Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("Unexpected error", err)
		}
	})

	t.Run("Duplex", func(t *testing.T) {
		inS := NewChanneledInput[int](testDataSize)
		dupS := NewDuplexer[int](testDataSize)
		inS.Pipe(dupS)

		data := make([]int, testDataSize)
		for i := 0; i < testDataSize; i++ {
			data[i] = rand.Int()
			inS.Write(data[i])
		}
		inS.Close()

		for i := 0; i < testDataSize; i++ {
			value1, valid, err := dupS.First().Get()
			value2, valid2, err2 := dupS.Second().Get()
			if !valid || !valid2 {
				t.Error("Unexpected invalid data red")
			}
			if err != nil || err2 != nil {
				t.Error("Unexpected error:", err, "and", err2)
			}
			expected := data[i]
			if expected != value1 || expected != value2 {
				t.Error("Unexpected value (wanted", expected, "got", value1, "and", value2, ")")
			}

		}
		value, valid, err := dupS.First().Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}

		value, valid, err = dupS.Second().Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}
	})

	t.Run("Split", func(t *testing.T) {
		predicate := func(a int) bool { return a%2 == 0 }
		inS := NewChanneledInput[int](testDataSize)
		outS := NewSplitter(10, predicate)
		inS.Pipe(outS)

		data := make([]int, testDataSize)
		for i := 0; i < testDataSize; i++ {
			data[i] = rand.Int()
			inS.Write(data[i])
		}
		inS.Close()

		posChan := outS.Positive().(ChanneledProducer[int])
		negChan := outS.Negative().(ChanneledProducer[int])
		count := 0
		posValid := true
		negValid := true
		for posValid || negValid {
			select {
			case p, valid := <-posChan.Channel():
				if !valid {
					posValid = false
					continue
				}
				count++
				if !predicate(p) {
					t.Error("Value should be in negative but is in positive.", p)
				}
			case n, valid := <-negChan.Channel():
				if !valid {
					negValid = false
					continue
				}
				count++
				if predicate(n) {
					t.Error("Value should be in positive but is in negative.", n)
				}
			}
		}
		if count != testDataSize {
			t.Error(testDataSize, "values on input but", count, "values on output.")
		}

		value, valid, err := outS.Positive().Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}

		value, valid, err = outS.Negative().Get()
		if valid {
			t.Error("Unexpected valid data red", value)
		}
		if err != nil {
			t.Error("unexpected error", err)
		}
	})

	t.Run("Merge (general functionality, autoclose)", func(t *testing.T) {
		inputsCount := 10
		inS := make([]ChanneledInput[int], inputsCount)
		mergeS := NewChanneledMerger[int](true, inputsCount)
		for i := 0; i < inputsCount; i++ {
			inS[i] = NewChanneledInput[int](testDataSize)
			inS[i].Pipe(mergeS)
		}

		data := make([][]int, inputsCount)
		for i := 0; i < inputsCount; i++ {
			data[i] = make([]int, testDataSize)
			for j := 0; j < testDataSize; j++ {
				data[i][j] = rand.Int()%100 + (i+1)*1000 //each input stream has random values betwenn i*1000 and (i+1)*1000 -> easy to check which value came from which stream
				inS[i].Write(data[i][j])
			}
			inS[i].Close()
		}
		// If data sent into input streams and received from merger are the same (and with same relative order for each input stream), we are ok.
		receivedData := make([]int, 0)
		for {
			value, ok, err := mergeS.Get()
			if err != nil {
				t.Error(err)
			}
			if !ok {
				break
			}
			sourceIndex := (value / 1000) - 1
			if len(data[sourceIndex]) == 0 {
				t.Error("Received more data than it was sent into the source.")
			}
			expectedData := data[sourceIndex][0]
			data[sourceIndex] = data[sourceIndex][1:]
			if expectedData != value {
				t.Error("Got wrong data:", value, "expected:", expectedData)
			}
			receivedData = append(receivedData, value)

		}
		for i := 0; i < inputsCount; i++ {
			if len(data[i]) > 0 {
				t.Error("Some data were sent but not received.", i, data[i])
			}
		}
	})

	t.Run("Merge force close", func(t *testing.T) {
		inputsCount := 10
		inS := make([]ChanneledInput[int], inputsCount)
		mergeS := NewChanneledMerger[int](false, inputsCount)
		for i := 0; i < inputsCount; i++ {
			inS[i] = NewChanneledInput[int](testDataSize)
			inS[i].Pipe(mergeS)
		}

		for i := 0; i < inputsCount; i++ {
			t := i
			go func() {
				for {
					inS[t].Write(rand.Int()%100 + (t+1)*1000)
					//each input stream has random values betwenn i*1000 and (i+1)*1000 -> easy to check which value came from which stream
				}
			}()
		}
		k := 0
		for !mergeS.Closed() {
			mergeS.Get()
			if rand.Int()%100 == 0 {
				mergeS.Close()
				k++
				//println(mergeS.Closed())
			}
		}

	})

	t.Run("Pipeline", func(t *testing.T) {
		//data operations
		filterEven := func(a int) bool { return a%2 == 0 }
		filterOdd := func(a int) bool { return a%2 != 0 }
		transformEven := func(a int) int { return 10 * a }
		transformOdd := func(a int) int { return 1000 * a }

		//streams (pipeline mebers)
		inS := NewChanneledInput[int](10)
		dupS := NewDuplexer[int](10)
		filEven := NewFilter(filterEven)
		filOdd := NewFilter(filterOdd)
		traEven := NewTransformer(transformEven)
		traOdd := NewTransformer(transformOdd)
		merS := NewChanneledMerger[int](true, 10)

		//prepare pipeline
		inS.Pipe(dupS)
		dupS.First().Pipe(filEven).(Producer[int]).Pipe(traEven).(Producer[int]).Pipe(merS)
		//or
		//		dupS.First().Pipe(filEven)
		//              filEven.Pipe(traEven)
		//              traEven.Pipe(merS)
		dupS.Second().Pipe(filOdd).(Producer[int]).Pipe(traOdd).(Producer[int]).Pipe(merS)
		data := make([]int, testDataSize)

		//feed data into the start of the pipeline (has to be in different goroutine (channel-buffered input stream would casue main thread to wait on Write if buffer is not big enough)
		go func() {
			for i := 0; i < testDataSize; i++ {
				data[i] = i + 1
				inS.Write(data[i])
			}
			inS.Close()
		}()

		// read data from the end of the pipeline
		dataRed := make([]int, testDataSize)
		j := 0
		for {
			value, ok, err := merS.Get()
			//			println("Got value", value, inS.Closed(), filOdd.Closed(), traOdd.Closed(), merS.Closed())
			if err != nil {
				t.Error(err)
			}
			if !ok {
				break
			}
			dataRed[j] = value
			j++

		}

		// sort data red fro the pipeline for easy comaprison with expectations
		sort.Slice(dataRed, func(i, j int) bool {
			return dataRed[i] < dataRed[j]
		})

		//same transoframtions made by hand on original data
		for i := 0; i < testDataSize; i++ {
			if i%2 == 0 {
				data[i] = transformOdd(data[i])
			} else {
				data[i] = transformEven(data[i])
			}
		}
		sort.Slice(data, func(i, j int) bool {
			return data[i] < data[j]
		})

		//check the outcome
		for i := 0; i < testDataSize; i++ {
			if data[i] != dataRed[i] {
				t.Error("Unexpected value red", dataRed[i], data[i])
			}
		}

	})
}

func TestBuffer(t *testing.T) {

	t.Run("read", func(t *testing.T) {

		result := make([]int, 3)

		is := NewChanneledInput[int](3)
		ts := NewTransformer(func(x int) int {
			return x * x
		})

		is.Write(1, 2, 3)
		is.Close()
		is.Pipe(ts)
		n, err := ts.Read(result)

		if err != nil || n != 3 || len(result) != 3 {
			t.Error("Reading the results was unsuccessful.")
		}

	})

	t.Run("collect", func(t *testing.T) {

		is := NewChanneledInput[int](3)
		ts := NewTransformer(func(x int) int {
			return x * x
		})

		is.Write(1, 2, 3)
		is.Close()
		is.Pipe(ts)

		result, err := ts.Collect()
		if err != nil || len(result) != 3 {
			t.Error("Collecting the results was unsuccessful.")
		}

	})

	t.Run("forEach", func(t *testing.T) {

		is := NewChanneledInput[int](3)

		is.Write(1, 2, 3)
		is.Close()

		err := is.ForEach(func(x int) error { return nil })
		if err != nil {
			t.Error("Iterating over elements was unsuccessful.")
		}

	})

	t.Run("async", func(t *testing.T) {

		is := NewChanneledInput[int](100)

		var wg sync.WaitGroup
		wg.Add(2)

		write := func() {
			defer wg.Done()
			defer is.Close()
			is.Write(make([]int, 1000000)...)
		}

		read := func() {
			defer wg.Done()
			result, err := is.Collect()
			if err != nil || len(result) != 1000000 {
				t.Error("Collecting the results in parallel was unsuccessful.")
			}
		}

		go write()
		go read()
		wg.Wait()

	})

	t.Run("errBuffer", func(t *testing.T) {
		is := NewChanneledInput[int](5)

		is.Write(1, 2, 3)
		is.Close()
		if _, err := is.Write(4, 5); err == nil {
			t.Error("Can be written into the stream even though it shouldn't be possible.")
		}

		if _, err := is.Read(nil); err == nil {
			t.Error("Can read from the stream even if it has nil input slice.")
		}

		if _, err := is.Collect(); err != nil {
			t.Error("Nothing was collected from the stream.")
		}

		// p := make([]int, 5)
		// if _, err := is.Read(p); err == nil {
		// 	t.Error("Can read the stream even when the stream is closed.")
		// }

		// if _, err := is.Collect(); err == nil {
		// 	t.Error("Can collect from the stream even when the stream is closed.")
		// }

		// if err := is.ForEach(func(x int) error { return nil }); err == nil {
		// 	t.Error("Can iterate over elements of the stream even when the stream is closed.")
		// }
	})

	t.Run("panicBuffer", func(t *testing.T) {
		is := NewChanneledInput[int](5)
		var a []int
		if _, err := is.Write(a...); err == nil {
			t.Error("Should be error.")
		}
	})

}

func shouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Error("Should have paniced")
}
