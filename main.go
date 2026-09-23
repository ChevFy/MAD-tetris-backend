package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mqtt-tetris-backend/config"
	mqttclient "mqtt-tetris-backend/mqtt"
	"mqtt-tetris-backend/services"
)

func main() {
	appConfig, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	client := mqttclient.NewClient(appConfig)
	if err := client.Connect(); err != nil {
		panic(err)
	}

	messageService := services.NewMessageService()

	for _, topic := range mqttclient.Topics {
		if err := client.Subscribe(topic, messageService.HandleMessage); err != nil {
			fmt.Printf("Subscribe error for %s: %v\n", topic.Name, err)
			return
		}
	}

	keepAlive := make(chan os.Signal, 1)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)
	<-keepAlive

	for _, topic := range mqttclient.Topics {
		if err := client.Unsubscribe(topic); err != nil {
			fmt.Printf("Unsubscribe error for %s: %v\n", topic.Name, err)
		}
	}
	client.Disconnect(250)
	fmt.Println("Disconnected")
}
