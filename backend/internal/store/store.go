// Package store implements a tiny persistence layer for the CRM.
//
// Design decision: a single JSON file guarded by a mutex, instead of a real
// database. For a scoped take-home demo this keeps the project dependency
// free (stdlib only), trivial to inspect ("cat data.json"), and still gives
// us real persistence across restarts/refreshes. It would NOT be the right
// choice for concurrent multi-user production traffic (file writes serialize
// under one lock, no transactions across processes) - see README for what a
// production version would use instead.
package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type data struct {
	Leads map[string]*Lead `json:"leads"`
}

// Store is a thread-safe, file-persisted lead store.
type Store struct {
	mu   sync.RWMutex
	path string
	data data
}

// New loads the store from path, seeding sample data if the file doesn't exist.
func New(path string) (*Store, error) {
	s := &Store{path: path, data: data{Leads: map[string]*Lead{}}}
	if b, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		return s, nil
	}
	s.seed()
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return s, nil
}

func newID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%x", prefix, b)
}

// saveLocked writes the current state to disk. Caller must hold s.mu.
func (s *Store) saveLocked() error {
	tmp := s.path + ".tmp"
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path) // atomic on the same filesystem
}

// ListLeads returns all leads, optionally filtered by a case-insensitive
// substring match against name/company/email, sorted newest first.
func (s *Store) ListLeads(query string) []*Lead {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]*Lead, 0, len(s.data.Leads))
	for _, l := range s.data.Leads {
		if q == "" ||
			strings.Contains(strings.ToLower(l.Name), q) ||
			strings.Contains(strings.ToLower(l.Company), q) ||
			strings.Contains(strings.ToLower(l.Email), q) {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *Store) GetLead(id string) (*Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.data.Leads[id]
	if !ok {
		return nil, ErrNotFound
	}
	return l, nil
}

func (s *Store) CreateLead(l *Lead) (*Lead, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l.ID = newID("lead")
	l.CreatedAt = time.Now()
	if l.Stage == "" {
		l.Stage = StageNew
	}
	if l.Notes == nil {
		l.Notes = []Note{}
	}
	if l.Tasks == nil {
		l.Tasks = []Task{}
	}
	s.data.Leads[l.ID] = l
	return l, s.saveLocked()
}

// UpdateLead applies a mutation function to a lead under the write lock and persists.
func (s *Store) UpdateLead(id string, mutate func(*Lead) error) (*Lead, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.data.Leads[id]
	if !ok {
		return nil, ErrNotFound
	}
	if err := mutate(l); err != nil {
		return nil, err
	}
	return l, s.saveLocked()
}

func (s *Store) AddNote(leadID string, n Note) (*Lead, error) {
	n.ID = newID("note")
	n.CreatedAt = time.Now()
	return s.UpdateLead(leadID, func(l *Lead) error {
		l.Notes = append([]Note{n}, l.Notes...) // newest first
		return nil
	})
}

func (s *Store) AddTask(leadID string, t Task) (*Lead, error) {
	t.ID = newID("task")
	return s.UpdateLead(leadID, func(l *Lead) error {
		l.Tasks = append(l.Tasks, t)
		return nil
	})
}

func (s *Store) SetTaskDone(leadID, taskID string, done bool) (*Lead, error) {
	return s.UpdateLead(leadID, func(l *Lead) error {
		for i := range l.Tasks {
			if l.Tasks[i].ID == taskID {
				l.Tasks[i].Done = done
				return nil
			}
		}
		return ErrNotFound
	})
}

func (s *Store) SetStage(leadID string, stage Stage) (*Lead, error) {
	return s.UpdateLead(leadID, func(l *Lead) error {
		l.Stage = stage
		return nil
	})
}

func (s *Store) SetSummary(leadID string, sum *Summary) (*Lead, error) {
	return s.UpdateLead(leadID, func(l *Lead) error {
		l.Summary = sum
		return nil
	})
}

// seed populates a small, deliberately uneven set of sample leads so the
// demo can show: a lead with rich history, a brand-new lead with nothing on
// it, and a lead with obvious missing fields (for the "missing info" part of
// the AI summary).
func (s *Store) seed() {
	now := time.Now()
	mk := func(daysAgo int) time.Time { return now.Add(-time.Duration(daysAgo) * 24 * time.Hour) }
	due := func(daysFromNow int) *time.Time {
		t := now.Add(time.Duration(daysFromNow) * 24 * time.Hour)
		return &t
	}

	leads := []*Lead{
		{
			ID: newID("lead"), Name: "Priya Menon", Title: "Head of Ops", Company: "Northwind Logistics",
			Email: "priya.menon@northwindlog.com", Phone: "+91 98765 43210", Source: "Inbound demo request",
			DealValue: 480000, Stage: StageQualified, CreatedAt: mk(14),
			Notes: []Note{
				{ID: newID("note"), Kind: "call", Content: "Discovery call. They run 40 delivery routes/day, current tool has no live tracking. Budget owner is Priya, technical eval by their IT lead Ramesh.", CreatedAt: mk(13)},
				{ID: newID("note"), Kind: "email", Content: "Sent pricing for the Growth tier (up to 50 vehicles). Asked for a security questionnaire.", CreatedAt: mk(9)},
				{ID: newID("note"), Kind: "meeting", Content: "Demo with Ramesh (IT) went well. Main concern: SSO support and data residency in India.", CreatedAt: mk(4)},
			},
			Tasks: []Task{
				{ID: newID("task"), Title: "Send security questionnaire answers", DueDate: due(-1), Done: true},
				{ID: newID("task"), Title: "Follow up on SSO/data residency concerns", DueDate: due(2), Done: false},
			},
		},
		{
			ID: newID("lead"), Name: "David Okafor", Company: "Solace Retail Group",
			Email: "d.okafor@solaceretail.com", Source: "LinkedIn outbound", Stage: StageContacted,
			DealValue: 150000, CreatedAt: mk(6),
			Notes: []Note{
				{ID: newID("note"), Kind: "note", Content: "Cold outreach reply: interested but says timing is 'maybe next quarter'.", CreatedAt: mk(5)},
			},
			Tasks: []Task{
				{ID: newID("task"), Title: "Check back in 3 weeks re: Q on timing", DueDate: due(5), Done: false},
			},
		},
		{
			// Deliberately sparse: no phone, no title, no notes, no tasks.
			ID: newID("lead"), Name: "Aiko Tanaka", Company: "Kiraboshi Foods",
			Email: "aiko.tanaka@kiraboshi.co.jp", Source: "Website form", Stage: StageNew,
			CreatedAt: mk(1), Notes: []Note{}, Tasks: []Task{},
		},
		{
			ID: newID("lead"), Name: "Marcus Lindqvist", Title: "VP Sales", Company: "Nordfrost AB",
			Email: "marcus.l@nordfrost.se", Phone: "+46 70 123 4567", Source: "Trade show",
			DealValue: 920000, Stage: StageProposal, CreatedAt: mk(21),
			Notes: []Note{
				{ID: newID("note"), Kind: "meeting", Content: "Met at LogiTech Expo. Strong interest in the pipeline automation module specifically.", CreatedAt: mk(20)},
				{ID: newID("note"), Kind: "call", Content: "Walked through proposal. They want a 90-day pilot with 3 warehouses before committing company-wide.", CreatedAt: mk(10)},
				{ID: newID("note"), Kind: "email", Content: "Sent revised proposal reflecting the pilot structure and reduced first-quarter pricing.", CreatedAt: mk(2)},
			},
			Tasks: []Task{
				{ID: newID("task"), Title: "Confirm pilot start date", DueDate: due(3), Done: false},
			},
		},
		{
			ID: newID("lead"), Name: "Sofia Reyes", Title: "Procurement Manager", Company: "Estrella Manufacturing",
			Email: "sreyes@estrellamfg.com", Phone: "+52 55 4433 2211", Source: "Referral",
			DealValue: 60000, Stage: StageLost, CreatedAt: mk(30),
			Notes: []Note{
				{ID: newID("note"), Kind: "call", Content: "Went with a competitor already embedded in their SAP stack. Left the door open for a 2027 re-eval.", CreatedAt: mk(8)},
			},
			Tasks: []Task{},
		},
	}
	for _, l := range leads {
		s.data.Leads[l.ID] = l
	}
}
