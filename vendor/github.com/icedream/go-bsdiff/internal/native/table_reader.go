package native

import (
	"io"
	"sync"
)

type readerTable struct {
	nextIndex int
	table     map[int]io.Reader
	mutex     sync.Mutex
}

func (bt *readerTable) Add(reader io.Reader) (index int) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	if bt.table == nil {
		bt.table = map[int]io.Reader{}
	}

	index = bt.nextIndex
	bt.table[index] = reader

	// TODO - Handle int overflow

	bt.nextIndex++

	return
}

func (bt *readerTable) Get(index int) io.Reader {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	return bt.table[index]
}

func (bt *readerTable) Free(index int) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	delete(bt.table, index)
}
