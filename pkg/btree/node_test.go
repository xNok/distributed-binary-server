package btree

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNodeMessagePropagation(t *testing.T) {
	// Create a parent node
	parent := NewNode("parent")
	parent.Start()
	defer parent.Stop()

	// Create tracking for received messages
	var leftReceived, rightReceived []Message
	var leftMu, rightMu sync.Mutex

	// Create mock child receivers
	leftReceiver := make(chan Message, 10)
	rightReceiver := make(chan Message, 10)

	// Wire up the channels - parent's output to receivers
	go func() {
		for msg := range parent.GetLeftChannel() {
			leftReceiver <- msg
		}
	}()

	go func() {
		for msg := range parent.GetRightChannel() {
			rightReceiver <- msg
		}
	}()

	// Collect messages from receivers
	go func() {
		for msg := range leftReceiver {
			leftMu.Lock()
			leftReceived = append(leftReceived, msg)
			leftMu.Unlock()
		}
	}()

	go func() {
		for msg := range rightReceiver {
			rightMu.Lock()
			rightReceived = append(rightReceived, msg)
			rightMu.Unlock()
		}
	}()

	// Create test message
	testMsg := Message{
		Content: "Hello, Binary Tree!",
		ID:      "test-1",
	}

	// Send message to parent
	ctx := context.Background()
	err := parent.HandleMessage(ctx, testMsg)
	if err != nil {
		t.Fatalf("Failed to handle message: %v", err)
	}

	// Wait for message propagation
	time.Sleep(50 * time.Millisecond)

	// Verify message reached left child
	leftMu.Lock()
	if len(leftReceived) != 1 {
		t.Errorf("Left child should have received 1 message, got %d", len(leftReceived))
	} else if leftReceived[0].Content != testMsg.Content {
		t.Errorf("Left child received wrong message. Expected: %s, Got: %s",
			testMsg.Content, leftReceived[0].Content)
	}
	leftMu.Unlock()

	// Verify message reached right child
	rightMu.Lock()
	if len(rightReceived) != 1 {
		t.Errorf("Right child should have received 1 message, got %d", len(rightReceived))
	} else if rightReceived[0].Content != testMsg.Content {
		t.Errorf("Right child received wrong message. Expected: %s, Got: %s",
			testMsg.Content, rightReceived[0].Content)
	}
	rightMu.Unlock()
}

func TestNodeWithoutChildren(t *testing.T) {
	// Create a node without children
	node := NewNode("standalone")
	node.Start()
	defer node.Stop()

	// Create test message
	testMsg := Message{
		Content: "Standalone message",
		ID:      "test-2",
	}

	// Send message to node
	ctx := context.Background()
	err := node.HandleMessage(ctx, testMsg)
	if err != nil {
		t.Fatalf("Failed to handle message: %v", err)
	}

	// Node should not block even without children listening to its output channels
	// This test ensures the node handles the case gracefully
}

func TestMultipleMessages(t *testing.T) {
	// Create a parent node
	parent := NewNode("parent")
	parent.Start()
	defer parent.Stop()

	// Create tracking for received messages
	var received []Message
	var mu sync.Mutex

	// Create mock child receiver
	receiver := make(chan Message, 10)

	// Wire up left channel only
	go func() {
		for msg := range parent.GetLeftChannel() {
			receiver <- msg
		}
	}()

	// Collect messages from receiver
	go func() {
		for msg := range receiver {
			mu.Lock()
			received = append(received, msg)
			mu.Unlock()
		}
	}()

	ctx := context.Background()
	messages := []Message{
		{Content: "Message 1", ID: "1"},
		{Content: "Message 2", ID: "2"},
		{Content: "Message 3", ID: "3"},
	}

	// Send multiple messages
	for _, msg := range messages {
		err := parent.HandleMessage(ctx, msg)
		if err != nil {
			t.Fatalf("Failed to handle message %s: %v", msg.ID, err)
		}
	}

	// Wait for message propagation
	time.Sleep(100 * time.Millisecond)

	// Verify all messages are received
	mu.Lock()
	if len(received) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(received))
	}

	for i, expectedMsg := range messages {
		if i < len(received) {
			if received[i].Content != expectedMsg.Content {
				t.Errorf("Message %d: Expected %s, Got %s",
					i+1, expectedMsg.Content, received[i].Content)
			}
		}
	}
	mu.Unlock()
}

func TestChannelBasedNodeIntegration(t *testing.T) {
	// Create nodes for a simple tree: parent -> left, right
	parent := NewNode("parent")
	left := NewNode("left")
	right := NewNode("right")

	parent.Start()
	left.Start()
	right.Start()

	defer parent.Stop()
	defer left.Stop()
	defer right.Stop()

	// Connect parent to children via channels
	go func() {
		for msg := range parent.GetLeftChannel() {
			select {
			case left.GetInboundChannel() <- msg:
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Timeout sending to left child")
			}
		}
	}()

	go func() {
		for msg := range parent.GetRightChannel() {
			select {
			case right.GetInboundChannel() <- msg:
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Timeout sending to right child")
			}
		}
	}()

	// Send message directly to parent's inbound channel
	testMsg := Message{Content: "Integration test", ID: "int-1"}

	select {
	case parent.GetInboundChannel() <- testMsg:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout sending message to parent")
	}

	// Wait for propagation
	time.Sleep(100 * time.Millisecond)

	// This test demonstrates that we can wire up nodes using only channels
	// without any TCP connections, making testing much easier
}
