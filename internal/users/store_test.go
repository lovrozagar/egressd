package users

import (
	"path/filepath"
	"testing"
)

func TestAddAuthenticateList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := store.Add("alice")
	if err != nil {
		t.Fatal(err)
	}
	if plain == "" {
		t.Fatal("expected non-empty password")
	}
	if !store.Authenticate("alice", plain) {
		t.Fatal("expected auth success")
	}
	if store.Authenticate("alice", "wrong") {
		t.Fatal("expected auth failure for wrong password")
	}
	if store.Authenticate("bob", plain) {
		t.Fatal("expected auth failure for missing user")
	}
	names := store.List()
	if len(names) != 1 || names[0] != "alice" {
		t.Fatalf("list = %v", names)
	}
	// Reload from disk — only hash should persist.
	store2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !store2.Authenticate("alice", plain) {
		t.Fatal("reload: expected auth success")
	}
	if _, err := store.Add("alice"); err == nil {
		t.Fatal("expected duplicate error")
	}
}
