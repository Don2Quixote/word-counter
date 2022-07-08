package sema

// Sema is a semaphore.
type Sema struct {
	ch chan struct{}
}

// New returns new Sema with specified slots count.
func New(slots int) *Sema {
	return &Sema{
		ch: make(chan struct{}, slots),
	}
}

// Acquire locks a slot in semaphore.
func (s *Sema) Acquire() {
	s.ch <- struct{}{}
}

// Release releases a slot in semaphore.
func (s *Sema) Release() {
	<-s.ch
}
