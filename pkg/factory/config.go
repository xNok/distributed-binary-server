package factory

import (
	"flag"
	"fmt"
)

// ParseNodeConfig parses command line flags and returns a NodeConfig
func ParseNodeConfig() (NodeConfig, error) {
	port := flag.String("port", "", "Server port argument")
	rightPort := flag.String("right", "", "Right child server port string argument")
	leftPort := flag.String("left", "", "Left child server port string argument")

	flag.Parse()

	if *port == "" {
		return NodeConfig{}, fmt.Errorf("port is required")
	}

	config := NodeConfig{
		Port: *port,
	}

	// Only set child ports if they're provided
	if *rightPort != "" {
		config.RightPort = rightPort
	}
	if *leftPort != "" {
		config.LeftPort = leftPort
	}

	return config, nil
}

// NewNodeConfigFromPorts creates a NodeConfig from explicit port values
func NewNodeConfigFromPorts(port string, leftPort, rightPort *string) NodeConfig {
	return NodeConfig{
		Port:      port,
		LeftPort:  leftPort,
		RightPort: rightPort,
	}
}
