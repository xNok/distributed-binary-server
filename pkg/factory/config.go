package factory

import (
	"flag"
	"fmt"
)

// NodeConfig holds the configuration for a tree node
type NodeConfig struct {
	Port          string
	ChildrenPorts []string // Indexed children ports (0=left, 1=right for binary trees)
}

// ParseNodeConfig parses command line flags and returns a NodeConfig for binary tree
func ParseNodeConfig() (NodeConfig, error) {
	port := flag.String("port", "", "Server port argument")
	rightPort := flag.String("right", "", "Right child server port string argument")
	leftPort := flag.String("left", "", "Left child server port string argument")

	flag.Parse()

	if *port == "" {
		return NodeConfig{}, fmt.Errorf("port is required")
	}

	config := NodeConfig{
		Port:          *port,
		ChildrenPorts: make([]string, 2), // Binary tree has 2 children
	}

	// Set child ports if provided (index 0 = left, index 1 = right)
	if *leftPort != "" {
		config.ChildrenPorts[0] = *leftPort
	}
	if *rightPort != "" {
		config.ChildrenPorts[1] = *rightPort
	}

	return config, nil
}

// NewNodeConfigFromPorts creates a NodeConfig from explicit port values (binary tree)
func NewNodeConfigFromPorts(port string, leftPort, rightPort *string) NodeConfig {
	config := NodeConfig{
		Port:          port,
		ChildrenPorts: make([]string, 2),
	}

	if leftPort != nil {
		config.ChildrenPorts[0] = *leftPort
	}
	if rightPort != nil {
		config.ChildrenPorts[1] = *rightPort
	}

	return config
}

// NewNodeConfigWithChildren creates a NodeConfig with an arbitrary number of children
func NewNodeConfigWithChildren(port string, childrenPorts []string) NodeConfig {
	return NodeConfig{
		Port:          port,
		ChildrenPorts: childrenPorts,
	}
}

// GetLeftPort returns the left child port (index 0) for binary trees
func (c *NodeConfig) GetLeftPort() string {
	if len(c.ChildrenPorts) > 0 {
		return c.ChildrenPorts[0]
	}
	return ""
}

// GetRightPort returns the right child port (index 1) for binary trees
func (c *NodeConfig) GetRightPort() string {
	if len(c.ChildrenPorts) > 1 {
		return c.ChildrenPorts[1]
	}
	return ""
}

// GetChildPort returns the port for the specified child index
func (c *NodeConfig) GetChildPort(index int) string {
	if index >= 0 && index < len(c.ChildrenPorts) {
		return c.ChildrenPorts[index]
	}
	return ""
}

// GetNumChildren returns the number of configured children
func (c *NodeConfig) GetNumChildren() int {
	return len(c.ChildrenPorts)
}
