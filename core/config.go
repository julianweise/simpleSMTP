package core

import (
	"github.com/joho/godotenv"
	"strconv"
	"os"
)

type SMTPServerConfig struct {
	Port 		int
	MailDirectory 	string
	MaxMailSize   	int
	MaxRecipients 	int
}

func NewSMTPServerConfig() (error, SMTPServerConfig) {
	// Load environment variables
	err := godotenv.Load()
	config := SMTPServerConfig{MailDirectory:  os.Getenv("MAIL_DIRECTORY")}

	config.Port, err = strconv.Atoi(os.Getenv("PORT"))
	config.MaxMailSize, err = strconv.Atoi(os.Getenv("MAX_MAIL_SIZE"))
	config.MaxRecipients, err = strconv.Atoi(os.Getenv("MAX_RECIPIENTS"))

	return err, config
}