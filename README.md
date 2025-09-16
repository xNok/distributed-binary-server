# btree-server-msg

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

## Features

##  Message Broadcasting

By default when a node recieves a message it broadcast that message to all childrens (left and right in binary tree)

### Setup a Tree Structure

```bash
# Terminal 1 - Root node (will broadcast to children)
make node1  # Starts root on port 3030, connects to 3031 and 3032

# Terminal 2 - Left child
make node2  # Starts left child on port 3031

# Terminal 3 - Right child  
make node3  # Starts right child on port 3032
```

### Send Messages and Observe Broadcasting

```bash
# Send a message to the root node
echo "Broadcasting test message!" | nc localhost 3030
```

**Expected Output:**
```
# Root node (3030) logs:
[node-3030] Received message: Broadcasting test message!
[node-3030] Forwarded to child 0
[node-3030] Forwarded to child 1

# Left child (3031) logs:
[node-3031] Received message: Broadcasting test message!
[node-3031] Forwarded to child 0  # (if it has children)
[node-3031] Forwarded to child 1  # (if it has children)

# Right child (3032) logs:
[node-3032] Received message: Broadcasting test message!
[node-3032] Forwarded to child 0  # (if it has children)  
[node-3032] Forwarded to child 1  # (if it has children)
```

### Multiple Messages

```bash
# Send multiple messages rapidly
(echo "Message 1"; echo "Message 2"; echo "Message 3") | nc localhost 3030
```

All messages will be broadcast to every node in the tree, demonstrating the complete propagation behavior.