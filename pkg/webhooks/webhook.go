package webhooks

import (
"Rshell/pkg/database"
"Rshell/pkg/logger"
"encoding/json"
)

type WecomConfig struct {
	URL     string `json:"url"` // for compatibility, url can store the key
	Enabled bool   `json:"enabled"`
}
type DingtalkConfig struct {
	Webhook string `json:"webhook"`
	Secret  string `json:"secret"`
	Enabled bool   `json:"enabled"`
}
type TelegramConfig struct {
	Token   string `json:"token"`
	ChatId  string `json:"chat_id"`
	Enabled bool   `json:"enabled"`
}
type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	To       string `json:"to"`
	Enabled  bool   `json:"enabled"`
}

func NotifyOnline(client database.Clients) {
	var settings []database.Settings
	err := database.Engine.In("name", []string{"wecom", "dingtalk", "telegram", "email"}).Find(&settings)
	if err != nil {
		logger.Error("Failed to fetch notification settings: " + err.Error())
		return
	}

	for _, setting := range settings {
		if setting.Value == "" || setting.Value == "{}" {
			continue
		}
		switch setting.Name {
		case "wecom":
			// try to parse as WecomConfig, but fallback if it's just the old key string
var wecom WecomConfig
if err := json.Unmarshal([]byte(setting.Value), &wecom); err != nil {
// Old format, the Value was just the key, enabled by default if not empty
SendWecom(client, setting.Value)
} else {
if wecom.Enabled && wecom.URL != "" {
SendWecom(client, wecom.URL) 
}
}
case "dingtalk":
var ding DingtalkConfig
if err := json.Unmarshal([]byte(setting.Value), &ding); err == nil && ding.Enabled && ding.Webhook != "" {
SendDingtalk(client, ding.Webhook, ding.Secret)
}
case "telegram":
var tg TelegramConfig
if err := json.Unmarshal([]byte(setting.Value), &tg); err == nil && tg.Enabled && tg.Token != "" && tg.ChatId != "" {
SendTelegram(client, tg.Token, tg.ChatId)
}
case "email":
var email EmailConfig
if err := json.Unmarshal([]byte(setting.Value), &email); err == nil && email.Enabled && email.Host != "" {
SendEmail(client, email)
}
}
}
}
