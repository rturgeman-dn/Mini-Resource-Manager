package core

import "errors"

var (
	ErrPoolNotFound = errors.New("pool not found")
	ErrNoFreeItems  = errors.New("no free items available")
)
