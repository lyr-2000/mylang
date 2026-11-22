package api

import (
	"fmt"
	"sync"
	"testing"
)

func TestPoolRun(t *testing.T) {
	p, err := NewPool("M:CLOSE*2;")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	if p.Program() == nil {
		t.Fatal("Program() = nil")
	}

	outs, err := p.Run(map[string][]float64{"CLOSE": {1, 2, 3}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(outs) != 1 || outs[0].Name != "M" {
		t.Fatalf("outputs = %+v, want [M]", outs)
	}
	for i, want := range []float64{2, 4, 6} {
		if outs[0].Values[i] != want {
			t.Errorf("M[%d] = %v, want %v", i, outs[0].Values[i], want)
		}
	}
}

func TestPoolRunWithContext(t *testing.T) {
	p, err := NewPool("M:CLOSE*2;")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	ctx := Ctx{Symbol: "sh600519", Ts: []int64{1, 2, 3}}
	outs, err := p.RunWithContext(ctx, map[string][]float64{"CLOSE": {1, 2, 3}})
	if err != nil {
		t.Fatalf("RunWithContext: %v", err)
	}
	if len(outs) != 1 || outs[0].Values[2] != 6 {
		t.Fatalf("outputs = %+v", outs)
	}
}

// TestPoolConcurrent 用 -race 验证：多 goroutine 并发取实例，各自结果不串味。
func TestPoolConcurrent(t *testing.T) {
	p, err := NewPool("M:CLOSE*2;")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}

	const N = 200
	var wg sync.WaitGroup
	errCh := make(chan error, N)

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := float64(i)
			vars := map[string][]float64{"CLOSE": {c, c + 1, c + 2}}
			outs, err := p.Run(vars)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: %v", i, err)
				return
			}
			if len(outs) != 1 || len(outs[0].Values) != 3 {
				errCh <- fmt.Errorf("goroutine %d: unexpected shape %+v", i, outs)
				return
			}
			for j := 0; j < 3; j++ {
				want := 2*c + 2*float64(j)
				if outs[0].Values[j] != want {
					errCh <- fmt.Errorf("goroutine %d[%d] = %v, want %v", i, j, outs[0].Values[j], want)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}
