package mini

import (
	"testing"

	"github.com/enriquebris/goconcurrentqueue"
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/suite"
)

type QueueSuite struct {
	suite.Suite
	fifo *goconcurrentqueue.FIFO
}

func TestQueue(t *testing.T) {
	suite.Run(t, new(QueueSuite))
}

func (s *QueueSuite) SetupSuite() {
	s.fifo = goconcurrentqueue.NewFIFO()
}

func (s *QueueSuite) TearDownSuite() {
}

type AnyStruct struct {
	Field1 string
	Field2 int
}

func (s *QueueSuite) Test_00_enqueue() {
	s.fifo.Enqueue("str1")
	s.fifo.Enqueue(5)
	s.fifo.Enqueue(AnyStruct{Field1: "001"})

	total := s.fifo.GetLen()
	s.Equal(3, total)

	for i := 0; i < total; i++ {
		item1, err := s.fifo.Dequeue()
		s.Nil(err)
		pp.Println(item1)
	}
}
