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
	"regexp"
)

type SMTPSession struct {
	Connection 			net.Conn
	Mail 				Mail
	Reader				*textproto.Reader
	Writer 				*textproto.Writer
	active				bool
	Configuration 		SMTPServerConfig
	MeasuringService 	SessionMeasuringService
	client				string
}

const mailFromRegex = "^MAIL\\s+FROM\\s*:\\s*<(?P<from>(?P<local>(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\\\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f ]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f ])*\\\"))(?:@(?P<host>(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])|(?:(?:[0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9]))))?)>$"
const mailRcptRegex = "^RCPT\\s+TO\\s*:\\s*<(?P<receiver>(?:(?:(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\\\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f ]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f ])*\\\"))(?:@(?:(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])|(?:(?:[0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9]))))?\\s*,?\\s*)+)>$"
const mailRcptIndividualRegex = "^(?P<receiver>(?:(?:(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\\\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f ]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f ])*\\\"))(?:@(?:(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\]))|(?:(?:[0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])))?))(?:\\s*,?\\s*)(?P<rest>.*?)$"

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
		if len(msg) < 4 {
			s.sendResponse("500 invalid command")
			continue
		}
		keyword := strings.ToUpper(msg[:4])
		if s.Configuration.ShouldMeasurePerformance {
			s.MeasuringService.StartMeasuring(keyword)
		}
		switch keyword {
		case "DATA":
			s.handleData()
		case "HELO":
			s.handleHelo(msg)
		case "MAIL":
			s.handleMail(msg)
		case "RCPT":
			s.handleRCPT(msg)
		case "NOOP":
			s.handleNoop()
		case "QUIT":
			s.handleQuit()
		case "RSET":
			s.handleReset()
		case "VRFY":
			s.handleVerify()
		default:
			fmt.Printf("Command not recognized: %s \n", msg)
			s.sendResponse("500 unrecognized command")
		}
	}
}

// SMTP Keywords

func (s *SMTPSession) handleHelo(line string) {
	arguments := strings.Fields(line)
	if len(arguments) < 2 {
		s.sendResponse("503 please provide your identifier")
		return
	}
	s.client = arguments[1]
	s.sendResponse("250 " + s.client + " - I am glad to meet you")
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

func (s *SMTPSession) handleMail(line string) {
	if len(s.client) < 1 {
		s.sendResponse("503 session not correct established. Issue HELO command first")
		return
	}
	var ok bool
	ok, s.Mail.Sender = matchesValidMailFromAddress(line)
	if !ok {
		s.sendResponse("501 invalid arguments")
		return
	}
	s.sendResponse("250 Sender OK")
}

func (s *SMTPSession) handleRCPT(line string) {
	if len(s.Mail.Sender) < 1 {
		s.sendResponse("503 sender missing. Issue MAIL command first")
		return
	}
	ok, receiver := matchesValidMailRcptAddresses(line)
	if !ok {
		s.sendResponse("501 invalid arguments")
		return
	}

	s.Mail.Recipient = append(s.Mail.Recipient, receiver...)
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

func matchesValidMailFromAddress(line string) (success bool, sender string) {
	r, err := regexp.Compile(mailFromRegex)
	if err != nil {
		log.Print("Error in Regex for MAIL FROM validation")
		return
	}
	match := r.FindStringSubmatch(line)
	if len(match) > 1 {
		return true, match[1]
	}
	return
}

func matchesValidMailRcptAddresses(line string) (success bool, receiverList []string) {
	r, err := regexp.Compile(mailRcptRegex)
	if err != nil {
		log.Print("Error in Regex for RCPT standard validation")
		return
	}

	match := r.FindStringSubmatch(line)
	if len(match) < 1 {
		return
	}

	r, err = regexp.Compile(mailRcptIndividualRegex)
	if err != nil {
		log.Print("Error in Regex for RCPT individual address validation")
		return
	}

	for {
		receiver := r.FindStringSubmatch(match[1])
		if len(receiver) <= 1 {
			break
		}
		if receiver[1] != "" {
			receiverList = append(receiverList, receiver[1])
		}
		match[1] = receiver[len(receiver) -1]
	}
	success = true
	return
}