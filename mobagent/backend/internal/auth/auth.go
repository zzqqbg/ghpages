package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

type Store struct {
	mu       sync.RWMutex
	byToken  map[string]*Account
	byID     map[string]*Account
	defaultID string
}

type fileConfig struct {
	DefaultAccount string    `json:"defaultAccount"`
	Accounts       []Account `json:"accounts"`
}

func Load(dataDir string) *Store {
	s := &Store{
		byToken:   make(map[string]*Account),
		byID:      make(map[string]*Account),
		defaultID: "demo",
	}
	if dataDir == "" {
		dataDir = filepath.Join("..", "data")
	}
	path := os.Getenv("MOBAGENT_ACCOUNTS")
	if path == "" {
		path = filepath.Join(dataDir, "accounts.json")
	}
	if raw, err := os.ReadFile(path); err == nil {
		var fc fileConfig
		if json.Unmarshal(raw, &fc) == nil {
			for i := range fc.Accounts {
				s.add(&fc.Accounts[i])
			}
			if fc.DefaultAccount != "" {
				s.defaultID = fc.DefaultAccount
			}
		}
	}
	if len(s.byToken) == 0 {
		s.add(&Account{ID: "demo", Name: "Demo", Token: "demo-token-change-me"})
	}
	if env := os.Getenv("MOBAGENT_TOKENS"); env != "" {
		for _, pair := range strings.Split(env, ",") {
			parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
			if len(parts) != 2 {
				continue
			}
			s.add(&Account{ID: parts[0], Name: parts[0], Token: parts[1]})
		}
	}
	return s
}

func (s *Store) add(a *Account) {
	if a.ID == "" || a.Token == "" {
		return
	}
	cp := *a
	s.byToken[cp.Token] = &cp
	s.byID[cp.ID] = &cp
}

func (s *Store) ValidateToken(token string) (*Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.byToken[token]
	return a, ok
}

func (s *Store) DefaultAccountID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defaultID
}

func (s *Store) Get(id string) (*Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.byID[id]
	return a, ok
}

func ExtractToken(authHeader, queryToken, headerToken string) string {
	if t := strings.TrimSpace(queryToken); t != "" {
		return t
	}
	if t := strings.TrimSpace(headerToken); t != "" {
		return t
	}
	h := strings.TrimSpace(authHeader)
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return h
}
