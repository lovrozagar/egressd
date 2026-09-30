package users

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

// Record is a single client identity. Password is never stored in plaintext.
type Record struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// Store is a file-backed user database (JSON with bcrypt hashes).
type Store struct {
	path string
	mu   sync.RWMutex
	by   map[string]Record
}

type fileShape struct {
	Users []Record `json:"users"`
}

// Open loads or creates an empty store at path.
func Open(path string) (*Store, error) {
	s := &Store{
		path: path,
		by:   make(map[string]Record),
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("users: read %s: %w", path, err)
	}
	if len(data) == 0 {
		return s, nil
	}
	var shape fileShape
	if err := json.Unmarshal(data, &shape); err != nil {
		return nil, fmt.Errorf("users: parse %s: %w", path, err)
	}
	for _, r := range shape.Users {
		s.by[r.Username] = r
	}
	return s, nil
}

func (s *Store) saveLocked() error {
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("users: mkdir: %w", err)
		}
	}
	shape := fileShape{Users: make([]Record, 0, len(s.by))}
	for _, r := range s.by {
		shape.Users = append(shape.Users, r)
	}
	sort.Slice(shape.Users, func(i, j int) bool {
		return shape.Users[i].Username < shape.Users[j].Username
	})
	data, err := json.MarshalIndent(shape, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("users: write: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("users: rename: %w", err)
	}
	return nil
}

// Add creates a user with a random password. Returns the plaintext password once.
func (s *Store) Add(username string) (plaintext string, err error) {
	if username == "" {
		return "", fmt.Errorf("users: username required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.by[username]; exists {
		return "", fmt.Errorf("users: %q already exists", username)
	}
	plain, err := randomPassword(24)
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("users: hash: %w", err)
	}
	s.by[username] = Record{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.saveLocked(); err != nil {
		delete(s.by, username)
		return "", err
	}
	return plain, nil
}

// List returns usernames sorted (no secrets).
func (s *Store) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.by))
	for n := range s.by {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Authenticate checks username/password against bcrypt hashes.
func (s *Store) Authenticate(username, password string) bool {
	s.mu.RLock()
	rec, ok := s.by[username]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(rec.PasswordHash), []byte(password)) == nil
}

// Valid implements socks5.CredentialStore.
func (s *Store) Valid(user, password, _ string) bool {
	return s.Authenticate(user, password)
}

// Count returns number of users.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.by)
}

// Path returns the store file path.
func (s *Store) Path() string {
	return s.path
}

func randomPassword(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("users: random: %w", err)
	}
	// URL-safe, no padding — easy to paste into proxy URLs.
	return base64.RawURLEncoding.EncodeToString(b), nil
}
