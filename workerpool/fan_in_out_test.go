package workerpool

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFanIn(t *testing.T) {
	expectedData, actualData, n := []int{}, []int{}, 10_000
	for range n {
		expectedData = append(expectedData, rand.Int())
	}

	input, inputRec := []chan int{}, []<-chan int{}
	for range 10 {
		ch := make(chan int)
		input = append(input, ch)
		inputRec = append(inputRec, ch)
	}

	for _, v := range expectedData {
		go func() { input[rand.Intn(10)] <- v }()
	}
	for v := range FanIn(context.Background(), inputRec) {
		actualData = append(actualData, v)
		if len(actualData) == n {
			for _, v := range input {
				close(v)
			}
		}
	}

	assert.ElementsMatch(t, expectedData, actualData)
}

func TestFanOut(t *testing.T) {
	expectedData, actualData, mu, n := []int{}, []int{}, sync.Mutex{}, 10_000
	for range n {
		expectedData = append(expectedData, rand.Int())
	}

	input, wg := make(chan int), sync.WaitGroup{}
	for _, ch := range FanOut(context.Background(), input, 10) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for v := range ch {
				mu.Lock()
				actualData = append(actualData, v)
				mu.Unlock()
			}
		}()
	}
	for _, v := range expectedData {
		input <- v
	}
	close(input)
	wg.Wait()

	assert.ElementsMatch(t, expectedData, actualData)
}
