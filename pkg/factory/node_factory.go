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

// NodeConfig holds the configuration for a btree node
type NodeConfig struct {
	Port      string
	LeftPort  *string
	RightPort *string
}

// BTreeNode represents a complete btree node with transport and wiring
type BTreeNode struct {
	Node        *btree.Node
	Server      *transport.Server
	LeftClient  *transport.Client
	RightClient *transport.Client
	ctx         context.Context
	cancel      context.CancelFunc
}

// TransportFactory defines a function that creates transport instances
type TransportFactory func() transport.Transport

// NewBTreeNode creates a fully wired btree node with the specified transport
func NewBTreeNode(config NodeConfig, transportFactory TransportFactory) (*BTreeNode, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create the btree node
	nodeName := fmt.Sprintf("node-%s", config.Port)
	node := btree.NewNode(nodeName)

	// Create and configure the server with the specified transport
	serverTransport := transportFactory()
	server := transport.NewServer(serverTransport, config.Port)

	btreeNode := &BTreeNode{
		Node:   node,
		Server: server,
		ctx:    ctx,
		cancel: cancel,
	}

	// Create child clients if configured
	if config.LeftPort != nil {
		leftTransport := transportFactory()
		btreeNode.LeftClient = transport.NewClient(leftTransport, *config.LeftPort)
	}

	if config.RightPort != nil {
		rightTransport := transportFactory()
		btreeNode.RightClient = transport.NewClient(rightTransport, *config.RightPort)
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
	if bn.LeftClient != nil {
		go bn.connectToChild(bn.LeftClient, "left")
		go bn.wireLeftOutbound()
	}

	if bn.RightClient != nil {
		go bn.connectToChild(bn.RightClient, "right")
		go bn.wireRightOutbound()
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

	// Close clients
	if bn.LeftClient != nil {
		bn.LeftClient.Close()
	}
	if bn.RightClient != nil {
		bn.RightClient.Close()
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

// wireLeftOutbound connects node left channel to left client
func (bn *BTreeNode) wireLeftOutbound() {
	for {
		select {
		case msg := <-bn.Node.GetLeftChannel():
			if bn.LeftClient != nil {
				select {
				case bn.LeftClient.GetOutboundChannel() <- msg:
				case <-bn.ctx.Done():
					return
				}
			}
		case <-bn.ctx.Done():
			return
		}
	}
}

// wireRightOutbound connects node right channel to right client
func (bn *BTreeNode) wireRightOutbound() {
	for {
		select {
		case msg := <-bn.Node.GetRightChannel():
			if bn.RightClient != nil {
				select {
				case bn.RightClient.GetOutboundChannel() <- msg:
				case <-bn.ctx.Done():
					return
				}
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
			log.Printf("Failed to connect to %s child (attempt %d): %v", childName, i+1, err)
			select {
			case <-time.After(time.Second):
			case <-bn.ctx.Done():
				return
			}
			continue
		}

		log.Printf("Connected to %s child", childName)
		return
	}

	log.Printf("Failed to connect to %s child after 10 attempts", childName)
}
