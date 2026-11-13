package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type App struct {
	db *gorm.DB
}

type ctxKey string

const userIDKey ctxKey = "userID"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := a.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		writeErr(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (a *App) listItems(w http.ResponseWriter, r *http.Request) {
	owner := userIDFrom(r)
	var items []Item
	q := a.db.Order("created_at desc").Limit(100)
	// Ownership is enforced here, in Go. The backend connects with a
	// privileged role, so Supabase RLS does NOT protect these queries.
	if owner != "" {
		q = q.Where("owner_id = ?", owner)
	}
	if err := q.Find(&items).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) createItem(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		writeErr(w, http.StatusBadRequest, "title is required")
		return
	}
	item := Item{Title: in.Title, Body: in.Body, OwnerID: userIDFrom(r)}
	if err := a.db.Create(&item).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "insert failed")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	q := a.db.Delete(&Item{}, id)
	if owner := userIDFrom(r); owner != "" {
		q = a.db.Where("owner_id = ?", owner).Delete(&Item{}, id)
	}
	if q.Error != nil || q.RowsAffected == 0 {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
