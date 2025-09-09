


node1:
	go run ./cmd/node/main.go -port 3030 -right 3031 -left 3032

node2:
	go run ./cmd/node/main.go -port 3031

node3:
	go run ./cmd/node/main.go -port 3032