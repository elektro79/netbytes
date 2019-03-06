package netBytes

import (
	"fmt"
	"net"
	"testing"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Server struct {
	ln net.Listener
}
type Client struct{}

func (s *Server) Connected(nbc *NetBytesConn) {
	fmt.Printf("Sconnected: %v\n", nbc.Conn.RemoteAddr())
}
func (s *Server) Msg(nbc *NetBytesConn, by []byte) {
	fmt.Printf("Sreceive %s\n", string(by))
	nbc.Send(by)
}
func (s *Server) Disconected(nbc *NetBytesConn, err error) {
	fmt.Printf("SDisconected: %v %s\n", nbc.Conn.RemoteAddr(), err.Error())
	s.ln.Close()
}

func (c *Client) Connected(nbc *NetBytesConn) {
	fmt.Printf("connected: %v\n", nbc.Conn.RemoteAddr())
	nbc.Send([]byte("test"))
}
func (c *Client) Msg(nbc *NetBytesConn, s []byte) {
	fmt.Printf("receive %s\n", s)
	nbc.Conn.Close()
}
func (c *Client) Disconected(nbc *NetBytesConn, err error) {
	fmt.Printf("Disconected: %v %s\n", nbc.Conn.RemoteAddr(), err.Error())
}

func TestMain(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:18000")
	check(err)
	ns := NewNetBytes(&Server{ln})
	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:18000")
		check(err)
		nsc := NewNetBytes(&Client{})
		go nsc.Connect(conn)
		fmt.Println("client exit")
	}()
	ns.Listen(ln)
}
