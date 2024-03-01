package donepattern

import (
    "fmt"
    "sync"
)

// Greeting&Farewell Done pattern
func ExampleDonePattern() {
    var wg sync.WaitGroup
    done := make(chan interface{})
    defer close(done)
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := printGreeting(done); err != nil {
            fmt.Printf("%v", err)
            return
        }
    }()
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := printFarewell(done); err != nil {
            fmt.Printf("%v", err)
            return
        }
    }()
    wg.Wait()
    // Output:
    // hello world!
    // goodbye world!
}
