package main

import (
	"log"
	"crypto/tls"
	"simpleSMTP/core"
	"os"
)

func main() {
	// load server certificate
	cer, err := tls.LoadX509KeyPair(os.Getenv("CERTIFICATE"), os.Getenv("KEY"))
	if err != nil {
		log.Fatal(err)
	}
	err, serverConfig := core.NewSMTPServerConfig()
	if err != nil {
		log.Fatal("Failed to setup server configuration")
	}
	server := core.TcpServer{Port: serverConfig.Port, Certificate: cer, Configuration: serverConfig}
	server.Serve()
}