package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// auth verifies the Supabase-issued JWT and puts the user id on the request
// context. In dev (SUPABASE_JWT_SECRET unset) it falls back to the
// X-Dev-User header so the frontend works before auth is wired up.
func (a *App) auth(next http.Handler) http.Handler {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			dev := r.Header.Get("X-Dev-User")
			if dev == "" {
				dev = "dev-user"
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, dev)))
			return
		}
		raw := bearer(r)
		if raw == "" {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		claims, err := verifyHS256(raw, secret)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			writeErr(w, http.StatusUnauthorized, "token has no subject")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, sub)))
	})
}

func userIDFrom(r *http.Request) string {
	if v, ok := r.Context().Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// verifyHS256 checks a Supabase legacy JWT (HS256, signed with the project's
// JWT secret). If your project uses the newer asymmetric signing keys, switch
// this to fetching <project>/.well-known/jwks.json and verifying with RS256.
func verifyHS256(token, secret string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errString("malformed token")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if json.Unmarshal(headerJSON, &header) != nil || header.Alg != "HS256" {
		return nil, errString("unsupported alg")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	if !hmacEqual([]byte(parts[0]+"."+parts[1]), sig, []byte(secret)) {
		return nil, errString("signature mismatch")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				writeErr(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withCORS allows the Vite dev server to call the API directly. When you use
// the Vite proxy (default in this scaffold) CORS is not needed.
func withCORS(next http.Handler) http.Handler {
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Dev-User")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
