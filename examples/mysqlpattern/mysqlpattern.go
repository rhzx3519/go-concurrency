package mysqlpattern

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/go-sql-driver/mysql"
    "log"
    "os"
)

type MysqlClient struct {
    db         *sql.DB
    add1Stream chan Add1Param
}

func NewMysqlClient() *MysqlClient {
    return &MysqlClient{
        add1Stream: make(chan Add1Param),
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

func getenv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
