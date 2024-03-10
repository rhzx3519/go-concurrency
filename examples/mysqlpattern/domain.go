package mysqlpattern

type Counter struct {
    ID    int64  `json:"id,omitempty"`
    Name  string `json:"name,omitempty"`
    Count int    `json:"count,omitempty"`
}

type Add1Param struct {
    Name  string
    Count chan int
}
