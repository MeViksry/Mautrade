package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

func main() {
	smtpHost := "smtp.hostinger.com"
	smtpPort := "465"
	smtpUsername := "support@mautrade.com"
	smtpPassword := "Mautradebareng20072026!"
	fromAddress := "verify@mautrade.com" // Testing the alias
	toAddress := "support@mautrade.com"

	serverName := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	tlsConfig := &tls.Config{InsecureSkipVerify: false, ServerName: smtpHost}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Test Email\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n<p>Test</p>", fromAddress, toAddress)

	fmt.Println("Connecting to", serverName)
	conn, err := tls.Dial("tcp", serverName, tlsConfig)
	if err != nil {
		fmt.Println("Dial Error:", err)
		os.Exit(1)
	}

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		fmt.Println("Client Error:", err)
		os.Exit(1)
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		fmt.Println("Auth Error:", err)
		os.Exit(1)
	}
	fmt.Println("Auth Successful")

	if err = client.Mail(fromAddress); err != nil {
		fmt.Println("Mail(From) Error:", err)
		os.Exit(1)
	}
	fmt.Println("Mail(From) Successful")

	if err = client.Rcpt(toAddress); err != nil {
		fmt.Println("Rcpt(To) Error:", err)
		os.Exit(1)
	}
	fmt.Println("Rcpt(To) Successful")

	w, err := client.Data()
	if err != nil {
		fmt.Println("Data Error:", err)
		os.Exit(1)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		fmt.Println("Write Error:", err)
		os.Exit(1)
	}

	if err = w.Close(); err != nil {
		fmt.Println("Close Error:", err)
		os.Exit(1)
	}

	fmt.Println("Success!")
}
