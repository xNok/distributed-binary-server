package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xnok/btree-server-msg/pkg/factory"
)

func main() {
	// Parse configuration from command line
	config, err := factory.ParseNodeConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting node on port %s with config: %+v\n", config.Port, config)

	// Create and start the btree node with TCP transport
	node, err := factory.NewBTreeNodeWithTCP(config)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	log.Printf("Node is running and ready to accept connections on port %s", config.Port)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	if err := node.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
