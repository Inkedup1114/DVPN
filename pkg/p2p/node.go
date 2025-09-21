package p2p

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Node struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	LastSeen  int64  `json:"last_seen"`
	Version   uint32 `json:"version"`
}

type NodeID string

func (n NodeID) String() string {
	return string(n)
}

type Message struct {
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Signature []byte          `json:"signature"`
}

type MessageRouter struct {
	handlers map[string]func(*Message, *Peer)
	mutex    sync.RWMutex
}

func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[string]func(*Message, *Peer)),
	}
}

func (mr *MessageRouter) Route(msg *Message, peer *Peer) {
	mr.mutex.RLock()
	handler, exists := mr.handlers[msg.Type]
	mr.mutex.RUnlock()

	if exists {
		handler(msg, peer)
	}
}

type P2PNetwork struct {
	localNode *Node
	peers     map[NodeID]*Peer
	listener  net.Listener
	router    *MessageRouter
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type Peer struct {
	Node       *Node
	Connection net.Conn
	LastPing   time.Time
	Connected  bool
	mutex      sync.RWMutex
}

func NewP2PNetwork(listenAddr string, privateKey ed25519.PrivateKey) (*P2PNetwork, error) {
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Create local node ID from public key
	nodeID := NodeID(fmt.Sprintf("%x", publicKey[:8]))

	localNode := &Node{
		ID:        string(nodeID),
		PublicKey: fmt.Sprintf("%x", publicKey),
		Version:   1,
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	localNode.Address = listener.Addr().String()

	ctx, cancel := context.WithCancel(context.Background())

	return &P2PNetwork{
		localNode: localNode,
		peers:     make(map[NodeID]*Peer),
		listener:  listener,
		router:    NewMessageRouter(),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (n *P2PNetwork) Start() error {
	go n.acceptLoop()
	go n.maintenanceLoop()
	return nil
}

func (n *P2PNetwork) Stop() error {
	n.cancel()
	return n.listener.Close()
}

func (n *P2PNetwork) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	go n.handleConnection(conn, false)
	return nil
}

func (n *P2PNetwork) acceptLoop() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			select {
			case <-n.ctx.Done():
				return
			default:
				continue
			}
		}

		go n.handleConnection(conn, true)
	}
}

func (n *P2PNetwork) handleConnection(conn net.Conn, incoming bool) {
	defer conn.Close()

	// Perform handshake
	peer, err := n.handshake(conn, incoming)
	if err != nil {
		return
	}

	// Add peer
	n.addPeer(peer)
	defer n.removePeer(NodeID(peer.Node.ID))

	// Handle messages
	n.messageLoop(peer)
}

func (n *P2PNetwork) handshake(conn net.Conn, incoming bool) (*Peer, error) {
	fmt.Printf("DEBUG: Starting handshake with %s (incoming: %v)\n", conn.RemoteAddr(), incoming)

	// Read peer info
	fmt.Printf("DEBUG: Attempting to decode peer info...\n")
	var nodeInfo Node
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&nodeInfo); err != nil {
		fmt.Printf("DEBUG: Failed to decode peer info: %v\n", err)
		return nil, err
	}

	// Send our info
	fmt.Printf("DEBUG: Sending our node info...\n")
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(n.localNode); err != nil {
		fmt.Printf("DEBUG: Failed to encode our info: %v\n", err)
		return nil, err
	}

	peer := &Peer{
		Node:       &nodeInfo,
		Connection: conn,
		Connected:  true,
		LastPing:   time.Now(),
	}

	fmt.Printf("DEBUG: Handshake successful, adding peer %s\n", nodeInfo.ID)

	return peer, nil
}

func (n *P2PNetwork) messageLoop(peer *Peer) {
	decoder := json.NewDecoder(peer.Connection)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			break
		}

		// Update last seen
		peer.mutex.Lock()
		peer.LastPing = time.Now()
		peer.mutex.Unlock()

		// Route message
		n.router.Route(&msg, peer)
	}
}

func (n *P2PNetwork) addPeer(peer *Peer) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.peers[NodeID(peer.Node.ID)] = peer
}

func (n *P2PNetwork) removePeer(nodeID NodeID) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	delete(n.peers, nodeID)
}

func (n *P2PNetwork) maintenanceLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.maintainPeers()
		}
	}
}

func (n *P2PNetwork) maintainPeers() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	now := time.Now()
	for nodeID, peer := range n.peers {
		peer.mutex.RLock()
		lastPing := peer.LastPing
		peer.mutex.RUnlock()

		// Disconnect stale peers
		if now.Sub(lastPing) > 2*time.Minute {
			peer.Connection.Close()
			delete(n.peers, nodeID)
		}
	}
}

// AddBootstrapPeer connects to a bootstrap peer
func (n *P2PNetwork) AddBootstrapPeer(address string) error {
	return n.ConnectToPeer(address)
}

// GetConnectedPeers returns a list of connected peer addresses
func (n *P2PNetwork) GetConnectedPeers() []string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	var peers []string
	for _, peer := range n.peers {
		if peer.Connected {
			peers = append(peers, peer.Connection.RemoteAddr().String())
		}
	}
	return peers
}

// GetPeerCount returns the number of connected peers
func (n *P2PNetwork) GetPeerCount() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	count := 0
	for _, peer := range n.peers {
		if peer.Connected {
			count++
		}
	}
	return count
}
