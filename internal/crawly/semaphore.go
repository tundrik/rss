package crawly

type Semaphore struct {
	sem chan struct{}
}

func newSemaphore(n int) Semaphore {
	return Semaphore{
		sem: make(chan struct{}, n),
	}
}

func (s *Semaphore) Acquire() {
	s.sem <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.sem
}

func (s *Semaphore) Len() int {
	return len(s.sem)
}