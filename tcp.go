package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"unicode"
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

func (w *Writer) WriteInline(s string) {
	fmt.Fprintf(&w.w, "+%s\r\n", toInline(s))
}

func handleClient(conn net.Conn) {
	//defer conn.Close()

	// Read data from the client
	reader := bufio.NewReader(conn)

	bufWriter := bufio.NewWriter(conn)

	buf := &Writer{w: *bufWriter}

	for {
		message, err := reader.ReadString('\n')

		if message == "" {
			break
		}
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}

		fmt.Print("Received message: ", message)

		//writer.WriteInline("PONG")

		//writer.w.Flush()

		//conn.

		//conn.Close()

		//fmt.Fprintf(conn, "+%s\r\n", toInline("PONG"))

		buf.WriteInline("PONG")

		buf.w.Flush()
		/*
			arr := make([]byte, 0)

			arr = append(arr, '+')

			arr = append(arr, (stripNewlines("PONG"))...)

			arr = append(arr, '\r', '\n') */

		//fmt.Println(arr)

		//_, err = conn.Write(arr)

		if err != nil {
			fmt.Println("Error writing to connection:", err)
			break
		}
	}

	/* 	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	} */
	//receivedData := string(buffer[:n])

	// Create an instance of Writer

	// Respond back to the client

	//_, _ := conn.Write([]byte(response))

	//defer conn.Close()

	/* 	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return
	} */
	//fmt.Println("Sent response to client:", response)
}

func stripNewlines(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' || s[i] == '\n' {
			s = strings.Replace(s, "\r", " ", -1)
			s = strings.Replace(s, "\n", " ", -1)
			break
		}
	}
	return s
}

func toInline(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
}

type Writer struct {
	w     bufio.Writer
	resp3 bool
}
