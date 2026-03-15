package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	repo *Repository
}

func NewServer(repo *Repository) *Server {
	return &Server{repo: repo}
}

func parseIntQuery(r *http.Request, key string) (*int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *Server) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	pagePtr, err := parseIntQuery(r, "page")
	if err != nil {
		http.Error(w, "invalid page", http.StatusBadRequest)
		return
	}
	pageSizePtr, err := parseIntQuery(r, "pageSize")
	if err != nil {
		http.Error(w, "invalid pageSize", http.StatusBadRequest)
		return
	}
	page := 1
	pageSize := 10
	if pagePtr != nil {
		page = *pagePtr
	}
	if pageSizePtr != nil {
		pageSize = *pageSizePtr
	}

	filter := UserFilter{}

	if idPtr, err := parseIntQuery(r, "id"); err == nil && idPtr != nil {
		filter.ID = idPtr
	}
	if name := strings.TrimSpace(r.URL.Query().Get("name")); name != "" {
		filter.Name = &name
	}
	if email := strings.TrimSpace(r.URL.Query().Get("email")); email != "" {
		filter.Email = &email
	}
	if gender := strings.TrimSpace(r.URL.Query().Get("gender")); gender != "" {
		g := Gender(strings.ToLower(gender))
		filter.Gender = &g
	}
	if birth := strings.TrimSpace(r.URL.Query().Get("birth_date")); birth != "" {
		// Expect format YYYY-MM-DD
		t, err := time.Parse("2006-01-02", birth)
		if err != nil {
			http.Error(w, "invalid birth_date, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		filter.BirthDate = &t
	}

	orderBy := strings.TrimSpace(r.URL.Query().Get("order_by"))
	orderDir := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("order_dir")))
	orderDesc := orderDir == "DESC"

	resp, err := s.repo.GetPaginatedUsers(ctx, page, pageSize, filter, orderBy, orderDesc)
	if err != nil {
		http.Error(w, "failed to get users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(ctx, w, resp)
}

func (s *Server) handleGetCommonFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	id1Ptr, err := parseIntQuery(r, "user1")
	if err != nil || id1Ptr == nil {
		http.Error(w, "invalid or missing user1", http.StatusBadRequest)
		return
	}
	id2Ptr, err := parseIntQuery(r, "user2")
	if err != nil || id2Ptr == nil {
		http.Error(w, "invalid or missing user2", http.StatusBadRequest)
		return
	}

	users, err := s.repo.GetCommonFriends(ctx, *id1Ptr, *id2Ptr)
	if err != nil {
		http.Error(w, "failed to get common friends: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(ctx, w, users)
}

func writeJSON(_ context.Context, w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

