package health

import (
	"sync/atomic"
)

//#################
//Thread Safe Types
//#################

type sBool struct {
	n int32
}

func (b *sBool) set(v bool) {
	if v {
		atomic.SwapInt32(&b.n, 1)
		return
	}
	atomic.SwapInt32(&b.n, 0)
}

func (b *sBool) setFalse() {
	b.set(false)
}

func (b *sBool) setTrue() {
	b.set(true)
}

func (b sBool) val() bool {
	n := atomic.LoadInt32(&b.n)
	if n == 1 {
		return true
	}
	return false
}

func (b sBool) String() string {
	if b.val() {
		return "true"
	}
	return "false"
}
