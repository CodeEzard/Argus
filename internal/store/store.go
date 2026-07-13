package store

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// create table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS events (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp  TEXT NOT NULL,
        metric     TEXT NOT NULL,
        value      REAL NOT NULL,
        zscore     REAL NOT NULL,
        severity   TEXT NOT NULL,
        diagnosis  TEXT NOT NULL,
        commands   TEXT NOT NULL,
        fix        TEXT NOT NULL,
        confidence REAL NOT NULL
    )`)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

type Event struct {
	ID         int
	Timestamp  string
	Metric     string
	Value      float64
	ZScore     float64
	Severity   string
	Diagnosis  string
	Commands   []string
	Fix        string
	Confidence float64
}

func (s *Store) Save(e *Event) error {
	cmdsJSON, err := json.Marshal(e.Commands)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT INTO events (timestamp, metric, value, zscore, severity, diagnosis, commands, fix, confidence)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Timestamp, e.Metric, e.Value, e.ZScore, e.Severity, e.Diagnosis, string(cmdsJSON), e.Fix, e.Confidence)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		e.ID = int(id)
	}
	return nil
}

func (s *Store) List() ([]Event, error) {
	rows, err := s.db.Query(`SELECT id, timestamp, metric, value, zscore, severity, diagnosis, commands, fix, confidence FROM events ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var cmdsStr string
		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Metric,
			&e.Value,
			&e.ZScore,
			&e.Severity,
			&e.Diagnosis,
			&cmdsStr,
			&e.Fix,
			&e.Confidence,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(cmdsStr), &e.Commands); err != nil {
			e.Commands = []string{}
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) GetByID(id int) (*Event, error) {
	var e Event
	var cmdsStr string
	err := s.db.QueryRow(`SELECT id, timestamp, metric, value, zscore, severity, diagnosis, commands, fix, confidence FROM events WHERE id = ?`, id).Scan(
		&e.ID, &e.Timestamp, &e.Metric, &e.Value, &e.ZScore,
		&e.Severity, &e.Diagnosis, &cmdsStr, &e.Fix, &e.Confidence,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(cmdsStr), &e.Commands)
	return &e, nil
}

func (s *Store) GetById(id int) (*Event, error) {
	return s.GetByID(id)
}
