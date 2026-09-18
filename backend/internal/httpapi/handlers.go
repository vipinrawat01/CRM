// Package httpapi wires the store and the AI client to HTTP handlers.
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"crm/internal/ai"
	"crm/internal/store"
)

type API struct {
	store *store.Store
	ai    *ai.Client
}

func New(s *store.Store, c *ai.Client) *API {
	return &API{store: s, ai: c}
}

// Routes registers all API routes on mux using Go 1.22's method+pattern matching.
func (a *API) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/leads", a.listLeads)
	mux.HandleFunc("POST /api/leads", a.createLead)
	mux.HandleFunc("GET /api/leads/{id}", a.getLead)
	mux.HandleFunc("PATCH /api/leads/{id}/stage", a.setStage)
	mux.HandleFunc("POST /api/leads/{id}/notes", a.addNote)
	mux.HandleFunc("POST /api/leads/{id}/tasks", a.addTask)
	mux.HandleFunc("PATCH /api/leads/{id}/tasks/{taskId}", a.setTaskDone)
	mux.HandleFunc("POST /api/leads/{id}/summarize", a.summarize)
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("encode error: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func handleStoreErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	log.Printf("store error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

// ---- handlers ----

func (a *API) listLeads(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	writeJSON(w, http.StatusOK, a.store.ListLeads(q))
}

func (a *API) getLead(w http.ResponseWriter, r *http.Request) {
	l, err := a.store.GetLead(r.PathValue("id"))
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

type createLeadReq struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Company   string `json:"company"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Source    string `json:"source"`
	DealValue int    `json:"dealValue"`
}

func (a *API) createLead(w http.ResponseWriter, r *http.Request) {
	var req createLeadReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Company = strings.TrimSpace(req.Company)
	if req.Name == "" || req.Company == "" {
		writeError(w, http.StatusUnprocessableEntity, "name and company are required")
		return
	}
	l := &store.Lead{
		Name: req.Name, Title: req.Title, Company: req.Company,
		Email: req.Email, Phone: req.Phone, Source: req.Source, DealValue: req.DealValue,
	}
	created, err := a.store.CreateLead(l)
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

type setStageReq struct {
	Stage string `json:"stage"`
}

var validStages = map[string]bool{
	string(store.StageNew): true, string(store.StageContacted): true,
	string(store.StageQualified): true, string(store.StageProposal): true,
	string(store.StageWon): true, string(store.StageLost): true,
}

func (a *API) setStage(w http.ResponseWriter, r *http.Request) {
	var req setStageReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validStages[req.Stage] {
		writeError(w, http.StatusUnprocessableEntity, "unknown stage")
		return
	}
	l, err := a.store.SetStage(r.PathValue("id"), store.Stage(req.Stage))
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

type addNoteReq struct {
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

var validNoteKinds = map[string]bool{"note": true, "call": true, "email": true, "meeting": true}

func (a *API) addNote(w http.ResponseWriter, r *http.Request) {
	var req addNoteReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		writeError(w, http.StatusUnprocessableEntity, "content is required")
		return
	}
	if req.Kind == "" {
		req.Kind = "note"
	}
	if !validNoteKinds[req.Kind] {
		writeError(w, http.StatusUnprocessableEntity, "unknown note kind")
		return
	}
	l, err := a.store.AddNote(r.PathValue("id"), store.Note{Kind: req.Kind, Content: req.Content})
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

type addTaskReq struct {
	Title   string  `json:"title"`
	DueDate *string `json:"dueDate"`
}

func (a *API) addTask(w http.ResponseWriter, r *http.Request) {
	var req addTaskReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}
	task := store.Task{Title: req.Title}
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := parseDate(*req.DueDate)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "dueDate must be YYYY-MM-DD")
			return
		}
		task.DueDate = &t
	}
	l, err := a.store.AddTask(r.PathValue("id"), task)
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

type setTaskDoneReq struct {
	Done bool `json:"done"`
}

func (a *API) setTaskDone(w http.ResponseWriter, r *http.Request) {
	var req setTaskDoneReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	l, err := a.store.SetTaskDone(r.PathValue("id"), r.PathValue("taskId"), req.Done)
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (a *API) summarize(w http.ResponseWriter, r *http.Request) {
	l, err := a.store.GetLead(r.PathValue("id"))
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	summary := a.ai.Summarize(r.Context(), l)
	updated, err := a.store.SetSummary(l.ID, summary)
	if err != nil {
		handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
