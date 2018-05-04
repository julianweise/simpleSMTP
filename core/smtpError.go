package core

type SMTPError struct {
	code 	int
	message string
}

func (e *SMTPError) Error() string {
	return string(e.code) + " " + e.message
}