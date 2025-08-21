package alloc

import (
	"context"
	"Mini-Resource-Manager/internal/core"
	"sync"
)

type Allocator struct {
	AllocCh chan core.AllocateRequest
	ReleaseCh chan core.ReleaseRequest
	store *core.Store // access to the shared in-memory store
}

// NewAllocator creates a new Allocator instance with the given store
func NewAllocator(store *core.Store) *Allocator {
	return &Allocator{
		AllocCh:   make(chan core.AllocateRequest, 64),
		ReleaseCh: make(chan core.ReleaseRequest, 64),
		store:     store,
	}
}

func (a *Allocator) Allocate(request core.AllocateRequest) {
	// here we know that the pool exists because we checked it in the handler

	pool, ok := a.store.PoolExists(request.Pool)
	if !ok {
		request.Reply <- core.AllocateResponse{
			Err: core.ErrPoolNotFound,
		}
		return
	}

	// lock the pool
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	// try to allocate item
	for item := pool.Next ; item <= pool.Max ; item++ {
		if !pool.InUse[item] {
			pool.InUse[item] = true
			pool.Next = item + 1
			request.Reply <- core.AllocateResponse{
				Value: item,
				Err: nil,
			}
			return
		}
	}

	// no free items found
	request.Reply <- core.AllocateResponse{
		Err: core.ErrNoFreeItems,
	}
}

func (a *Allocator) Release(request core.ReleaseRequest) {
	// Check if the pool exists
	pool, ok := a.store.PoolExists(request.Pool)
	if !ok {
		request.Reply <- core.ErrPoolNotFound
		return
	}

	// Lock the pool
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	// Check if the value is within the pool range
	if request.Value < pool.Min || request.Value > pool.Max {
		request.Reply <- core.ErrInvalidValue
		return
	}

	// Check if the value is currently allocated
	if !pool.InUse[request.Value] {
		request.Reply <- core.ErrValueNotAllocated
		return
	}

	// Release the value
	pool.InUse[request.Value] = false
	
	// Update Next pointer if this value is before the current Next
	if request.Value < pool.Next {
		pool.Next = request.Value
	}

	request.Reply <- nil
}

func (a *Allocator) Start(ctx context.Context, numWorkers int, wg *sync.WaitGroup) {
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for{
				select{
				case <-ctx.Done():
					// get request to stop so exit the loop
					return
				case req, ok := <-a.AllocCh:
					if !ok {
						return
					}
					a.Allocate(req)
				case req, ok := <-a.ReleaseCh:
					if !ok {
						return
					}
					a.Release(req)
				}
			}
		}()
	}
}