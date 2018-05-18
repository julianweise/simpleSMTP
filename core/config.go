package core

import (
	"github.com/joho/godotenv"
	"strconv"
	"os"
	"time"
	"log"
)

type SMTPServerConfig struct {
	Port                     int
	MailDirectory            string
	MailWriteInterval		 time.Duration
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
	config.MailWriteInterval, err = time.ParseDuration(os.Getenv("MAIL_WRITE_INTERVAL"))

	if err != nil {
		log.Printf("[ERR] Error parsing time duration %s: ", os.Getenv("MAIL_WRITE_INTERVAL"))
		log.Println(err.Error())
		config.MailWriteInterval = 5 * time.Minute
	}

	return err, config
}