package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var SendHeartbeatSeconds time.Duration = 1
var SendClientHeartbeatSeconds time.Duration = 5

var SendMutex *sync.Mutex = &sync.Mutex{}

type PushServer struct {
	UpdateMux      *sync.Mutex
	CntUpdateMux   *sync.Mutex

	HeartCh         chan int
	ClientsHeartCh  chan int

	ItemsCnt        int64
	Items        []*PushItem
	PushedCouter    int64
}

type PushItem struct {
	Id      int64
	Query   string
}

func NewPushServer() *PushServer {
	s := &PushServer{}

	s.UpdateMux      = new(sync.Mutex)
	s.HeartCh        = make(chan int, 1)
	s.ClientsHeartCh = make(chan int, 1)

	return s
}

func (s *PushServer) CreatePushItem() (*PushItem, error) {
	item := &PushItem{}

	s.CntUpdateMux.Lock()
	s.ItemsCnt += 1
	item.Id = s.ItemsCnt
	s.CntUpdateMux.Unlock()

	return item, nil
}

func (s *PushServer) AddPushItem(item *PushItem) error {
	newItems := append(s.Items, item)

	s.UpdateMux.Lock()
	s.Items = newItems
	s.UpdateMux.Unlock()

	return nil
}

func (s *PushServer) RunHeartbeatServer() error {
	for {
		s.HeartCh <- 1
		time.Sleep(SendHeartbeatSeconds * time.Second)
	}

	return nil
}

func (s *PushServer) RunClientHeartbeatServer() error {
	for {
		s.ClientsHeartCh <- 1
		time.Sleep(SendClientHeartbeatSeconds * time.Second)
	}

	return nil
}

func (s *PushServer) RunSendServer() error {
	for {
		_, ok := <-s.HeartCh

		if ok == false {
			fmt.Println("[sender] finished")
			break
		}

		cnt := len(s.Items)
		fmt.Printf("[sender] checking count for send -> %v", cnt)
	}

	return nil
}

func (s *PushServer) RunClientsServer() error {
	for {
		_, ok := <-s.ClientsHeartCh

		if ok == false {
			fmt.Println("[clients server] finished")
			break
		}

		fmt.Println("[clients server] clients pool rework")

		item, err := s.CreatePushItem()
		if err != nil {
			return err
		}

		item.Query = "TestQuery"
		s.AddPushItem(item)
	}
	return nil
}

// RunPushServer run each's N seconds and produce requests to clickhouse,
// then fetched results will go throgh socets to end users channels.
func (s *PushServer) Run() error {
	go s.RunHeartbeatServer()
	go s.RunClientHeartbeatServer()

	go s.RunSendServer()
	go s.RunClientsServer()

	time.Sleep(100 * time.Second)

	err := errors.New("not implemented")
	return err
}
