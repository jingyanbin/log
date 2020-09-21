package log

import (
	"sync"
	"time"
)

var queueNodeSize int32 = 64
var queueNodeFree = sync.Pool{
	New: func() interface{} { return &queueNode{data:make([]interface{}, queueNodeSize), size:queueNodeSize} },
}

type queueNode struct {
	data [] interface{}
	next *queueNode
	pos int32
	end int32
	size int32
}

func (node *queueNode)free() {
	node.pos = 0
	node.end = 0
	node.next = nil
	queueNodeFree.Put(node)
}

func newQueueNode() *queueNode{
	return queueNodeFree.Get().(*queueNode)
}

type queue struct {
	first *queueNode
	last *queueNode
	mutex sync.Mutex
	len int
	closed bool
}

func newQueue() *queue {
	q := &queue{}
	n := newQueueNode()
	q.first = n
	q.last = n
	return q
}

func (q *queue)Closed() bool {
	return q.closed
}

func (q *queue)Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.closed == false{
		q.closed = true
	}
}

func (q *queue)Open() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.closed == true{
		q.closed = false
	}
}

func (q *queue)push(v interface{}) {
	q.last.data[q.last.end] = v
	q.last.end++
	q.len++
	if q.last.end == q.last.size{
		n := newQueueNode()
		q.last.next = n
		q.last = n
	}
}

func (q *queue)Push(v interface{}) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.closed == true{
		return false
	}
	q.push(v)
	return true
}

func (q *queue)PushForce(v interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.push(v)
}

func (q *queue)Pop() (v interface{}, closed bool) {
	q.mutex.Lock()
	closed = q.closed
	for{
		if q.first.pos < q.first.end{
			v = q.first.data[q.first.pos]
			q.first.pos++
			q.len--
			q.mutex.Unlock()
			closed = false
			return
		} else if q.first.pos == q.first.size{
			if q.first.next == nil{
				q.mutex.Unlock()
				return
			}
			first := q.first
			q.first = first.next
			first.free()
			continue
		}
		q.mutex.Unlock()
		return
	}
}

func (q *queue)PopBlock() (v interface{}, closed bool) {
	q.mutex.Lock()
	for {
		if q.first.pos < q.first.end{
			v = q.first.data[q.first.pos]
			q.first.pos++
			q.len--
			q.mutex.Unlock()
			return
		} else if q.first.pos == q.first.size{
			for q.first.next == nil{
				if q.closed{
					q.mutex.Unlock()
					return nil, true
				}
				q.mutex.Unlock()
				//runtime.Gosched()
				time.Sleep(time.Microsecond)
				q.mutex.Lock()
			}
			first := q.first
			q.first = first.next
			first.free()
			continue
		}
		if q.closed{
			q.mutex.Unlock()
			return nil, true
		}
		q.mutex.Unlock()
		//runtime.Gosched()
		time.Sleep(time.Microsecond)
		q.mutex.Lock()
	}
}

func (q *queue)Len() int {
	return q.len
}
