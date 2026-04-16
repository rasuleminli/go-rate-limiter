package ratelimit

import (
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
