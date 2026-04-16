package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Tokens   float64
	LastSeen time.Time
}

const MaxTokens = 5.0
const RefillRate = 1.0  // refill 1 token per second
const RequestCost = 1.0 // each API hit costs 1 token

var memoryStore = map[string]*Client{}
var mutex sync.Mutex

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		allowed := allow(ip)

		if !allowed {
			http.Error(w, "[429] Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allow(ip string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	now := time.Now()

	client, ok := memoryStore[ip]
	if !ok {
		client = &Client{
			Tokens:   MaxTokens,
			LastSeen: now,
		}
		memoryStore[ip] = client
	}

	deltaSec := now.Sub(client.LastSeen).Seconds()
	client.Tokens += deltaSec * RefillRate

	if client.Tokens > MaxTokens {
		client.Tokens = MaxTokens
	}

	client.LastSeen = now

	if client.Tokens >= RequestCost {
		client.Tokens -= RequestCost
		return true
	}

	return false
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
