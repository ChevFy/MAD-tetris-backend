package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mqtt-tetris-backend/config"
	database "mqtt-tetris-backend/db"
	mqttclient "mqtt-tetris-backend/mqtt"
	"mqtt-tetris-backend/services"
)

func main() {
	appConfig, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	db, err := database.Open(appConfig)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	client := mqttclient.NewClient(appConfig)
	if err := client.Connect(); err != nil {
		panic(err)
	}
	defer client.Disconnect(250)

	messageService := services.NewMessageService(db)

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
	fmt.Println("Disconnected")
}
