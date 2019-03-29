package native

import (
	"io"
	"sync"
)

type writerTable struct {
	nextIndex int
	table     map[int]io.Writer
	mutex     sync.Mutex
}

func (bt *writerTable) Add(writer io.Writer) (index int) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	if bt.table == nil {
		bt.table = map[int]io.Writer{}
	}

	index = bt.nextIndex
	bt.table[index] = writer

	// TODO - Handle int overflow

	bt.nextIndex++

	return
}

func (bt *writerTable) Get(index int) io.Writer {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	return bt.table[index]
}

func (bt *writerTable) Free(index int) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	delete(bt.table, index)
}
