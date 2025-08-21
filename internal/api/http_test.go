// File: internal/api/http_test.go
package api

import (
	"Mini-Resource-Manager/internal/alloc"
	"Mini-Resource-Manager/internal/core"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func setupTestServer() *httptest.Server {
	store := core.NewStore()
	allocator := alloc.NewAllocator(store)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	allocator.Start(ctx, 3, &wg)

	h := NewHandler(store, allocator)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	ts := httptest.NewServer(mux)

	ts.Config.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	ts.Config.RegisterOnShutdown(func() {
		cancel()
		wg.Wait()
	})

	return ts
}

func TestCreateTemplateAndPool_AndAllocateRelease(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// A1. Create template
	template := map[string]interface{}{
		"name": "vlan",
		"min":  100,
		"max":  105,
	}
	templateBytes, _ := json.Marshal(template)
	r1, err := http.Post(ts.URL+"/templates", "application/json", bytes.NewReader(templateBytes))
	if err != nil || r1.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create template: %v", err)
	}

	// A2. Create pool
	pool := map[string]interface{}{
		"name":     "vlan-pool",
		"template": "vlan",
	}
	poolBytes, _ := json.Marshal(pool)
	r2, err := http.Post(ts.URL+"/pools", "application/json", bytes.NewReader(poolBytes))
	if err != nil || r2.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create pool: %v", err)
	}

	// B1. Allocate value
	allocReq := map[string]interface{}{
		"pool": "vlan-pool",
	}
	allocBytes, _ := json.Marshal(allocReq)
	r3, err := http.Post(ts.URL+"/allocate", "application/json", bytes.NewReader(allocBytes))
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	if r3.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", r3.StatusCode)
	}

	// C3. Allocate with unknown pool (should fail 404)
	badReq := map[string]interface{}{"pool": "nope"}
	badBytes, _ := json.Marshal(badReq)
	r4, _ := http.Post(ts.URL+"/allocate", "application/json", bytes.NewReader(badBytes))
	if r4.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", r4.StatusCode)
	}
}

func TestTimeout_Allocation(t *testing.T) {
	store := core.NewStore()
	allocator := alloc.NewAllocator(store)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	allocator.Start(ctx, 1, &wg)

	h := NewHandler(store, allocator)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer func() {
		cancel()
		wg.Wait()
		ts.Close()
	}()

	// Create template and pool
	template := map[string]interface{}{"name": "vlan", "min": 1, "max": 1}
	tb, _ := json.Marshal(template)
	http.Post(ts.URL+"/templates", "application/json", bytes.NewReader(tb))
	pool := map[string]interface{}{"name": "p", "template": "vlan"}
	pb, _ := json.Marshal(pool)
	http.Post(ts.URL+"/pools", "application/json", bytes.NewReader(pb))

	// Fill pool (single value)
	http.Post(ts.URL+"/allocate", "application/json", bytes.NewReader([]byte(`{"pool":"p"}`)))

	// This should now block and timeout
	client := &http.Client{Timeout: 2 * time.Second}
	r, err := client.Post(ts.URL+"/allocate", "application/json", bytes.NewReader([]byte(`{"pool":"p"}`)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StatusCode != http.StatusConflict && r.StatusCode != http.StatusRequestTimeout {
		t.Errorf("expected 409 or 408, got %d", r.StatusCode)
	}
}