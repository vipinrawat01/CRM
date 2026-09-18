// Command server runs the CRM backend REST API.
//
// In dev, run the Vite dev server separately (frontend/, npm run dev on
// :5173) - it proxies /api to this server on :8080. For the demo/prod build,
// this server also serves frontend/dist as static files, so `npm run build`
// + `go run ./cmd/server` is a single backend to run and one URL to open.
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"crm/internal/ai"
	"crm/internal/env"
	"crm/internal/httpapi"
	"crm/internal/store"
)

func main() {
	_ = env.Load(".env")
	_ = env.Load(filepath.Join("..", ".env"))

	dataPath := envOr("DATA_FILE", "data.json")
	addr := envOr("ADDR", ":8080")
	frontendDir := envOr("FRONTEND_DIR", filepath.Join("..", "frontend", "dist"))

	s, err := store.New(dataPath)
	if err != nil {
		log.Fatalf("loading store: %v", err)
	}

	client := ai.NewClient()
	api := httpapi.New(s, client)

	mux := http.NewServeMux()
	api.Routes(mux)
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	log.Printf("CRM listening on %s (data file: %s, frontend: %s)", addr, dataPath, frontendDir)
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Printf("OPENAI_API_KEY not set - lead summaries will use the local fallback")
	}
	log.Fatal(http.ListenAndServe(addr, mux))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
