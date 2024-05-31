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

	map_con := make(map[net.Conn]bool)

	items := make(map[string][]byte)

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

		go handleClient(conn, &map_con, &items)

		//defer conn.Close()
		//fmt.Print(res)
	}
}

func (w *Writer) WriteInline(s string) {
	fmt.Fprintf(&w.w, "+%s\r\n", toInline(s))
}

func handleClient(conn net.Conn, map_con *map[net.Conn]bool, items *map[string][]byte) {

	reader := bufio.NewReader(conn)

	//writer := bufio.NewWriter(conn)

	var args []string

	for {

		b, err := reader.ReadBytes('\n')

		//m, _ := reader.ReadString('\n')

		fmt.Println(b, string(b))

		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}

		//conn.Write([]byte("+PONG\r\n"))

		//fmt.Println()

		if len(b) > 0 {
			switch b[0] {
			default:
				args = append(args, string(b))
			case '$':
				n, _ := parseInt([]byte{b[1]})
				fmt.Println("Captured dollar sign, potentially can know the size of the buffer", n)
			case '*':
				marks := make([]int, 0, 16)
				fmt.Println(marks)
			case '+':
				//simple string
				args = append(args, string(b[1:]))
			case '-':
				fmt.Println("should capture errors")
			}

		}

		//we have arguments, time to answer

		if len(args) > 0 {
			fmt.Println("has commands", args)
			if len(args) == 4 {
				conn.Write([]byte("+OK\r\n"))
			}
		}
	}
}

func parseInt(b []byte) (int, bool) {
	if len(b) == 1 && b[0] >= '0' && b[0] <= '9' {
		return int(b[0] - '0'), true
	}
	var n int
	var sign bool
	var i int

	if len(b) > 0 && b[0] == '-' {
		sign = true
		i++
	}
	for ; i < len(b); i++ {
		if b[i] < '0' || b[i] > '9' {
			return 0, false
		}
		n = n*10 + int(b[i]-'0')
	}
	if sign {
		n *= -1
	}
	return n, true
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
