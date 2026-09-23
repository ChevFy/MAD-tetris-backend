package mqtt

import (
	"crypto/tls"
	"fmt"

	"mqtt-tetris-backend/config"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client paho.Client
}

func NewClient(appConfig config.Config) *Client {
	opts := paho.NewClientOptions()
	opts.AddBroker(appConfig.Broker)
	opts.SetClientID(appConfig.ClientID)
	opts.SetUsername(appConfig.Username)
	opts.SetPassword(appConfig.Password)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: false})
	opts.OnConnect = func(paho.Client) {
		fmt.Println("Connected to HiveMQ Broker")
	}
	opts.OnConnectionLost = func(_ paho.Client, err error) {
		fmt.Printf("Connection lost: %v\n", err)
	}

	return &Client{client: paho.NewClient(opts)}
}

func (client *Client) Connect() error {
	token := client.client.Connect()
	token.Wait()
	return token.Error()
}

func (client *Client) Subscribe(topic Topic, handler paho.MessageHandler) error {
	token := client.client.Subscribe(topic.Name, topic.QoS, handler)
	token.Wait()
	if err := token.Error(); err != nil {
		return err
	}
	fmt.Printf("Subscribed to topic: %s\n", topic.Name)
	return nil
}

func (client *Client) Unsubscribe(topic Topic) error {
	token := client.client.Unsubscribe(topic.Name)
	token.Wait()
	return token.Error()
}

func (client *Client) Disconnect(quiesce uint) {
	client.client.Disconnect(quiesce)
}
