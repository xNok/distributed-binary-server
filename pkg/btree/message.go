package btree

import "context"

// Message represents a message that flows through the binary tree
type Message struct {
	Content string
	ID      string // Optional message ID for tracking
}

// MessageHandler defines the interface for handling messages in a btree node
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg Message) error
}

// MessageSender defines the interface for sending messages to child nodes
type MessageSender interface {
	SendToLeft(ctx context.Context, msg Message) error
	SendToRight(ctx context.Context, msg Message) error
}

// MessageReceiver defines the interface for receiving messages
type MessageReceiver interface {
	Receive(ctx context.Context) <-chan Message
}

// MessageService combines all message operations
type MessageService interface {
	MessageHandler
	MessageSender
	MessageReceiver
}
