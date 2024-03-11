package mysqlpattern

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/go-sql-driver/mysql"
    "github.com/stretchr/testify/assert"
    "log"
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

var (
    db     *sql.DB
    client *MysqlClient
)

func init() {
    var err error
    db, err = initConnection()
    if err != nil {
        log.Fatalln(err)
    }

    client = NewMysqlClient()
    ctx, _ := context.WithTimeout(context.TODO(), time.Minute)
    //defer cancel()
    client.Run(ctx)
}

func initConnection() (*sql.DB, error) {
    cfg := mysql.Config{
        User:                 "root",
        Passwd:               "",
        Net:                  "tcp",
        Addr:                 fmt.Sprintf("%s:%s", "127.0.0.1", "3306"),
        DBName:               "demo-brokers",
        AllowNativePasswords: true,
        ParseTime:            true,
    }

    db, err := sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        return nil, err
    }
    pingErr := db.Ping()
    if pingErr != nil {
        return nil, pingErr
    }
    fmt.Println("Mysql Connected...")
    return db, nil
}

func add1Transaction(name string, db *sql.DB) (int, error) {
    // Get a Tx for making transaction requests.
    tx, err := db.BeginTx(context.TODO(), nil)
    if err != nil {
        return 0, err
    }
    // Defer a rollback in case anything fails.
    defer tx.Rollback()

    row := tx.QueryRow("SELECT count FROM counters WHERE name = ? FOR UPDATE", name)
    var count int
    if err := row.Scan(&count); err != nil {
        return 0, err
    }
    if _, err := tx.Exec("UPDATE counters SET count = ? WHERE name = ?",
        count+1, name); err != nil {
        return 0, err
    }

    // Commit the transaction.
    if err = tx.Commit(); err != nil {
        return 0, err
    }
    return count + 1, nil
}

func BenchmarkMysqlClient_Add1(b *testing.B) {
    b.Run("transaction bench", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            add1Transaction("reading", db)
        }
    })

    b.Run("mysql client bench", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            client.Add1("reading")
        }
    })
}
