package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type cached struct {
	status    int
	body      []byte
	completed bool
}

type store struct {
	mu   sync.Mutex
	data map[string]*cached
}

func newStore() *store { return &store{data: map[string]*cached{}} }

func (s *store) start(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return false
	}
	s.data[key] = &cached{}
	return true
}

func (s *store) get(key string) (*cached, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.data[key]
	if !ok {
		return nil, false
	}
	cp := &cached{status: c.status, body: append([]byte(nil), c.body...), completed: c.completed}
	return cp, true
}

func (s *store) finish(key string, status int, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &cached{status: status, body: append([]byte(nil), body...), completed: true}
}

func idempotencyMiddleware(s *store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key required", http.StatusBadRequest)
			return
		}

		if c, ok := s.get(key); ok {
			if c.completed {
				w.WriteHeader(c.status)
				_, _ = w.Write(c.body)
				return
			}
			http.Error(w, "processing", http.StatusConflict)
			return
		}

		if !s.start(key) {
			http.Error(w, "processing", http.StatusConflict)
			return
		}

		time.Sleep(500 * time.Millisecond)
		respBody, _ := json.Marshal(map[string]any{"status": "ok", "key": key})
		s.finish(key, http.StatusOK, respBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBody)
		_ = next
	})
}

func main() {
	s := newStore()
	server := httptest.NewServer(idempotencyMiddleware(s, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	defer server.Close()

	client := http.Client{Timeout: 2 * time.Second}
	send := func(name string) {
		req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
		req.Header.Set("Idempotency-Key", "demo-key")
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%s error: %v", name, err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Printf("%s status=%d body=%s", name, resp.StatusCode, string(body))
	}

	go send("first")
	time.Sleep(50 * time.Millisecond)
	send("second")
	time.Sleep(700 * time.Millisecond)
	send("third-repeat")
}
