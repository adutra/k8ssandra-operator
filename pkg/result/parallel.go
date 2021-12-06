package result

import (
	"context"
	"sync"
)

// A ReconcileGroup is a collection of goroutines working on subtasks that are part of
// the same overall reconciliation task.
//
// A zero ReconcileGroup is valid and does not cancel on error.
type ReconcileGroup struct {
	cancel func()
	wg     sync.WaitGroup
	lock   sync.Mutex
	result ReconcileResult
}

// NewGroupWithContext returns a new ReconcileGroup and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go returns a terminal result (that is,
// result.Completed() returns true) or the first time Wait returns, whichever occurs first.
func NewGroupWithContext(ctx context.Context) (*ReconcileGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &ReconcileGroup{cancel: cancel}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then returns the merged result from them.
func (g *ReconcileGroup) Wait() ReconcileResult {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.result
}

// Go calls the given function in a new goroutine.
//
// The first call to return a terminal result (that is, result.Completed() returns true) cancels the group.
func (g *ReconcileGroup) Go(f func() ReconcileResult) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		result := f()
		g.lock.Lock()
		defer g.lock.Unlock()
		g.result = Merge(g.result, result)
		if g.result.Completed() && g.cancel != nil {
			g.cancel()
			g.cancel = nil
		}
	}()
}
