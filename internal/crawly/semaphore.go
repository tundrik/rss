package crawly

type Semaphore struct {
	sem chan struct{}
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