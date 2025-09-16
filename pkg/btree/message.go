package btree

import (
	"context"
	"time"
)

// Message represents a message that flows through the tree
type Message struct {
	Content   string
	ID        string    // Optional message ID for tracking
	Timestamp time.Time // When the message was created
	Source    string    // Optional source node identifier
}

// NewMessage creates a new message with timestamp
func NewMessage(content, id string) Message {
	return Message{
		Content:   content,
		ID:        id,
		Timestamp: time.Now(),
	}
}

// MessageHandler defines the interface for handling messages in a tree node
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg Message) error
}

// MessageSender defines the interface for sending messages to child nodes
type MessageSender interface {
	// Send to specific child by index
	SendToChild(ctx context.Context, index int, msg Message) error

	// Convenience methods for binary trees
	SendToLeft(ctx context.Context, msg Message) error
	SendToRight(ctx context.Context, msg Message) error
}

// MessageBroadcaster defines the interface for broadcasting messages
type MessageBroadcaster interface {
	// BroadcastToChildren sends a message to all children
	BroadcastToChildren(ctx context.Context, msg Message) error

	// GetNumChildren returns the number of children
	GetNumChildren() int
}

// MessageReceiver defines the interface for receiving messages
type MessageReceiver interface {
	Receive(ctx context.Context) <-chan Message
}

// MessageService combines all message operations
type MessageService interface {
	MessageHandler
	MessageSender
	MessageBroadcaster
	MessageReceiver
}
