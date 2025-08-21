package main
//The Application Entry Point

import (
	"context"
	"fmt"
	"Mini-Resource-Manager/internal/api"
	"Mini-Resource-Manager/internal/core"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"Mini-Resource-Manager/internal/alloc"
)

func main() {
	store := core.NewStore() // in-memory store (map)

	allocator := alloc.NewAllocator(store)
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker pool with 3 workers
	var wg sync.WaitGroup
	allocator.Start(ctx, 3, &wg)

	// Creates a new API handler and injects the shared in-memory store
	// This allows all handler methods (e.g., HandleCreateTemplate, HandleCreatePool) 
	// to access the store via h.store without passing it explicitly to every function.
	// Improves modularity, testability, and keeps the codebase clean and extensible.
	handler := api.NewHandler(store, allocator)

	// Create mux and register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Listening on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Cancel context to stop workers
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
	}

	// Wait for all workers to finish
	wg.Wait()

	// Print metrics
	fmt.Println("--------------------------------")
	fmt.Println("Metrics:")
	allocations, releases, timeouts := store.Metrics.GetMetrics()
	fmt.Printf("Allocations: %d, Releases: %d, Timeouts: %d\n", allocations, releases, timeouts)
	fmt.Println("--------------------------------")
	
	fmt.Println("Server stopped gracefully")
}