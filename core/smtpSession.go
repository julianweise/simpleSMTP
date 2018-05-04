package core

import (
	"net"
	"strings"
	"fmt"
	"bufio"
	"log"
	"net/textproto"
)

type SMTPSession struct {
	Connection 		net.Conn
	Mail 			Mail
	Reader			*textproto.Reader
	Writer 			*textproto.Writer
	active			bool
	Configuration 	SMTPServerConfig
}

func (s *SMTPSession) handle() {
	defer s.Connection.Close()
	s.active = true
	s.Reader = textproto.NewReader(bufio.NewReader(s.Connection))
	s.Writer = textproto.NewWriter(bufio.NewWriter(s.Connection))

	fmt.Printf("New s.Connectionection esablished for: %s \n", s.Connection.RemoteAddr().String())
	s.sendResponse("220 service ready")

	for s.active {
		msg, err := s.Reader.ReadLine()
		if err != nil {
			log.Println(err)
		}
		command := strings.Fields(msg)
		if len(command) < 1 {
			continue
		}
		switch command[0] {
		case "DATA":
			s.handleData()
		case "HELO":
			s.handleHelo()
		case "MAIL":
			s.handleMail(command[1:])
		case "RCPT":
			s.handleRCPT(command[1:])
		case "NOOP":
			s.handleNoop()
		case "QUIT":
			s.handleQuit()
			break
		case "RSET":
			s.handleReset()
		case "VRFY":
			s.handleVerify()
		default:
			fmt.Printf("Command not recognized: %s \n", command)
			s.sendResponse("500 unrecognized command")
		}
	}
}

// SMTP Keywords

func (s *SMTPSession) handleHelo() {
	s.sendResponse("250 I am glad to meet you")
}

func (s *SMTPSession) handleNoop() {
	s.sendResponse("250 OK")
}

func (s *SMTPSession) handleQuit() {
	s.sendResponse("221 closing channel")
	fmt.Printf("Closing s.Connectionection to %s as requested by client", s.Connection.RemoteAddr().String())
	s.Connection.Close()
}

func (s *SMTPSession) handleReset() {
	s.Mail = Mail{}
	s.sendResponse("250 OK")
}

func (s *SMTPSession) handleVerify() {
	// TODO
	fmt.Println("[ERR] VRFY was called but is not implemented yet")
	s.sendResponse("500 unrecognized command")
}

func (s *SMTPSession) handleMail(arguments []string) {
	s.checkNumberOfArguments(arguments, 1)
	if arguments[1][:3] != "FROM" {
		s.sendResponse("501 invalid arguments")
	}

	s.Mail.Sender = arguments[1][5:]
	s.sendResponse("250 Sender OK")
}

func (s *SMTPSession) handleRCPT(arguments []string) {
	s.checkNumberOfArguments(arguments, 1)
	if arguments[1][:1] != "TO" {
		s.sendResponse("501 invalid arguments")
	}

	s.Mail.Recipient = append(s.Mail.Recipient, arguments[1][3:])
	s.sendResponse("250 Sender OK")
}

func (s *SMTPSession) handleData() {
	s.sendResponse("354 End data with <CR><LF>.<CR><LF>")

	dataReader := newSMTPDataReader(s.Reader, 200000)
	mailData := make([]byte, 200000)
	n, err := dataReader.Read(mailData)

	if err != nil {
		s.sendResponse(err.Error())
		return
	}

	s.Mail.Data = string(mailData[0:n])

	s.sendResponse("250 OK")
	s.Mail.writeToFile(s.Configuration.MailDirectory)
}

// helper functions
func (s *SMTPSession) checkNumberOfArguments(arguments []string, numberRequired int) {
	if len(arguments) < numberRequired {
		s.sendResponse("501 arguments missing")
	}
}

func (s *SMTPSession) sendResponse(response string) {
	err := s.Writer.PrintfLine(response)
	if err != nil {
		log.Println(err)
	}
}