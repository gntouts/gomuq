package main

import (
	"context"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func messagePubHandle(client mqtt.Client, msg mqtt.Message) {
	logMsg := fmt.Sprintf("Received: %s from: %s\n", msg.Payload(), string(msg.Topic()))
	Log.Info(logMsg)
	topic := string(msg.Topic())
	payload := string(msg.Payload())
	ctx := context.Background()
	_, err := get(ctx, "b0")
	if err == nil {
		// then db is populated
		dbMessage := MessageFromDB()
		key := strings.ReplaceAll(topic, "hass/listen_go/", "")
		dbMessage.data[key] = payload
		ctx := context.Background()
		set(ctx, key, payload)
		uartMsg, err := dbMessage.ToString()
		if err != nil {
			Log.WithError(err).Error("Unable to convert Message to string")
		} else {
			uartOut <- uartMsg
		}
	} else {
		Log.WithError(err).Error("Unable to find b0 in Redis")
	}
}

func connectionHandler(client mqtt.Client) {
	Log.Info("Connected to MQTT")
	sub_to_hass(client)
}

func connectionLostHandler(client mqtt.Client, err error) {
	Log.WithError(err).Warning("MQTT connection lost")
}

func mqtt_client(broker string, port int) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("uamq_client")
	opts.SetDefaultPublishHandler(messagePubHandle)
	opts.OnConnect = connectionHandler
	opts.OnConnectionLost = connectionLostHandler
	return mqtt.NewClient(opts)
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	msg := fmt.Sprintf("Subscribed to : %s\n", topic)
	Log.Info(msg)
}

func pub(client mqtt.Client, topic string, payload string) {
	token := client.Publish("topic/test", 0, false, payload)
	token.Wait()
}

func sub_to_hass(client mqtt.Client) {
	for _, t := range MessageKeys {
		topic := "hass/listen_go/" + t
		sub(client, topic)
	}
}

func publish_mqtt(client mqtt.Client) {
	for {
		if len(mqttKill) > 0 {
			<-mqttKill
			Log.Info("Received MQTT kill signal")
			return
		}
		if len(mqttOut) > 0 {
			msg := <-mqttOut
			parts := strings.Split(msg, ":")
			if len(parts) == 2 {
				topic := parts[0]
				payload := parts[1]
				pub(client, topic, payload)
			}
		}
	}
}

func mqttHandler(host string, port int) {
	client := mqtt_client(host, port)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		Log.Error(token.Error())
		panic(token.Error())
	}
	publish_mqtt(client)
	Log.Info("MQTT exiting")
	client.Disconnect(1)
}
