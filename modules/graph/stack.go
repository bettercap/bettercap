package graph

import "sync"

type element struct {
	data interface{}
	next *element
}

type stack struct {
	lock *sync.Mutex
	head *element
	Size int
}

func (stk *stack) Push(data interface{}) {
	stk.lock.Lock()

	element := new(element)
	element.data = data
	temp := stk.head
	element.next = temp
	stk.head = element
	stk.Size++

	stk.lock.Unlock()
}

func (stk *stack) Pop() interface{} {
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

func NewStack() *stack {
	stk := new(stack)
	stk.lock = &sync.Mutex{}
	return stk
}
