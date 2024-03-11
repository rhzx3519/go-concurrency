/**
CREATE TABLE `counters` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(45) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT '""',
  `count` int DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name_UNIQUE` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin

*/
package mysqlpattern

import (
    "context"
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
    var ctx = context.TODO()
    // Get a Tx for making transaction requests.
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return 0, err
    }
    // Defer a rollback in case anything fails.
    defer tx.Rollback()

    result, err := tx.Exec("INSERT INTO counters (name) VALUES (?)",
        param.Name)
    if err != nil {
        return 0, fmt.Errorf("addCounter: %v", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("addCounter: %v", err)
    }

    // Commit the transaction.
    if err = tx.Commit(); err != nil {
        return 0, err
    }

    return id, nil
}

func queryByName(name string, db *sql.DB) (Counter, error) {
    var counter Counter

    // Get a Tx for making transaction requests.
    tx, err := db.BeginTx(context.TODO(), nil)
    if err != nil {
        return counter, err
    }
    // Defer a rollback in case anything fails.
    defer tx.Rollback()

    row := tx.QueryRow("SELECT id, name, count FROM counters WHERE name = ?", name)
    if err := row.Scan(&counter.ID, &counter.Name, &counter.Count); err != nil {
        if err == sql.ErrNoRows {
            return counter, fmt.Errorf("queryByName %v: no such counter", name)
        }
        return counter, fmt.Errorf("queryByName %v: %v", name, err)
    }

    // Commit the transaction.
    if err = tx.Commit(); err != nil {
        return counter, err
    }
    return counter, nil
}

func updateCounter(id int64, count int, db *sql.DB) error {
    // Get a Tx for making transaction requests.
    tx, err := db.BeginTx(context.TODO(), nil)
    if err != nil {
        return err
    }
    // Defer a rollback in case anything fails.
    defer tx.Rollback()

    result, err := tx.Exec("UPDATE SET counters count = ? WHERE id = ?",
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

    // Commit the transaction.
    if err = tx.Commit(); err != nil {
        return err
    }
    return nil
}

func deleteByName(name string, db *sql.DB) error {
    // Get a Tx for making transaction requests.
    tx, err := db.BeginTx(context.TODO(), nil)
    if err != nil {
        return err
    }
    // Defer a rollback in case anything fails.
    defer tx.Rollback()

    result, err := tx.Exec("DELETE FROM counters WHERE name = ?", name)
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

    // Commit the transaction.
    if err = tx.Commit(); err != nil {
        return err
    }
    return nil
}
