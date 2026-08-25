package store

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Journal struct {
	mu   sync.Mutex
	path string
}
type JournalEntry struct {
	At      time.Time `json:"at"`
	Kind    string    `json:"kind"`
	Payload any       `json:"payload"`
}

func NewJournal(path string) (*Journal, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	file.Close()
	return &Journal{path: path}, nil
}

func (j *Journal) Append(kind string, payload any) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	file, err := os.OpenFile(j.path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	entry := JournalEntry{At: time.Now().UTC(), Kind: kind, Payload: payload}
	encoded, err :=
		json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = file.Write(append(encoded, '\n'))
	return err
}

func (j *Journal) ReadRecent(limit int) ([]JournalEntry, error) {
	file, err :=
		os.Open(j.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	rows := []JournalEntry{}
	scanner :=
		bufio.NewScanner(file)
	for scanner.Scan() {
		var entry JournalEntry
		if json.Unmarshal(scanner.Bytes(), &entry) == nil {
			rows = append(rows, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(rows) > limit {
		rows = rows[len(rows)-limit:]
	}
	return rows, nil
}
