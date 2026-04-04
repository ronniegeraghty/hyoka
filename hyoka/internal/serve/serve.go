// Package serve provides a local web server for browsing evaluation reports.
package serve

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ronniegeraghty/hyoka/internal/prompt"
)

// Options configures the serve command.
type Options struct {
	ReportsDir string
	DocsDir    string
	SiteDir    string
	PromptsDir string
	Port       int
}

// DocInfo holds metadata about a documentation file.
type DocInfo struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

// internalDocs lists documentation files that should be excluded from the API.
var internalDocs = map[string]bool{
	"cleanup-plan":   true,
	"eval-tool-plan": true,
}

// Start launches a local HTTP server for browsing reports.
func Start(opts Options) error {
	if opts.Port == 0 {
		opts.Port = 8080
	}

	abs, err := filepath.Abs(opts.ReportsDir)
	if err != nil {
		return fmt.Errorf("resolving reports dir: %w", err)
	}
	opts.ReportsDir = abs

	if opts.DocsDir != "" {
		if d, err := filepath.Abs(opts.DocsDir); err == nil {
			opts.DocsDir = d
		}
	}
	if opts.SiteDir != "" {
		if d, err := filepath.Abs(opts.SiteDir); err == nil {
			opts.SiteDir = d
		}
	}
	if opts.PromptsDir != "" {
		if d, err := filepath.Abs(opts.PromptsDir); err == nil {
			opts.PromptsDir = d
		}
	}

	mux := buildMux(opts)

	addr := fmt.Sprintf(":%d", opts.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)
	fmt.Printf("🌐 Serving evaluation reports at %s\n", url)
	fmt.Printf("   Reports directory: %s\n", opts.ReportsDir)
	if opts.SiteDir != "" {
		fmt.Printf("   Site directory:    %s\n", opts.SiteDir)
	}
	if opts.DocsDir != "" {
		fmt.Printf("   Docs directory:    %s\n", opts.DocsDir)
	}
	if opts.PromptsDir != "" {
		fmt.Printf("   Prompts directory: %s\n", opts.PromptsDir)
	}
	fmt.Printf("   Press Ctrl+C to stop\n\n")

	return http.Serve(listener, corsMiddleware(mux))
}

// buildMux creates the HTTP handler with all routes.
func buildMux(opts Options) *http.ServeMux {
	mux := http.NewServeMux()

	// --- API: runs ---
	mux.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		handleAPIRuns(w, r, opts.ReportsDir)
	})
	mux.HandleFunc("/api/runs/", func(w http.ResponseWriter, r *http.Request) {
		handleAPIRunDetail(w, r, opts.ReportsDir)
	})

	// --- API: docs ---
	if opts.DocsDir != "" {
		mux.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
			handleAPIDocs(w, r, opts.DocsDir)
		})
		mux.HandleFunc("/api/docs/", func(w http.ResponseWriter, r *http.Request) {
			handleAPIDocDetail(w, r, opts.DocsDir)
		})
	}

	// --- API: prompts ---
	if opts.PromptsDir != "" {
		mux.HandleFunc("/api/prompts", func(w http.ResponseWriter, r *http.Request) {
			handleAPIPrompts(w, r, opts.PromptsDir)
		})
		mux.HandleFunc("/api/prompts/", func(w http.ResponseWriter, r *http.Request) {
			handleAPIPromptDetail(w, r, opts.PromptsDir)
		})
	}

	// --- Static file serving for raw report files ---
	reportFS := http.FileServer(http.Dir(opts.ReportsDir))
	mux.Handle("/reports/", http.StripPrefix("/reports/", reportFS))

	// --- SPA fallback / site serving ---
	mux.HandleFunc("/", spaHandler(opts.SiteDir))

	return mux
}

// corsMiddleware adds CORS headers to all responses for dev-server compatibility.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Runs API handlers ---

func handleAPIRuns(w http.ResponseWriter, _ *http.Request, reportsDir string) {
	runs, err := listRunSummaries(reportsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, runs)
}

func handleAPIRunDetail(w http.ResponseWriter, r *http.Request, reportsDir string) {
	// Route: /api/runs/{runId} or /api/runs/{runId}/eval?path=...
	rest := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	if rest == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.SplitN(rest, "/", 2)
	runID := parts[0]

	// /api/runs/{runId}/eval?path=...
	if len(parts) == 2 && parts[1] == "eval" {
		handleAPIEval(w, r, reportsDir, runID)
		return
	}

	// /api/runs/{runId} — return full summary.json
	if len(parts) == 1 {
		summaryPath := filepath.Join(reportsDir, runID, "summary.json")
		serveJSONFile(w, r, summaryPath)
		return
	}

	http.NotFound(w, r)
}

