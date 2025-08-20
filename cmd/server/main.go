package main
//The Application Entry Point


import (
	"fmt"
	"Mini-Resource-Manager/internal/api"
	"Mini-Resource-Manager/internal/core"
	"net/http"
)

func main() {
	store := core.NewStore() // in-memory store (map)
	
	// Create handler with store dependency
	handler := api.NewHandler(store)
	
	// Create mux and register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", mux)
}