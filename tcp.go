package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
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

		fmt.Print(commands)

		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}
	}
}

func readRESP(reader *bufio.Reader) ([]Command, error) {

	buf := make([]byte, 512)

	n, err := reader.Read(buf)

	if err != nil {
		return []Command{}, nil
	}

	switch buf[0] {
	case '*':
		//handle bulk array *4\r\n$3\r\nSET...
		for i := 0; i < n; i++ {
			if buf[i] == '\n' && buf[i-1] != '\r' {
				panic("Incorrect Format")
			}

			if buf[i] == '\r' {
				line := buf[:i-1]

				fmt.Println(line)

				switch line[0] {

				case '$':
					fmt.Print("handle bulk string")
					//gotta find out the size
					//welp, let's ASSUME the digit is 1 byte long

					str_size, _ := parseInt([]byte{line[1]})

					command_arr := make([]byte, str_size)

					for j := 0; j < str_size; j++ {
						command_arr = append(command_arr, line[j])
					}

				}
			}
		}
	}

	return []Command{}, nil

}

func readBulkString(line string, reader *bufio.Reader) (string, error) {

	length, err := strconv.Atoi(line[1:])

	if err != nil {
		return "", err
	}

	if length == -1 {
		return "", nil
	}

	bulk := make([]byte, length+2)

	_, err = reader.Read(bulk)

	if err != nil {
		return "", err
	}

	return string(bulk[:length]), nil
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
	args [][]byte
	Raw  []byte
}
