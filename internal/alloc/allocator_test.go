// Unit tests

package alloc

import (
	"Mini-Resource-Manager/internal/core"
	"context"
	"testing"
)

func TestAllocate_Success(t *testing.T) {
	store := core.NewStore()
	store.CreateTemplate(core.Template{Name: "vlan", Min: 100, Max: 101})
	store.CreatePool(&core.Pool{
		Name:  "vlan-pool",
		Tmpl:  "vlan",
		Min:   100,
		Max:   101,
		InUse: make(map[int]bool),
		Next:  100,
	})

	alloc := NewAllocator(store)
	reply := make(chan core.AllocateResponse, 1)
	alloc.Allocate(core.AllocateRequest{
		Pool:  "vlan-pool",
		Reply: reply,
		Ctx:   context.Background(),
	})

	res := <-reply
	if res.Err != nil || res.Value != 100 {
		t.Errorf("expected 100, got %v, err %v", res.Value, res.Err)
	}
}

func TestAllocate_NoFreeItems(t *testing.T) {
	store := core.NewStore()
	store.CreateTemplate(core.Template{Name: "vlan", Min: 100, Max: 100})
	store.CreatePool(&core.Pool{
		Name:  "vlan-pool",
		Tmpl:  "vlan",
		Min:   100,
		Max:   100,
		InUse: map[int]bool{100: true},
		Next:  100,
	})

	alloc := NewAllocator(store)
	reply := make(chan core.AllocateResponse, 1)
	alloc.Allocate(core.AllocateRequest{
		Pool:  "vlan-pool",
		Reply: reply,
		Ctx:   context.Background(),
	})

	res := <-reply
	if res.Err != core.ErrNoFreeItems {
		t.Errorf("expected ErrNoFreeItems, got %v", res.Err)
	}
}
