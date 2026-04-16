package main

import (
	"fmt"
	"net/http"

	"go-rate-limiter/internal/ratelimit"
)

func main() {
	mux := http.NewServeMux()

	limitedHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Limited, don't overuse me!")
		},
	)

	mux.Handle("/limited", ratelimit.RateLimitMiddleware(limitedHandler))

	mux.HandleFunc("/unlimited", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Unlimited! Let's go!")
	})

	http.ListenAndServe(":8080", mux)
}
