package core

import (
	"net"
	"strings"
	"fmt"
	"bufio"
	"log"
	"net/textproto"
	"io"
	"time"
)

type SMTPSession struct {
	Connection 			net.Conn
	Mail 				Mail
	Reader				*textproto.Reader
	Writer 				*textproto.Writer
	active				bool
	Configuration 		SMTPServerConfig
	MeasuringService 	SessionMeasuringService
}

func (s *SMTPSession) handle() {
	defer s.Connection.Close()
	if s.Configuration.ShouldMeasurePerformance {
		s.MeasuringService = NewSessionMeasuringService()
		defer s.MeasuringService.PrintResults()
	}
	s.active = true
	maxLineLength := int64(s.Configuration.MaxLengthLine)

	s.Reader = textproto.NewReader(bufio.NewReader(io.LimitReader(io.Reader(s.Connection), maxLineLength)))
	s.Writer = textproto.NewWriter(bufio.NewWriter(s.Connection))

	fmt.Printf("New s.Connectionection esablished for: %s \n", s.Connection.RemoteAddr().String())
	s.sendResponse("220 service ready")

	timeoutDuration := time.Duration(s.Configuration.Timeout) * time.Second

	for s.active {
		// reset timeout
		s.Connection.SetReadDeadline(time.Now().Add(timeoutDuration))
		// read client input
		msg, err := s.Reader.ReadLine()

		if err != nil {
			if err == io.EOF {
				log.Printf("Lost connection to %s\n", s.Connection.RemoteAddr().String())
			} else if err, ok := err.(net.Error); ok && err.Timeout() {
				s.sendResponse("221 idle timeout - closing channel")
			} else {
				log.Println(err)
			}
			return
		}
		command := strings.Fields(msg)
		keyword := strings.ToUpper(command[0])
		if len(command) < 1 {
			continue
		}
		if s.Configuration.ShouldMeasurePerformance {
			s.MeasuringService.StartMeasuring(keyword)
		}
		switch keyword {
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
	fmt.Printf("Closing s.Connectionection to %s as requested by client \n", s.Connection.RemoteAddr().String())
	s.active = false
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
	if len(arguments) < 1 || len(arguments[0]) <= len("FROM:") {
		s.sendResponse("501 arguments missing")
		return
	}
	if arguments[0][:4] != "FROM" {
		fmt.Println(arguments[0][:4])
		s.sendResponse("501 invalid arguments")
		return
	}

	s.Mail.Sender = arguments[0][5:]
	s.sendResponse("250 Sender OK")
}

func (s *SMTPSession) handleRCPT(arguments []string) {
	if len(s.Mail.Sender) < 1 {
		s.sendResponse("503 sender missing. Issue MAIL command first")
		return
	}
	if len(arguments) < 1 || len(arguments[0]) <= len("TO:") {
		s.sendResponse("501 arguments missing")
		return
	}
	if arguments[0][:2] != "TO" {
		s.sendResponse("501 invalid arguments")
		return
	}

	s.Mail.Recipient = append(s.Mail.Recipient, arguments[0][3:])
	s.sendResponse("250 Sender OK")
}

func (s *SMTPSession) handleData() {
	if len(s.Mail.Sender) < 1 {
		s.sendResponse("503 sender missing. Issue MAIL command first")
		return
	}
	if len(s.Mail.Recipient) < 1 {
		s.sendResponse("503 at least one recipient is required")
		return
	}
	s.sendResponse("354 End data with <CR><LF>.<CR><LF>")

	dataReader := s.Reader.DotReader()
	mailData := make([]byte, s.Configuration.MaxMailSize)
	n, err := dataReader.Read(mailData)

	if err != io.EOF {
		s.sendResponse("552 Maximum message size exceeded")
		s.Reader.R.Discard(s.Reader.R.Buffered())
		s.Reader = textproto.NewReader(s.Reader.R)
		return
	}

	s.Mail.Data = string(mailData[0:n])

	s.sendResponse("250 OK")
	s.Mail.writeToFile(s.Configuration.MailDirectory)
}

// helper functions

func (s *SMTPSession) sendResponse(response string) {
	err := s.Writer.PrintfLine(response)
	if s.Configuration.ShouldMeasurePerformance {
		s.MeasuringService.FinalizeMeasuring()
	}
	if err != nil {
		log.Println(err)
	}
}