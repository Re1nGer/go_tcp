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

//commands := &map[string]bool{""}

func handleClient(conn net.Conn, map_con *map[net.Conn]bool, items *map[string][]byte) {

	reader := bufio.NewReader(conn)

	//var commands []Command

	//handle client handles bulk array of messages (most of the time)

	//fmt.Print(string(buf[:n]))

	for {
		//fmt.Println("whole message", string(b))

		commands, err := readRESP(reader)

		for _, el := range commands.args {
			//for now just ignore CLIENT SETINFO bulshit
			fmt.Println(el, len(el), el == "PING")

			if el == "redis-py" {
				conn.Write([]byte("+OK\r\n"))
			}

			if el == "5.0.4" {
				conn.Write([]byte("+OK\r\n"))
			}

			if el == "PING" {
				conn.Write([]byte("+PONG\r\n"))
			}
		}

		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}
	}
}

func readRESP(reader *bufio.Reader) (Command, error) {

	res := Command{}

	buf := make([]byte, 512)

	n, err := reader.Read(buf)

	if err != nil {
		return Command{}, nil
	}

	switch buf[0] {
	case '*':
		//handle bulk array *4\r\n$3\r\nSET...
		command := string(buf[:n])

		arr := strings.Split(command, "\r\n")

		fmt.Println(len(arr))
		for _, el := range arr {
			if len(el) > 0 && (el[0] == '*' || el[0] == '$') {
				continue
			}
			res.args = append(res.args, el)
		}
	}

	return res, nil

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
	args []string
	Raw  []byte
}
