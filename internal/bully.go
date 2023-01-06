package bully

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Bully struct {
	*net.TCPListener

	ID           string
	addr         string
	coordinator  string
	nodeMap      *NodeMap
	mu           *sync.RWMutex
	receiveChan  chan Message
	electionChan chan Message
}

func NewBully(ID, addr, proto string, configNodes map[string]string) (*Bully, error) {
	b := &Bully{
		ID:           ID,
		addr:         addr,
		coordinator:  ID,
		nodeMap:      NewNodeMap(),
		mu:           &sync.RWMutex{},
		electionChan: make(chan Message, 1),
		receiveChan:  make(chan Message),
	}

	if err := b.Listen(proto, addr); err != nil {
		return nil, fmt.Errorf("New Bully err %v", err)
	}
	b.Connect(proto, configNodes)
	return b, nil
}

func (b *Bully) Listen(proto, addr string) error {
	localAddr, err := net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return fmt.Errorf("conver tcp address error %v", err)
	}
	b.TCPListener, err = net.ListenTCP(proto, localAddr)
	if err != nil {
		return fmt.Errorf("listen tcp error %v", err)
	}
	go b.listen()
	return nil
}

func (b *Bully) listen() {
	for {
		conn, err := b.AcceptTCP()
		if err != nil {
			log.Printf("accetp tcp error %v", err)
			continue
		}
		go b.receive(conn)
	}
}

func (b *Bully) receive(rwc io.ReadCloser) {
	var msg Message
	dec := gob.NewDecoder(rwc)

	for {
		if err := dec.Decode(&msg); err == io.EOF || msg.Type == CLOSE {
			_ = rwc.Close()
			if msg.NodeID == b.GetCoordinator() {
				b.nodeMap.Delete(msg.NodeID)
				b.SetCoordinator(b.ID)
				b.Elect()
			}
			break
		} else if msg.Type == OK {
			select {
			case b.electionChan <- msg:
				continue
			case <-time.After(200 * time.Millisecond):
				continue
			}
		} else {
			b.receiveChan <- msg
		}
	}
}

func (b *Bully) SetCoordinator(ID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// 为什么 ID == b.ID 的时候需要设置为coordinator
	if ID > b.coordinator || ID == b.ID {
		b.coordinator = ID
	}
}

func (b *Bully) GetCoordinator() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.coordinator
}

func (b *Bully) Elect() {
	for _, remoteBullyNode := range b.nodeMap.GetAllNode() {
		if remoteBullyNode.ID > b.ID {
			_ = b.Send(remoteBullyNode.ID, remoteBullyNode.addr, ELECTION)
		}
	}

	select {
	// TODO 这两个case看不懂
	case <-b.electionChan:
		return
	case <-time.After(time.Second):
		b.SetCoordinator(b.ID)
		for _, remoteBullyNode := range b.nodeMap.GetAllNode() {
			_ = b.Send(remoteBullyNode.ID, remoteBullyNode.addr, COORDINATOR)
		}
		return
	}
}

func (b *Bully) Send(toID, addr string, what int) error {
	maxRetries := 5

	if !b.nodeMap.Find(toID) {
		_ = b.connect("tcp4", addr, toID)
	}

	for attempts := 0; attempts < maxRetries; attempts++ {
		err := b.nodeMap.Send(toID, &Message{NodeID: b.ID, Addr: b.addr, Type: what})
		if err == nil {
			break
		} else {
			fmt.Println("send msg to %s error %v", toID, err)
		}
		// 为什么这里需要一次连接？
		//_ = b.connect("tcp4", addr, toID)
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (b *Bully) Connect(proto string, configNodes map[string]string) {
	for ID, addr := range configNodes {
		if b.ID == ID {
			continue
		}
		if err := b.connect(proto, addr, ID); err != nil {
			log.Printf("connect %s error %v", addr, err)
			b.nodeMap.Delete(ID)
		}
	}
}

func (b *Bully) connect(proto, addr, ID string) error {
	remoteAddr, err := net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return fmt.Errorf("convert tcp address %s error %v", addr, err)
	}
	sock, err := net.DialTCP(proto, nil, remoteAddr)
	if err != nil {
		return fmt.Errorf("dial tcp with addr %s error %v", addr, err)
	}
	b.nodeMap.Add(ID, addr, sock)
	return nil
}

// Run NOTE: This function is an infinite loop.
func (b *Bully) Run(workFunc func()) {
	go workFunc()

	b.Elect()
	for msg := range b.receiveChan {
		if msg.Type == ELECTION && msg.NodeID < b.ID {
			_ = b.Send(msg.NodeID, msg.Addr, OK)
			b.Elect()
		} else if msg.Type == COORDINATOR {
			b.SetCoordinator(msg.NodeID)
		}
	}
}
