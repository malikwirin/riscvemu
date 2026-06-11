package cpu

// instructionQueue holds instructions that have been fetched but not yet
// issued to a reservation station. The queue acts as a small decoupled
// buffer between the front-end (Machine) and the issue stage.
type instructionQueue struct {
	entries []iqEntry
	head    int
	tail    int
	count   int
	cap     int
}

type iqEntry struct {
	word uint32
	pc   uint32
}

func newInstructionQueue(capacity int) *instructionQueue {
	return &instructionQueue{
		entries: make([]iqEntry, capacity),
		cap:     capacity,
	}
}

// Enqueue stores an instruction at the tail of the queue. Returns false
// if the queue is full so the caller can stall the PC.
func (q *instructionQueue) Enqueue(word, pc uint32) bool {
	if q.count == q.cap {
		return false
	}
	q.entries[q.tail] = iqEntry{word: word, pc: pc}
	q.tail = (q.tail + 1) % q.cap
	q.count++
	return true
}

// Dequeue removes and returns the head of the queue. The boolean is false
// when the queue is empty.
func (q *instructionQueue) Dequeue() (uint32, uint32, bool) {
	if q.count == 0 {
		return 0, 0, false
	}
	e := q.entries[q.head]
	q.entries[q.head] = iqEntry{}
	q.head = (q.head + 1) % q.cap
	q.count--
	return e.word, e.pc, true
}

// RequeueHead puts an entry back at the head of the queue so it is the
// next one dequeued. Used when issue stalled (e.g. structural stall) and
// the instruction must be retried on the next cycle.
func (q *instructionQueue) RequeueHead(word, pc uint32) {
	if q.count == 0 {
		q.entries[0] = iqEntry{word: word, pc: pc}
		q.head = 0
		q.tail = 1
		q.count = 1
		return
	}
	q.head = (q.head - 1 + q.cap) % q.cap
	q.entries[q.head] = iqEntry{word: word, pc: pc}
	q.count++
}

func (q *instructionQueue) Len() int { return q.count }
func (q *instructionQueue) Cap() int { return q.cap }
