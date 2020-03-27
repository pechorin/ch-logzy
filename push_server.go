package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// time between fetching requests from clickhouse
	// for client
	FetchIntervals = [...]int16{0, 5, 10, 15, 30, 60, 120, 240}

	// if null interval provided
	NullFetchInterval int16 = int16(5)

	SendHeartbeatSeconds time.Duration       = 1
	SendClientHeartbeatSeconds time.Duration = 5

	SendMutex *sync.Mutex = &sync.Mutex{}
)

// ClientStream represents connected with websocket client
// this struct hold all connected information, like current query, interval
// last ping time, configs etc.
type ClientSession struct {
	Id								int64
	Query							string
	Active						bool
	FetchInterval     int16

	CreatedAt         time.Time
	LastKeepaliveAt   time.Time
}

// TODO: if N empty loops reached then sleep for
//       N * (iterations * ratio)
func (cs *ClientSession) Start() (clientCh chan string, err error) {
	timer := time.Tick((time.Duration)(NullFetchInterval) * time.Second)

	limit := 10
	cur   := 0

	go func() {
		for {
			if cur == limit {
				fmt.Println("limit reached")
				break
			}

			select {
			case <-timer:
				cur += 1
				fmt.Println("ClientSession Start tick")
			}
		}
	}()

	return
}

type PushServer struct {
	UpdateMux      *sync.Mutex
	CntUpdateMux   *sync.Mutex

	HeartCh         chan int
	ClientsHeartCh  chan int

	Clients     []*ClientSession
	ClientsCnt     int64
}

func NewPushServer() *PushServer {
	s := &PushServer{}

	s.UpdateMux      = new(sync.Mutex)
	s.CntUpdateMux   = new(sync.Mutex)
	s.HeartCh        = make(chan int, 1)
	s.ClientsHeartCh = make(chan int, 1)

	return s
}

func (s *PushServer) CreateClient() (item *ClientSession, err error) {
	// increase global id
	s.CntUpdateMux.Lock()
	s.ClientsCnt += 1
	item.Id = s.ClientsCnt
	s.CntUpdateMux.Unlock()

	// other defaults
	item.Active    = false
	item.CreatedAt = time.Now()

	return
}

func (s *PushServer) ConnectClient(client *ClientSession) error {
	client.Active = true
	newColl := append(s.Clients, client)

	s.UpdateMux.Lock()
	s.Clients = newColl
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

		cnt := len(s.Clients)
		fmt.Printf("[sender] clients count for send -> %v\n", cnt)
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

		item, err := s.CreateClient()
		if err != nil {
			return err
		}

		item.Query = "TestQuery"
		s.ConnectClient(item)
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
