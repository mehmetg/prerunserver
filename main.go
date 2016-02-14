package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Staring server!")
	router := mux.NewRouter.StrictSlash(true)
	router.HandleFunc("/", Index)
	log.Fatal(http.ListenAndServe(":4444", nil))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %q", html.EscapeString(r.URL.Path))
}
