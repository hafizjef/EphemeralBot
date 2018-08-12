package runners

import (
	"context"
	"sync"
)

// CancelMap used to make map
type CancelMap struct {
	sync.Mutex
	internal map[string]context.CancelFunc
}

// NewCancelMap export
func NewCancelMap() *CancelMap {
	return &CancelMap{
		internal: make(map[string]context.CancelFunc),
	}
}

// Get value based on key
func (c *CancelMap) Get(key string) (value context.CancelFunc, ok bool) {
	c.Lock()
	result, ok := c.internal[key]
	c.Unlock()
	return result, ok
}

// Set value using key
func (c *CancelMap) Set(key string, value context.CancelFunc) {
	c.Lock()
	c.internal[key] = value
	c.Unlock()
}

// Delete from cancelmap using key
func (c *CancelMap) Delete(key string) {
	c.Lock()
	delete(c.internal, key)
	c.Unlock()
}
