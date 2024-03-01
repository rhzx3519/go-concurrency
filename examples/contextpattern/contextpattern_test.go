package contextpattern

import (
	"context"
	"fmt"
	"sync"
)

// Greeting&Farewell Context pattern
// Let’s say that genGreeting only wants to wait one second before abandoning
// the call to locale — a timeout of one second. We also want to build some
// smart logic into main. If printGreeting is unsuccessful, we also want to
// cancel our call to printFarewell. After all, it wouldn’t make sense to say
// goodbye if we don’t say hello!
func ExampleContextPattern() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()
	wg.Wait()
	// Output:
	// cannot print greeting: context deadline exceeded
	// cannot print farewell: context canceled
}

func ExampleContextValue() {
	ProcessRequest("jane", "abc123")
	// Output:
	// handling response for jane (auth: abc123)
}
