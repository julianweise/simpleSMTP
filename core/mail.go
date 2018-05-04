package core

import (
	"time"
	"io/ioutil"
	"log"
	"strings"
)

type Mail struct {
	Sender 		string
	Recipient	[]string
	Data		string
}

func (mail Mail) writeToFile(location string) error {
	fileName := time.Now().Format("2006-01-02 15:04:05") +
				"_" + mail.Sender +
				"_" + strings.Join(mail.Recipient[:], ",") +
				".txt"
	fileContent := []byte(mail.Data)
	err := ioutil.WriteFile(location + fileName, fileContent, 0644)
	if err != nil {
		log.Printf("[ERR] Error creating file %s for mail %s. \n", fileName, mail.Data)
		log.Println(err)
	}
	return err
}