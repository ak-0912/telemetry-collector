package postgres

import "testing"

func TestNewBunDB(t *testing.T) {
	db := NewBunDB("postgres://telemetry:telemetry@localhost:5432/telemetry?sslmode=disable")
	if db == nil {
		t.Fatal("expected bun DB")
	}
	_ = db.Close()
}
