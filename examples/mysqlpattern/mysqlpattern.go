package mysqlpattern

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "github.com/go-sql-driver/mysql"
    "log"
    "os"
)

type InsertOp struct {
    Table  string
    Query  string
    Values []any
    ID     chan int64
}

type QueryOp struct {
    Table string
    Query string
    Where []any
}

type UpdateOp struct {
    Table string
    Query string
    Obj   any
}

type DeleteOp struct {
    Table string
    ID    int64
}

type MysqlClient struct {
    db                 *sql.DB
    add1Stream         chan Add1Param
    addCounterParam    chan AddCounterParam
    queryByNameParam   chan QueryByNameParam
    updateCounterParam chan UpdateCounterParam
    deleteByNameParam  chan DeleteByNameParam
    insertOpStream     chan InsertOp
}

func NewMysqlClient() *MysqlClient {
    return &MysqlClient{
        add1Stream:         make(chan Add1Param),
        addCounterParam:    make(chan AddCounterParam),
        queryByNameParam:   make(chan QueryByNameParam),
        updateCounterParam: make(chan UpdateCounterParam),
        deleteByNameParam:  make(chan DeleteByNameParam),
        insertOpStream:     make(chan InsertOp),
    }
}

func (c *MysqlClient) Run(ctx context.Context) (err error) {
    cfg := mysql.Config{
        User:                 getenv("DBUSER", "root"),
        Passwd:               getenv("DBPASS", ""),
        Net:                  "tcp",
        Addr:                 fmt.Sprintf("%s:%s", getenv("DBHOST", "127.0.0.1"), getenv("DBPORT", "3306")),
        DBName:               "demo-brokers",
        AllowNativePasswords: true,
        ParseTime:            true,
    }

    c.db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        return err
    }
    pingErr := c.db.Ping()
    if pingErr != nil {
        return pingErr
    }
    fmt.Println("Mysql Connected...")

    go func() {
        defer c.exit()
        for {
            select {
            case param := <-c.addCounterParam:
                id, err := addCounter(param, c.db)
                if err != nil {
                    param.Result <- CounterResult{Err: err}
                } else {
                    param.Result <- CounterResult{Counter: Counter{ID: id, Name: param.Name}}
                }
            case param := <-c.queryByNameParam:
                counter, err := queryByName(param.Name, c.db)
                param.Result <- CounterResult{Counter: counter, Err: err}
            case param := <-c.updateCounterParam:
                err := updateCounter(param.ID, param.Count, c.db)
                param.Result <- CounterResult{Err: err}
            case param := <-c.deleteByNameParam:
                err := deleteByName(param.Name, c.db)
                param.Result <- CounterResult{Err: err}
            case param := <-c.add1Stream:
                count, err := c.doSomeSql(param.Name)
                if err != nil {
                    log.Fatalln(err)
                }
                param.Count <- count
            case <-ctx.Done():
                return
            }
        }
    }()

    return
}

func (c *MysqlClient) Add1(name string) int {
    param := Add1Param{
        Name:  name,
        Count: make(chan int),
    }
    c.add1Stream <- param
    return <-param.Count
}

func (c *MysqlClient) QueryByName(name string) Counter {
    param := QueryByNameParam{
        Name:   name,
        Result: make(chan CounterResult),
    }
    c.queryByNameParam <- param
    return (<-param.Result).Counter
}

func (c *MysqlClient) AddCounter(name string) Counter {
    param := AddCounterParam{
        Name:   name,
        Result: make(chan CounterResult),
    }
    c.addCounterParam <- param
    return (<-param.Result).Counter
}

func (c *MysqlClient) DeleteByName(name string) error {
    param := DeleteByNameParam{
        Name:   name,
        Result: make(chan CounterResult),
    }
    c.deleteByNameParam <- param
    return (<-param.Result).Err
}

func (c *MysqlClient) exit() {
    close(c.add1Stream)
    if err := c.db.Close(); err != nil {
        log.Panicln(err)
    }
    fmt.Println("Mysql Disconnected...")
}

func (c *MysqlClient) doSomeSql(name string) (int, error) {
    row := c.db.QueryRow("SELECT count FROM counters WHERE name = ?", name)
    var count int
    if err := row.Scan(&count); err != nil {
        return 0, err
    }
    if _, err := c.db.Exec("UPDATE counters SET count = ? WHERE name = ?",
        count+1, name); err != nil {
        return 0, err
    }
    return count + 1, nil
}

func (c *MysqlClient) insert(op InsertOp) (int64, error) {
    result, err := c.db.Exec(op.Query, op.Values...)
    if err != nil {
        return 0, fmt.Errorf("insert %v: %v", op.Table, err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("insert %v: %v", op.Table, err)
    }
    return id, nil
}

func (c *MysqlClient) query(op QueryOp) (any, error) {
    var counter Counter
    row := c.db.QueryRow(op.Query, op.Where...)
    if err := row.Scan(&counter.ID, &counter.Name, &counter.Count); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return counter, fmt.Errorf("query %v %v: no such counter", op.Table, op.Where)
        }
        return counter, fmt.Errorf("query %v %v: %v", op.Table, op.Where, err)
    }
    return counter, nil
}

func getenv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
