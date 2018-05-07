package main

import (
	"log"
	"crypto/tls"
	"simpleSMTP/core"
	"os"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	// load server certificate
	cer, err := tls.LoadX509KeyPair(os.Getenv("CERTIFICATE"), os.Getenv("KEY"))
	if err != nil {
		log.Fatal("Certificate loading: " + err.Error())
	}
	err, serverConfig := core.NewSMTPServerConfig()
	if err != nil {
		log.Fatal("Failed to setup server configuration")
	}
	server := core.TcpServer{Certificate: cer, Configuration: serverConfig}
	server.Serve()
}