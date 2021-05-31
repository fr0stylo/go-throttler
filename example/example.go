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

	//Add throttler as middleware function to your route to unleash beast power
	mux.Handle("/", throttler.Middleware(throttler.NewOpts(100).WithVerbose(false))(Test))

	http.ListenAndServe(":3000", mux)
}
