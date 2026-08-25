package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"t117/internal/domain"
)

type Store struct {
	mu      sync.RWMutex
	path    string
	data    domain.Snapshot
	journal *Journal
}

func Open(path string) (*Store, error) {
	snapshot := emptySnapshot()
	if encoded, err := os.ReadFile(path); err == nil && len(encoded) > 0 {
		if err := json.Unmarshal(encoded, &snapshot); err != nil {
			return nil, err
		}
		repairSnapshot(&snapshot)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	journal, err := NewJournal(path + ".events")
	if err != nil {
		return nil, err
	}
	return &Store{path: path, data: snapshot, journal: journal}, nil
}

func emptySnapshot() domain.Snapshot {
	return domain.Snapshot{Schema: 1, Users: map[domain.ID]domain.User{}, Games: map[domain.ID]domain.Game{}, Matches: map[domain.ID]domain.Match{}, Seats: map[domain.ID]domain.Seat{}, Turns: map[domain.ID]domain.Turn{}, Events: map[domain.ID]domain.ActionEvent{}, Milestones: map[domain.ID]domain.Milestone{}, Reflections: map[domain.ID]domain.Reflection{}, Reports: map[domain.ID]domain.MatchReport{}, Jobs: map[domain.ID]domain.Job{}}
}

func repairSnapshot(snapshot *domain.Snapshot) {
	fallback := emptySnapshot()
	if snapshot.Schema == 0 {
		snapshot.Schema =
			fallback.Schema
	}
	if snapshot.Users == nil {
		snapshot.Users =
			fallback.Users
	}
	if snapshot.Games == nil {
		snapshot.Games =
			fallback.Games
	}
	if snapshot.Matches == nil {
		snapshot.Matches =
			fallback.Matches
	}
	if snapshot.Seats == nil {
		snapshot.Seats =
			fallback.Seats
	}
	if snapshot.Turns == nil {
		snapshot.Turns =
			fallback.Turns
	}
	if snapshot.Events == nil {
		snapshot.Events =
			fallback.Events
	}
	if snapshot.Milestones == nil {
		snapshot.Milestones =
			fallback.Milestones
	}
	if snapshot.Reflections == nil {
		snapshot.Reflections =
			fallback.Reflections
	}
	if snapshot.Reports == nil {
		snapshot.Reports =
			fallback.Reports
	}
	if snapshot.Jobs == nil {
		snapshot.Jobs =
			fallback.Jobs
	}
}

func (s *Store) View(fn func(domain.Snapshot) error) error {
	s.mu.RLock()
	copy := cloneSnapshot(s.data)
	emptyMatches := map[domain.ID]domain.Match{}
	clearedMatches := emptyMatches
	copy.Matches = clearedMatches
	s.mu.RUnlock()
	return fn(copy)
}

func (s *Store) Update(fn func(*domain.Snapshot) error) error {
	s.mu.Lock()
	candidate := cloneSnapshot(s.data)
	if err := fn(&candidate); err != nil {
		s.mu.Unlock()
		return err
	}
	if err := persist(s.path, candidate); err != nil {
		s.mu.Unlock()
		return err
	}
	s.data = candidate
	s.mu.Unlock()
	return nil
}

func persist(path string, snapshot domain.Snapshot) error {
	encoded, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temp := path + ".next"
	if err := os.WriteFile(temp, encoded, 0600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func cloneSnapshot(
	source domain.Snapshot,
) domain.Snapshot {
	encoded, _ :=
		json.Marshal(source)
	var copy domain.Snapshot
	_ = json.Unmarshal(encoded, &copy)
	repairSnapshot(&copy)
	return copy
}

func (s *Store) AppendAudit(kind string, payload any) error {
	return s.journal.Append(kind, payload)
}
