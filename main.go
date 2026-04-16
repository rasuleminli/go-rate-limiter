package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/limited", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Limited, don't over use me!")
	})

	http.HandleFunc("/unlimited", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Unlimited! Let's go!")
	})

	http.ListenAndServe(":8080", nil)
}
