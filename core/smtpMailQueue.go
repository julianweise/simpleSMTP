package core

import (
	"time"
	"hash/crc64"
	"strconv"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"net/http"
	"encoding/json"
	"bytes"
	"errors"
)

type SMTPMailQueue struct {
	mails         []*SMTPMail
	writeInterval time.Duration
	crc64Table    *crc64.Table
	IsWriting     bool
	Configuration *SMTPServerConfig
	mutex		  sync.Mutex
}

type SMTPJsonMail struct {
	Received		string		`json:"received"`
	ReceivedFrom	string		`json:"received_from"`
	ReceivedBy		string		`json:"received_by"`
	MailFrom		string		`json:"mail_from"`
	RcptTo			[]string	`json:"rcpt_to"`
	Data			string		`json:"data"`
}

func NewMailQueue(serverConfiguration *SMTPServerConfig)  (err error, queue SMTPMailQueue) {
	queue = SMTPMailQueue{
		writeInterval: serverConfiguration.MailWriteInterval,
		IsWriting:     false,
		crc64Table:    crc64.MakeTable(crc64.ECMA),
		Configuration: serverConfiguration,
		mutex:		   sync.Mutex{},
	}

	return
}

func (q *SMTPMailQueue) startWriting() {
	if q.IsWriting {
		return
	}

	q.IsWriting = true
	go q.run()
}

func (q *SMTPMailQueue) stopWriting() {
	q.IsWriting = false
}

func (q *SMTPMailQueue) run() {
	for q.IsWriting {
		q.saveAll()

		startWaiting := time.Now()
		for q.IsWriting && time.Now().Sub(startWaiting) < q.writeInterval {
			/*
			this loop is used to keep the thread response time for a change of the
			IsWriting flag as low as possible
			 */
			time.Sleep(100 * time.Millisecond)
		}
	}

	q.saveAll()

	if len(q.mails) > 0 {
		log.Println("[INFO] " + strconv.Itoa(len(q.mails)) + " e-mails were not saved upon stopping the mail queue!")
	}
}

func (q *SMTPMailQueue) push(mail *SMTPMail) {
	q.mutex.Lock()
	q.mails = append(q.mails, mail)
	q.mutex.Unlock()
}

func (q *SMTPMailQueue) pop() (mail *SMTPMail) {
	if len(q.mails) <= 0 {
		return
	}
	q.mutex.Lock()
	mail = q.mails[0]
	q.mails = q.mails[1:]
	q.mutex.Unlock()
	return
}

func (q *SMTPMailQueue) getFileLocation(mail *SMTPMail) (location string, err error) {
	location = q.Configuration.MailDirectory

	if !strings.HasSuffix(location, "/") {
		location += "/"
	}

	senderHash := crc64.Checksum([]byte(mail.Sender), q.crc64Table)
	senderDirectory := strconv.FormatUint(senderHash, 16)

	location += senderDirectory + "/"

	if _, err := os.Stat(location); os.IsNotExist(err) {
		err = os.MkdirAll(location, 0755)
	}

	if err != nil {
		log.Print("[ERR] Error creating mail location: ")
		log.Println(err)
	}

	return
}

func (q *SMTPMailQueue) getFileName(mail *SMTPMail) (name string) {
	fileName := mail.Sender + strconv.FormatInt(time.Now().Unix(), 16)
	fileHash := crc64.Checksum([]byte(fileName), q.crc64Table)

	ending := ".mail"
	name = strconv.FormatUint(fileHash, 16) + ending
	return
}

func (q *SMTPMailQueue) saveRemote(mail *SMTPMail) (err error) {
	postUrl := q.Configuration.MailStorageService

	jsonMail := SMTPJsonMail{
		Received: time.Now().UTC().Format(time.RFC3339),
		ReceivedFrom: "smtp-client",
		ReceivedBy: "localhost",
		MailFrom: mail.Sender,
		RcptTo: mail.Recipients,
		Data: mail.Data,
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(jsonMail)

	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", postUrl, buffer)

	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		err = errors.New("storage service did not store the mail: " + strconv.Itoa(response.StatusCode) + " " + response.Status)
	}

	return nil
}

func (q *SMTPMailQueue) save(mail *SMTPMail) (err error) {

	/*
	Mail file format:

	<sender>
	<recipient1>
	<recipient2>
	...
	<recipientN>

	<data>
	 */


 	// try to save mail remotely
 	err = q.saveRemote(mail)

 	if err == nil {
 		return nil
	}

	// remote storage did not work... so store it locally
	fileLocation, err := q.getFileLocation(mail)

	if err != nil {
		// cannot create mail directory
		return
	}

	fileName := q.getFileName(mail)
	var recipients string

	for i := 0; i < len(mail.Recipients); i++ {
		recipients += mail.Recipients[i] + "\n"
	}

	fileContent := []byte(mail.Sender + "\n" + recipients + mail.Data)
	err = ioutil.WriteFile(fileLocation + fileName, fileContent, 0644)

	if err != nil {
		// cannot access mail directory
		log.Printf("[ERR] Error saving mail in file %s%s: ", fileLocation, fileName)
		log.Println(err)
	}

	return
}

func (q *SMTPMailQueue) saveAll() {
	mailsLen := len(q.mails)

	for i := 0; i < mailsLen; i++ {
		mail := q.pop()

		if mail == nil {
			continue
		}

		err := q.save(mail)
		if err != nil {
			// retry later: push mail back into queue
			q.push(mail)
		}
	}
}
