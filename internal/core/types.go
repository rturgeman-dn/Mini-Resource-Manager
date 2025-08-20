// Package models provides the core types for the application.
package core

// Template defines a named range of integers (inclusive: [Min, Max])
// for example: { "name": "vlan", "min": 100, "max": 105 }
type Template struct {
	Name      string `json:"name"` //must be unique and non-empty
	Min       int    `json:"min"`
	Max       int    `json:"max"` //min <= max
}
