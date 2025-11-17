package service

import (
	"database/sql"
	"time"
)

type sqlNullString struct{ String string; Valid bool }
type sqlNullInt64 struct{ Int64 int64; Valid bool }
type sqlNullFloat64 struct{ Float64 float64; Valid bool }
type sqlNullTime struct{ Time time.Time; Valid bool }

func (s sqlNullString) ToSQLNullString() sql.NullString { return sql.NullString{String: s.String, Valid: s.Valid} }
func (i sqlNullInt64) ToSQLNullInt64() sql.NullInt64   { return sql.NullInt64{Int64: i.Int64, Valid: i.Valid} }
func (f sqlNullFloat64) ToSQLNullFloat64() sql.NullFloat64 { return sql.NullFloat64{Float64: f.Float64, Valid: f.Valid} }
func (t sqlNullTime) ToSQLNullTime() sql.NullTime       { return sql.NullTime{Time: t.Time, Valid: t.Valid} }

func sqlNullStringFromMap(m map[string]interface{}, key string) sql.NullString {
	if v, ok := m[key].(string); ok && v != "" {
		return sql.NullString{String: v, Valid: true}
	}
	// also sometimes project is object
	if p, ok := m["project"].(map[string]interface{}); ok {
		if pid, ok2 := p["id"].(string); ok2 {
			return sql.NullString{String: pid, Valid: true}
		}
	}
	return sql.NullString{}
}
