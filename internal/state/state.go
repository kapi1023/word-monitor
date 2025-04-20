package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ExamSlot struct {
	Day         string   `json:"day"`
	Time        string   `json:"time"`
	PracticeIDs []string `json:"practice_ids"`
	TheoryIDs   []string `json:"theory_ids"`
}

type Storage struct {
	path   string
	mu     sync.Mutex
	latest map[string][]ExamSlot
}

func New(path, _ string) (*Storage, error) {
	abs, _ := filepath.Abs(path)
	_ = os.MkdirAll(filepath.Dir(abs), 0755)
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		_ = os.WriteFile(abs, []byte("{}"), 0644)
	}
	s := &Storage{
		path:   abs,
		latest: make(map[string][]ExamSlot),
	}
	_ = s.load()
	return s, nil
}

func (s *Storage) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil
	}
	return json.Unmarshal(data, &s.latest)
}

func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s.latest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Storage) Get(key string) []ExamSlot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.latest[key]
}

func (s *Storage) Add(key string, slot ExamSlot) {
	s.mu.Lock()
	examSlots := s.latest[key]
	for _, examSlot := range examSlots {
		if examSlot.Day == slot.Day && examSlot.Time == slot.Time {
			s.mu.Unlock()
			return
		}
	}
	s.latest[key] = append(s.latest[key], slot)
	s.mu.Unlock()
	_ = s.Save()
}

func (s *Storage) Exists(key, day, time string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, slot := range s.latest[key] {
		if slot.Day == day && slot.Time == time {
			return true
		}
	}
	return false
}

func Key(wordID, category string) string {
	return wordID + ":" + category
}
