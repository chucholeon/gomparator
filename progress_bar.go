package main

import (
	"sync"

	"gopkg.in/cheggaaa/pb.v1"
)

type ProgressBar struct {
	okPb      *pb.ProgressBar
	errorPb   *pb.ProgressBar
	status2XX *pb.ProgressBar
	status4XX *pb.ProgressBar
	status401 *pb.ProgressBar
	pool      *pb.Pool
}

func NewProgressBar(total int) *ProgressBar {
	okPb := makeProgressBar(total, "ok")
	errorPb := makeProgressBar(total, "error")
	status2XX := makeProgressBar(total, "status2XX")
	status4XX := makeProgressBar(total, "status4XX")
	status401 := makeProgressBar(total, "status401")

	return &ProgressBar{
		okPb,
		errorPb,
		status2XX,
		status4XX,
		status401,
		nil,
	}
}

func (p *ProgressBar) IncrementOk() {
	p.okPb.Add(1)
}

func (p *ProgressBar) IncrementError() {
	p.errorPb.Add(1)
}

func (p *ProgressBar) IncrementStatus2XX() {
	p.status2XX.Add(1)
}

func (p *ProgressBar) IncrementStatus4XX() {
	p.status4XX.Add(1)
}

func (p *ProgressBar) IncrementStatus401() {
	p.status401.Add(1)
}

func (p *ProgressBar) Start() {
	pool, err := pb.StartPool(p.okPb, p.errorPb, p.status2XX, p.status401, p.status4XX)
	if err != nil {
		panic(err)
	}
	p.pool = pool
	p.okPb.Start()
}

func (p *ProgressBar) Stop() {
	wg := new(sync.WaitGroup)
	for _, bar := range []*pb.ProgressBar{p.okPb, p.errorPb, p.status2XX, p.status401, p.status4XX} {
		wg.Add(1)
		go func(cb *pb.ProgressBar) {
			cb.Finish()
			wg.Done()
		}(bar)
	}
	wg.Wait()
	// close pool
	_ = p.pool.Stop()
}

func makeProgressBar(total int, prefix string) *pb.ProgressBar {
	bar := pb.New(total)
	bar.Prefix(prefix)
	bar.SetMaxWidth(120)
	bar.ShowElapsedTime = true
	bar.ShowTimeLeft = false
	return bar
}
