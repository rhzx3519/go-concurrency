package mysqlpattern

import (
    "database/sql"
    "fmt"
)

type Counter struct {
    ID    int64  `json:"id,omitempty"`
    Name  string `json:"name,omitempty"`
    Count int    `json:"count,omitempty"`
}

type Add1Param struct {
    Name  string
    Count chan int
}

type AddCounterParam struct {
    Name   string
    Result chan CounterResult
}

type CounterResult struct {
    Counter Counter
    Err     error
}

type QueryByNameParam struct {
    Name   string
    Result chan CounterResult
}

type UpdateCounterParam struct {
    ID     int64
    Count  int
    Result chan CounterResult
}

type DeleteByNameParam struct {
    Name   string
    Result chan CounterResult
}

func addCounter(param AddCounterParam, db *sql.DB) (int64, error) {
    result, err := db.Exec("INSERT INTO counters (name) VALUES (?)",
        param.Name)
    if err != nil {
        return 0, fmt.Errorf("addCounter: %v", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("addCounter: %v", err)
    }
    return id, nil
}

func queryByName(name string, db *sql.DB) (Counter, error) {
    var counter Counter
    row := db.QueryRow("SELECT id, name, count FROM counters WHERE name = ?", name)
    if err := row.Scan(&counter.ID, &counter.Name, &counter.Count); err != nil {
        if err == sql.ErrNoRows {
            return counter, fmt.Errorf("queryByName %v: no such counter", name)
        }
        return counter, fmt.Errorf("queryByName %v: %v", name, err)
    }
    return counter, nil
}

func updateCounter(id int64, count int, db *sql.DB) error {
    result, err := db.Exec("UPDATE SET counters count = ? WHERE id = ?",
        count, id)
    if err != nil {
        return fmt.Errorf("updateCounter: %v", err)
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("updateCounter: %v", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("updateCounter %v: no such counter", id)
    }
    return nil
}

func deleteByName(name string, db *sql.DB) error {
    result, err := db.Exec("DELETE FROM counters WHERE name = ?", name)
    if err != nil {
        return fmt.Errorf("deleteByName: %v", err)
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("deleteByName: %v", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("deleteByName %v: no such counter", name)
    }
    return nil
}
