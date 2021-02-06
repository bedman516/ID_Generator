package main

// A PriorityQueue implements a minHeap for IDs
type PriorityQueue []*ID

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// same time, smaller id first
	if pq[i].timestamp == pq[j].timestamp {
		return pq[i].IDNum < pq[j].IDNum
	}
	// minHeap
	return pq[i].timestamp < pq[j].timestamp
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Push - pushes the item into the queue
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*ID)
	*pq = append(*pq, item)
}

// Pop - pops the min(time) item
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}