func handleAPIEval(w http.ResponseWriter, r *http.Request, reportsDir, runID string) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		http.Error(w, `missing "path" query parameter`, http.StatusBadRequest)
		return
	}

	// Prevent directory traversal
	cleaned := filepath.Clean(relPath)
	if strings.Contains(cleaned, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(reportsDir, runID, cleaned)
	serveJSONFile(w, r, fullPath)
}

// --- Docs API handlers ---

func handleAPIDocs(w http.ResponseWriter, _ *http.Request, docsDir string) {
	docs, err := listDocs(docsDir)
	if err != nil {
		slog.Error("listing docs", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, docs)
}

func handleAPIDocDetail(w http.ResponseWriter, r *http.Request, docsDir string) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/docs/")
	if slug == "" || strings.Contains(slug, "/") || strings.Contains(slug, "..") {
		http.NotFound(w, r)
		return
	}

	if internalDocs[slug] {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join(docsDir, slug+".md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	title := extractMarkdownTitle(string(content))
	doc := DocInfo{
		Slug:    slug,
		Title:   title,
		Content: string(content),
	}
	writeJSON(w, doc)
}

// --- Prompts API handlers ---

func handleAPIPrompts(w http.ResponseWriter, _ *http.Request, promptsDir string) {
	prompts, err := prompt.LoadPrompts(promptsDir)
	if err != nil {
		slog.Error("loading prompts", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, prompts)
}

func handleAPIPromptDetail(w http.ResponseWriter, r *http.Request, promptsDir string) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/prompts/")
	if slug == "" || strings.Contains(slug, "..") {
		http.NotFound(w, r)
		return
	}

	prompts, err := prompt.LoadPrompts(promptsDir)
	if err != nil {
		slog.Error("loading prompts", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, p := range prompts {
		if p.ID == slug {
			writeJSON(w, p)
			return
		}
	}

	http.NotFound(w, r)
}

// --- SPA handler ---

func spaHandler(siteDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no site directory configured, return a minimal fallback
		if siteDir == "" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Site directory not configured. Use --site-dir to point to the built React site.")
			return
		}

		// Try to serve static file from site dir
		filePath := filepath.Join(siteDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}

		// SPA fallback: serve index.html for client-side routing
		indexPath := filepath.Join(siteDir, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	}
}

// --- Data helpers ---

// listRunSummaries reads run directories and returns their full summary.json content.
func listRunSummaries(reportsDir string) ([]json.RawMessage, error) {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("reading reports dir: %w", err)
	}

	type summaryEntry struct {
		runID string
		data  json.RawMessage
	}

	var items []summaryEntry
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "trends" {
			continue
		}

		summaryPath := filepath.Join(reportsDir, e.Name(), "summary.json")
		data, err := os.ReadFile(summaryPath)
		if err != nil {
			// Include a minimal entry for runs without summary.json
			minimal, _ := json.Marshal(map[string]string{"run_id": e.Name()})
			items = append(items, summaryEntry{runID: e.Name(), data: minimal})
			continue
		}
		items = append(items, summaryEntry{runID: e.Name(), data: data})
	}

	// Sort newest first
	sort.Slice(items, func(i, j int) bool {
		return items[i].runID > items[j].runID
	})

	result := make([]json.RawMessage, len(items))
	for i, item := range items {
		result[i] = item.data
	}
	return result, nil
}

// listDocs reads the docs directory and returns metadata for each public doc.
func listDocs(docsDir string) ([]DocInfo, error) {
	entries, err := os.ReadDir(docsDir)
	if err != nil {
		return nil, fmt.Errorf("reading docs dir: %w", err)
	}

	var docs []DocInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		slug := strings.TrimSuffix(e.Name(), ".md")
		if internalDocs[slug] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(docsDir, e.Name()))
		if err != nil {
			continue
		}

		docs = append(docs, DocInfo{
			Slug:  slug,
			Title: extractMarkdownTitle(string(content)),
		})
	}

	return docs, nil
}

// extractMarkdownTitle returns the text of the first `# ` heading, or the slug.
func extractMarkdownTitle(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// --- JSON helpers ---

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encoding JSON response", "error", err)
	}
}

func serveJSONFile(w http.ResponseWriter, _ *http.Request, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
