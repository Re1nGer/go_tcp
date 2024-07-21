package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

// test requires spinning up a real redis instance
func TestSetGetWithRedis(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer rdb.Close()

	key := "test_key"
	value := "test_value"

	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		t.Errorf("Error setting key: %v", err)
		return
	}

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		t.Errorf("Error getting key: %v", err)
		return
	}

	if val != value {
		t.Errorf("Expected value %s, got %s", value, val)
	}
}

func TestReadRESP2_EchoCommand(t *testing.T) {
	// Create a buffer with a simple command
	data := []byte("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")

	rd := bytes.NewReader(data)

	reader := NewReader(rd)

	n := copy(reader.buf, data)

	reader.buf = reader.buf[:n]

	// Call the function
	cmd, err := reader.readRESP2()

	// Assert no error and expected command/argument
	if err != nil {
		t.Errorf("Error reading RESP: %v", err)
		return
	}

	if string(cmd.args[0]) != "ECHO" {
		t.Errorf("Error reading RESPPPPRRRRp: %v", string(cmd.args[0]))
		t.Log(len(string(cmd.args[0])))
	}

}

func TestCommand(t *testing.T) {
	data := []byte("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")

	rd := bytes.NewReader(data)

	reader := NewReader(rd)

	n := copy(reader.buf, data)

	reader.buf = reader.buf[:n]

	output, _ := reader.readRESP2()

	t.Log(string(output.args[0]), string(output.args[1]))

	if len(output.args) != 2 {
		t.Errorf("Error reading RESPPPPRRRRp: %v", string(output.args[0]))
	}
}

func TestReadRESP_EchoCommand(t *testing.T) {
	// Create a buffer with a simple command
	data := []byte("*2\r\n$3\r\nGET\r\n$5\r\nhello\r\n")

	rd := bytes.NewReader(data)

	reader := NewReader(rd)

	n := copy(reader.buf, data)

	reader.buf = reader.buf[:n]

	// Call the function
	cmd, err := reader.readRESP()

	// Assert no error and expected command/argument
	if err != nil {
		t.Errorf("Error reading RESPPPPPPP: %v", err)
		return
	}

	if cmd.args[0] != "echo" || len(cmd.args) != 2 || cmd.args[1] != "hello" {
		t.Errorf("Unexpected command: %v, args: %v", cmd, cmd.args)
	}
}

func BenchmarkResp(b *testing.B) {
	data := []byte("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")

	rd := bytes.NewReader(data)

	reader := NewReader(rd)

	n := copy(reader.buf, data)

	reader.buf = reader.buf[:n]

	for i := 0; i < b.N; i++ {
		reader.readRESP()
	}
}

func BenchmarkResp2(b *testing.B) {
	data := []byte("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")

	rd := bytes.NewReader(data)

	reader := NewReader(rd)

	n := copy(reader.buf, data)

	reader.buf = reader.buf[:n]

	for i := 0; i < b.N; i++ {
		reader.readRESP2()
	}
}

func BenchmarkRespDiff(b *testing.B) {
	data := []byte("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")

	rd := bytes.NewReader(data)

	reader := NewReader(rd)

	n := copy(reader.buf, data)

	reader.buf = reader.buf[:n]

	b.Run("old version", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader.readRESP()
		}
	})

	b.Run("new version", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader.readRESP2()
		}
	})
}
