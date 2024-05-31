package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Notifier interface {
	Send(message string) error
}

type DingTalkNotifier struct {
	WebhookURL string
}

// NewDingTalkNotifier 创建一个新的 DingTalkNotifier 实例
func NewDingTalkNotifier(webhookURL string) *DingTalkNotifier {
	return &DingTalkNotifier{WebhookURL: webhookURL}
}

// DingTalkMessage 是钉钉消息格式
type DingTalkMessage struct {
	MsgType string              `json:"msgtype"`
	Text    DingTalkMessageText `json:"text"`
}

type DingTalkMessageText struct {
	Content string `json:"content"`
}

// Send 发送钉钉通知
func (d *DingTalkNotifier) Send(message string) error {
	msg := DingTalkMessage{
		MsgType: "text",
		Text: DingTalkMessageText{
			Content: message,
		},
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
