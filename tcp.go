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

	map_con := map[net.Conn]bool{}

	var items = make(map[string][]byte)

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

		go handleClient(conn, map_con, items)

		//defer conn.Close()
		//fmt.Print(res)
	}
}

func (w *Writer) WriteInline(s string) {
	fmt.Fprintf(&w.w, "+%s\r\n", toInline(s))
}

func handleClient(conn net.Conn, map_con map[net.Conn]bool, items map[string][]byte) {
	//defer conn.Close()

	// Read data from the client
	//bufWriter := bufio.NewWriter(conn)

	//buf := &Writer{w: *bufWriter}

	reader := bufio.NewReader(conn)

	/* 	res, err := buf.ParseRESP(conn)
	   	if err != nil {
	   		panic(err)
	   	}

	   	fmt.Println(res) */

	//var args []CommandArg

	for {

		message, err := reader.ReadString('\n')

		fmt.Print(message)

		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}

		if err != nil {
			panic(err)
		}

		if message == "" {
			break
		}

		/* 		value, par_err := parseLine(message)

		   		if par_err != nil {
		   			break
		   		}

		   		args = append(args, CommandArg{Value: value}) */
	}

	/* 	for {

		message, err := reader.ReadString('\n')

		if strings.Contains(message, "*") {
			counter += 1
		}

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

		buf.WriteInline("PONG")

		buf.w.Flush()

		if err != nil {
			fmt.Println("Error writing to connection:", err)
			break
		}
	} */

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

func (w *Writer) ParseRESP(reader net.Conn) ([]CommandArg, error) {

	scanner := bufio.NewReader(reader)

	var args []CommandArg

	for {
		line, err := scanner.ReadString('\n')

		fmt.Print(line)

		if err != nil {
			return nil, err
		}

		if line == "" {
			break
		}

		/* 		value, par_err := parseLine(line)

		   		if par_err != nil {
		   			return nil, err
		   		}

		   		args = append(args, CommandArg{Value: value}) */
	}
	return args, nil
}

type CommandArg struct {
	Type  string
	Value string
}

type Writer struct {
	w   bufio.Writer
	b   []byte
	err error
}

type Reader struct {
	rd    *bufio.Reader
	buf   []byte
	start int
	end   int
	cmds  []Command
}

type Command struct {
	args [][]byte
	Raw  []byte
}
