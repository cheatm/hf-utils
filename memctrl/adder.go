package memctrl

import "sync/atomic"

type Adder struct {
	w int64
	r int64
	f int64
	s int64
}

func (a *Adder) Push() {
	w := atomic.AddInt64(&a.w, 1)
	r := atomic.LoadInt64(&a.r)
	if w-r > a.s {
		atomic.AddInt64(&a.f, 1)
	}
}

func (a *Adder) Pop() int64 {
	r := atomic.AddInt64(&a.r, 1) - 1
	w := atomic.LoadInt64(&a.w)
	if r >= w {
		atomic.AddInt64(&a.f, 1)
	}
	return r
}

type Pointer struct {
	w      int64
	r      int64
	cr     int64
	cw     int64
	size   int64
	failed int64
}

func (p *Pointer) Push() {
	w := atomic.AddInt64(&p.w, 1)
	// r := atomic.LoadInt64(&p.cr)
	r := atomic.LoadInt64(&p.r)
	if w-r > p.size {
		atomic.AddInt64(&p.failed, 1)
	} else {
		// atomic.AddInt64(&p.cw, 1)
	}
}

func (p *Pointer) Pop() int64 {
	r := atomic.AddInt64(&p.r, 1) - 1
	w := atomic.LoadInt64(&p.w)
	// w := atomic.LoadInt64(&p.cw)
	if r >= w {
		atomic.AddInt64(&p.failed, 1)
	} else {
		// atomic.AddInt64(&p.cr, 1)
	}
	return r
}
