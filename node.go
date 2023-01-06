package bully

import (
	"encoding/gob"
	"fmt"
	"io"
	"sync"
)

// Node 表示一个单纯的节点，只包含ID、地址和
type Node struct {
	ID   string
	addr string
	sock *gob.Encoder
}

func NewNode(ID, addr string, fd io.Writer) *Node {
	return &Node{
		ID:   ID,
		addr: addr,
		sock: gob.NewEncoder(fd),
	}
}

// NodeMap 表示当前BullyNode保存的集群所有节点的信息
type NodeMap struct {
	mu    *sync.RWMutex
	nodes map[string]*Node
}

func NewNodeMap() *NodeMap {
	return &NodeMap{mu: &sync.RWMutex{}, nodes: make(map[string]*Node)}
}

func (nm *NodeMap) Add(ID, addr string, fd io.Writer) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.nodes[ID] = NewNode(ID, addr, fd)
}

func (nm *NodeMap) Delete(ID string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	delete(nm.nodes, ID)
}

func (nm *NodeMap) Find(ID string) bool {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	_, exist := nm.nodes[ID]
	return exist
}

func (nm *NodeMap) Send(ID string, msg interface{}) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if node, ok := nm.nodes[ID]; !ok {
		return fmt.Errorf("node %s not in NodeMap", ID)
	} else if err := node.sock.Encode(msg); err != nil {
		return fmt.Errorf("send msg to Node %s err [%v]", ID, err)
	}
	return nil
}

func (nm *NodeMap) GetAllNode() []struct {
	ID   string
	addr string
} {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	var AllNodes []struct {
		ID   string
		addr string
	}

	for _, node := range nm.nodes {
		AllNodes = append(AllNodes, struct {
			ID   string
			addr string
		}{
			node.ID,
			node.addr,
		})
	}
	return AllNodes
}
