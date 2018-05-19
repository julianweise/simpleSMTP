package core

import (
	"log"
	"crypto/tls"
	"net"
	"strconv"
	"fmt"
)

type TcpServer struct {
	Certificate		tls.Certificate
	Configuration	SMTPServerConfig
	MailQueue 		*SMTPMailQueue
}

func (s *TcpServer) Serve() {
	config := &tls.Config{Certificates: []tls.Certificate{s.Certificate}}
	ln, err := tls.Listen("tcp", ":" + strconv.Itoa(s.Configuration.Port), config)
	if err != nil {
		log.Fatal("Error setting up server: " + err.Error())
	}
	s.listen(ln)
}

func (s *TcpServer) listen(ln net.Listener) {
	defer ln.Close()
	s.MailQueue.startWriting()
	defer s.MailQueue.stopWriting()
	fmt.Printf("Server is up and running on port %d.\n", s.Configuration.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		session := SMTPSession{
			Connection: conn,
			Configuration: s.Configuration,
			mailQueue: s.MailQueue,
		}
		go session.handle()
	}
}