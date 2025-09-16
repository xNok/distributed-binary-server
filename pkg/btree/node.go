package btree

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Node represents a node in a tree structure
type Node struct {
	name        string
	inbound     chan Message
	childrenOut []chan Message
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewNode creates a new tree node with the specified number of children
func NewNode(name string, numChildren int) *Node {
	ctx, cancel := context.WithCancel(context.Background())

	// Create channels for each child
	childrenOut := make([]chan Message, numChildren)
	for i := range childrenOut {
		childrenOut[i] = make(chan Message, 100)
	}

	return &Node{
		name:        name,
		inbound:     make(chan Message, 100),
		childrenOut: childrenOut,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// NewBinaryNode creates a new binary tree node (convenience function)
func NewBinaryNode(name string) *Node {
	return NewNode(name, 2)
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

// GetChildChannel returns the channel for the specified child index
func (n *Node) GetChildChannel(index int) (<-chan Message, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if index < 0 || index >= len(n.childrenOut) {
		return nil, fmt.Errorf("child index %d out of range [0, %d)", index, len(n.childrenOut))
	}

	return n.childrenOut[index], nil
}

// GetLeftChannel returns the channel for left child (index 0) - convenience for binary trees
func (n *Node) GetLeftChannel() <-chan Message {
	if len(n.childrenOut) > 0 {
		return n.childrenOut[0]
	}
	return nil
}

// GetRightChannel returns the channel for right child (index 1) - convenience for binary trees
func (n *Node) GetRightChannel() <-chan Message {
	if len(n.childrenOut) > 1 {
		return n.childrenOut[1]
	}
	return nil
}

// GetNumChildren returns the number of children this node supports
func (n *Node) GetNumChildren() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.childrenOut)
}

// HandleMessage processes an incoming message and broadcasts to all children
func (n *Node) HandleMessage(ctx context.Context, msg Message) error {
	log.Printf("[%s] Received message: %s (ID: %s)", n.name, msg.Content, msg.ID)

	// Update message source for tracking
	msg.Source = n.name

	// Broadcast to all children
	return n.BroadcastToChildren(ctx, msg)
}

// BroadcastToChildren sends a message to all children
func (n *Node) BroadcastToChildren(ctx context.Context, msg Message) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.childrenOut) == 0 {
		log.Printf("[%s] No children to broadcast to (leaf node)", n.name)
		return nil
	}

	successCount := 0
	for i, childOut := range n.childrenOut {
		select {
		case childOut <- msg:
			log.Printf("[%s] Broadcast to child %d successful", n.name, i)
			successCount++
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Child channel is full or not being read, continue
			log.Printf("[%s] Child %d channel full, skipping broadcast", n.name, i)
		}
	}

	log.Printf("[%s] Broadcast complete: %d/%d children reached", n.name, successCount, len(n.childrenOut))
	return nil
}

// SendToChild sends a message to the specified child index
func (n *Node) SendToChild(ctx context.Context, index int, msg Message) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if index < 0 || index >= len(n.childrenOut) {
		return fmt.Errorf("child index %d out of range [0, %d)", index, len(n.childrenOut))
	}

	select {
	case n.childrenOut[index] <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SendToLeft sends a message to the left child (index 0) - convenience for binary trees
func (n *Node) SendToLeft(ctx context.Context, msg Message) error {
	return n.SendToChild(ctx, 0, msg)
}

// SendToRight sends a message to the right child (index 1) - convenience for binary trees
func (n *Node) SendToRight(ctx context.Context, msg Message) error {
	return n.SendToChild(ctx, 1, msg)
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
