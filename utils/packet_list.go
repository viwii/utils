package utils

import (
	"sync"
)

type PacketList struct {
	list      []interface{}
	listGuard sync.Mutex
	listCond  *sync.Cond
	isDone    bool
}

func (pl *PacketList) Add(p interface{}) {
	if !pl.isDone {
		pl.listGuard.Lock()
		pl.list = append(pl.list, p)
		pl.listGuard.Unlock()

		pl.listCond.Signal()
	}

}

func (pl *PacketList) Reset() {
	pl.list = pl.list[0:0]
}

func (pl *PacketList) Len() int {
	pl.listGuard.Lock()
	defer pl.listGuard.Unlock()

	return len(pl.list)
}

func (pl *PacketList) BeginPick() []interface{} {

	pl.listGuard.Lock()

	for len(pl.list) == 0 {
		pl.listCond.Wait()
	}

	pl.listGuard.Unlock()

	pl.listGuard.Lock()

	return pl.list
}

func (pl *PacketList) EndPick() {
	pl.Reset()
	pl.listGuard.Unlock()
}

func (pl *PacketList) Close() {
	pl.isDone = true
	pl.listCond.Signal()

}

func NewPacketList() *PacketList {
	p := &PacketList{
		isDone: false,
	}
	p.listCond = sync.NewCond(&p.listGuard)

	return p
}
