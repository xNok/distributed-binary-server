# Distributed Binary Tree with Channel-based Architecture

This project implements a distributed binary tree where each node can propagate messages to its children. 

## Architecture Overview

The application now follows a clean, layered architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │     BTree       │    │   Transport     │
│     Layer       │◄──►│     Layer       │◄──►│     Layer       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
      main.go              pkg/btree/           pkg/transport/
```

### Key Components

#### 1. BTree Layer (`pkg/btree/`)
- **Message**: Defines the message structure flowing through the tree
- **Node**: Implements tree node logic using channels for communication
- **Interfaces**: `MessageHandler`, `MessageSender`, `MessageReceiver` for clean abstractions

#### 2. Transport Layer (`pkg/transport/`)
- **Transport Interface**: Abstract interface for different transport protocols
- **TCP Implementation**: Concrete TCP transport in `pkg/transport/tcp/`
- **Server/Client Wrappers**: Higher-level abstractions for network communication

#### 3. Application Layer (`cmd/node/`)
- **main.go**: Wires together btree nodes with transport layers
- **Configuration**: Command-line based configuration for node topology

## Benefits of the New Architecture

### 1. **Easy Testing**
```go
// No TCP connections needed for testing!
parent := btree.NewNode("parent")
child := btree.NewNode("child")

// Wire them up with channels
go func() {
    for msg := range parent.GetLeftChannel() {
        child.GetInboundChannel() <- msg
    }
}()

// Test message propagation
parent.HandleMessage(ctx, btree.Message{Content: "test"})
```

### 2. **Transport Agnostic**
- Current: TCP transport
- Future: WebSocket, gRPC, message queues, or any other protocol
- Swap transports without changing business logic

### 3. **Type Safety**
- Messages are strongly typed through Go channels
- Compile-time verification of message flow

### 4. **Concurrent Processing**
- Each node processes messages concurrently
- Non-blocking message propagation to children

## Usage

### Running a Single Node
```bash
go run cmd/node/main.go -port 3030
```

### Running a Tree
```bash
# Root node
go run cmd/node/main.go -port 3030 -left 3031 -right 3032

# Left child  
go run cmd/node/main.go -port 3031

# Right child
go run cmd/node/main.go -port 3032
```

### Sending Messages
```bash
echo "Hello, Binary Tree!" | nc localhost 3030
```

## Testing

### Unit Tests
```bash
go test ./pkg/btree/ -v
```

### Channel-based Example
```bash
go run examples/channel_example.go
```

## Project Structure

```
.
├── cmd/
│   └── node/
│       └── main.go              # Application entry point
├── pkg/
│   ├── btree/
│   │   ├── message.go           # Message definitions and interfaces
│   │   ├── node.go              # BTree node implementation
│   │   └── node_test.go         # Channel-based tests
│   └── transport/
│       ├── transport.go         # Transport interfaces and wrappers
│       └── tcp/
│           └── tcp.go           # TCP transport implementation
├── examples/
│   └── channel_example.go       # Testing demonstration
├── go.mod
├── Makefile
└── README.md
```

## Key Interfaces

### Message Interface
```go
type MessageHandler interface {
    HandleMessage(ctx context.Context, msg Message) error
}

type MessageSender interface {
    SendToLeft(ctx context.Context, msg Message) error
    SendToRight(ctx context.Context, msg Message) error
}
```

### Transport Interface
```go
type Transport interface {
    Listen(ctx context.Context, address string) error
    Connect(ctx context.Context, address string) error
    Close() error
    GetInboundChannel() <-chan btree.Message
    GetOutboundChannel() chan<- btree.Message
}
```

## Future Enhancements

1. **Additional Transport Protocols**
   - WebSocket transport for browser clients
   - gRPC transport for high-performance scenarios
   - Message queue transport for reliability

2. **Advanced Features**
   - Message acknowledgments
   - Message routing based on content
   - Load balancing across children
   - Persistent message storage

3. **Monitoring**
   - Metrics collection
   - Health checks
   - Distributed tracing

## Migration from MVP

The original MVP directly used TCP connections in the business logic. The new architecture:

1. ✅ **Separated concerns**: Business logic (btree) from transport (TCP)
2. ✅ **Enabled testing**: Can test message flow without network
3. ✅ **Improved maintainability**: Clean interfaces and modular design
4. ✅ **Enhanced flexibility**: Easy to swap transport protocols

This refactoring transforms the MVP into a production-ready, clean code application that follows SOLID principles and enables comprehensive testing.