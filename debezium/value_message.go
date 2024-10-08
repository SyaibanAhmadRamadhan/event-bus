//go:build go1.22

package debezium

import "github.com/guregu/null/v5"

type Operation string

const (
	Create   Operation = "c"
	Update   Operation = "u"
	Delete   Operation = "d"
	Truncate Operation = "t"
	Message  Operation = "m"
)

type Source struct {
	Version   string     `json:"version"`
	Connector string     `json:"connector"`
	Name      string     `json:"name"`
	TsMs      int64      `json:"ts_ms"`
	Snapshot  string     `json:"snapshot"`
	Db        string     `json:"db"`
	Sequence  string     `json:"sequence"`
	Schema    string     `json:"schema"`
	Table     string     `json:"table"`
	TxID      int64      `json:"txId"`
	Lsn       int64      `json:"lsn"`
	Xmin      null.Int64 `json:"xmin"`
}

type Payload[before, after any] struct {
	Before null.Value[before]      `json:"before"`
	After  null.Value[after]       `json:"after"`
	Source Source                  `json:"source"`
	Op     Operation               `json:"op"`
	TsMs   int64                   `json:"ts_ms"`
	Tx     null.Value[Transaction] `json:"transaction"`
}

type Transaction struct {
	ID                  string `json:"id"`
	TotalOrder          int64  `json:"total_order"`
	DataCollectionOrder int64  `json:"data_collection_order"`
}

type Envelope[payloadBefore, payloadAfter any] struct {
	Schema  map[string]any                       `json:"schema"`
	Payload Payload[payloadBefore, payloadAfter] `json:"payload"`
}
