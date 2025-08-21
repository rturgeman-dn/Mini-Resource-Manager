package api

import (
	"encoding/json"
	"net/http"
	"Mini-Resource-Manager/internal/core"
	"context"
	"time"
)

// HandleCreateTemplate handles the incoming HTTP POST request to create a new template
func (h *Handler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template core.Template
	
	// decode JSON from request body to core.Template struct
	err := json.NewDecoder(r.Body).Decode(&template)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate template
	if template.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name is required"})
		return
	}

	if template.Min > template.Max {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "min_greater_than_max"})
		return
	}

	// 409 Conflict with { "error": "template_exists" }
	_, exists := h.store.TemplateExists(template.Name)
	if exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "template_exists"})
		return
	}

	// save template to database (in-memory store)
	h.store.CreateTemplate(template)

	// return 201 Created with the new template in the response body
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// HandleCreatePool handles the incoming HTTP POST request to create a new pool
func (h *Handler) HandleCreatePool(w http.ResponseWriter, r *http.Request) {
	var pool core.Pool

	// decode JSON from request body to core.Pool struct
	err := json.NewDecoder(r.Body).Decode(&pool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if pool.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name is required"})
		return
	}

	template, exists := h.store.TemplateExists(pool.Tmpl)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "template_not_found"})
		return
	}

	// 409 Conflict with { "error": "pool_exists" }
	_, poolExists := h.store.PoolExists(pool.Name)  // Fixed: handle both return values
	if poolExists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "pool_exists"})
		return
	}

	// Pull min/max from template and init the pool
	pool.Min = template.Min
	pool.Max = template.Max
	pool.InUse = make(map[int]bool)
	pool.Next = template.Min

	// Save pool 
	h.store.CreatePool(&pool)  

	// return 201 Created with the new pool in the response body
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pool)
}

func (h *Handler) HandleAllocate(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Pool string `json:"pool"`
	}
	
	// decode JSON from request body to core.AllocateRequest struct
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate request
	if reqBody.Pool == "" {
		http.Error(w, "Pool is required", http.StatusBadRequest)
		return
	}

	// check that pool exists
	_, exists := h.store.PoolExists(reqBody.Pool) 
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "pool_not_found"})
		return
	}

	// check timeout (10 seconds)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	replyCh := make(chan core.AllocateResponse, 1)
	req := core.AllocateRequest{
		Pool: reqBody.Pool,
		Ctx: ctx,
		Reply: replyCh,  
	}

	// Send request to allocator
	select {
	case h.allocator.AllocCh <- req:  // Send to allocator channel
		// Request sent successfully
	case <-ctx.Done():
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(map[string]string{"error": "request_timeout"})
		return
	}

	// Wait for response from allocator
	select {
	case <-ctx.Done():
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(map[string]string{"error": "request_timeout"})
		return
	case reply := <-replyCh:
		if reply.Err == core.ErrNoFreeItems {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "no_free_items"})
			return
		} 

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"value": reply.Value})
		return
	}
}