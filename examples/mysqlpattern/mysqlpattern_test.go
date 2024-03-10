package mysqlpattern

import (
    "context"
    "sync"
    "testing"
    "time"
)

func TestMysqlClient_Run(t *testing.T) {
    client := NewMysqlClient()
    ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
    defer cancel()
    client.Run(ctx)

    var wg sync.WaitGroup
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            client.Add1("reading")
        }()
    }

    wg.Wait()
}
