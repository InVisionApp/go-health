package health

import "sync"

//#################
//Thread Safe Types
//#################

//
// counter

type counter struct {
	num int
	mu  *sync.Mutex
}

//New counter stating at 0
func NewCounter() *counter {
	return &counter{num: 0, mu: &sync.Mutex{}}
}

func (c *counter) inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.num += 1
}

func (c *counter) dec() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.num -= 1
}

func (c *counter) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.num = 0
}

func (c *counter) val() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.num
}

//
//Boolean

type sBool struct {
	v  bool
	mu *sync.Mutex
}

//New false
func NewBool() *sBool {
	return &sBool{v: false, mu: &sync.Mutex{}}
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
