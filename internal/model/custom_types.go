package model

import (
	"database/sql"
	"encoding/json"
)

// JsonNullInt64 adalah tipe alias untuk sql.NullInt64 yang akan di-marshal
// menjadi nilai integer atau null dalam format JSON.
type JsonNullInt64 struct {
	sql.NullInt64
}

// MarshalJSON mengimplementasikan antarmuka json.Marshaler.
func (v JsonNullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	}
	return json.Marshal(nil)
}