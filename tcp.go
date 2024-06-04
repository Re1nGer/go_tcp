package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
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

		go handleClient(conn, &map_con, items)

		//defer conn.Close()
		//fmt.Print(res)
	}
}

//commands := &map[string]bool{""}

func handleClient(conn net.Conn, map_con *map[net.Conn]bool, items map[string][]byte) {

	reader := bufio.NewReader(conn)

	//var commands []Command

	//handle client handles bulk array of messages (most of the time)

	//fmt.Print(string(buf[:n]))

	for {
		//fmt.Println("whole message", string(b))

		commands, err := readRESP(reader)

		for idx, el := range commands.args {

			fmt.Println(commands.args)

			//for now just ignore CLIENT SETINFO bulshit
			if el == "redis-py" {
				conn.Write([]byte("+OK\r\n"))
			}

			if el == "5.0.4" {
				conn.Write([]byte("+OK\r\n"))
			}

			if el == "PING" {
				conn.Write([]byte("+PONG\r\n"))
			}

			if el == "EXISTS" {

				counter := 0

				for i := range len(commands.args) - 1 {

					_, ok := items[commands.args[i+1]]

					if ok {
						counter += 1
					}

					counter_str := strconv.Itoa(counter)

					conn.Write([]byte(":" + counter_str + "\r\n"))
				}

			}

			if el == "GET" {
				key := commands.args[idx+1]
				val, ok := items[key]
				if ok {
					ans := "+" + string(val) + "\r\n"
					conn.Write([]byte(ans))
				} else {
					conn.Write([]byte("$-1\r\n"))
				}
			}

			if el == "SET" {
				if idx+1 < len(commands.args) {
					key := commands.args[idx+1]
					items[key] = []byte(commands.args[idx+2])
					conn.Write([]byte("+OK\r\n"))
				} else {
					conn.Write([]byte("-Invalid command\r\n"))
				}
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

/* func (rd *Reader) readCommands(leftover *int) ([]Command, error) {

	var cmds []Command

		marks := make([]int, 0, 16)
		for i := 1; i < len(b); i++ {
			if b[i] == '\n' {
				if b[i-1] != '\r' {
					return nil, errInvalidMultiBulkLength
				}
				count, ok := parseInt(b[1 : i-1])
				if !ok || count <= 0 {
					return nil, errInvalidMultiBulkLength
				}
				marks = marks[:0]
				for j := 0; j < count; j++ {
					// read bulk length
					i++
					if i < len(b) {
						if b[i] != '$' {
							return nil, &errProtocol{"expected '$', got '" +
								string(b[i]) + "'"}
						}
						si := i
						for ; i < len(b); i++ {
							if b[i] == '\n' {
								if b[i-1] != '\r' {
									return nil, errInvalidBulkLength
								}
								size, ok := parseInt(b[si+1 : i-1])
								if !ok || size < 0 {
									return nil, errInvalidBulkLength
								}
								if i+size+2 >= len(b) {
									// not ready
									break outer2
								}
								if b[i+size+2] != '\n' ||
									b[i+size+1] != '\r' {
									return nil, errInvalidBulkLength
								}
								i++
								marks = append(marks, i, i+size)
								i += size + 1
								break
							}
						}
					}
				}
				if len(marks) == count*2 {
					var cmd Command
					if rd.rd != nil {
						// make a raw copy of the entire command when
						// there's a underlying reader.
						cmd.Raw = append([]byte(nil), b[:i+1]...)
					} else {
						// just assign the slice
						cmd.Raw = b[:i+1]
					}
					cmd.Args = make([][]byte, len(marks)/2)
					// slice up the raw command into the args based on
					// the recorded marks.
					for h := 0; h < len(marks); h += 2 {
						cmd.Args[h/2] = cmd.Raw[marks[h]:marks[h+1]]
					}
					cmds = append(cmds, cmd)
					b = b[i+1:]
					if len(b) > 0 {
						goto next
					} else {
						goto done
					}
				}
			}
		}
	if leftover != nil {
		*leftover = rd.end - rd.start
	}
	if len(cmds) > 0 {
		return cmds, nil
	}
	if rd.rd == nil {
		return nil, errIncompleteCommand
	}
	if rd.end == len(rd.buf) {
		// at the end of the buffer.
		if rd.start == rd.end {
			// rewind the to the beginning
			rd.start, rd.end = 0, 0
		} else {
			// must grow the buffer
			newbuf := make([]byte, len(rd.buf)*2)
			copy(newbuf, rd.buf)
			rd.buf = newbuf
		}
	}
	n, err := rd.rd.Read(rd.buf[rd.end:])
	if err != nil {
		return nil, err
	}
	rd.end += n
	return rd.readCommands(leftover)
} */

func (wr *Writer) writeSimpleString(s string) {
	wr.b = append(wr.b, '+')
	wr.b = append(wr.b, s...)
	wr.b = append(wr.b, '\r', '\n')
}

/* func Parse(raw []byte) (Command, error) {
	rd := Reader{buf: raw, end: len(raw)}
	var leftover int
	cmds, err := rd.readCommands(&leftover)
	if err != nil {
		return Command{}, err
	}
	if leftover > 0 {
		return Command{}, errors.New("too much data")
	}
	return cmds[0], nil
} */

type Writer struct {
	w   bufio.Writer
	b   []byte
	err error
}

// buffers need to be reset after reading it to the end
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
