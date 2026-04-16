package main

import (
	"fmt"
	"net/http"

	"go-rate-limiter/internal/ratelimit"
)

func main() {
	limited := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Limited, don't overuse me!")
		},
	)
	http.Handle("/limited", ratelimit.RateLimitMiddleware(limited))

	http.HandleFunc("/unlimited", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Unlimited! Let's go!")
	})

	http.ListenAndServe(":8080", nil)
}
