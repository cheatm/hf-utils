package sharedmemory

import "reflect"

type Level struct {
	Price  int64
	Volume int64
}

type FullDepth struct {
	Price     int64
	Timestamp int64
	Asks      [100]Level
	Bids      [100]Level
}

var fullDepthSize = reflect.TypeOf(FullDepth{}).Size()

func FullDepthSize() int {
	return int(fullDepthSize)
}
