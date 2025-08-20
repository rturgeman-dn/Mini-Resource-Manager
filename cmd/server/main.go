package main

import (
	"fmt"
	"Mini-Resource-Manager/internal/api"
	"Mini-Resource-Manager/internal/core"
	"net/http"
)

func main() {
	store := core.NewStore() // in-memory store (map)

	// POST /templates
	http.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
		api.HandleCreateTemplate(w, r, store)
	})

	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}