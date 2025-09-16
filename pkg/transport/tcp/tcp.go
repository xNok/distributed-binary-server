package tcp

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/xnok/btree-server-msg/pkg/btree"
)

// TCPTransport implements the Transport interface using TCP
type TCPTransport struct {
	inbound  chan btree.Message
	outbound chan btree.Message
	listener net.Listener
	conn     net.Conn
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex
	isServer bool
	isClient bool
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport() *TCPTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPTransport{
		inbound:  make(chan btree.Message, 100),
		outbound: make(chan btree.Message, 100),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Listen starts listening for incoming TCP connections
func (t *TCPTransport) Listen(ctx context.Context, address string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isServer {
		return fmt.Errorf("already listening")
	}

	// Ensure address has port format
	if !strings.Contains(address, ":") {
		address = ":" + address
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", address, err)
	}

	t.listener = listener
	t.isServer = true

	log.Printf("TCP transport listening on %s", address)

	// Start accepting connections
	t.wg.Add(1)
	go t.acceptConnections(ctx)

	// Start processing outbound messages
	t.wg.Add(1)
	go t.processOutbound()

	return nil
}

// Connect establishes a TCP connection to the specified address
func (t *TCPTransport) Connect(ctx context.Context, address string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isClient {
		return fmt.Errorf("already connected")
	}

	// Ensure address has localhost prefix if just port
	if !strings.Contains(address, ":") {
		address = "localhost:" + address
	} else if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}

	t.conn = conn
	t.isClient = true

	log.Printf("TCP transport connected to %s", address)

	// Start processing outbound messages
	t.wg.Add(1)
	go t.processOutbound()

	return nil
}

// Close closes the TCP transport
func (t *TCPTransport) Close() error {
	t.cancel()

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.listener != nil {
		t.listener.Close()
	}

	if t.conn != nil {
		t.conn.Close()
	}

	// Wait for goroutines to finish
	t.wg.Wait()

	// Close channels
	close(t.inbound)
	close(t.outbound)

	return nil
}

// GetInboundChannel returns the channel for incoming messages
func (t *TCPTransport) GetInboundChannel() <-chan btree.Message {
	return t.inbound
}

// GetOutboundChannel returns the channel for outgoing messages
func (t *TCPTransport) GetOutboundChannel() chan<- btree.Message {
	return t.outbound
}

// acceptConnections accepts incoming TCP connections
func (t *TCPTransport) acceptConnections(ctx context.Context) {
	defer t.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Failed to accept connection: %v", err)
					continue
				}
			}

			// Handle each connection in a separate goroutine
			t.wg.Add(1)
			go t.handleConnection(conn)
		}
	}
}

// handleConnection handles a single TCP connection
func (t *TCPTransport) handleConnection(conn net.Conn) {
	defer t.wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
			text := scanner.Text()
			if text != "" {
				msg := btree.Message{
					Content: text,
					ID:      "", // Could generate UUID here if needed
				}

				select {
				case t.inbound <- msg:
					log.Printf("TCP: Received message: %s", text)
				case <-t.ctx.Done():
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("TCP: Connection scan error: %v", err)
	}
}

// processOutbound sends outbound messages over TCP
func (t *TCPTransport) processOutbound() {
	defer t.wg.Done()

	for {
		select {
		case msg := <-t.outbound:
			if err := t.sendMessage(msg); err != nil {
				log.Printf("TCP: Failed to send message: %v", err)
			}
		case <-t.ctx.Done():
			return
		}
	}
}

// sendMessage sends a message over the TCP connection
func (t *TCPTransport) sendMessage(msg btree.Message) error {
	t.mu.RLock()
	conn := t.conn
	t.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("no active connection")
	}

	message := msg.Content
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	_, err := conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	log.Printf("TCP: Sent message: %s", strings.TrimSpace(message))
	return nil
}
