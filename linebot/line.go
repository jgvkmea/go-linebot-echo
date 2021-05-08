package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type line struct {
	channelID     string
	channelSecret string
	client        *linebot.Client
}

func newLine() (*line, error) {
	line := line{
		channelID:     os.Getenv("CHANNEL_TOKEN"),
		channelSecret: os.Getenv("CHANNEL_SECRET"),
	}

	client, err := linebot.New(
		line.channelSecret,
		line.channelID,
	)
	if err != nil {
		return nil, err
	}
	line.client = client

	return &line, nil
}

func parseRequest(channelSecret string, r events.APIGatewayProxyRequest) ([]*linebot.Event, error) {
	if !validateSignature(channelSecret, r.Headers["x-line-signature"], []byte(r.Body)) {
		return nil, linebot.ErrInvalidSignature
	}
	request := &struct {
		Events []*linebot.Event `json:"events"`
	}{}
	if err := json.Unmarshal([]byte(r.Body), request); err != nil {
		return nil, err
	}
	return request.Events, nil
}

func validateSignature(channelSecret, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(body)
	return hmac.Equal(decoded, hash.Sum(nil))
}

func (l *line) replyMessages(event *linebot.Event, message *linebot.TextMessage) error {
	ts := []linebot.SendingMessage{}
	ts = append(ts, linebot.NewTextMessage(message.Text))
	ts = append(ts, linebot.NewTextMessage(message.Text))

	_, err := l.client.ReplyMessage(event.ReplyToken, ts...).Do()
	if err != nil {
		return err
	}
	return nil
}
