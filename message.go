package bully

const (
	ELECTION = iota
	OK
	COORDINATOR
	CLOSE
)

type Message struct {
	NodeID string
	Addr   string
	Type   int
}
