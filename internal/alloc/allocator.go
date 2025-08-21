package alloc

import (
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

func (a *Allocator) Start(numWorkers int, wg *sync.WaitGroup) {
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for request := range a.AllocCh {
				a.Allocate(request)
			}
		}()
	}
}