package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*CachedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*CachedResponse)}
}

func (m *MemoryStore) Get(key string) (*CachedResponse, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	resp, ok := m.data[key]
	if !ok {
		return nil, false
	}
	copyResp := &CachedResponse{
		StatusCode: resp.StatusCode,
		Body:       append([]byte(nil), resp.Body...),
		Completed:  resp.Completed,
	}
	return copyResp, true
}

func (m *MemoryStore) StartProcessing(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return false
	}
	m.data[key] = &CachedResponse{Completed: false}
	return true
}

func (m *MemoryStore) Finish(key string, status int, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if resp, exists := m.data[key]; exists {
		resp.StatusCode = status
		resp.Body = append([]byte(nil), body...)
		resp.Completed = true
		return
	}
	m.data[key] = &CachedResponse{
		StatusCode: status,
		Body:       append([]byte(nil), body...),
		Completed:  true,
	}
}

type captureWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newCaptureWriter() *captureWriter {
	return &captureWriter{header: make(http.Header), status: http.StatusOK}
}

func (c *captureWriter) Header() http.Header {
	return c.header
}

func (c *captureWriter) Write(b []byte) (int, error) {
	return c.body.Write(b)
}

func (c *captureWriter) WriteHeader(statusCode int) {
	c.status = statusCode
}

func IdempotencyMiddleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		if cached, exists := store.Get(key); exists {
			if cached.Completed {
				w.WriteHeader(cached.StatusCode)
				_, _ = w.Write(cached.Body)
				return
			}
			http.Error(w, "Duplicate request in progress", http.StatusConflict)
			return
		}

		if !store.StartProcessing(key) {
			if cached, exists := store.Get(key); exists && cached.Completed {
				w.WriteHeader(cached.StatusCode)
				_, _ = w.Write(cached.Body)
				return
			}
			http.Error(w, "Duplicate request in progress", http.StatusConflict)
			return
		}

		recorder := newCaptureWriter()
		next.ServeHTTP(recorder, r)

		store.Finish(key, recorder.status, recorder.body.Bytes())

		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.status)
		_, _ = w.Write(recorder.body.Bytes())
	})
}

func paymentHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("Processing started")
	time.Sleep(2 * time.Second)

	response := map[string]any{
		"status":         "paid",
		"amount":         1000,
		"transaction_id": fmt.Sprintf("uuid-%d", time.Now().UnixNano()),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	store := NewMemoryStore()
	handler := IdempotencyMiddleware(store, http.HandlerFunc(paymentHandler))
	server := httptest.NewServer(handler)
	defer server.Close()

	const key = "loan-payment-key-123"
	const workers = 7

	var wg sync.WaitGroup
	wg.Add(workers)

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
			req.Header.Set("Idempotency-Key", key)

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("[req-%d] error: %v", id, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			log.Printf("[req-%d] status=%d body=%s", id, resp.StatusCode, string(body))
		}(i + 1)
	}

	wg.Wait()

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
	req.Header.Set("Idempotency-Key", key)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("final repeated request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[post-complete-repeat] status=%d body=%s", resp.StatusCode, string(body))
}
