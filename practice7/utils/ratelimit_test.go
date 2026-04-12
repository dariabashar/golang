package utils

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddleware_Concurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("RATE_LIMIT_MAX", "5")
	t.Setenv("RATE_LIMIT_WINDOW_SEC", "60")

	r := gin.New()
	r.Use(RateLimitMiddleware())
	r.GET("/x", func(c *gin.Context) { c.Status(200) })

	var wg sync.WaitGroup
	ok := 0
	var mu sync.Mutex
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/x", nil)
			req.RemoteAddr = "192.0.2.1:1234"
			r.ServeHTTP(w, req)
			mu.Lock()
			if w.Code == 200 {
				ok++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	if ok > 5 {
		t.Fatalf("expected at most 5 OK, got %d", ok)
	}
}
