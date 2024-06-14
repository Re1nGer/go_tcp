package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// this is incorrect because each connection should have its reader and writer in conn struct
// but Server struct should have a map of connections, each connection is treated separately
type Server struct {
	listenAddr string
	conMap     map[*conn]bool
	ln         net.Listener
	handler    func(c Conn, cmd Command)
}

type conn struct {
	rd   *Reader
	wr   *Writer
	conn net.Conn
	cmds Command
}

type Conn interface {
	WriteError(msg string)
	// WriteString writes a string to the client.
	WriteString(str string)
	// WriteBulk writes bulk bytes to the client.
	WriteBulk(bulk []byte)
	// WriteBulkString writes a bulk string to the client.
	WriteBulkString(bulk string)
	// WriteInt writes an integer to the client.
	WriteInt(num int)
}

// this is a communication channel from client to our library
func (c conn) WriteBulk(bulk []byte)       {}
func (c conn) WriteBulkString(bulk string) {}
func (c conn) WriteInt(num int)            {}
func (c conn) WriteError(msg string)       {}
func (c conn) WriteString(str string)      { c.wr.WriteString(str) }

func NewServer(listenAddr string, ln net.Listener) *Server {
	return &Server{
		listenAddr: listenAddr,
		ln:         ln,
		conMap:     make(map[*conn]bool),
	}
}

func newServer() *Server {
	return &Server{
		conMap: make(map[*conn]bool),
	}
}

func NewServerNetwork(addr string, handler func(conn Conn, cmd Command)) *Server {
	if handler == nil {
		panic("handler is nil")
	}
	s := newServer()
	s.handler = handler
	s.listenAddr = addr
	return s
}

func ListenAndServe(addr string, handler func(conn Conn, cmd Command)) error {
	return ListenAndServeNetwork(addr, handler)
}

func (s *Server) ListenAndServe() error {
	return s.Serve(nil)
}

func ListenAndServeNetwork(addr string, handler func(c Conn, cmd Command)) error {
	return NewServerNetwork(addr, handler).ListenAndServe()
}

// we may pass the channel to notify if there are any errors trying to listen to port over tcp
// TODO: add a custom handler
func (s *Server) Serve(ch chan error) error {
	ln, err := net.Listen("tcp", s.listenAddr)
	fmt.Println("Listen on port", s.listenAddr)
	if err != nil && ch != nil {
		ch <- err
		return err
	}

	//gotta handle concurrency issues here
	s.ln = ln

	return serve(s)
}

func serve(s *Server) error {
	//for now let's just attempt to close the listener this way
	defer s.ln.Close()

	for {
		ln, err := s.ln.Accept()
		if err != nil {
			return err
		}

		con := &conn{
			conn: ln, //necessary to close connection or to get any kind of information about current connection
			wr:   NewWriter(ln),
			rd:   NewReader(ln), //filling in buffers, all incoming messages are in rd reader
		}

		s.conMap[con] = true

		//handle requets in a separate goroutine
		go handle(s, *con)
	}
}

func handle(s *Server, c conn) {

	defer func() {
		c.conn.Close()
	}()

	_ = func() error {
		for {

			commands, err := readRESP(c.rd)

			if err != nil {
				//write to the client
				fmt.Print("Error ?")
				c.wr.WriteError("ERR " + err.Error()) //write error
				c.wr.Flush()
				return err
			}

			c.cmds = commands

			s.handler(c, commands)

			fmt.Println("Do we get to flushing ?")

			//fmt.Println(string(c.wr.b), "Empty??")

			err = c.wr.Flush()

			if err != nil {
				return err
			}
		}
	}()

}

// handler which allows for users

func NewReader(r io.Reader) *Reader {
	return &Reader{
		rd:  bufio.NewReader(r),
		buf: make([]byte, 1024), //for now let's just ASSUME it's gonna be enough
	}
}

func NewWriter(wr io.Writer) *Writer {
	return &Writer{
		w: wr,
	}
}

func (wr *Writer) Flush() error {
	_, wr.err = wr.w.Write(wr.b)
	wr.b = wr.b[:0] //emptying the slice
	return wr.err
}

func (wr *Writer) WriteError(err string) {
	wr.b = append(wr.b, '-')
	wr.b = append(wr.b, []byte(err)...)
	wr.b = append(wr.b, '\r', '\n')
}

func (wr *Writer) WriteString(s string) {
	wr.b = append(wr.b, '+')
	wr.b = append(wr.b, []byte(s)...)
	wr.b = append(wr.b, '\r', '\n')
	fmt.Println(string(wr.b))
}

/* func main() {
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
} */

func main() {
	err := ListenAndServe("localhost:6379", func(conn Conn, cmd Command) {

		//items := make(map[string]string)

		switch cmd.args[0] {
		case "CLIENT":
			fmt.Print("Hit client")
			conn.WriteString("OK")
		case "ping":
			conn.WriteString("PONG")

		}

		fmt.Print(cmd)
	})

	if err != nil {
		panic(err)
	}
}

//commands := &map[string]bool{""}

func handleClient(conn net.Conn, items map[string][]byte) {

	for {

		commands, err := readRESP(&Reader{})

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

			/* 			//hasn't been checked yet
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
			   			} */
		}

		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}
	}
}

func readRESP(r *Reader) (Command, error) {

	res := Command{}

	buf := make([]byte, len(r.buf))

	copy(buf, r.buf)

	//handle bulk array *4\r\n$3\r\nSET...

	command := string(buf)

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
	w   io.Writer
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
