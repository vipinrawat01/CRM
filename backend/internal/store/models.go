package store

import "time"

// Stage is a pipeline stage for a deal.
type Stage string

const (
	StageNew        Stage = "New"
	StageContacted  Stage = "Contacted"
	StageQualified  Stage = "Qualified"
	StageProposal   Stage = "Proposal"
	StageWon        Stage = "Won"
	StageLost       Stage = "Lost"
)

// StageOrder defines the display/kanban order of pipeline stages.
var StageOrder = []Stage{StageNew, StageContacted, StageQualified, StageProposal, StageWon, StageLost}

// Note is a free-form note or logged activity (call, email, meeting) on a lead.
type Note struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"` // "note" | "call" | "email" | "meeting"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// Task is a follow-up task tied to a lead.
type Task struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	DueDate *time.Time `json:"dueDate,omitempty"`
	Done    bool       `json:"done"`
}

// Summary is the AI-generated lead summary, cached on the lead.
type Summary struct {
	Who           string    `json:"who"`
	WhatMatters   string    `json:"whatMatters"`
	WhatHappened  string    `json:"whatHappened"`
	MissingInfo   []string  `json:"missingInfo"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Source        string    `json:"source"` // "model" | "fallback"
	FallbackNote  string    `json:"fallbackNote,omitempty"`
}

// Lead is a sales lead / account.
type Lead struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title,omitempty"`
	Company     string    `json:"company"`
	Email       string    `json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Source      string    `json:"source,omitempty"`
	DealValue   int       `json:"dealValue,omitempty"`
	Stage       Stage     `json:"stage"`
	CreatedAt   time.Time `json:"createdAt"`
	Notes       []Note    `json:"notes"`
	Tasks       []Task    `json:"tasks"`
	Summary     *Summary  `json:"summary,omitempty"`
}
