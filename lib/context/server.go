package context

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/WangYihang/Platypus/lib/util/hash"
	"github.com/WangYihang/Platypus/lib/util/log"
	humanize "github.com/dustin/go-humanize"
)

type Server struct {
	Host      string
	Port      int16
	Clients   map[string](*Client)
	TimeStamp time.Time
	Hash      string
}

func CreateServer(host string, port int16) *Server {
	ts := time.Now()
	return &Server{
		Host:      host,
		Port:      port,
		Clients:   make(map[string](*Client)),
		TimeStamp: ts,
		Hash:      hash.MD5(fmt.Sprintf("%s:%s:%s", host, port, ts)),
	}
}

func (s *Server) Run() {
	service := fmt.Sprintf("%s:%d", s.Host, s.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		log.Error("Resolve TCP address failed: ", err)
		return
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Error("Listen failed: ", err)
		return
	}
	log.Info(fmt.Sprintf("Server running at: %s", s.FullDesc()))

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		client := CreateClient(conn)
		log.Info("New client %s Connected", client.Desc())
		s.AddClient(client)
	}
}

func (s *Server) OnelineDesc() string {
	var buffer bytes.Buffer
	buffer.WriteString(
		fmt.Sprintf(
			"%s:%d (%d online clients)",
			s.Host,
			s.Port,
			len(s.Clients),
		),
	)
	return buffer.String()
}

func (s *Server) FullDesc() string {
	var buffer bytes.Buffer
	buffer.WriteString(
		fmt.Sprintf(
			"[%s] %s:%d (%d online clients) (started at: %s)",
			s.Hash,
			s.Host,
			s.Port,
			len(s.Clients),
			humanize.Time(s.TimeStamp),
		),
	)
	var descs []string
	for _, client := range s.Clients {
		descs = append(descs, fmt.Sprintf("\t%s", client.Desc()))
	}
	if len(descs) > 0 {
		buffer.WriteString("\n")
	}
	buffer.WriteString(strings.Join(descs, "\n"))
	return buffer.String()
}

func (s *Server) Stop() {
	log.Info(fmt.Sprintf("Stopping server: %s", s.OnelineDesc()))
	for _, client := range s.Clients {
		s.DeleteClient(client)
	}
}

func (s *Server) AddClient(client *Client) {
	s.Clients[client.Hash] = client
}

func (s *Server) DeleteClient(client *Client) {
	client.Close()
	delete(s.Clients, client.Hash)
}

func (s *Server) GetAllClients() map[string](*Client) {
	return s.Clients
}