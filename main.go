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
)

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
	handleCriticalError(err)

	// load server certificate
	cer, err := tls.LoadX509KeyPair(os.Getenv("CERTIFICATE"), os.Getenv("KEY"))
	handleCriticalError(err)

	// load config and start up
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":" + os.Getenv("PORT"), config)
	handleCriticalError(err)
	fmt.Printf("Listening on port %s \n", os.Getenv("PORT"))
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
	fmt.Printf("New connection esablished for: %s \n", conn.RemoteAddr().String())
	sendResponse(conn, "220 service ready")
	for {
		msg, err := r.ReadString('\n')
		handleNonCriticalError(err)
		switch command := strings.Replace(msg, "\n", "", -1); command {
		case "QUIT":
			sendResponse(conn, "221 closing channel")
			fmt.Printf("Closing connection to %s as requested by client", conn.RemoteAddr().String())
			conn.Close()
			return
		default:
			fmt.Printf("Command not recognized: %s \n", command)
			sendResponse(conn, "500 unrecognized command")
		}
	}
}