package graph

import "sync"

type entry struct {
	data any
	next *entry
}

type Stack struct {
	lock *sync.Mutex
	head *entry
	Size int
}

func (stk *Stack) Push(data any) {
	stk.lock.Lock()

	element := new(entry)
	element.data = data
	temp := stk.head
	element.next = temp
	stk.head = element
	stk.Size++

	stk.lock.Unlock()
}

func (stk *Stack) Pop() any {
	if stk.head == nil {
		return nil
	}
	stk.lock.Lock()
	r := stk.head.data
	stk.head = stk.head.next
	stk.Size--

	stk.lock.Unlock()

	return r
}

func NewStack() *Stack {
	stk := new(Stack)
	stk.lock = &sync.Mutex{}
	return stk
}
