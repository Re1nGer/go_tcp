package main

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestSetGetWithRedis(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Replace with your Redis server address
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

func TestReadRESP_EchoCommand(t *testing.T) {
	// Create a buffer with a simple command
	data := []byte("*1\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")
	reader := bufio.NewReader(bytes.NewReader(data))

	// Call the function
	cmd, err := readRESP(reader)

	// Assert no error and expected command/argument
	if err != nil {
		t.Errorf("Error reading RESP: %v", err)
		return
	}

	if cmd.args[0] != "echo" || len(cmd.args) != 2 || cmd.args[1] != "hello" {
		t.Errorf("Unexpected command: %v, args: %v", cmd, cmd.args)
	}
}
