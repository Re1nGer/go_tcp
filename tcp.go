package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

// this is incorrect because each connection should have its reader and writer in conn struct
// but Server struct should have a map of connections, each connection is treated separately
type Server struct {
	listenAddr string
	conMap     map[*conn]bool
	ln         net.Listener
	handler    func(c Conn, cmd C2)
	mu         sync.Mutex
}

type conn struct {
	rd   *Reader
	wr   *Writer
	conn net.Conn
	cmds C2
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
	WriteRaw(b []byte)
	WriteByte(b []byte)
}

// this is a communication channel from client to our library
func (c conn) WriteBulk(bulk []byte)       {}
func (c conn) WriteBulkString(bulk string) {}
func (c conn) WriteInt(num int)            { c.wr.WriteInt(num) }
func (c conn) WriteError(msg string)       {}
func (c conn) WriteByte(b []byte)          { c.wr.WriteByte(b) }
func (c conn) WriteString(str string)      { c.wr.WriteString(str) }
func (c conn) WriteRaw(b []byte)           { c.wr.WriteRaw(b) }

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

func NewServerNetwork(addr string, handler func(conn Conn, cmd C2)) *Server {
	if handler == nil {
		panic("handler is nil")
	}
	s := newServer()
	s.mu.Lock()
	s.handler = handler
	s.listenAddr = addr
	s.mu.Unlock()
	return s
}

func ListenAndServe(addr string, handler func(conn Conn, cmd C2)) error {
	return ListenAndServeNetwork(addr, handler)
}

func (s *Server) ListenAndServe() error {
	return s.Serve(nil)
}

func ListenAndServeNetwork(addr string, handler func(c Conn, cmd C2)) error {
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
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

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

			commands, err := c.rd.readRESP2()

			if err != nil {
				//write to the client
				c.wr.WriteError("ERR " + err.Error()) //write error
				c.wr.Flush()
				return err
			}

			c.cmds = commands

			s.handler(c, commands)

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
}

func (wr *Writer) WriteByte(b []byte) {
	wr.b = append(wr.b, '+')
	wr.b = append(wr.b, b...)
	wr.b = append(wr.b, '\r', '\n')
}

func (wr *Writer) WriteInt(i int) {
	//wr.b = append(wr.b, ':')
	if i >= 0 && i <= 9 {
		wr.b = append(wr.b, ':', byte('0'+i), '\r', '\n')
		return
	}
	wr.b = append(wr.b, ':')
	wr.b = strconv.AppendInt(wr.b, int64(i), 10) //you cant just go ahead and append int
	wr.b = append(wr.b, '\r', '\n')
}

func (wr *Writer) WriteRaw(b []byte) {
	copy(wr.b, b)
}

func main() {
	var mu sync.RWMutex
	items := make(map[string][]byte)
	err := ListenAndServe("localhost:6379", func(conn Conn, cmd C2) {
		fmt.Println(string(cmd.args[0]))
		switch strings.ToLower(string(cmd.args[0])) {
		case "client":
			conn.WriteString("OK")
		case "ping":
			conn.WriteString("PONG")
		case "set":
			mu.Lock()
			items[string(cmd.args[1])] = cmd.args[2]
			mu.Unlock()
			conn.WriteString("OK")
		case "get":
			mu.RLock()
			val, ok := items[string(cmd.args[1])]
			mu.RUnlock()
			if ok {
				conn.WriteByte(val)
			} else {
				conn.WriteRaw([]byte("$-1\r\n"))
			}
		case "exists":
			counter := 0
			for i := range len(cmd.args) - 1 {
				mu.RLock()
				_, ok := items[string(cmd.args[i+1])]
				mu.RUnlock()
				if ok {
					counter += 1
				}
				fmt.Println(counter)
			}
			conn.WriteInt(counter)
		case "del":
			counter := 0
			for _, el := range cmd.args[1:] {
				key := string(el)
				_, ok := items[key]
				delete(items, key)
				if ok {
					counter += 1
				}
				fmt.Println(counter)
			}
			conn.WriteInt(counter)
		}
	})

	if err != nil {
		panic(err)
	}
}

//commands := &map[string]bool{""}

func handleClient(conn net.Conn, items map[string][]byte) {

	for {

		rd := NewReader(bytes.NewReader([]byte("test")))

		commands, err := rd.readRESP()

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

// had to refactor into something more efficient, using Split which uses regex is inefficient
// TODO: provide some benchmarking for this
func (r *Reader) readRESP() (Command, error) {

	res := Command{}

	buf := make([]byte, 512)

	n, err := r.rd.Read(buf)

	if err != nil {
		return Command{}, err
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

type C2 struct {
	args [][]byte
	Raw  []byte
}

func (rd *Reader) readRESP2() (C2, error) {

	b, err := rd.rd.ReadByte()

	if err != nil {
		return C2{}, err
	}

	var c C2

	if b == '*' {
		arg_counter, err := rd.rd.ReadBytes('\n')
		if err != nil {
			return C2{}, err
		}

		counter, ok := parseInt(arg_counter[:len(arg_counter)-2])

		if !ok {
			return C2{}, errors.New("couldn't parse number")
		}

		in := 0

		c.args = make([][]byte, counter)

		for {
			line, err := rd.rd.ReadBytes('\n')

			if err != nil {
				break
			}

			if len(line) > 0 && line[0] != '$' {
				c.args[in] = append(c.args[in], line[:len(line)-2]...)
				in += 1
			}
			if counter == in {
				break
			}
		}
	}
	return c, nil

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
