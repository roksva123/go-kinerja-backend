package model

import (
	"encoding/json"
	"time"
)

type SyncHistory struct {
	ID         int64           `json:"id"`
	SyncTime   time.Time       `json:"sync_time"`
	SyncType   string          `json:"sync_type"`
	Status     string          `json:"status"`
	DurationMs int64           `json:"duration_ms"`
	Details    json.RawMessage `json:"details,omitempty"`
}
