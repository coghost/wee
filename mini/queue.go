package mini

type IQueue interface {
	Enqueue(interface{}) error
	Dequeue() (interface{}, error)
	GetLen() int
	DequeueOrWaitForNextElement() (interface{}, error)
}
