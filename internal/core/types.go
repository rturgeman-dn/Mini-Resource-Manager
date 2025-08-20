// Package models provides the core types for the application.
package core

// Template defines a named range of integers (inclusive: [Min, Max])
// for example: { "name": "vlan", "min": 100, "max": 105 }
type Template struct {
	Name      string `json:"name"` //must be unique and non-empty
	Min       int    `json:"min"`
	Max       int    `json:"max"` //min <= max
}

type Pool struct {
	Name   string // name of the pool
	Tmpl   string // name of the template to use
	Min, Max int // min and max values of the pool
	InUse  map[int]bool // allocated flag
	Next   int          // next scan start (inclusive)
  }
  
type AllocateRequest struct {
	Pool  string
	Reply chan AllocateResponse
	// Ctx   context.Context
}
type AllocateResponse struct {
	Value int
	Err   error
}

type ReleaseRequest struct {
	Pool  string
	Value int
	Reply chan error
	// Ctx   context.Context
}
