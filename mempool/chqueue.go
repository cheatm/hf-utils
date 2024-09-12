package mempool

type chQueue struct {
	ch chan int64
}

func (q *chQueue) Init(size int64) {
	q.ch = make(chan int64, size)
}

func (q *chQueue) Push(idx int64) bool {
	select {
	case q.ch <- idx:
		return true
	default:
		return false
	}
}

func (q *chQueue) Pop() (idx int64, ok bool) {
	select {
	case idx = <-q.ch:
		return idx, true
	default:
		return -1, false
	}
}
