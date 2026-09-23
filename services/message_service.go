package services

import (
	"fmt"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (service *MessageService) HandleMessage(_ paho.Client, message paho.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}
