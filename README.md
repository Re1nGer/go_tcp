## Custom Redis Server (Golang)

This lightweight Redis server, built with Golang, implements core functionalities for a key-value store. Ideal for learning and experimentation.

**Features:**

* In-memory key-value storage
* Basic string commands (SET, GET, DEL)
* RESP protocol communication

**Getting Started:**

1. **Prerequisites:** Golang (version 1.18 or later recommended)
2. **Build:** `go build`
3. **Run:** `./tcp.go`

**Usage:**

The server listens on port 6379 by default. Use any Redis client to interact with it using the RESP protocol.

**Development:**

This project serves as a foundation for further exploration. Consider extending functionalities with additional data structures, persistence, and advanced commands.

**License:**

MIT License
