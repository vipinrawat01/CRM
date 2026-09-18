# Northstar CRM

A small B2B sales CRM with an AI-generated lead summary feature. Go backend
(stdlib only, JSON-file persistence) and a React + TypeScript frontend built
with Vite, in separate `backend/` and `frontend/` folders. Ships with five
realistic seed leads so the full workflow can be demonstrated immediately.

## Features

- Browse and search leads
- Open a lead and see its full detail record
- Add notes / activities (note, call, email, meeting) to a lead's timeline
- Move a lead through a sales pipeline (New → Contacted → Qualified →
  Proposal → Won/Lost)
- Create and complete follow-up tasks
- Generate an AI lead summary (who they are, what matters, what's happened,
  what's missing) via OpenAI, with a deterministic local fallback

## Project layout

```
backend/
  cmd/server/main.go        entrypoint: HTTP server, serves API + built frontend
  internal/store/           models + JSON-file persistence (CRUD, seed data)
  internal/httpapi/         REST API handlers
  internal/ai/              AI summary: OpenAI call + local rule-based fallback
  internal/env/             minimal .env file loader
  .env.example              copy to .env and fill in your OpenAI key
frontend/
  src/                      React + TypeScript app (Vite)
  src/components/           Sidebar, LeadDetail, SummarySection, TasksSection, NotesSection, NewLeadModal
  src/api.ts                typed API client
  src/types.ts              shared TypeScript types (mirrors backend/internal/store/models.go)
```

## Run it

Requires Go 1.22+ and Node 18+.

**1. Configure the backend**

```bash
cd backend
cp .env.example .env
# edit .env and set OPENAI_API_KEY=sk-...
```

**2. Build the frontend**

```bash
cd frontend
npm install
npm run build
```

**3. Run the backend**

```bash
cd backend
go run ./cmd/server
```

Open **http://localhost:8080** — this serves both the API and the built
frontend (`frontend/dist`). Five seeded leads load automatically the first
time (a fresh `backend/data.json` is created on first run).

### Frontend dev mode (optional)

For hot-reload while working on the UI, run the Vite dev server alongside
the backend instead of building:

```bash
# terminal 1
cd backend && go run ./cmd/server

# terminal 2
cd frontend && npm run dev
```

Open **http://localhost:5173** — Vite proxies `/api/*` requests to the Go
server on `:8080`.

### Environment variables (`backend/.env`)

| Var                 | Default              | Purpose                                                          |
|---------------------|----------------------|-------------------------------------------------------------------|
| `OPENAI_API_KEY`    | *(unset)*            | enables real AI summaries; without it, summaries use the local rule-based fallback |
| `OPENAI_MODEL`      | `gpt-4o-mini`        | override the model used for summaries                            |
| `ADDR`              | `:8080`              | listen address                                                    |
| `DATA_FILE`         | `data.json`          | where state is persisted                                          |
| `FRONTEND_DIR`      | `../frontend/dist`   | directory served as static files                                  |

Without an API key the app still fully works — every summary is generated
locally from the lead's own data and clearly labeled as a local fallback.

## Product decisions

- **Go stdlib only, no ORM/framework.** Go 1.22's `http.ServeMux` already
  does method+path-param routing, so a router library wasn't needed.
- **JSON file instead of a database.** Real persistence (survives restarts
  and refreshes) without needing a database server to run the demo.
  Explicitly not what would be used for concurrent multi-user production
  traffic.
- **Separate backend/frontend folders**, each independently runnable/
  deployable. The Go server also serves the built frontend so the whole app
  can run as a single process for a demo.
- **Missing-info detection is deterministic, not just AI-guessed**: the
  fallback (and the prompt given to the model) is built from an explicit
  field/notes/tasks check, so "what's missing" is grounded in the actual
  record, not a hallucination risk.
- **The AI feature never hard-fails.** Every summary request always
  returns *something* — model output, or a local fallback with the reason
  shown — because a sales rep shouldn't hit a dead end.

## API reference

| Method | Path                              | Purpose                     |
|--------|------------------------------------|------------------------------|
| GET    | `/api/leads?q=`                   | list/search leads            |
| POST   | `/api/leads`                      | create a lead                |
| GET    | `/api/leads/{id}`                 | lead detail                  |
| PATCH  | `/api/leads/{id}/stage`           | move pipeline stage          |
| POST   | `/api/leads/{id}/notes`           | log a note/call/email/meeting|
| POST   | `/api/leads/{id}/tasks`           | add a follow-up task         |
| PATCH  | `/api/leads/{id}/tasks/{taskId}`  | mark task done/undone        |
| POST   | `/api/leads/{id}/summarize`       | generate the AI lead summary |

All endpoints validated end-to-end during development: creation rejects
missing name/company (422), unknown IDs return 404, stage changes reject
invalid stage values, and summaries generate correctly both with and
without an API key.

## What's intentionally left out

Authentication/multi-user accounts, a real database, pagination (fine at
seed-data scale, not at 10k leads), editing/deleting existing leads and
notes, mobile layout polish.

## Next improvements

- Swap the JSON file for Postgres with row-level locking once more than one
  rep needs to write concurrently.
- Stream the AI summary response instead of blocking, so it doesn't read
  as "frozen" on a slow model call.

## Security/reliability risk to address before launch

The JSON file has no access control and no encryption at rest — anyone
with filesystem access can read every lead's email/phone/deal value in
plain text. Before launch this needs a real datastore with per-tenant
auth, and `OPENAI_API_KEY` needs to live in a secrets manager, not a
plain `.env` file on a shared machine.
