package mysqlpattern

import (
    "context"
    "github.com/stretchr/testify/assert"
    "sync"
    "testing"
    "time"
)

func TestMysqlClient_Run(t *testing.T) {
    client := NewMysqlClient()
    ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
    defer cancel()
    client.Run(ctx)
    const COUNTER_NAME = "reading"
    client.DeleteByName(COUNTER_NAME)
    client.AddCounter(COUNTER_NAME)

    var wg sync.WaitGroup
    const N = 10000
    for i := 0; i < N; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            client.Add1(COUNTER_NAME)
        }()
    }

    wg.Wait()

    count := client.QueryByName(COUNTER_NAME).Count
    assert.Equal(t, count, N)
}
