package goconc_test

import (
	"go-conc.googlecode.com/svn/trunk/goconc"
	"testing"
)

func TestPipeline(t *testing.T) {
	var p goconc.Pipeline
	mulStage := p.Stage(10)
	addStage := p.Stage(1)
	const N = 100
	x := 0
	for i := 0; i < N; i++ {
		i := i
		mulStage.In <- func() {
			ii := i*i
			addStage.In <- func() {
				x += ii
			}
		}
	}
	p.Wait()
	
	x0 := 0
	for i := 0; i < N; i++ {
		x0 += i*i
	}
	if x0 != x {
		t.Fatalf("expected/got: %d/%d", x0, x)
	}
}
