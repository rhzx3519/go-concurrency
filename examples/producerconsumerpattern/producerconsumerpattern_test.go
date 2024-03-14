package producerconsumerpattern

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand/v2"
	"sync"
	"testing"
)

func TestProducer_Run(t *testing.T) {
	producer := NewProducer()
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	producer.Run(ctx)

	var wg sync.WaitGroup
	const N = 1000
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			outStream := make(chan int)
			defer close(outStream)
			param := &Parameter{outStream: outStream}
			producer.paramStream <- param
			<-outStream

			l := rand.IntN(10)
			for j := 0; j < l; j++ {
				producer.anotherStream <- struct{}{}
			}
		}()
	}

	wg.Wait()
	fmt.Println(<-producer.countStream)

	// Output:
	// doProduce... 1 0
}

func TestRWProducer_Run(t *testing.T) {
	producer := NewRWProducer()
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	producer.Run(ctx)

	var wg sync.WaitGroup
	const N = 100000
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			producer.writeStream <- struct{}{}
			<-producer.readStream
		}()
	}
	wg.Wait()
	assert.Equal(t, N, <-producer.readStream)
}

func TestChan(t *testing.T) {
	stream := make(chan int)
	stream <- 1
	fmt.Println(<-stream)
}
