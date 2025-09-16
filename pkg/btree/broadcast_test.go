package btree

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMessageBroadcasting(t *testing.T) {
	// Create a tree: root -> left, right
	root := NewBinaryNode("root")
	left := NewBinaryNode("left")
	right := NewBinaryNode("right")

	root.Start()
	left.Start()
	right.Start()

	defer root.Stop()
	defer left.Stop()
	defer right.Stop()

	// Track received messages
	var leftReceived, rightReceived []Message
	var leftMu, rightMu sync.Mutex

	// Wire root to children and collect messages
	go func() {
		for msg := range root.GetLeftChannel() {
			leftMu.Lock()
			leftReceived = append(leftReceived, msg)
			leftMu.Unlock()
			// Also forward to left node's inbound for processing
			select {
			case left.GetInboundChannel() <- msg:
			default:
			}
		}
	}()

	go func() {
		for msg := range root.GetRightChannel() {
			rightMu.Lock()
			rightReceived = append(rightReceived, msg)
			rightMu.Unlock()
			// Also forward to right node's inbound for processing
			select {
			case right.GetInboundChannel() <- msg:
			default:
			}
		}
	}()

	// Send test message to root
	testMsg := NewMessage("Broadcast test!", "broadcast-1")
	err := root.HandleMessage(context.Background(), testMsg)
	if err != nil {
		t.Fatalf("Failed to handle message: %v", err)
	}

	// Wait for propagation
	time.Sleep(100 * time.Millisecond)

	// Verify both children received the message
	leftMu.Lock()
	if len(leftReceived) != 1 {
		t.Errorf("Left child should have received 1 message, got %d", len(leftReceived))
	} else if leftReceived[0].Content != testMsg.Content {
		t.Errorf("Left child received wrong message. Expected: %s, Got: %s",
			testMsg.Content, leftReceived[0].Content)
	}
	leftMu.Unlock()

	rightMu.Lock()
	if len(rightReceived) != 1 {
		t.Errorf("Right child should have received 1 message, got %d", len(rightReceived))
	} else if rightReceived[0].Content != testMsg.Content {
		t.Errorf("Right child received wrong message. Expected: %s, Got: %s",
			testMsg.Content, rightReceived[0].Content)
	}
	rightMu.Unlock()
}

func TestBroadcastToChildren(t *testing.T) {
	// Create a node with 3 children (ternary tree)
	parent := NewNode("parent", 3)
	parent.Start()
	defer parent.Stop()

	// Track messages sent to each child
	var child0, child1, child2 []Message
	var mu0, mu1, mu2 sync.Mutex

	// Collect messages from each child channel
	go func() {
		for msg := range parent.childrenOut[0] {
			mu0.Lock()
			child0 = append(child0, msg)
			mu0.Unlock()
		}
	}()

	go func() {
		for msg := range parent.childrenOut[1] {
			mu1.Lock()
			child1 = append(child1, msg)
			mu1.Unlock()
		}
	}()

	go func() {
		for msg := range parent.childrenOut[2] {
			mu2.Lock()
			child2 = append(child2, msg)
			mu2.Unlock()
		}
	}()

	// Broadcast a message
	testMsg := NewMessage("Ternary broadcast!", "ternary-1")
	err := parent.BroadcastToChildren(context.Background(), testMsg)
	if err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Wait for delivery
	time.Sleep(50 * time.Millisecond)

	// Verify all children received the message
	mu0.Lock()
	if len(child0) != 1 || child0[0].Content != testMsg.Content {
		t.Errorf("Child 0 didn't receive correct message")
	}
	mu0.Unlock()

	mu1.Lock()
	if len(child1) != 1 || child1[0].Content != testMsg.Content {
		t.Errorf("Child 1 didn't receive correct message")
	}
	mu1.Unlock()

	mu2.Lock()
	if len(child2) != 1 || child2[0].Content != testMsg.Content {
		t.Errorf("Child 2 didn't receive correct message")
	}
	mu2.Unlock()
}

func TestLeafNodeBroadcast(t *testing.T) {
	// Test broadcasting on a leaf node (no children)
	leaf := NewNode("leaf", 0)
	leaf.Start()
	defer leaf.Stop()

	// Broadcasting should work without errors even with no children
	testMsg := NewMessage("Leaf test", "leaf-1")
	err := leaf.BroadcastToChildren(context.Background(), testMsg)
	if err != nil {
		t.Fatalf("Leaf node broadcast should not fail: %v", err)
	}

	if leaf.GetNumChildren() != 0 {
		t.Errorf("Leaf node should have 0 children, got %d", leaf.GetNumChildren())
	}
}

func TestMessageSourceTracking(t *testing.T) {
	// Test that message source is updated as it flows through the tree
	root := NewBinaryNode("root")
	root.Start()
	defer root.Stop()

	var received Message
	var mu sync.Mutex

	// Capture message from left channel
	go func() {
		msg := <-root.GetLeftChannel()
		mu.Lock()
		received = msg
		mu.Unlock()
	}()

	// Send message to root
	testMsg := NewMessage("Source tracking test", "source-1")
	err := root.HandleMessage(context.Background(), testMsg)
	if err != nil {
		t.Fatalf("Failed to handle message: %v", err)
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Verify source was updated
	mu.Lock()
	if received.Source != "root" {
		t.Errorf("Expected source to be 'root', got '%s'", received.Source)
	}
	if received.Content != testMsg.Content {
		t.Errorf("Message content changed during broadcast")
	}
	mu.Unlock()
}
