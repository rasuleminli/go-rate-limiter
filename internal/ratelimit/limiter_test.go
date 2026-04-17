package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllow_BurstLimit(t *testing.T) {
	testIP := "192.168.1.1"
	memoryStore = make(map[string]*Client)

	for i := range int(MaxTokens) {
		if !allow(testIP) {
			t.Fatalf("Request #%d was incorrectly rejected", i+1)
		}
	}

	if allow(testIP) {
		t.Fatalf("Request #%d should've been rejected, but it was allowed", int(MaxTokens)+1)
	}
}

func TestAllow_Refill(t *testing.T) {
	testIP := "192.168.1.2"
	memoryStore = make(map[string]*Client)

	mockTime := time.Now()
	clock = func() time.Time { return mockTime }
	t.Cleanup(func() { clock = time.Now })

	for range int(MaxTokens) {
		allow(testIP)
	}

	if allow(testIP) {
		t.Fatalf("Request #%d should've been rejected, but it was allowed", int(MaxTokens)+1)
	}

	mockTime = mockTime.Add(time.Second)

	for i := range int(RequestCost) {
		if !allow(testIP) {
			t.Fatalf("Request #%d was incorrectly rejected", i+1)
		}
	}

	if allow(testIP) {
		t.Fatalf("Request after the refill should've been rejected, but it was allowed")
	}
}

func TestAllow_Concurrent(t *testing.T) {
	testIP := "192.168.1.3"
	memoryStore = make(map[string]*Client)

	var wg sync.WaitGroup
	var counter atomic.Int32
	totalRequests := 100

	for range totalRequests {
		wg.Go(func() {
			if allow(testIP) {
				counter.Add(1)
			}
		})
	}

	wg.Wait()

	if int(counter.Load()) != int(MaxTokens) {
		t.Fatalf("Expected %d requests to be made, but instead allowed to make %d", int(MaxTokens), int(counter.Load()))
	}
}

func TestRateLimitMiddleware_Returns429(t *testing.T) {
	testIP := "192.168.1.4"
	memoryStore = make(map[string]*Client)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := RateLimitMiddleware(nextHandler)

	// Exhaust the limit
	for range int(MaxTokens) {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = testIP + ":12345" // need the port because getIP splits it
		recorder := httptest.NewRecorder()

		handlerToTest.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Expected status OK, but got %d", recorder.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = testIP + ":12345"
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429 (Too Many Requests), but got %d", recorder.Code)
	}

}
