package factory

import (
	"testing"
	"time"

	"github.com/xnok/btree-server-msg/pkg/transport"
	"github.com/xnok/btree-server-msg/pkg/transport/tcp"
)

func TestNewBTreeNode(t *testing.T) {
	// Test creating a node without children
	config := NewNodeConfigFromPorts("8080", nil, nil)
	
	// Use TCP transport factory
	node, err := NewBTreeNodeWithTCP(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	if node == nil {
		t.Fatal("Node should not be nil")
	}

	if node.Node == nil {
		t.Fatal("BTree node should not be nil")
	}

	if node.Server == nil {
		t.Fatal("Server should not be nil")
	}

	if node.LeftClient != nil {
		t.Error("LeftClient should be nil when no left port configured")
	}

	if node.RightClient != nil {
		t.Error("RightClient should be nil when no right port configured")
	}
}

func TestNewBTreeNodeWithChildren(t *testing.T) {
	// Test creating a node with children
	leftPort := "8081"
	rightPort := "8082"
	config := NewNodeConfigFromPorts("8080", &leftPort, &rightPort)
	
	node, err := NewBTreeNodeWithTCP(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	if node.LeftClient == nil {
		t.Error("LeftClient should not be nil when left port configured")
	}

	if node.RightClient == nil {
		t.Error("RightClient should not be nil when right port configured")
	}
}

func TestBTreeNodeLifecycle(t *testing.T) {
	// Test full lifecycle without actual network connections
	config := NewNodeConfigFromPorts("8080", nil, nil)
	
	node, err := NewBTreeNodeWithTCP(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	// Start the node
	err = node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop the node
	err = node.Stop()
	if err != nil {
		t.Fatalf("Failed to stop node: %v", err)
	}
}

func TestNewNodeConfigFromPorts(t *testing.T) {
	tests := []struct {
		name      string
		port      string
		leftPort  *string
		rightPort *string
	}{
		{
			name:      "node without children",
			port:      "3030",
			leftPort:  nil,
			rightPort: nil,
		},
		{
			name:      "node with left child only",
			port:      "3030",
			leftPort:  stringPtr("3031"),
			rightPort: nil,
		},
		{
			name:      "node with right child only",
			port:      "3030",
			leftPort:  nil,
			rightPort: stringPtr("3032"),
		},
		{
			name:      "node with both children",
			port:      "3030",
			leftPort:  stringPtr("3031"),
			rightPort: stringPtr("3032"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewNodeConfigFromPorts(tt.port, tt.leftPort, tt.rightPort)

			if config.Port != tt.port {
				t.Errorf("Expected port %s, got %s", tt.port, config.Port)
			}

			if (config.LeftPort == nil) != (tt.leftPort == nil) {
				t.Errorf("LeftPort mismatch: expected %v, got %v", tt.leftPort, config.LeftPort)
			}

			if (config.RightPort == nil) != (tt.rightPort == nil) {
				t.Errorf("RightPort mismatch: expected %v, got %v", tt.rightPort, config.RightPort)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// TestNewBTreeNodeWithCustomTransport tests using a custom transport factory
func TestNewBTreeNodeWithCustomTransport(t *testing.T) {
	config := NewNodeConfigFromPorts("8080", nil, nil)
	
	// Use custom transport factory
	customTransportFactory := func() transport.Transport {
		return tcp.NewTCPTransport() // In real scenario, this could be WebSocket, gRPC, etc.
	}
	
	node, err := NewBTreeNode(config, customTransportFactory)
	if err != nil {
		t.Fatalf("Failed to create node with custom transport: %v", err)
	}
	
	if node == nil {
		t.Fatal("Node should not be nil")
	}
	
	// Test that we can start and stop it
	err = node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	
	time.Sleep(50 * time.Millisecond)
	
	err = node.Stop()
	if err != nil {
		t.Fatalf("Failed to stop node: %v", err)
	}
}
