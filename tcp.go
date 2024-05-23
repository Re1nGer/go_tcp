package main

import (
	"fmt"
	"net"
)

func main() {
	// Listen for incoming connections
	listener, err := net.Listen("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	//var m map[string]int

	//m = make(map[string]int)

	fmt.Println("Server is listening on port 6379")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Handle client connection in a goroutine
		//go handleClient(conn)

		//line := bufio.NewReader(conn)

		//res, err := line.ReadBytes('\n')

		if err != nil {
			panic(err)
		}

		go handleClient(conn)

		//defer conn.Close()
		//fmt.Print(res)
	}
}

func handleClient(conn net.Conn) {
	//defer conn.Close()

	// Read data from the client
	/* 	buffer := make([]byte, 1024)
	   	n, err := conn.Read(buffer)
	   	if err != nil {
	   		fmt.Println("Error reading:", err.Error())
	   		return
	   	}
	   	receivedData := string(buffer[:n])
	   	fmt.Println("Received message from client:", receivedData) */

	// Respond back to the client
	response := "+PONG\r\n"

	_, _ = conn.Write([]byte(response))

	/* 	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return
	} */
	//fmt.Println("Sent response to client:", response)
}
