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