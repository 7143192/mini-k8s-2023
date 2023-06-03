package queue

import "sync"

type Queue struct {
	queue []any
	lock  sync.Mutex
}

func (q *Queue) Enqueue(item any) {
	q.lock.Lock()
	q.queue = append(q.queue, item)
	q.lock.Unlock()
}

func (q *Queue) Dequeue() any {
	q.lock.Lock()
	if len(q.queue) == 0 {
		q.lock.Unlock()
		return nil
	} else {
		obj := q.queue[0]
		q.queue = q.queue[1:]
		q.lock.Unlock()
		return obj
	}
}

func (q *Queue) Size() int {
	q.lock.Lock()
	size := len(q.queue)
	q.lock.Unlock()
	return size
}
