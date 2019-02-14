package godo

import "sync"

// WaitGroupN is a custom wait group that tracks the number added
// so it can be stopped.
type WaitGroupN struct {
	sync.WaitGroup
	sync.Mutex
	N int
}

// Add adds to counter.
func (wg *WaitGroupN) Add(n int) {
	wg.Lock()
	wg.N += n
	wg.Unlock()
	wg.WaitGroup.Add(n)
}

// Done removes from counter.
func (wg *WaitGroupN) Done() {
	wg.Lock()
	wg.N--
	wg.Unlock()

	wg.WaitGroup.Done()
}

// Stop calls done on remaining counter.
func (wg *WaitGroupN) Stop() {
	wg.Lock()
	for i := 0; i < wg.N; i++ {
		wg.WaitGroup.Done()
	}
	wg.N = 0
	wg.Unlock()
}
