package main
//The Application Entry Point

import (
	"fmt"
	"Mini-Resource-Manager/internal/api"
	"Mini-Resource-Manager/internal/core"
	"net/http"
	"sync"
	"Mini-Resource-Manager/internal/alloc"
)

func main() {
	store := core.NewStore() // in-memory store (map)

	allocator := alloc.NewAllocator(store)
	
	//  start worker pool with 3 workers
	var wg sync.WaitGroup
	allocator.Start(3, &wg)

	// Creates a new API handler and injects the shared in-memory store
	// This allows all handler methods (e.g., HandleCreateTemplate, HandleCreatePool) 
	// to access the store via h.store without passing it explicitly to every function.
	// Improves modularity, testability, and keeps the codebase clean and extensible.
	handler := api.NewHandler(store, allocator)

	
	
	// Create mux and register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// start server
	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", mux)

	wg.Wait()
}