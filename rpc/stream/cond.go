package stream

import (
	"context"
	"sync"
)

// Cond implements conditional variable with a channel
type Cond struct {
	mu sync.Mutex // guards ch
	ch chan struct{}
}

func NewCond() *Cond {
	return &Cond{ch: make(chan struct{})}
}

// NotifyChan returns the current notification channel. Capture it while still
// holding the data lock that Broadcast serializes against, then release the
// data lock and wait on the returned channel with WaitChan. This closes the
// lost-wakeup window: a Broadcast that races in after the data lock is released
// closes exactly this channel, so the waiter is woken instead of blocking on a
// freshly-created channel it will never see closed.
func (c *Cond) NotifyChan() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ch
}

// WaitChan blocks until ch is closed (signaled) or ctx is canceled. It returns
// true on signal, false on cancellation.
func WaitChan(ctx context.Context, ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	case <-ctx.Done():
		return false
	}
}

// Wait returns true if the condition is signaled, false if the context is
// canceled. Callers that hold a lock which Broadcast also acquires should
// instead capture NotifyChan() before releasing that lock and then WaitChan on
// it, to avoid missing a Broadcast that fires in the release window.
func (c *Cond) Wait(ctx context.Context) bool {
	return WaitChan(ctx, c.NotifyChan())
}

func (c *Cond) Broadcast() {
	c.mu.Lock()
	defer c.mu.Unlock()
	close(c.ch)
	c.ch = make(chan struct{})
}
