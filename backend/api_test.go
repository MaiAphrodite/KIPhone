package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newAPITestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL unset: skipping DB-backed httptest contract; run backend/test_api.sh against a live server")
	}
	t.Setenv("SUPABASE_JWT_SECRET", "")

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open TEST_DATABASE_URL: %v", err)
	}
	if err := db.AutoMigrate(&Item{}, &Student{}, &Account{}, &Disbursement{}, &Finding{}, &FindingEvent{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	app := &App{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)
	mux.Handle("POST /api/seed/load", app.auth(http.HandlerFunc(app.loadSeed)))
	mux.Handle("POST /api/detect", app.auth(http.HandlerFunc(app.runDetect)))
	mux.Handle("GET /api/findings", app.auth(http.HandlerFunc(app.listFindings)))
	mux.Handle("GET /api/findings/{id}", app.auth(http.HandlerFunc(app.getFinding)))
	mux.Handle("POST /api/findings/{id}/status", app.auth(http.HandlerFunc(app.setFindingStatus)))
	mux.Handle("GET /api/stats", app.auth(http.HandlerFunc(app.stats)))
	mux.Handle("GET /api/eval", app.auth(http.HandlerFunc(app.eval)))

	srv := httptest.NewServer(withRecover(withCORS(mux)))
	t.Cleanup(srv.Close)
	return srv
}

func apiCall(t *testing.T, client *http.Client, method, url string, body any) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("X-Dev-User", "qa")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	out := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("%s %s returned non-JSON (%d): %s", method, url, resp.StatusCode, raw)
		}
	}
	return resp.StatusCode, out
}

func wantStatus(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: status %d, want %d", name, got, want)
	}
}

func wantNumber(t *testing.T, name string, v any, want float64) {
	t.Helper()
	got, ok := v.(float64)
	if !ok {
		t.Fatalf("%s: %v is not a number", name, v)
	}
	if got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func TestAPIContract(t *testing.T) {
	srv := newAPITestServer(t)
	client := srv.Client()
	base := srv.URL

	code, body := apiCall(t, client, http.MethodGet, base+"/health", nil)
	wantStatus(t, "GET /health", code, http.StatusOK)
	if body["ok"] != "true" {
		t.Fatalf("health.ok = %v, want true", body["ok"])
	}

	code, body = apiCall(t, client, http.MethodPost, base+"/api/seed/load", nil)
	wantStatus(t, "POST /api/seed/load", code, http.StatusOK)
	wantNumber(t, "seed.students", body["students"], 215)
	wantNumber(t, "seed.accounts", body["accounts"], 215)
	wantNumber(t, "seed.disbursements", body["disbursements"], 215)
	wantNumber(t, "seed.planted", body["planted"], 15)

	code, body = apiCall(t, client, http.MethodGet, base+"/api/stats", nil)
	wantStatus(t, "GET /api/stats", code, http.StatusOK)
	wantNumber(t, "stats.records after seed", body["records"], 215)
	wantNumber(t, "stats.findings before detect", body["findings"], 0)

	code, body = apiCall(t, client, http.MethodPost, base+"/api/detect", nil)
	wantStatus(t, "POST /api/detect", code, http.StatusOK)
	if created, ok := body["findings_created"].(float64); !ok || created <= 0 {
		t.Fatalf("detect.findings_created = %v, want > 0", body["findings_created"])
	}

	code, body = apiCall(t, client, http.MethodGet, base+"/api/findings", nil)
	wantStatus(t, "GET /api/findings", code, http.StatusOK)
	rows, ok := body["findings"].([]any)
	if !ok {
		t.Fatalf("findings is not an array: %T", body["findings"])
	}
	if len(rows) != 15 {
		t.Fatalf("findings rows = %d, want 15", len(rows))
	}
	first, ok := rows[0].(map[string]any)
	if !ok {
		t.Fatalf("first finding is not an object: %T", rows[0])
	}
	if first["status"] != StatusBaru {
		t.Fatalf("first finding status = %v, want %s", first["status"], StatusBaru)
	}
	findingID, ok := first["id"].(float64)
	if !ok {
		t.Fatalf("first finding id is not a number: %v", first["id"])
	}

	code, body = apiCall(t, client, http.MethodPost, base+"/api/findings/"+formatID(findingID)+"/status", map[string]string{
		"to_status": StatusSalahDeteksi,
		"note":      "qa contract",
	})
	wantStatus(t, "transition baru->salah_deteksi", code, http.StatusOK)
	event, ok := body["event"].(map[string]any)
	if !ok {
		t.Fatalf("transition response missing event: %v", body)
	}
	if event["from_status"] != StatusBaru {
		t.Fatalf("transition event.from_status = %v, want %s", event["from_status"], StatusBaru)
	}

	code, body = apiCall(t, client, http.MethodGet, base+"/api/findings/"+formatID(findingID), nil)
	wantStatus(t, "GET /api/findings/{id}", code, http.StatusOK)
	persisted, ok := body["finding"].(map[string]any)
	if !ok {
		t.Fatalf("get finding response missing finding: %v", body)
	}
	if persisted["status"] != StatusSalahDeteksi {
		t.Fatalf("persisted status = %v, want %s", persisted["status"], StatusSalahDeteksi)
	}

	code, _ = apiCall(t, client, http.MethodPost, base+"/api/findings/"+formatID(findingID)+"/status", map[string]string{
		"to_status": StatusBaru,
	})
	wantStatus(t, "invalid transition salah_deteksi->baru", code, http.StatusConflict)

	code, body = apiCall(t, client, http.MethodGet, base+"/api/eval", nil)
	wantStatus(t, "GET /api/eval", code, http.StatusOK)
	if recall, ok := body["recall"].(float64); !ok || recall < 0.7 {
		t.Fatalf("eval.recall = %v, want >= 0.7", body["recall"])
	}
	if p10, ok := body["precision_at_10"].(float64); !ok || p10 < 0.8 {
		t.Fatalf("eval.precision_at_10 = %v, want >= 0.8", body["precision_at_10"])
	}
}

func formatID(v float64) string {
	return strconv.FormatUint(uint64(v), 10)
}
