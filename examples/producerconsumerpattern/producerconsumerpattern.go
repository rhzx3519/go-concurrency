package producerconsumerpattern

import (
	"context"
	"fmt"
	"math/rand"
)

type Parameter struct {
	outStream chan int
}

type Producer struct {
	paramStream   chan *Parameter
	anotherStream chan struct{}
	countStream   chan int
	count         int
}

func NewProducer() *Producer {
	return &Producer{
		paramStream:   make(chan *Parameter),
		anotherStream: make(chan struct{}),
		countStream:   make(chan int),
	}
}

func (p *Producer) Run(ctx context.Context) {
	go func() {
		defer close(p.paramStream)
		defer close(p.anotherStream)
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.anotherStream:
				fmt.Println("another stream...")
			case param := <-p.paramStream:
				fmt.Println("param stream...")
				param.outStream <- p.doProduce()
			case p.countStream <- p.count:
			}
		}
	}()
}

func (p *Producer) doProduce() int {
	p.count++
	fmt.Println("doProduce...", p.count)
	return rand.Int()
}

///

type RWProducer struct {
	writeStream chan struct{}
	readStream  chan int
	count       int
}

func NewRWProducer() *RWProducer {
	return &RWProducer{
		writeStream: make(chan struct{}),
		readStream:  make(chan int),
	}
}

func (p *RWProducer) Run(ctx context.Context) <-chan int {
	go func() {
		defer close(p.readStream)
		for {
			select {
			case <-ctx.Done():
				return
			case p.readStream <- p.get():
				//fmt.Println("read stream...")
			}
		}
	}()

	go func() {
		defer close(p.writeStream)
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.writeStream:
				p.plus1()
				//fmt.Println("write stream...")
			}
		}
	}()

	return p.readStream
}

func (p *RWProducer) get() int {
	//fmt.Println("get...")
	return p.count
}

func (p *RWProducer) plus1() int {
	p.count++
	return p.count
}
