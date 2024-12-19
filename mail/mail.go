package mail

import (
	"os"
	"strconv"

	gomail "gopkg.in/mail.v2"
)

func SendMail() error {
	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", "bots.jhamm@gmail.com")
	m.SetHeader("To", "jon@hammerskov.dk")
	m.SetHeader("Subject", "Hello from Go!")
	m.SetBody("text/plain", "This is the body of the email.")

	// Set up the SMTP dialer
	//d := gomail.NewDialer("smtp.gmail.com", 587, "bots.jhamm@gmail.com", "iels qkjy gitt bcxc")
	port, err := strconv.Atoi(os.Getenv("SMTP_SERVER_PORT"))
	if err != nil {
		port = 587
	}
	d := gomail.NewDialer(os.Getenv("SMTP_SERVER"), port, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"))

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
