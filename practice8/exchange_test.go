package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/convert" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(RateResponse{Base: "USD", Target: "EUR", Rate: 0.92})
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	rate, err := svc.GetRate("USD", "EUR")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rate != 0.92 {
		t.Fatalf("rate %v want 0.92", rate)
	}
}

func TestGetRate_APIBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	_, err := svc.GetRate("USD", "XXX")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "api error: invalid currency pair" {
		t.Fatalf("got %q", err.Error())
	}
}

func TestGetRate_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestGetRate_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	svc.Client.Timeout = 5 * time.Millisecond
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected network/timeout error")
	}
}

func TestGetRate_Server500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"unexpected":`))
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRate_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error on empty body")
	}
}

func TestGetRate_NonOK_WithoutErrorField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	svc := NewExchangeService(srv.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected unexpected status error")
	}
}
