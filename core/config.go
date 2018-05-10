package core

import (
	"github.com/joho/godotenv"
	"strconv"
	"os"
)

type SMTPServerConfig struct {
	Port                     int
	MailDirectory            string
	MaxMailSize              int
	MaxRecipients            int
	Timeout                  int
	MaxLengthLine            int
	ShouldMeasurePerformance bool
}

func NewSMTPServerConfig() (error, SMTPServerConfig) {
	// Load environment variables
	err := godotenv.Load()
	config := SMTPServerConfig{MailDirectory:  os.Getenv("MAIL_DIRECTORY")}

	config.Port, err = strconv.Atoi(os.Getenv("PORT"))
	config.Timeout, err = strconv.Atoi(os.Getenv("TIMEOUT"))
	config.MaxMailSize, err = strconv.Atoi(os.Getenv("MAX_MAIL_SIZE"))
	config.MaxRecipients, err = strconv.Atoi(os.Getenv("MAX_RECIPIENTS"))
	config.MaxLengthLine, err = strconv.Atoi(os.Getenv("MAX_LENGTH_LINE"))
	config.ShouldMeasurePerformance, err = strconv.ParseBool(os.Getenv("MEASURE_PERFORMANCE"))

	return err, config
}