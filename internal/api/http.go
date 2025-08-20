package api

import (
	"encoding/json"
	"net/http"
	"Mini-Resource-Manager/internal/core"
)

// POST /templates
// HandleCreateTemplate handles the incoming HTTP POST request to create a new template
func HandleCreateTemplate(w http.ResponseWriter, r *http.Request, store *core.Store) {
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
	if store.TemplateExists(template.Name) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "template_exists"})
		return
	}

	// save template to database (in-memory store)
	store.CreateTemplate(template)

	// return 201 Created with the new template in the response body
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

func HandleCreatePool(w http.ResponseWriter, r *http.Request, store *core.Store) {
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

	exists := store.TemplateExists(pool.Tmpl)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "template_not_found"})
		return
	}

	// 409 Conflict with { "error": "pool_exists" }
	if store.PoolExists(pool.Name) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "pool_exists"})
		return
	}

	// save pool to database (in-memory store)
	store.CreatePool(pool)

	// return 201 Created with the new template in the response body
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pool)
}