package store

import (
	"reflect"
	"testing"
)

func TestStore(t *testing.T) {
	// Use an in-memory SQLite database for testing
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// 1. Verify list is empty initially
	events, err := s.List()
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}

	// 2. Save a test event
	e1 := &Event{
		Timestamp:  "2026-07-13T12:00:00Z",
		Metric:     "cpu_usage",
		Value:      0.85,
		ZScore:     3.2,
		Severity:   "medium",
		Diagnosis:  "High CPU consumption from Docker containers",
		Commands:   []string{"docker stats", "top"},
		Fix:        "Scale horizontally or optimize processes",
		Confidence: 0.90,
	}

	err = s.Save(e1)
	if err != nil {
		t.Fatalf("failed to save event: %v", err)
	}

	if e1.ID != 1 {
		t.Errorf("expected event ID to be set to 1, got %d", e1.ID)
	}

	// 3. Save a second event to test ordering and retrieval
	e2 := &Event{
		Timestamp:  "2026-07-13T12:05:00Z",
		Metric:     "memory_usage",
		Value:      0.95,
		ZScore:     4.5,
		Severity:   "high",
		Diagnosis:  "Out of memory risk detected",
		Commands:   []string{"free -m", "ps aux --sort=-%mem"},
		Fix:        "Restart target process or upgrade RAM",
		Confidence: 0.98,
	}

	err = s.Save(e2)
	if err != nil {
		t.Fatalf("failed to save event: %v", err)
	}

	if e2.ID != 2 {
		t.Errorf("expected event ID to be set to 2, got %d", e2.ID)
	}

	// 4. Verify List returns both events sorted by timestamp DESC
	events, err = s.List()
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Since they are ordered by timestamp DESC, e2 (12:05) should be first, and e1 (12:00) second
	if events[0].ID != e2.ID || events[0].Metric != e2.Metric {
		t.Errorf("expected first event to be %v, got %v", e2, events[0])
	}

	if events[1].ID != e1.ID || events[1].Metric != e1.Metric {
		t.Errorf("expected second event to be %v, got %v", e1, events[1])
	}

	// Deep equal check on Commands slice
	if !reflect.DeepEqual(events[0].Commands, e2.Commands) {
		t.Errorf("expected commands to be %v, got %v", e2.Commands, events[0].Commands)
	}
	if !reflect.DeepEqual(events[1].Commands, e1.Commands) {
		t.Errorf("expected commands to be %v, got %v", e1.Commands, events[1].Commands)
	}
}
