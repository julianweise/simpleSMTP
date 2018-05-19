package core

type SMTPMail struct {
	Sender     string
	Recipients []string
	Data       string
}