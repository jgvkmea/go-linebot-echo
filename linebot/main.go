package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sirupsen/logrus"
)

type line struct {
	channelID     string
	channelSecret string
	client        *linebot.Client
}

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log := logrus.New()

	line, err := newLine()
	if err != nil {
		log.Errorln("failed to create linebot client: ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to create linebot client: %v", err)
	}

	eves, err := parseRequest(line.channelSecret, request)
	if err != nil {
		log.Errorln("failed to parse events: ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to parse events: %v", err)
	}
	log.Infof("get events: %+v", eves)

	for _, eve := range eves {
		switch eve.Type {
		case linebot.EventTypeMessage:
			switch message := eve.Message.(type) {
			case *linebot.TextMessage:
				log.Infoln("start first reply")
				_, err = line.client.ReplyMessage(eve.ReplyToken, linebot.NewTextMessage(message.Text)).Do()
				if err != nil {
					log.Errorln("failed to reply message: ", err)
					return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to reply message: %v", err)
				}

				log.Infoln("start second reply")
				_, err = line.client.ReplyMessage(eve.ReplyToken, linebot.NewTextMessage(message.Text)).Do()
				if err != nil {
					log.Errorln("failed to reply message at second time: ", err)
					return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to reply message at second time: %v", err)
				}

			default:
				log.Errorln("received request is not TextMessage")
				return events.APIGatewayProxyResponse{StatusCode: 400}, fmt.Errorf("bad request")
			}

		default:
			log.Errorln("received request is not EventTypeMessage")
			return events.APIGatewayProxyResponse{StatusCode: 400}, fmt.Errorf("bad request")
		}
	}

	log.Infoln("echo bot done")
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
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
