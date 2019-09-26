package health

import "sync"

//#################
//Thread Safe Types
//#################

type sBool struct {
	v  bool
	mu sync.Mutex
}

func newBool() *sBool {
	return &sBool{v: false}
}

func (b *sBool) setFalse() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.v = false
}

func (b *sBool) setTrue() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.v = true
}

func (b *sBool) val() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.v
}

func (b *sBool) String() string {
	if b.val() {
		return "true"
	}

	return "false"
}
