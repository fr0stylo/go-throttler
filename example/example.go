package main

import (
	"net/http"
	"throttler"
)

func Test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
func main() {
	mux := http.NewServeMux()

	mux.Handle("/", throttler.Middleware(&throttler.ThrottlerOpts{
		ExhaustionCount: 100,
	})(Test))

	http.ListenAndServe(":3000", mux)
}
