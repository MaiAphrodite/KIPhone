package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Item struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Body      string    `gorm:"type:text" json:"body"`
	OwnerID   string    `gorm:"index;size:64" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := envOr("PORT", "8080")

	// PreferSimpleProtocol keeps GORM/pgx working through Supabase's
	// transaction-mode pooler (port 6543), which does not support
	// server-side prepared statements.
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	if err := db.AutoMigrate(&Item{}, &Student{}, &Account{}, &Disbursement{}, &Finding{}, &FindingEvent{}, &Batch{}, &Transaction{}, &SpendFinding{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	app := &App{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)
	mux.Handle("GET /api/items", app.auth(http.HandlerFunc(app.listItems)))
	mux.Handle("POST /api/items", app.auth(http.HandlerFunc(app.createItem)))
	mux.Handle("DELETE /api/items/{id}", app.auth(http.HandlerFunc(app.deleteItem)))

	// Integra — PRD §7
	mux.Handle("POST /api/seed/load", app.auth(http.HandlerFunc(app.loadSeed)))
	mux.Handle("POST /api/detect", app.auth(http.HandlerFunc(app.runDetect)))
	mux.Handle("GET /api/findings", app.auth(http.HandlerFunc(app.listFindings)))
	mux.Handle("GET /api/findings/{id}", app.auth(http.HandlerFunc(app.getFinding)))
	mux.Handle("POST /api/findings/{id}/status", app.auth(http.HandlerFunc(app.setFindingStatus)))
	mux.Handle("GET /api/stats", app.auth(http.HandlerFunc(app.stats)))
	mux.Handle("GET /api/eval", app.auth(http.HandlerFunc(app.eval)))
	mux.Handle("GET /api/batch", app.auth(http.HandlerFunc(app.batchReceipt)))
	mux.Handle("POST /api/batch/anchor", app.auth(http.HandlerFunc(app.anchorBatch)))

	// Integra — transaction feed (dummy partner-bank data)
	mux.Handle("POST /api/tx/seed", app.auth(http.HandlerFunc(app.txSeed)))
	mux.Handle("POST /api/tx/detect", app.auth(http.HandlerFunc(app.txDetect)))
	mux.Handle("GET /api/tx/findings", app.auth(http.HandlerFunc(app.txFindings)))
	mux.Handle("GET /api/tx/findings/{id}", app.auth(http.HandlerFunc(app.txFindingDetail)))
	mux.Handle("POST /api/tx/findings/{id}/status", app.auth(http.HandlerFunc(app.txStatus)))
	mux.Handle("GET /api/tx/students/{id}/statement", app.auth(http.HandlerFunc(app.txStatement)))
	mux.Handle("GET /api/tx/stats", app.auth(http.HandlerFunc(app.txStats)))
	mux.Handle("GET /api/tx/eval", app.auth(http.HandlerFunc(app.txEval)))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           withRecover(withCORS(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("backend listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
