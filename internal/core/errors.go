package core

import "errors"

var (
	ErrPoolNotFound      = errors.New("pool not found")
	ErrNoFreeItems       = errors.New("no free items available")
	ErrInvalidValue      = errors.New("invalid value for pool")
	ErrValueNotAllocated = errors.New("value is not currently allocated")
)
