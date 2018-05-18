package core

import (
	"time"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
)

type SMTPMail struct {
	Sender     string
	Recipients []string
	Data       string
}

func (mail SMTPMail) writeToFile(location string) error {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	fileName := time.Now().Format("2006-01-02 15:04:05") +
				"_" + strconv.Itoa(r1.Int()) + "_" +
				".txt"
	fileContent := []byte(mail.Data)
	err := ioutil.WriteFile(location + fileName, fileContent, 0644)
	if err != nil {
		log.Printf("[ERR] Error creating file %s for mail %s. \n", fileName, mail.Data)
		log.Println(err)
	}
	return err
}