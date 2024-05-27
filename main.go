package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type Client struct {
	Name string
	Conn net.Conn
}

type Server struct {
	ListenAddr string
	Listener   net.Listener
	Clients    map[string]*Client
	Signal     os.Signal
}

func NewServer(listenAddr string) *Server {
	clients := make(map[string]*Client)
	return &Server{
		ListenAddr: listenAddr,
		Clients:    clients,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		fmt.Printf("failed to create listener:\n  %v\n", err)
		return err
	}
	s.Listener = listener

	log.Printf("listening on: %s", s.Listener.Addr().String())

	s.acceptConns()

	return err
}

func (s *Server) acceptConns() error {
	for {

		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection: %v", err)
			return err
		}
		go s.handleConns(&conn)

	}
}


func (s *Server) handleConns(conn *net.Conn) {
	s.handleJoin(conn)
	defer (*conn).Close()
	for {
		err := handleMessages(s.Clients, *conn)
		if err == io.EOF {
			(*conn).Close()
			log.Println(s.Clients[(*conn).RemoteAddr().String()].Name, " left!")
			delete(s.Clients, (*conn).RemoteAddr().String())
			break
		}
	}

}

func (s *Server) handleJoin(conn *net.Conn) error {

	_, err := (*conn).Write([]byte("username:"))
	if err != nil {
		fmt.Printf("failed to write to connection:\n  %v\n", err)
		return err
	}

	nameBuffer := make([]byte, 512)
	n, err := (*conn).Read(nameBuffer)
	if err != nil {
		fmt.Printf("failed to read from connection:\n  %v\n", err)
		return err
	}

	var name string = string(nameBuffer[:n-1])

	for {
		for _, client := range s.Clients {
			if name == client.Name {
				(*conn).Write([]byte("name taken try another one\n"))

				_, err := (*conn).Write([]byte("username:"))
				if err != nil {
					fmt.Printf("failed to write to connection:\n  %v\n", err)
					return err
				}
				n, err := (*conn).Read(nameBuffer)
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("failed to read from connection:\n  %v\n", err)
					return err
				}
				name = string(nameBuffer[:n-1])

				continue
			} else {
				break
			}
		}
		break
	}

	s.Clients[(*conn).RemoteAddr().String()] = &Client{
		Name: name,
		Conn: *conn,
	}
	log.Printf("%s joined!\n", name)

	return nil
}

func handleMessages(clients map[string]*Client, conn net.Conn) error {
	messageBuffer := make([]byte, 1024)
	addr := conn.RemoteAddr().String()
	n, err := conn.Read(messageBuffer)
	var message string

	if err != nil {
		return err
	}
	message = string(messageBuffer[:n-1])
	if message == "clear" {
		conn.Write([]byte("\033[3J"))
	}
	for _, client := range clients {
		if client.Conn != conn {
			text := fmt.Sprintf("%s:%s\n", clients[addr].Name, message)
			client.Conn.Write([]byte(text))
		}
	}
	return nil
}

func main() {
	server := NewServer("localhost:8080")
	server.Start()
	fmt.Println("clients num", len(server.Clients))
}
