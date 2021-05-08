package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sirupsen/logrus"
)

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
				log.Infoln("start reply")
				err = line.replyMessages(eve, message)
				if err != nil {
					log.Errorln("faild to reply messages: ", err)
					return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("internal error")
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
