package storage

import (
	"testing"
	"time"
)

func TestMemoryStorage_SaveAssignsIDAndCreatedAt(t *testing.T) {
	t.Parallel()

	s := New()
	in := &Message{Message: "hello"}

	got, err := s.Save(in)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if got.ID != 1 {
		t.Fatalf("ID = %d, want 1", got.ID)
	}

	if got.Message != "hello" {
		t.Fatalf("Message = %q, want %q", got.Message, "hello")
	}

	if got.CreatedAt.IsZero() {
		t.Fatal("CreatedAt is zero")
	}

	if time.Since(got.CreatedAt) > time.Second {
		t.Fatalf("CreatedAt too old: %v", got.CreatedAt)
	}
}

func TestMemoryStorage_SaveUniqueIDs(t *testing.T) {
	t.Parallel()

	s := New()

	first, err := s.Save(&Message{Message: "a"})
	if err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second, err := s.Save(&Message{Message: "b"})
	if err != nil {
		t.Fatalf("Save second: %v", err)
	}

	if first.ID == second.ID {
		t.Fatalf("IDs collided: both %d", first.ID)
	}
}

func TestMemoryStorage_GetAll(t *testing.T) {
	t.Parallel()

	s := New()

	all, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll empty: %v", err)
	}

	if len(all) != 0 {
		t.Fatalf("GetAll empty len = %d, want 0", len(all))
	}

	if _, err = s.Save(&Message{Message: "one"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err = s.Save(&Message{Message: "two"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	all, err = s.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("GetAll len = %d, want 2", len(all))
	}

	seen := map[string]bool{}
	for _, m := range all {
		seen[m.Message] = true
	}

	if !seen["one"] || !seen["two"] {
		t.Fatalf("GetAll missing messages: %+v", seen)
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	t.Parallel()

	s := New()

	saved, err := s.Save(&Message{Message: "gone"})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err = s.Delete(saved.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	all, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(all) != 0 {
		t.Fatalf("after Delete len = %d, want 0", len(all))
	}
}

func TestMemoryStorage_DeleteMissingID(t *testing.T) {
	t.Parallel()

	s := New()

	if err := s.Delete(999); err != nil {
		t.Fatalf("Delete missing id: %v", err)
	}
}
