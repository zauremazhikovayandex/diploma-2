package repositories

import (
	"testing"
	"time"
)

func TestScanPasswordRow_OK(t *testing.T) {
	now := time.Now()

	meta := "note"
	row := []any{
		int64(1),    // user_id
		"id1",       // id
		"login",     // login
		"encrypted", // password
		meta,        // meta
		now,         // created_at
		now,         // updated_at
		false,       // is_deleted
	}

	rec, err := scanPasswordRow(row)
	if err != nil {
		t.Fatalf("scanPasswordRow error: %v", err)
	}
	if rec.UserID != 1 || rec.ID != "id1" || rec.Login != "login" || rec.Password != "encrypted" {
		t.Fatalf("unexpected record: %+v", rec)
	}
	if rec.Meta == nil || *rec.Meta != meta {
		t.Fatalf("unexpected meta: %#v", rec.Meta)
	}
}

func TestScanPasswordRow_BadTypes(t *testing.T) {
	now := time.Now()

	// user_id неправильного типа
	row := []any{
		"not-int64", // user_id
		"id1",
		"login",
		"encrypted",
		nil,
		now,
		now,
		false,
	}
	if _, err := scanPasswordRow(row); err == nil {
		t.Fatal("expected error for bad user_id type")
	}
}
