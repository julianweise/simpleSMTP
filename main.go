package main

import (
	"log"
	"crypto/tls"
	"net"
	"bufio"
	"github.com/joho/godotenv"
	"fmt"
	"os"
	"strings"
	"strconv"
)

var Config struct {
	port int
	serverCertificateName string
	serverKeyName string
	mailDirectory string
	maxMailSize int
}

func handleCriticalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func handleNonCriticalError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	Config.port, err = strconv.Atoi(os.Getenv("PORT"))
	Config.maxMailSize, err = strconv.Atoi(os.Getenv("MAX_MAIL_SIZE"))
	Config.serverCertificateName = os.Getenv("CERTIFICATE")
	Config.serverKeyName = os.Getenv("KEY")
	Config.mailDirectory = os.Getenv("MAIL_DIRECTORY")

	handleCriticalError(err)

	// prepare local environment
	if _, err := os.Stat(Config.mailDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(Config.mailDirectory, os.ModePerm)
		handleCriticalError(err)
	}

	// load server certificate
	cer, err := tls.LoadX509KeyPair(Config.serverCertificateName, Config.serverKeyName)
	handleCriticalError(err)

	// load config and start up
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":" + string(Config.port), config)
	handleCriticalError(err)
	fmt.Printf("Listening on port %s \n", string(Config.port))
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func sendResponse(conn net.Conn, response string) {
	n, err := conn.Write([]byte(response))
	if err != nil {
		log.Println(n, err)
		return
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	mail := Mail{}
	dataMode := false
	fmt.Printf("New connection esablished for: %s \n", conn.RemoteAddr().String())
	sendResponse(conn, "220 service ready")
	for {
		msg, err := r.ReadString('\n')
		handleNonCriticalError(err)
		if dataMode {
			// max mail length in byte
			if len(mail.Data) > Config.maxMailSize {
				sendResponse(conn, "556 Message too large")
				mail.Data = ""
				dataMode = false
				continue
			}
			if msg != ".\n" {
				mail.Data += msg
			} else {
				sendResponse(conn, "250 OK")
				mail.writeToFile(Config.mailDirectory)
				dataMode = false
			}
			continue
		}
		command := strings.Split(strings.Replace(msg, "\n", "", -1), " ")
		if len(command) > 0 {
			switch {
			case command[0] == "DATA":
				sendResponse(conn, "354 End data with <CR><LF>.<CR><LF>")
				dataMode = true
			case command[0] == "HELO":
				sendResponse(conn, "250 I am glad to meet you")
			case command[0] == "MAIL":
				if len(command) > 2 {
					mail.Sender = command[2]
					sendResponse(conn, "250 Sender OK")
				} else {
					sendResponse(conn, "500 unrecognized command")
				}
			case command[0] == "RCT":
				if len(command) > 2 {
					mail.Recipient = command[2]
					sendResponse(conn, "250 Recipient OK")
				} else {
					sendResponse(conn, "500 unrecognized command")
				}
			case command[0] == "NOOP":
				sendResponse(conn, "250 OK")
			case command[0] == "QUIT":
				sendResponse(conn, "221 closing channel")
				fmt.Printf("Closing connection to %s as requested by client", conn.RemoteAddr().String())
				conn.Close()
				return
			case command[0] == "RSET":
				dataMode = false
				mail = Mail{}
				sendResponse(conn, "250 OK")
			case command[0] == "VRFY":
				// TODO
				fmt.Println("[ERR] VRFY was called but is not implemented yet")
				sendResponse(conn, "500 unrecognized command")
			default:
				fmt.Printf("Command not recognized: %s \n", command)
				sendResponse(conn, "500 unrecognized command")
			}
		}
	}
}