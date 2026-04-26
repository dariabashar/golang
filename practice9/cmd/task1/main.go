package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type PaymentClient struct {
	Client *http.Client
	Config RetryConfig
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var netErr net.Error
		return errors.As(err, &netErr)
	}
	if resp == nil {
		return false
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	case http.StatusUnauthorized, http.StatusNotFound:
		return false
	default:
		return false
	}
}

func CalculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	base := cfg.BaseDelay << attempt
	if base > cfg.MaxDelay {
		base = cfg.MaxDelay
	}
	// Full jitter: random value in [0, backoff]
	return time.Duration(rand.Int63n(int64(base) + 1))
}

func (pc *PaymentClient) ExecutePayment(ctx context.Context, url string) (*http.Response, []byte, error) {
	var lastErr error

	for attempt := 0; attempt < pc.Config.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return nil, nil, err
		}

		resp, err := pc.Client.Do(req)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				return nil, nil, readErr
			}
			if !IsRetryable(resp, nil) || resp.StatusCode == http.StatusOK {
				return resp, body, nil
			}
			lastErr = fmt.Errorf("retriable status: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt == pc.Config.MaxRetries-1 || !IsRetryable(resp, err) {
			break
		}

		wait := CalculateBackoff(attempt, pc.Config)
		log.Printf("Attempt %d failed: waiting %v before retry...", attempt+1, wait)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, nil, ctx.Err()
		case <-timer.C:
		}
	}

	return nil, nil, fmt.Errorf("payment failed after %d attempts: %w", pc.Config.MaxRetries, lastErr)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		switch {
		case attempts <= 3:
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"gateway overloaded"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}
	}))
	defer server.Close()

	client := &PaymentClient{
		Client: &http.Client{Timeout: 3 * time.Second},
		Config: RetryConfig{
			MaxRetries: 5,
			BaseDelay:  500 * time.Millisecond,
			MaxDelay:   4 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, body, err := client.ExecutePayment(ctx, server.URL)
	if err != nil {
		log.Fatalf("ExecutePayment error: %v", err)
	}

	log.Printf("Attempt %d: Success! status=%d body=%s", attempts, resp.StatusCode, string(body))
}
