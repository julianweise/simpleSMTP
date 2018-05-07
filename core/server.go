package core

import (
	"os"
	"log"
	"crypto/tls"
	"net"
	"strconv"
	"fmt"
)

type TcpServer struct {
	Certificate		tls.Certificate
	Configuration	SMTPServerConfig
}

func (s *TcpServer) setUpFileSystem() {
	// prepare local environment
	if _, err := os.Stat(s.Configuration.MailDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(s.Configuration.MailDirectory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *TcpServer) Serve() {
	s.setUpFileSystem()
	config := &tls.Config{Certificates: []tls.Certificate{s.Certificate}}
	ln, err := tls.Listen("tcp", ":" + strconv.Itoa(s.Configuration.Port), config)
	if err != nil {
		log.Fatal("Error setting up server: " + err.Error())
	}
	s.listen(ln)
}

func (s *TcpServer) listen(ln net.Listener) {
	defer ln.Close()
	fmt.Printf("Server is up and running on port %d.\n", s.Configuration.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		session := SMTPSession{Connection: conn, Configuration: s.Configuration}
		go session.handle()
	}
}