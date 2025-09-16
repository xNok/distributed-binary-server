package factory

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xnok/btree-server-msg/pkg/btree"
	"github.com/xnok/btree-server-msg/pkg/transport"
	"github.com/xnok/btree-server-msg/pkg/transport/tcp"
)

// BTreeNode represents a complete btree node with transport and wiring
type BTreeNode struct {
	Node            *btree.Node
	Server          *transport.Server
	ChildrenClients []*transport.Client
	ctx             context.Context
	cancel          context.CancelFunc
}

// TransportFactory defines a function that creates transport instances
type TransportFactory func() transport.Transport

// NewBTreeNode creates a fully wired btree node with the specified transport
func NewBTreeNode(config NodeConfig, transportFactory TransportFactory) (*BTreeNode, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create the btree node with the number of children specified in config
	nodeName := fmt.Sprintf("node-%s", config.Port)
	node := btree.NewNode(nodeName, config.GetNumChildren())

	// Create and configure the server with the specified transport
	serverTransport := transportFactory()
	server := transport.NewServer(serverTransport, config.Port)

	btreeNode := &BTreeNode{
		Node:            node,
		Server:          server,
		ChildrenClients: make([]*transport.Client, config.GetNumChildren()),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Create child clients for each configured child port
	for i, childPort := range config.ChildrenPorts {
		if childPort != "" {
			childTransport := transportFactory()
			btreeNode.ChildrenClients[i] = transport.NewClient(childTransport, childPort)
		}
	}

	return btreeNode, nil
}

// NewBTreeNodeWithTCP creates a btree node using TCP transport (convenience function)
func NewBTreeNodeWithTCP(config NodeConfig) (*BTreeNode, error) {
	return NewBTreeNode(config, func() transport.Transport {
		return tcp.NewTCPTransport()
	})
}

// Start begins all components and wires them together
func (bn *BTreeNode) Start() error {
	// Start the btree node
	bn.Node.Start()

	// Start the server
	go func() {
		if err := bn.Server.Start(bn.ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wire inbound messages from server to node
	go bn.wireInbound()

	// Connect to children and wire outbound messages
	for i, client := range bn.ChildrenClients {
		if client != nil {
			go bn.connectToChild(client, fmt.Sprintf("child-%d", i))
			go bn.wireChildOutbound(i)
		}
	}

	return nil
}

// Stop gracefully shuts down the node
func (bn *BTreeNode) Stop() error {
	log.Println("Shutting down btree node...")

	// Cancel context to stop all goroutines
	bn.cancel()

	// Stop node
	bn.Node.Stop()

	// Close all child clients
	for _, client := range bn.ChildrenClients {
		if client != nil {
			client.Close()
		}
	}

	// Close server
	bn.Server.Close()

	return nil
}

// wireInbound connects server inbound messages to node
func (bn *BTreeNode) wireInbound() {
	for {
		select {
		case msg := <-bn.Server.GetInboundChannel():
			select {
			case bn.Node.GetInboundChannel() <- msg:
			case <-bn.ctx.Done():
				return
			}
		case <-bn.ctx.Done():
			return
		}
	}
}

// wireChildOutbound connects node child channel to corresponding client
func (bn *BTreeNode) wireChildOutbound(childIndex int) {
	childChannel, err := bn.Node.GetChildChannel(childIndex)
	if err != nil {
		log.Printf("Error getting child channel %d: %v", childIndex, err)
		return
	}

	client := bn.ChildrenClients[childIndex]
	if client == nil {
		return
	}

	for {
		select {
		case msg := <-childChannel:
			select {
			case client.GetOutboundChannel() <- msg:
			case <-bn.ctx.Done():
				return
			}
		case <-bn.ctx.Done():
			return
		}
	}
}

// connectToChild handles connection with retry logic
func (bn *BTreeNode) connectToChild(client *transport.Client, childName string) {
	for i := 0; i < 10; i++ {
		select {
		case <-bn.ctx.Done():
			return
		default:
		}

		if err := client.Connect(bn.ctx); err != nil {
			log.Printf("Failed to connect to %s (attempt %d): %v", childName, i+1, err)
			select {
			case <-time.After(time.Second):
			case <-bn.ctx.Done():
				return
			}
			continue
		}

		log.Printf("Connected to %s", childName)
		return
	}

	log.Printf("Failed to connect to %s after 10 attempts", childName)
}

// GetLeftClient returns the left child client (index 0) - convenience for binary trees
func (bn *BTreeNode) GetLeftClient() *transport.Client {
	if len(bn.ChildrenClients) > 0 {
		return bn.ChildrenClients[0]
	}
	return nil
}

// GetRightClient returns the right child client (index 1) - convenience for binary trees
func (bn *BTreeNode) GetRightClient() *transport.Client {
	if len(bn.ChildrenClients) > 1 {
		return bn.ChildrenClients[1]
	}
	return nil
}

// GetChildClient returns the client for the specified child index
func (bn *BTreeNode) GetChildClient(index int) *transport.Client {
	if index >= 0 && index < len(bn.ChildrenClients) {
		return bn.ChildrenClients[index]
	}
	return nil
}
