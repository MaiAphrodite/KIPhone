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
	if err := db.AutoMigrate(&Item{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	app := &App{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)
	mux.Handle("GET /api/items", app.auth(http.HandlerFunc(app.listItems)))
	mux.Handle("POST /api/items", app.auth(http.HandlerFunc(app.createItem)))
	mux.Handle("DELETE /api/items/{id}", app.auth(http.HandlerFunc(app.deleteItem)))

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
