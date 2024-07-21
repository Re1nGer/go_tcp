# Redis Server Implementation in Go

## Overview

This project is a lightweight Redis server implementation written in Go. It supports basic Redis commands and uses the Redis Serialization Protocol (RESP) for message parsing. The server is designed to be efficient and easy to extend with additional Redis commands.

## Features

- RESP (Redis Serialization Protocol) message parsing
- Concurrent handling of client connections
- Support for basic Redis commands: ECHO, PING, SET, GET, EXISTS, DEL
- Extensible architecture for adding new commands
- Thread-safe operations on shared data

## Getting Started

### Prerequisites

- Go 1.15 or higher

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/re1nger/redis-go-server.git
   ```
2. Navigate to the project directory:
   ```
   cd redis-go-server
   ```

### Running the Server

To start the Redis server, run the following command:

```
go run tcp.go
```

By default, the server listens on `localhost:6379`.

## Usage

You can interact with the server using any Redis client. Here are some example commands you can use:

```
SET key value
GET key
EXISTS key
DEL key
PING
ECHO message
```

## API

The server exposes two main interfaces: `Conn` and `C2`.

### Conn Interface

The `Conn` interface represents a client connection and provides methods for writing responses:

```go
type Conn interface {
    WriteError(msg string)
    WriteString(str string)
    WriteBulkString(bulk string)
    WriteInt(num int)
    WriteRaw(b []byte)
    WriteBytes(b []byte)
}
```

### C2 Struct

The `C2` struct represents a command received from a client:

```go
type C2 struct {
    args [][]byte
    Raw  []byte
}
```

- `args`: A slice of byte slices representing the command and its arguments
- `Raw`: The raw byte data of the entire command

## Extending the Server

To add new commands, modify the `main()` function in `tcp.go`. Add a new case to the switch statement, implementing the desired functionality. For example:

```go
case "newcommand":
    // Implement your new command here
    conn.WriteString("New command executed")
```

## Architecture

The server uses a concurrent model where each client connection is handled in a separate goroutine. It maintains a thread-safe map of key-value pairs for data storage.

Key components:
- `Server`: Manages the TCP listener and client connections
- `conn`: Represents an individual client connection
- `Reader` and `Writer`: Handle RESP protocol parsing and response formatting

## Limitations and Future Improvements
- No persistence mechanism
- Transactions, pub/sub, or clustering is not implemented yet

Future improvements could include:
- Adding data persistence
- Implementing Redis clustering capabilities
- Enhancing error handling and logging

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.