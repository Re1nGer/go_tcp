package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	rd         Reader
	wr         Writer
	conMap     map[*net.Conn]bool
}

func NewServer(listenAddr string, conn net.Conn, ln net.Listener) *Server {
	return &Server{
		listenAddr: listenAddr,
		rd:         *NewReader(conn),
		wr:         *NewWriter(conn),
		ln:         ln,
		conMap:     map[*net.Conn]bool{},
	}
}

// we may pass the channel to notify if there are any errors trying to listen to port over tcp
// TODO: add a custom handler
func (s *Server) Serve(ch chan error) (net.Listener, error) {
	ln, err := net.Listen("tcp", s.listenAddr)

	if err != nil && ch != nil {
		ch <- err
		return nil, err
	}

	//gotta handle concurrency issues here
	s.ln = ln

	return s.ln, nil
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		rd:  bufio.NewReader(r),
		buf: make([]byte, 1024), //for now let's just ASSUME it's gonna be enough
	}
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: bufio.NewWriter(w),
		b: make([]byte, 1024),
	}
}

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

	items := make(map[string][]byte)

	time := make(map[string]int64)

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

		go handleClient(conn, items, time)

		//defer conn.Close()
		//fmt.Print(res)
	}
}

//commands := &map[string]bool{""}

func handleClient(conn net.Conn, items map[string][]byte, time_map map[string]int64) {

	reader := bufio.NewReader(conn)

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

			if el == "ping" {
				conn.Write([]byte("+PONG\r\n"))
			}

			if el == "echo" {
				if idx+1 < len(commands.args) {
					echo_val := commands.args[idx+1]
					ans_arr := make([]byte, 0)
					ans_arr = append(ans_arr, '$')
					ans_arr = strconv.AppendInt(ans_arr, int64(len(echo_val)), 10)
					ans_arr = append(ans_arr, '\r', '\n')
					ans_arr = append(ans_arr, echo_val...)
					ans_arr = append(ans_arr, '\r', '\n')
					conn.Write(ans_arr)
				} else {
					conn.Write([]byte("-Invalid command\r\n"))
				}
			}

			if el == "exists" {
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

			if el == "get" {
				key := commands.args[idx+1]
				val, ok := items[key]
				if ok {
					ans_arr := make([]byte, 0)
					ans_arr = append(ans_arr, '+')
					ans_arr = append(ans_arr, string(val)...)
					ans_arr = append(ans_arr, '\r', '\n')
					conn.Write(ans_arr)
				} else {
					conn.Write([]byte("$-1\r\n"))
				}
			}

			if el == "set" {
				if idx+2 < len(commands.args) {
					key := commands.args[idx+1]
					items[key] = []byte(commands.args[idx+2])
					conn.Write([]byte("+OK\r\n"))
				} else {
					conn.Write([]byte("-Invalid command\r\n"))
				}
			}

			if el == "del" {
				counter := 0

				for _, el := range commands.args[1:] {
					_, ok := items[el]
					delete(items, el)

					if ok {
						counter += 1
					}
				}

				byte_res := make([]byte, 0)

				byte_res = appendPrefix(byte_res, ':', int64(counter))

				byte_res = append(byte_res, ':')

				byte_res = append(byte_res, '\r', '\n')

				conn.Write(byte_res)
			}

			//hasn't been checked yet
			if el == "setex" {
				if idx+3 < len(commands.args) {

					key := commands.args[idx+1]

					seconds, _ := strconv.Atoi(commands.args[idx+2])

					val := commands.args[idx+3]

					items[key] = []byte(val)

					time_map[key] = time.Now().Unix() + int64(seconds)

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

	//handle bulk array *4\r\n$3\r\nSET...

	command := string(buf[:n])

	arr := strings.Split(command, "\r\n")

	for _, el := range arr {
		if len(el) > 0 && (el[0] == '*' || el[0] == '$' || el[0] == '\r' || el[0] == '\n') || el == "" {
			continue
		}
		if isValidCommand(el) {
			res.args = append(res.args, strings.ToLower(el))
		} else { // it means it's an argument
			res.args = append(res.args, el)
		}
	}

	return res, nil
}

func appendPrefix(b []byte, c byte, n int64) []byte {
	if n >= 0 && n <= 9 {
		return append(b, c, byte('0'+n), '\r', '\n')
	}
	b = append(b, c)
	b = strconv.AppendInt(b, n, 10)
	return append(b, '\r', '\n')
}

func isValidCommand(command string) bool {
	return caseInvariant(command, "SET") ||
		caseInvariant(command, "GET") ||
		caseInvariant(command, "DEL") ||
		caseInvariant(command, "EXISTS") ||
		caseInvariant(command, "PING") ||
		caseInvariant(command, "SETEX") ||
		caseInvariant(command, "ECHO")
}

func caseInvariant(s string, c string) bool {
	return strings.ToLower(s) == c || s == c
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
	w   *bufio.Writer
	b   []byte
	err error
}

// buffers need to be reset after reading it to the end
type Reader struct {
	rd  *bufio.Reader
	buf []byte
}

type Command struct {
	args []string
	Raw  []byte
}
