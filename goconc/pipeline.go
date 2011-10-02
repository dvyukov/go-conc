package goconc

import (
	"sync"
)

type Pipeline struct {
	stages []*Stage
}

type Stage struct {
	q chan func()
	w sync.WaitGroup
}

func (p *Pipeline) Stage(conc int) *Stage {
	var s Stage
	s.q = make(chan func(), conc*10)
	s.w.Add(conc)
	for i := 0; i < conc; i++ {
		go func() {
			for g := range s.q {
				g()
			}
			s.w.Done()
		}()
	}
	p.stages = append(p.stages, &s)
	return &s
}

func (p *Pipeline) Wait() {
	for _, s := range p.stages {
		close(s.q)
		s.w.Wait()
	}
}

func (s *Stage) Go(g func()) {
	s.q <- g
}
