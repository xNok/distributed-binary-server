package btree

import (
	"context"
	"log"
	"sync"
)

// Node represents a node in the binary tree
type Node struct {
	name     string
	inbound  chan Message
	leftOut  chan Message
	rightOut chan Message
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewNode creates a new btree node
func NewNode(name string) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		name:     name,
		inbound:  make(chan Message, 100),
		leftOut:  make(chan Message, 100),
		rightOut: make(chan Message, 100),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins message processing for this node
func (n *Node) Start() {
	go n.messageLoop()
}

// Stop stops the node
func (n *Node) Stop() {
	n.cancel()
}

// GetInboundChannel returns the channel for receiving messages
func (n *Node) GetInboundChannel() chan<- Message {
	return n.inbound
}

// GetLeftChannel returns the channel for left child communication
func (n *Node) GetLeftChannel() <-chan Message {
	return n.leftOut
}

// GetRightChannel returns the channel for right child communication
func (n *Node) GetRightChannel() <-chan Message {
	return n.rightOut
}

// HandleMessage processes an incoming message
func (n *Node) HandleMessage(ctx context.Context, msg Message) error {
	log.Printf("[%s] Received message: %s", n.name, msg.Content)

	// Send to both children (if channels are being listened to)
	select {
	case n.leftOut <- msg:
		log.Printf("[%s] Forwarded to left child", n.name)
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Left channel is full or not being read, continue
	}

	select {
	case n.rightOut <- msg:
		log.Printf("[%s] Forwarded to right child", n.name)
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Right channel is full or not being read, continue
	}

	return nil
}

// SendToLeft sends a message to the left child
func (n *Node) SendToLeft(ctx context.Context, msg Message) error {
	select {
	case n.leftOut <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SendToRight sends a message to the right child
func (n *Node) SendToRight(ctx context.Context, msg Message) error {
	select {
	case n.rightOut <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Receive returns the channel to receive messages
func (n *Node) Receive(ctx context.Context) <-chan Message {
	return n.inbound
}

// messageLoop processes incoming messages
func (n *Node) messageLoop() {
	for {
		select {
		case msg := <-n.inbound:
			if err := n.HandleMessage(n.ctx, msg); err != nil {
				log.Printf("[%s] Error handling message: %v", n.name, err)
			}
		case <-n.ctx.Done():
			log.Printf("[%s] Node stopped", n.name)
			return
		}
	}
}
