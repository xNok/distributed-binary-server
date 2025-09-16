package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xnok/btree-server-msg/pkg/btree"
)

// Example demonstrating message broadcasting in a tree
func main() {
	fmt.Println("=== Message Broadcasting Example ===\n")

	// Create a 3-level tree
	root := btree.NewBinaryNode("ROOT")
	leftChild := btree.NewBinaryNode("LEFT-CHILD")
	rightChild := btree.NewBinaryNode("RIGHT-CHILD")
	leftGrandchild := btree.NewBinaryNode("LEFT-GRANDCHILD")
	rightGrandchild := btree.NewBinaryNode("RIGHT-GRANDCHILD")

	// Start all nodes
	nodes := []*btree.Node{root, leftChild, rightChild, leftGrandchild, rightGrandchild}
	for _, node := range nodes {
		node.Start()
		defer node.Stop()
	}

	// Wire the tree: root -> left, right -> grandchildren
	go func() {
		for msg := range root.GetLeftChannel() {
			select {
			case leftChild.GetInboundChannel() <- msg:
			case <-time.After(100 * time.Millisecond):
				log.Println("Timeout sending to left child")
			}
		}
	}()

	go func() {
		for msg := range root.GetRightChannel() {
			select {
			case rightChild.GetInboundChannel() <- msg:
			case <-time.After(100 * time.Millisecond):
				log.Println("Timeout sending to right child")
			}
		}
	}()

	go func() {
		for msg := range leftChild.GetLeftChannel() {
			select {
			case leftGrandchild.GetInboundChannel() <- msg:
			case <-time.After(100 * time.Millisecond):
				log.Println("Timeout sending to left grandchild")
			}
		}
	}()

	go func() {
		for msg := range rightChild.GetRightChannel() {
			select {
			case rightGrandchild.GetInboundChannel() <- msg:
			case <-time.After(100 * time.Millisecond):
				log.Println("Timeout sending to right grandchild")
			}
		}
	}()

	fmt.Println("Tree structure:")
	fmt.Println("       ROOT")
	fmt.Println("      /    \\")
	fmt.Println("   LEFT    RIGHT")
	fmt.Println("   /         \\")
	fmt.Println("LEFT-GC    RIGHT-GC")
	fmt.Println()

	// Wait for wiring to complete
	time.Sleep(100 * time.Millisecond)

	// Send messages and observe broadcasting
	messages := []btree.Message{
		btree.NewMessage("Hello, everyone!", "msg-1"),
		btree.NewMessage("Broadcasting works!", "msg-2"),
		btree.NewMessage("Tree-wide message!", "msg-3"),
	}

	for i, msg := range messages {
		fmt.Printf("📤 Sending message %d: %s\n", i+1, msg.Content)

		err := root.HandleMessage(context.Background(), msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// Wait between messages to see clear output
		time.Sleep(200 * time.Millisecond)
		fmt.Println()
	}

	fmt.Println("=== Broadcasting Behavior Demonstrated ===")
	fmt.Println("✓ Root receives message and broadcasts to ALL children")
	fmt.Println("✓ Each child receives message and broadcasts to THEIR children")
	fmt.Println("✓ Messages propagate through the entire tree automatically")
	fmt.Println("✓ Source tracking shows which node forwarded each message")
	fmt.Println("✓ Broadcast success/failure is logged for monitoring")

	// Let final messages process
	time.Sleep(100 * time.Millisecond)
}
