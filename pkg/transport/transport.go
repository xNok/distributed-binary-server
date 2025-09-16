package transport

import (
	"context"

	"github.com/xnok/btree-server-msg/pkg/btree"
)

// Transport defines the interface for network transport layers
type Transport interface {
	// Listen starts listening for incoming connections on the specified address
	Listen(ctx context.Context, address string) error

	// Connect establishes a connection to the specified address
	Connect(ctx context.Context, address string) error

	// Close closes the transport
	Close() error

	// GetInboundChannel returns the channel for incoming messages
	GetInboundChannel() <-chan btree.Message

	// GetOutboundChannel returns the channel for outgoing messages
	GetOutboundChannel() chan<- btree.Message
}

// Server wraps a transport and provides server functionality
type Server struct {
	transport Transport
	address   string
}

// NewServer creates a new transport server
func NewServer(transport Transport, address string) *Server {
	return &Server{
		transport: transport,
		address:   address,
	}
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	return s.transport.Listen(ctx, s.address)
}

// GetInboundChannel returns the inbound channel from the transport
func (s *Server) GetInboundChannel() <-chan btree.Message {
	return s.transport.GetInboundChannel()
}

// GetOutboundChannel returns the outbound channel to the transport
func (s *Server) GetOutboundChannel() chan<- btree.Message {
	return s.transport.GetOutboundChannel()
}

// Close closes the server
func (s *Server) Close() error {
	return s.transport.Close()
}

// Client wraps a transport and provides client functionality
type Client struct {
	transport Transport
	address   string
}

// NewClient creates a new transport client
func NewClient(transport Transport, address string) *Client {
	return &Client{
		transport: transport,
		address:   address,
	}
}

// Connect connects to the remote address
func (c *Client) Connect(ctx context.Context) error {
	return c.transport.Connect(ctx, c.address)
}

// GetOutboundChannel returns the outbound channel to send messages
func (c *Client) GetOutboundChannel() chan<- btree.Message {
	return c.transport.GetOutboundChannel()
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.transport.Close()
}
